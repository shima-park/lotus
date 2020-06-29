package pipeline

import (
	"context"
	"expvar"
	"reflect"
	"sync"
	"time"

	"github.com/pkg/errors"

	circuit "github.com/rubyist/circuitbreaker"
	"github.com/shima-park/lotus/pkg/common/inject"
	"github.com/shima-park/lotus/pkg/common/log"
	"github.com/shima-park/lotus/pkg/common/monitor"
)

type execContext struct {
	ctx      context.Context
	cancel   context.CancelFunc
	injector inject.Injector
	stream   *Stream
	monitor  monitor.Monitor
	breaker  *circuit.Breaker
	inputC   chan inject.Injector
	wg       sync.WaitGroup
}

func (c *execContext) Start() error {
	if c.isStopped() {
		return errors.New("Exec context is stopped")
	}

	c.run(c.stream, c.inputC)
	return nil
}

func (c *execContext) Stop() {
	if c.isStopped() {
		return
	}

	c.cancel()
	c.wg.Wait()
}

func (c *execContext) isStopped() bool {
	select {
	case <-c.ctx.Done():
		return true
	default:

	}
	return false
}

func (c *execContext) Run() {
	if c.isStopped() {
		return
	}

	select {
	case <-c.ctx.Done():
		close(c.inputC)
		return
	case c.inputC <- c.injector:
	}
}

func (c *execContext) run(s *Stream, inputC chan inject.Injector) {
	outputC := make(chan inject.Injector, s.config.BufferSize)
	var once sync.Once
	closeFunc := func() {
		once.Do(func() { close(outputC) })
	}

	for i := 0; i < s.config.Replica; i++ {
		c.runStream(s, inputC, outputC, closeFunc)
	}

	if len(s.childs) > 0 {
		childInputs := c.split(outputC, len(s.childs))
		for i := 0; i < len(s.childs); i++ {
			c.run(s.childs[i], childInputs[i])
		}
	} else {
		go func() {
			for range outputC {

			}
		}()
	}
}

func (c *execContext) runStream(s *Stream, inputC, outputC chan inject.Injector, closeFunc func()) {
	moni := c.monitor.With(s.Name())
	moni.Set(METRICS_KEY_STREAM_BUFFER_SIZE, expvar.Func(func() interface{} { return s.config.BufferSize }))
	moni.Set(METRICS_KEY_STREAM_REPLICA, expvar.Func(func() interface{} { return s.config.Replica }))

	c.wg.Add(1)
	go func() {
		defer s.Recover(func() {
			moni.Add(METRICS_KEY_STREAM_RUNNING_REPLICA, -1)
			moni.Set(METRICS_KEY_STREAM_EXIT_TIME, monitor.Time(time.Now()))
			closeFunc()
			c.wg.Done()
		})

		moni.Set(METRICS_KEY_STREAM_START_TIME, monitor.Time(time.Now()))
		moni.Add(METRICS_KEY_STREAM_RUNNING_REPLICA, 1)
		var elapsed time.Duration
		for inj := range inputC {
			if !c.breaker.Ready() { // 熔断器打开
				moni.Add(METRICS_KEY_STREAM_BREAKER_OPEN, 1)
				time.Sleep(time.Second)
				moni.Add(METRICS_KEY_STREAM_BREAKER_OPEN, -1)
			}
			moni.Set(METRICS_KEY_STREAM_LAST_START_TIME, monitor.Time(time.Now()))
			moni.Add(METRICS_KEY_STREAM_RUN_TIMES, 1)
			inj.MapTo(moni, "Monitor", (*monitor.Monitor)(nil))
			startTime := time.Now()

			val, err := s.Invoke(inj)

			elapsed += time.Since(startTime)
			moni.Set(METRICS_KEY_STREAM_ELAPSED, monitor.Elapsed(elapsed))
			moni.Set(METRICS_KEY_STREAM_LAST_END_TIME, monitor.Time(time.Now()))

			newInj, err := handleResult(s.Name(), inj, val, err)
			if err != nil {
				log.Error(err.Error())
				moni.Add(METRICS_KEY_STREAM_ERROR_COUNT, 1)
				moni.Set(METRICS_KEY_STREAM_ERROR, monitor.String(err.Error()))
				c.breaker.Fail()
				continue
			}
			c.breaker.Success()

			// 有些流程没有子流程, 不能根据塞入队列成功来判断
			moni.Add(METRICS_KEY_STREAM_SUCCESS_COUNT, 1)

			if len(s.childs) > 0 {
				select {
				case <-c.ctx.Done():
					return
				case outputC <- newInj:

				}
			}
		}

	}()
}

func handleResult(name string, inj inject.Injector, val reflect.Value, err error) (inject.Injector, error) {
	if err != nil {
		return nil, errors.Wrapf(err, "Stream: %s", name)
	}

	if !val.IsValid() {
		return nil, errors.Errorf("Stream: %s, Return values is not valid", name)
	}

	newInj := inject.New()
	newInj.SetParent(inj)
	if err := newInj.MapValues(val); err != nil {
		return nil, errors.Errorf("Stream: %s, MapValues error: %s", name, err)
	}

	return newInj, nil
}

func (c *execContext) split(in chan inject.Injector, n int) []chan inject.Injector {
	outChans := make([]chan inject.Injector, n)
	for i := 0; i < n; i++ {
		outChans[i] = make(chan inject.Injector)
	}
	go func() {
	Loop:
		for v := range in {
			for _, out := range outChans {
				select {
				case <-c.ctx.Done():
					break Loop
				case out <- v:
				}
			}
		}

		for _, out := range outChans {
			close(out)
		}
	}()
	return outChans
}
