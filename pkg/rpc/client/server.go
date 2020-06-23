package client

import "github.com/shima-park/lotus/pkg/rpc/proto"

var _ proto.Server = &server{}

type server struct {
	apiBuilder
}

func (s *server) Metadata() (proto.MetadataView, error) {
	var ret proto.MetadataView
	err := GetJSON(s.api("/metadata"), &ret)
	return ret, err
}
