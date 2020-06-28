package server

import (
	"github.com/gin-gonic/gin"
	"github.com/shima-park/lotus/pkg/pipeline"
	"github.com/shima-park/lotus/pkg/rpc/proto"
)

func (s *Server) listPipelines(c *gin.Context) {
	res, err := s.Pipeline.List()
	if err != nil {
		Failed(c, err)
		return
	}
	Success(c, res)
}

func (s *Server) addPipeline(c *gin.Context) {
	var conf pipeline.Config
	if err := c.BindYAML(&conf); err != nil {
		Failed(c, err)
		return
	}

	err := s.Pipeline.Add(conf)
	if err != nil {
		Failed(c, err)
		return
	}

	Success(c, nil)
}

func (s *Server) removePipeline(c *gin.Context) {
	var names []string
	if err := c.BindJSON(&names); err != nil {
		Failed(c, err)
		return
	}

	err := s.Pipeline.Remove(names...)
	if err != nil {
		Failed(c, err)
		return
	}

	Success(c, nil)
}

func (s *Server) ctrlPipeline(c *gin.Context) {
	err := s.Pipeline.Control(proto.ControlCommand(c.Query("cmd")), c.QueryArray("name"))
	if err != nil {
		Failed(c, err)
		return
	}

	Success(c, nil)
}

func (s *Server) generateConfig(c *gin.Context) {
	var req proto.GenerateConfigRequest
	if err := c.BindJSON(&req); err != nil {
		Failed(c, err)
		return
	}

	config, err := s.Pipeline.GenerateConfig(
		req.Name,
		proto.WithSchedule(req.Schedule),
		proto.WithBootstrap(req.Bootstrap),
		proto.WithComponents(req.Components),
		proto.WithProcessor(req.Processors),
		proto.WithCircuitBreakerRate(req.CircuitBreakerRate),
		proto.WithCircuitBreakerSamples(req.CircuitBreakerSamples),
	)
	if err != nil {
		Failed(c, err)
		return
	}
	Success(c, config)
}

func (s *Server) findPipeline(c *gin.Context) {
	pipe, err := s.Pipeline.Find(c.Query("name"))
	if err != nil {
		Failed(c, err)
		return
	}

	Success(c, pipe)
}

func (s *Server) recreatePipeline(c *gin.Context) {
	var conf pipeline.Config
	if err := c.BindYAML(&conf); err != nil {
		Failed(c, err)
		return
	}
	err := s.Pipeline.Recreate(conf)
	if err != nil {
		Failed(c, err)
		return
	}
	Success(c, nil)
}

func (s *Server) visualizePipeline(c *gin.Context) {
	data, err := s.Pipeline.Visualize(
		c.Query("name"),
		proto.VisualizeFormat(c.Query("format")),
	)
	if err != nil {
		Failed(c, err)
		return
	}

	Success(c, data)
}
