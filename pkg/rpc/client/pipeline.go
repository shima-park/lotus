package client

import (
	"net/url"
	"strings"

	pipe "github.com/shima-park/lotus/pkg/pipeline"
	"github.com/shima-park/lotus/pkg/rpc/proto"
)

var _ proto.Pipeline = &pipeline{}

type pipeline struct {
	apiBuilder
}

func (p *pipeline) List() ([]proto.PipelineView, error) {
	var res []proto.PipelineView
	err := GetJSON(p.api("/pipeline/list"), &res)
	return res, err
}

func (p *pipeline) Add(conf pipe.Config) error {
	return PostYaml(p.api("/pipeline/add"), conf, nil)
}

func (p *pipeline) Remove(names ...string) error {
	return PostJSON(p.api("/pipeline/remove"), &names, nil)
}

func (p *pipeline) Control(cmd proto.ControlCommand, names []string) error {
	vals := url.Values{}
	vals.Add("cmd", string(cmd))
	for _, name := range names {
		vals.Add("name", name)
	}
	return GetJSON(p.api("/pipeline/ctrl?"+vals.Encode()), nil)
}

func (p *pipeline) Find(name string) (*proto.PipelineView, error) {
	var res proto.PipelineView
	err := GetJSON(p.api("/pipeline?name="+name), &res)
	return &res, err
}

func (p *pipeline) GenerateConfig(name, schedule string, components, processors []string) (*pipe.Config, error) {
	vals := url.Values{}
	vals.Add("name", name)
	vals.Add("schedule", schedule)
	vals.Add("components", strings.Join(components, ","))
	vals.Add("processors", strings.Join(processors, ","))
	var config pipe.Config
	err := GetJSON(p.api("/pipeline/generate-config?"+vals.Encode()), &config)
	return &config, err
}

func (p *pipeline) Recreate(conf pipe.Config) error {
	return PostYaml(p.api("/pipeline/recreate"), conf, nil)
}
