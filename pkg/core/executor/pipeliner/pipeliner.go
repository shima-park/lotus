package pipeliner

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
	"github.com/shima-park/lotus/pkg/core/common/inject"
	"github.com/shima-park/lotus/pkg/core/common/log"
	"github.com/shima-park/lotus/pkg/core/common/monitor"
	errgrp "github.com/shima-park/lotus/pkg/util/errors"
	"github.com/shima-park/seed/executor"
)

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

func NewPipelineByConfig(conf Config) *pipeliner {
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

		state: int32(executor.Idle),
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
		p.errs = append(p.errs, errors.Wrapf(err, "Pipeline: %s NewComponents", conf.Name))
	}

	p.processors, err = conf.NewProcessors()
	if err != nil {
		p.errs = append(p.errs, errors.Wrapf(err, "Pipeline: %s NewProcessors", conf.Name))
	}

	var pm = map[string]Processor{}
	for _, p := range p.processors {
		pm[p.Name] = p
	}

	p.stream, err = NewStream(conf.Stream, pm)
	if err != nil {
		p.errs = append(p.errs, errors.Wrapf(err, "Pipeline: %s NewStream", conf.Name))
	}

	if p.config.Schedule != "" && p.parser != nil {
		p.schedule, err = p.parser(p.config.Schedule)
		if err != nil {
			p.errs = append(p.errs,
				errors.Wrapf(err, "Pipeline: %s parse schedule %s", conf.Name, p.config.Schedule))
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
	if len(p.errs) > 0 {
		return errgrp.ErrorGroup(p.errs).Error()
	}

	if !atomic.CompareAndSwapInt32(&p.state, int32(executor.Idle), int32(executor.Running)) {
		return nil
	}

	if err := p.start(); err != nil {
		p.errs = append(p.errs, err)
		atomic.CompareAndSwapInt32(&p.state, int32(executor.Running), int32(executor.Exited))
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
					p.name, r, string(debug.Stack()))
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

	atomic.StoreInt32(&p.state, int32(executor.Exited))
}

func (p *pipeliner) isStopped() bool {
	select {
	case <-p.ctx.Done():
		return true
	default:
	}
	return false
}

func (p *pipeliner) State() executor.State {
	return executor.State(atomic.LoadInt32(&p.state))
}

func (p *pipeliner) Visualize(w io.Writer, format string) error {
	v, ok := visualizers[format]
	if !ok {
		return fmt.Errorf("Unsupported visualize type: %s, supported visualize types: %s",
			format, supportedVisualizerTypes)
	}

	return v(w, p)
}
