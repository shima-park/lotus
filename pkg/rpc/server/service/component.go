package service

import (
	"fmt"
	"sort"

	"github.com/shima-park/lotus/pkg/component"
	"github.com/shima-park/lotus/pkg/rpc/proto"
	"gopkg.in/yaml.v2"
)

type componentService struct {
}

func NewComponentService() proto.Component {
	return &componentService{}
}

func (s *componentService) List() ([]proto.ComponentView, error) {
	var res []proto.ComponentView
	for _, c := range component.ListFactory() {
		res = append(res, *newComponentView(c.Name, c.Factory))
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})
	return res, nil
}

func (s *componentService) Find(name string) (*proto.ComponentView, error) {
	factory, err := component.GetFactory(name)
	if err != nil {
		return nil, err
	}

	return newComponentView(name, factory), nil
}

func newComponentView(name string, factory component.Factory) *proto.ComponentView {
	return &proto.ComponentView{
		Name:         name,
		SampleConfig: factory.SampleConfig(),
		Description:  factory.Description(),
		InjectName:   getInjectName(factory),
		ReflectType:  fmt.Sprint(factory.ExampleType()),
	}
}

func getInjectName(factory component.Factory) string {
	var sniff struct {
		Name string `yaml:"name"`
	}
	_ = yaml.Unmarshal([]byte(factory.SampleConfig()), &sniff)
	return sniff.Name
}
