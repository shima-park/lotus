package client

import (
	"net/url"

	"github.com/shima-park/lotus/pkg/rpc/proto"
	"github.com/shima-park/lotus/pkg/util/http"
)

var _ proto.Executor = &executor{}

type executor struct {
	apiBuilder
}

func (p *executor) Add(_type string, config []byte) error {
	return http.PostYaml(p.api("/executor/add?type="+_type), config, nil)
}

func (p *executor) List() ([]proto.ExecutorView, error) {
	var res []proto.ExecutorView
	err := http.GetJSON(p.api("/executor/list"), &res)
	return res, err
}

func (p *executor) Remove(ids ...string) error {
	return http.PostJSON(p.api("/executor/remove"), &ids, nil)
}

func (p *executor) Control(cmd proto.ControlCommand, ids ...string) error {
	vals := url.Values{}
	vals.Add("cmd", string(cmd))
	for _, id := range ids {
		vals.Add("id", id)
	}
	return http.GetJSON(p.api("/executor/ctrl?"+vals.Encode()), nil)
}

func (p *executor) Find(id string) (*proto.ExecutorView, error) {
	var res proto.ExecutorView
	err := http.GetJSON(p.api("/executor?id="+id), &res)
	return &res, err
}

func (p *executor) Recreate(_type string, config []byte) error {
	vals := url.Values{}
	vals.Add("type", _type)
	vals.Add("is_recreate", "true")
	return http.PostYaml(p.api("/executor/recreate?"+vals.Encode()), config, nil)
}

func (p *executor) Visualize(format proto.VisualizeFormat, id string) ([]byte, error) {
	vals := url.Values{}
	vals.Add("id", id)
	vals.Add("format", string(format))

	var data []byte
	err := http.GetJSON(p.api("/executor/visualize?"+vals.Encode()), &data)
	return data, err
}
