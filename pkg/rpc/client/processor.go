package client

import "github.com/shima-park/lotus/pkg/rpc/proto"

var _ proto.Processor = &processor{}

type processor struct {
	apiBuilder
}

func (c *processor) List() ([]proto.ProcessorView, error) {
	var res []proto.ProcessorView
	err := GetJSON(c.api("/processor/list"), &res)
	return res, err
}

func (c *processor) Find(name string) (*proto.ProcessorView, error) {
	var res proto.ProcessorView
	err := GetJSON(c.api("/processor?name="+name), &res)
	return &res, err
}
