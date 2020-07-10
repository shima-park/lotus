package server

import (
	"github.com/gin-gonic/gin"
	"github.com/shima-park/lotus/pkg/rpc/service"
)

type Server struct {
	options Options
	engine  *gin.Engine

	service.LotusService
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
	c.LotusService, err = service.NewLotusService(
		c.options.MetadataPath,
		c.options.HTTPAddr,
	)

	return c, err
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
	c.Stop()
}
