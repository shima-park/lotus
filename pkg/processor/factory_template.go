package processor

import (
	"fmt"
	"reflect"
)

type FactoryTemplate struct {
	sampleConfig string
	description  string
	factoryFunc  FactoryFunc
	example      interface{}
}

func NewFactory(sampleConfig interface{}, description string, example interface{}, factoryFunc FactoryFunc) Factory {
	var conf string
	if sampleConfig != nil {
		t := reflect.TypeOf(sampleConfig)
		if t.Kind() == reflect.String {
			conf = fmt.Sprint(sampleConfig)
		} else {
			m, ok := sampleConfig.(interface {
				Marshal() ([]byte, error)
			})
			if ok {
				data, _ := m.Marshal()
				conf = string(data)
			}
		}
	}

	return FactoryTemplate{
		sampleConfig: conf,
		description:  description,
		factoryFunc:  factoryFunc,
		example:      example,
	}
}

func NewFactoryWithProcessor(sampleConfig interface{}, description string, p Processor) Factory {
	return NewFactory(sampleConfig, description, p, func(string) (Processor, error) {
		return p, nil
	})
}

func (f FactoryTemplate) SampleConfig() string {
	return f.sampleConfig
}

func (f FactoryTemplate) Description() string {
	return f.description
}

func (f FactoryTemplate) New(config string) (Processor, error) {
	return f.factoryFunc(config)
}

func (f FactoryTemplate) Example() Processor {
	return f.example
}
