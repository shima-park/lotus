package server

import (
	"github.com/gin-gonic/gin"
	"github.com/shima-park/lotus/pkg/rpc/proto"
	"github.com/shima-park/lotus/pkg/rpc/service"
)

type Server struct {
	options Options
	engine  *gin.Engine

	metadata proto.Metadata
	*service.Service
}

func New(opts ...Option) (*Server, error) {
	c := &Server{
		options: defaultOptions,
		engine:  gin.Default(),
	}

	for _, opt := range opts {
		opt(&c.options)
	}

	var err error
	c.metadata, err = service.NewMetadata(c.options.MetadataPath)
	if err != nil {
		return nil, err
	}

	c.Service = service.NewService(c.metadata)

	return c, nil
}

func (c *Server) Serve() error {
	if err := c.Start(); err != nil {
		return err
	}

	if c.options.HTTPAddr != "" {
		c.setRouter()

		return c.engine.Run(c.options.HTTPAddr)
	}
	return nil
}

func (c *Server) Stop() {
	_ = c.Service.Stop()
}
