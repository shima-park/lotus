package client

import "github.com/shima-park/lotus/pkg/rpc/proto"

var _ proto.Component = &component{}

type component struct {
	apiBuilder
}

func (c *component) List() ([]proto.ComponentView, error) {
	var res []proto.ComponentView
	err := GetJSON(c.api("/component/list"), &res)
	return res, err
}

func (c *component) Find(name string) (*proto.ComponentView, error) {
	var res proto.ComponentView
	err := GetJSON(c.api("/component?name="+name), &res)
	return &res, err
}
