package server

import (
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/shima-park/lotus/pkg/rpc/proto"
)

func (s *Server) listExecutors(c *gin.Context) {
	res, err := s.Executor.List()
	if err != nil {
		Failed(c, err)
		return
	}
	Success(c, res)
}

func (s *Server) addExecutor(c *gin.Context) {
	_type := c.Query("type")
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Failed(c, err)
		return
	}

	isRecreate := c.Query("is_recreate") == "true"
	if isRecreate {
		err = s.Executor.Recreate(_type, body)
	} else {
		err = s.Executor.Add(_type, body)
	}
	if err != nil {
		Failed(c, err)
		return
	}

	Success(c, nil)
}

func (s *Server) removeExecutor(c *gin.Context) {
	var ids []string
	if err := c.BindJSON(&ids); err != nil {
		Failed(c, err)
		return
	}

	err := s.Executor.Remove(ids...)
	if err != nil {
		Failed(c, err)
		return
	}

	Success(c, nil)
}

func (s *Server) ctrlExecutor(c *gin.Context) {
	err := s.Executor.Control(proto.ControlCommand(c.Query("cmd")), c.QueryArray("name")...)
	if err != nil {
		Failed(c, err)
		return
	}

	Success(c, nil)
}

func (s *Server) generateConfig(c *gin.Context) {
	//var req proto.GenerateConfigRequest
	//if err := c.BindJSON(&req); err != nil {
	//	Failed(c, err)
	//	return
	//}
	//
	//config, err := s.Executor.GenerateConfig(
	//	req.Name,
	//	proto.WithSchedule(req.Schedule),
	//	proto.WithBootstrap(req.Bootstrap),
	//	proto.WithComponents(req.Components),
	//	proto.WithProcessor(req.Processors),
	//	proto.WithCircuitBreakerRate(req.CircuitBreakerRate),
	//	proto.WithCircuitBreakerSamples(req.CircuitBreakerSamples),
	//)
	//if err != nil {
	//	Failed(c, err)
	//	return
	//}
	//Success(c, config)
}

func (s *Server) findExecutor(c *gin.Context) {
	pipe, err := s.Executor.Find(c.Query("name"))
	if err != nil {
		Failed(c, err)
		return
	}

	Success(c, pipe)
}

func (s *Server) visualizeExecutor(c *gin.Context) {
	data, err := s.Executor.Visualize(
		proto.VisualizeFormat(c.Query("format")),
		c.Query("id"),
	)
	if err != nil {
		Failed(c, err)
		return
	}

	Success(c, data)
}
