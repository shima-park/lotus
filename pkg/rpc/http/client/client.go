package client

import (
	"github.com/shima-park/lotus/pkg/rpc/proto"
	"github.com/shima-park/lotus/pkg/util/http"
)

type Client struct {
	proto.Executor
	proto.Component
	proto.Processor
	proto.Plugin
	proto.Server
}

func NewClient(addr string) *Client {
	addr = http.NormalizeURL(addr)
	b := apiBuilder{addr}
	return &Client{
		Executor:  &executor{b},
		Component: &component{b},
		Processor: &processor{b},
		Plugin:    &plugin{b},
		Server:    &server{b},
	}
}

type apiBuilder struct {
	addr string
}

func (b *apiBuilder) api(path string) string {
	return b.addr + path
}
