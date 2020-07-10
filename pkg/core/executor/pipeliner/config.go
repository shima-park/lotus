package pipeliner

import (
	"github.com/pkg/errors"
	errgrp "github.com/shima-park/lotus/pkg/util/errors"
	"github.com/shima-park/seed/component"
	"github.com/shima-park/seed/processor"
)

type Config struct {
	Name                  string              `yaml:"name"`
	Kind                  string              `yaml:"kind"`
	Schedule              string              `yaml:"schedule"`                // 调度计划，为空时死循环调度，可以传入cron表达式调度
	CircuitBreakerSamples int64               `yaml:"circuit_breaker_samples"` // 熔断器采样数, 防止stream出现异常耗尽cpu资源
	CircuitBreakerRate    float64             `yaml:"circuit_breaker_rate"`    // 熔断器采样率
	Components            []map[string]string `yaml:"components"`              // key: name, value: rawConfig
	Processors            []map[string]string `yaml:"processors"`              // key: name, value: rawConfig
	Stream                StreamConfig        `yaml:"stream"`                  // key: name, value: StreamConfig
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

func (c Config) NewComponents() ([]Component, error) {
	var components []Component
	var eg errgrp.ErrorGroup
	for _, name2config := range c.Components {
		for componentName, rawConfig := range name2config {
			factory, err := component.GetFactory(componentName)
			if err != nil {
				eg = append(eg, errors.Wrapf(err, "Component: %s", componentName))
				continue
			}
			c, err := factory.New(rawConfig)
			if err != nil {
				eg = append(eg, errors.Wrapf(err, "Component: %s", componentName))
				continue
			}
			components = append(components, Component{
				Name:      componentName,
				RawConfig: rawConfig,
				Component: c,
				Factory:   factory,
			})
		}
	}

	return components, eg.Error()
}

func (c Config) NewProcessors() ([]Processor, error) {
	var processors []Processor
	var eg errgrp.ErrorGroup
	for _, name2config := range c.Processors {
		for processorName, rawConfig := range name2config {
			factory, err := processor.GetFactory(processorName)
			if err != nil {
				eg = append(eg, errors.Wrapf(err, "Processor: %s", processorName))
				continue
			}

			p, err := factory.New(rawConfig)
			if err != nil {
				eg = append(eg, errors.Wrapf(err, "Processor: %s", processorName))
				continue
			}
			processors = append(processors, Processor{
				Name:      processorName,
				RawConfig: rawConfig,
				Processor: p,
				Factory:   factory,
			})
		}
	}
	return processors, eg.Error()
}
