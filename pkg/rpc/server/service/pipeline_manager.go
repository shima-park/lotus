package service

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/shima-park/lotus/pkg/pipeline"
)

type PipelinerManager interface {
	AddPipeline(config pipeline.Config) (pipeline.Pipeliner, error)
	RemovePipeline(name ...string) error
	RecreatePipeline(config pipeline.Config) (pipeline.Pipeliner, error)
	List() []pipeline.Pipeliner
	Find(name string) pipeline.Pipeliner
	Restart(name ...string) error
	Start(name ...string) error
	Stop(name ...string) error
}

type pipelinerManager struct {
	rwlock    sync.RWMutex
	pipelines map[string]pipeline.Pipeliner // key: name value: Pipeliner
}

func NewPipelinerManager() PipelinerManager {
	pm := &pipelinerManager{
		pipelines: map[string]pipeline.Pipeliner{},
	}
	return pm
}

func (p *pipelinerManager) List() []pipeline.Pipeliner {
	var ps []pipeline.Pipeliner

	p.rwlock.RLock()
	for _, p := range p.pipelines {
		ps = append(ps, p)
	}
	p.rwlock.RUnlock()

	sort.Slice(ps, func(i, j int) bool {
		return ps[i].Name() < ps[j].Name()
	})

	return ps
}

func (p *pipelinerManager) Find(name string) pipeline.Pipeliner {
	p.rwlock.Lock()
	defer p.rwlock.Unlock()

	return p.find(name)
}

func (p *pipelinerManager) find(name string) pipeline.Pipeliner {
	return p.pipelines[name]
}

func (p *pipelinerManager) AddPipeline(config pipeline.Config) (pipeline.Pipeliner, error) {
	p.rwlock.Lock()
	defer p.rwlock.Unlock()

	return p.addPipeline(config)
}

func (p *pipelinerManager) addPipeline(config pipeline.Config) (pipeline.Pipeliner, error) {
	_, ok := p.pipelines[config.Name]
	if ok {
		return nil, fmt.Errorf("Pipeline: %s is already register", config.Name)
	}

	pipe := pipeline.NewPipelineByConfig(config)
	p.pipelines[config.Name] = pipe
	return pipe, nil
}

func (p *pipelinerManager) RemovePipeline(names ...string) error {
	return p.doByName(false, names, p.removePipeline)
}

func (p *pipelinerManager) removePipeline(pipe pipeline.Pipeliner) error {
	pipe.Stop()
	delete(p.pipelines, pipe.Name())
	return nil
}

func (p *pipelinerManager) RecreatePipeline(config pipeline.Config) (pipeline.Pipeliner, error) {
	name := config.Name
	var pipe pipeline.Pipeliner
	err := p.doByName(false, []string{name}, func(oldPipe pipeline.Pipeliner) error {
		err := p.removePipeline(oldPipe)
		if err != nil {
			return errors.Wrapf(err, "Pipeline: %s", name)
		}

		pipe, err = p.addPipeline(config)
		if err != nil {
			return errors.Wrapf(err, "Pipeline: %s", name)
		}

		if oldPipe.State() != pipeline.Running {
			return nil
		}
		return p.start(pipe)
	})

	return pipe, err
}

func (p *pipelinerManager) Restart(names ...string) error {
	return p.doByName(false, names, func(oldPipe pipeline.Pipeliner) error {
		name := oldPipe.Name()
		err := p.removePipeline(oldPipe)
		if err != nil {
			return errors.Wrapf(err, "Pipeline: %s", name)
		}

		pipe, err := p.addPipeline(oldPipe.GetConfig())
		if err != nil {
			return errors.Wrapf(err, "Pipeline: %s", name)
		}

		return p.start(pipe)
	})
}

func (p *pipelinerManager) Start(names ...string) error {
	return p.doByName(true, names, p.start)
}

func (p *pipelinerManager) start(pipe pipeline.Pipeliner) error {
	if pipe.Error() != nil {
		return pipe.Error()
	}
	switch pipe.State() {
	case pipeline.Idle:
		err := pipe.Start()
		if err != nil {
			return errors.Wrapf(err, "Pipeline: %s", pipe.Name())
		}
	case pipeline.Exited:
		return fmt.Errorf("Pipeline(%s)'s state is exited, please try to restart it", pipe.Name())
	}
	return nil
}

func (p *pipelinerManager) Stop(names ...string) error {
	return p.doByName(true, names, func(pipe pipeline.Pipeliner) error {
		switch pipe.State() {
		case pipeline.Running:
			pipe.Stop()
		}
		return nil
	})
}

func (p *pipelinerManager) doByName(isReadLock bool, names []string, callback func(pipe pipeline.Pipeliner) error) error {
	if isReadLock {
		p.rwlock.RLock()
		defer p.rwlock.RUnlock()
	} else {
		p.rwlock.Lock()
		defer p.rwlock.Unlock()
	}

	var errs []string
	for _, name := range names {
		pipe := p.find(name)
		if pipe == nil {
			errs = append(errs, fmt.Sprintf("Pipeline: %s is not found", name))
			continue
		}

		err := callback(pipe)
		if err != nil {
			errs = append(errs, fmt.Sprintf("Pipeline: %s %v", name, err))
			continue
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ""))
	}

	return nil
}
