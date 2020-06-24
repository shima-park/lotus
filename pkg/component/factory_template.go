package component

import (
	"fmt"
	"reflect"
)

type FactoryTemplate struct {
	sampleConfig string
	description  string
	factoryFunc  FactoryFunc
	exampleType  reflect.Type
}

func NewFactory(sampleConfig interface{}, description string, _type reflect.Type, factoryFunc FactoryFunc) Factory {
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
		exampleType:  _type,
	}
}

func NewFactoryWithProcessor(sampleConfig interface{}, description string, _type reflect.Type, p Component) Factory {
	return NewFactory(sampleConfig, description, _type, func(string) (Component, error) {
		return p, nil
	})
}

func (f FactoryTemplate) SampleConfig() string {
	return f.sampleConfig
}

func (f FactoryTemplate) Description() string {
	return f.description
}

func (f FactoryTemplate) New(config string) (Component, error) {
	return f.factoryFunc(config)
}

func (f FactoryTemplate) ExampleType() reflect.Type {
	return f.exampleType
}
