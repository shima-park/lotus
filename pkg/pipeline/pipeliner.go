package pipeline

import (
	"expvar"
	"fmt"
	"io"
	"reflect"

	"context"

	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	circuit "github.com/rubyist/circuitbreaker"
	"github.com/shima-park/lotus/pkg/common/inject"
	"github.com/shima-park/lotus/pkg/common/log"
	"github.com/shima-park/lotus/pkg/common/monitor"
	"github.com/shima-park/lotus/pkg/component"
	"github.com/shima-park/lotus/pkg/processor"
)

type Pipeliner interface {
	Name() string
	Start() error
	Stop()
	State() State
	ListComponents() []Component
	ListProcessors() []Processor
	Monitor() monitor.Monitor
	GetConfig() Config
	SetConfig(config Config) error
	Visualize(w io.Writer, format string) error
	CheckDependence() []error
	Error() error
}

type Component struct {
	Name      string
	RawConfig string
	Component component.Component
	Factory   component.Factory
}

type Processor struct {
	Name      string
	RawConfig string
	Processor processor.Processor
	Factory   processor.Factory
}

type pipeliner struct {
	config Config

	name       string
	components []Component
	processors []Processor

	ctx       context.Context
	cancel    context.CancelFunc
	parser    ScheduleParser
	schedule  Schedule
	injector  inject.Injector
	startTime time.Time

	stream  *Stream
	monitor monitor.Monitor

	state     int32
	runningWg sync.WaitGroup

	errs []error
}

func NewPipelineByConfig(conf Config) Pipeliner {
	ctx, cancel := context.WithCancel(context.Background())
	p := &pipeliner{
		config: conf,

		name: conf.Name,

		ctx:       ctx,
		cancel:    cancel,
		parser:    defaultScheduleParser,
		schedule:  defaultSchedule,
		injector:  inject.New(),
		startTime: time.Now(),

		monitor: monitor.NewMonitor(conf.Name),

		state: int32(Idle),
	}
	p.monitor.Set(METRICS_KEY_PIPELINE_STATE, expvar.Func(func() interface{} { return p.State() }))
	p.injector.MapTo(p.monitor, "Monitor", (*monitor.Monitor)(nil))
	p.injector.MapTo(p.ctx, "Context", (*context.Context)(nil))
	return p.init(conf)
}

func (p *pipeliner) init(conf Config) *pipeliner {
	if p.name == "" {
		p.errs = append(p.errs,
			errors.Wrap(errors.New("The pipeliner's name cannot be empty "), conf.Name),
		)
	}

	var err error
	p.components, err = conf.NewComponents()
	if err != nil {
		p.errs = append(p.errs, errors.Wrapf(err, "Pipeline: %s", conf.Name))
	}

	p.processors, err = conf.NewProcessors()
	if err != nil {
		p.errs = append(p.errs, errors.Wrapf(err, "Pipeline: %s", conf.Name))
	}

	var pm = map[string]Processor{}
	for _, p := range p.processors {
		pm[p.Name] = p
	}

	p.stream, err = NewStream(conf.Stream, pm)
	if err != nil {
		p.errs = append(p.errs, errors.Wrapf(err, "Pipeline: %s", conf.Name))
	}

	if p.stream == nil {
		p.errs = append(p.errs, errors.Wrapf(err, "Pipeline: %s", conf.Name))
	}

	if p.config.Schedule != "" && p.parser != nil {
		p.schedule, err = p.parser(p.config.Schedule)
		if err != nil {
			p.errs = append(p.errs, errors.Wrapf(err, "Pipeline: %s", conf.Name))
		}
	}

	distinct := map[reflect.Type]map[string]struct{}{}
	for _, c := range p.components {
		instance := c.Component.Instance()

		if _, ok := distinct[instance.Type()]; !ok {
			distinct[instance.Type()] = map[string]struct{}{}
		}

		if _, ok := distinct[instance.Type()][instance.Name()]; ok {
			err = fmt.Errorf("Pipeline: %s, Component: %s, Type: %s, Name: %s is already registered",
				conf.Name, c.Name, instance.Type(), instance.Name(),
			)
			p.errs = append(p.errs, err)
		}

		distinct[instance.Type()] = map[string]struct{}{
			instance.Name(): struct{}{},
		}

		p.injector.Set(instance.Type(), instance.Name(), instance.Value())
	}

	if errs := p.CheckDependence(); len(errs) > 0 {
		p.errs = append(p.errs, errs...)
	}

	return p
}

func (p *pipeliner) Name() string {
	return p.name
}

func (p *pipeliner) CheckDependence() []error {
	checkInj := inject.New()
	checkInj.SetParent(p.injector)
	return check(p.stream, checkInj)
}

func (p *pipeliner) newExecContext() *execContext {
	inj := inject.New()
	inj.SetParent(p.injector)

	ctx, cancel := context.WithCancel(p.ctx)
	inj.MapTo(ctx, "Context", (*context.Context)(nil))

	c := &execContext{
		ctx:      ctx,
		cancel:   cancel,
		injector: inj,
		stream:   p.stream,
		monitor:  p.monitor,
		breaker:  circuit.NewRateBreaker(p.config.CircuitBreakerRate, p.config.CircuitBreakerSamples),
		inputC:   make(chan inject.Injector, p.stream.config.BufferSize),
	}

	return c
}

func (p *pipeliner) Start() error {
	if !atomic.CompareAndSwapInt32(&p.state, int32(Idle), int32(Running)) {
		return nil
	}

	if err := p.start(); err != nil {
		p.errs = append(p.errs, err)
		atomic.CompareAndSwapInt32(&p.state, int32(Running), int32(Exited))
		return err
	}

	return nil
}

func (p *pipeliner) start() error {
	for _, c := range p.components {
		if err := c.Component.Start(); err != nil {
			return err
		}
	}

	c := p.newExecContext()
	if err := c.Start(); err != nil {
		return err
	}

	p.runningWg.Add(1)
	go func() {
		defer p.runningWg.Done()

		for !p.isStopped() {
			<-time.After(time.Second)
			p.monitor.Set(METRICS_KEY_PIPELINE_UPTIME, monitor.Elapsed(time.Since(p.startTime)))
		}
	}()

	p.runningWg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("Pipeline: %s, Panic: %s, Stack: %s",
					p.Name(), r, string(debug.Stack()))
			}
			c.Stop()

			p.monitor.Set(METRICS_KEY_PIPELINE_EXIT_TIME, monitor.Time(time.Now()))

			p.runningWg.Done()
		}()

		p.monitor.Set(METRICS_KEY_PIPELINE_START_TIME, monitor.Time(time.Now()))

		now := time.Now()
		next := p.schedule.Next(now)
		timer := time.NewTimer(next.Sub(now))
		p.monitor.Set(METRICS_KEY_PIPELINE_NEXT_RUN_TIME, monitor.Time(next))

		for {
			select {
			case <-p.ctx.Done():
				return
			case now = <-timer.C:
				next = p.schedule.Next(now)
				timer.Reset(next.Sub(now))
				p.monitor.Set(METRICS_KEY_PIPELINE_NEXT_RUN_TIME, monitor.Time(next))
				p.monitor.Set(METRICS_KEY_PIPELINE_LAST_START_TIME, monitor.Time(now))

				c.Run()

				p.monitor.Add(METRICS_KEY_PIPELINE_RUN_TIMES, 1)
				p.monitor.Set(METRICS_KEY_PIPELINE_LAST_END_TIME, monitor.Time(time.Now()))
			}
		}
	}()

	return nil
}

func (p *pipeliner) Stop() {
	if p.isStopped() {
		return
	}

	p.cancel()

	p.runningWg.Wait()

	for _, c := range p.components {
		if err := c.Component.Stop(); err != nil {
			log.Error("Failed to stop %s component error: %s", c.Component.Instance().Name(), err)
		}
	}

	atomic.StoreInt32(&p.state, int32(Exited))
}

func (p *pipeliner) isStopped() bool {
	select {
	case <-p.ctx.Done():
		return true
	default:
	}
	return false
}

func (p *pipeliner) State() State {
	return State(atomic.LoadInt32(&p.state))
}

func (p *pipeliner) ListComponents() []Component {
	return p.components
}

func (p *pipeliner) ListProcessors() []Processor {
	return p.processors
}

func (p *pipeliner) Visualize(w io.Writer, format string) error {
	v, ok := visualizers[format]
	if !ok {
		return fmt.Errorf("Unsupported visualize type: %s, supported visualize types: %s",
			format, supportedVisualizerTypes)
	}

	return v(w, p)
}

func (p *pipeliner) Monitor() monitor.Monitor {
	return p.monitor
}

func (p *pipeliner) GetConfig() Config {
	return p.config
}

func (p *pipeliner) SetConfig(Config) error {
	return errors.New("Unimplemented method")
}

func (p *pipeliner) Error() error {
	return ErrorGroup(p.errs).Error()
}
