package executor

import (
	"fmt"
)

type Factory interface {
	// 组件示例配置
	SampleConfig() string

	// 组件描述
	Description() string

	// 创建实例
	New(config string) (Executor, error)
}

type FactoryFunc func(config string) (Executor, error)

var registry = make(map[string]Factory)

func Register(name string, factory Factory) error {
	if name == "" {
		return fmt.Errorf("Error registering executor: name cannot be empty")
	}
	if factory == nil {
		return fmt.Errorf("Error registering executor '%v': factory cannot be empty", name)
	}
	if _, exists := registry[name]; exists {
		return fmt.Errorf("Error registering executor '%v': already registered", name)
	}

	registry[name] = factory

	return nil
}

func GetFactory(name string) (Factory, error) {
	if _, exists := registry[name]; !exists {
		return nil, fmt.Errorf("Error creating executor. No such executor type exist: '%v'", name)
	}
	return registry[name], nil
}

type NamedFactory struct {
	Name    string
	Factory Factory
}

func ListFactory() []NamedFactory {
	var list []NamedFactory
	for name, factory := range registry {
		list = append(list, NamedFactory{
			Name:    name,
			Factory: factory,
		})
	}
	return list
}
