package client

import (
	"net/url"

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

func (p *pipeline) GenerateConfig(name string, opts ...proto.ConfigOption) (*pipe.Config, error) {
	options := proto.NewConfigOptions(opts...)
	req := proto.GenerateConfigRequest{
		Name:                  name,
		Schedule:              options.Schedule,
		CircuitBreakerSamples: options.CircuitBreakerSamples,
		CircuitBreakerRate:    options.CircuitBreakerRate,
		Bootstrap:             options.Bootstrap,
		Components:            options.Components,
		Processors:            options.Processors,
	}
	var config pipe.Config
	err := PostJSON(p.api("/pipeline/generate-config"), req, &config)
	return &config, err
}

func (p *pipeline) Recreate(conf pipe.Config) error {
	return PostYaml(p.api("/pipeline/recreate"), conf, nil)
}

func (p *pipeline) Visualize(name string, format proto.VisualizeFormat) ([]byte, error) {
	vals := url.Values{}
	vals.Add("name", name)
	vals.Add("format", string(format))

	var data []byte
	err := GetJSON(p.api("/pipeline/visualize?"+vals.Encode()), &data)
	return data, err
}
