package client

import (
	"github.com/shima-park/lotus/pkg/rpc/proto"
	"github.com/shima-park/lotus/pkg/util/http"
)

var _ proto.Component = &component{}

type component struct {
	apiBuilder
}

func (c *component) List() ([]proto.ComponentView, error) {
	var res []proto.ComponentView
	err := http.GetJSON(c.api("/component/list"), &res)
	return res, err
}

func (c *component) Find(name string) (*proto.ComponentView, error) {
	var res proto.ComponentView
	err := http.GetJSON(c.api("/component?name="+name), &res)
	return &res, err
}
