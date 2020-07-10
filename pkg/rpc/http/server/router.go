package server

import (
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/shima-park/lotus/pkg/rpc/proto"
)

func (s *Server) setRouter() {
	r := s.engine
	r.POST("/executor/generate-config", s.generateExecutorConfig)
	r.POST("/executor/add", s.addExecutor)
	r.POST("/executor/put", s.putExecutor)
	r.GET("/executor/get", s.getExecutor)
	r.POST("/executor/remove", s.removeExecutor)
	r.GET("/executor/list", s.listExecutors)

	r.POST("/executor-process/register", s.registerExecutorProcess)
	r.POST("/executor-process/remove", s.removeExecutorProcess)
	r.GET("/executor-process/list", s.listExecutorProcesses)

	r.GET("/component/list", s.listComponents)

	r.GET("/processor/list", s.listProcessors)

	r.GET("/plugin/list", s.listPlugins)
	r.POST("/plugin/add", s.addPlugin)
	r.POST("/plugin/remove", s.removePlugin)
	r.POST("/plugin/put", s.putPlugin)
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, proto.Result{
		Data: data,
	})
}

func Failed(c *gin.Context, err error) {
	c.JSON(http.StatusOK, proto.Result{
		Code: http.StatusInternalServerError,
		Msg:  err.Error(),
	})
}

func (s *Server) generateExecutorConfig(c *gin.Context) {
	var req proto.GenerateExecutorConfigRequest
	if err := c.BindJSON(&req); err != nil {
		Failed(c, err)
		return
	}

	config, err := s.GenerateExecutorConfig(&req)
	if err != nil {
		Failed(c, err)
		return
	}
	Success(c, config)
}

func (s *Server) addExecutor(c *gin.Context) {
	var req proto.AddExecutorRequest
	if err := c.BindJSON(&req); err != nil {
		Failed(c, err)
		return
	}

	err := s.PutExecutorConfig(req.Config, false)
	if err != nil {
		Failed(c, err)
		return
	}
	Success(c, nil)
}

func (s *Server) putExecutor(c *gin.Context) {
	var req proto.PutExecutorRequest
	if err := c.BindJSON(&req); err != nil {
		Failed(c, err)
		return
	}

	err := s.PutExecutorConfig(req.Config, true)
	if err != nil {
		Failed(c, err)
		return
	}
	Success(c, nil)
}

func (s *Server) getExecutor(c *gin.Context) {
	config, err := s.GetExecutorConfig(c.Query("name"))
	if err != nil {
		Failed(c, err)
		return
	}
	Success(c, &proto.GetExecutorResponse{
		Config: config,
	})
}

func (s *Server) removeExecutor(c *gin.Context) {
	var req proto.RemoveExecutorRequest
	if err := c.BindJSON(&req); err != nil {
		Failed(c, err)
		return
	}

	err := s.RemoveExecutor(req.Names...)
	if err != nil {
		Failed(c, err)
		return
	}
	Success(c, nil)
}

func (s *Server) listExecutors(c *gin.Context) {
	res, err := s.ListExecutors(c.QueryArray("names")...)
	if err != nil {
		Failed(c, err)
		return
	}
	Success(c, res)
}

func (s *Server) registerExecutorProcess(c *gin.Context) {
	var req proto.RegisterExecutorProcessRequest
	if err := c.BindJSON(&req); err != nil {
		Failed(c, err)
		return
	}

	err := s.RegisterExecutorProcess(req.Name, req.ID, req.ServeAddr)
	if err != nil {
		Failed(c, err)
		return
	}
	Success(c, nil)
}

func (s *Server) removeExecutorProcess(c *gin.Context) {
	var req proto.RemoveExecutorProcessRequest
	if err := c.BindJSON(&req); err != nil {
		Failed(c, err)
		return
	}

	err := s.RemoveExecutorProcess(req.Name)
	if err != nil {
		Failed(c, err)
		return
	}
	Success(c, nil)
}

func (s *Server) listExecutorProcesses(c *gin.Context) {
	res, err := s.ListExecutorProcesses()
	if err != nil {
		Failed(c, err)
		return
	}
	Success(c, res)
}

func (s *Server) listComponents(c *gin.Context) {
	res, err := s.ListComponents(c.QueryArray("names")...)
	if err != nil {
		Failed(c, err)
		return
	}
	Success(c, res)
}

func (s *Server) listProcessors(c *gin.Context) {
	res, err := s.ListProcessors(c.QueryArray("names")...)
	if err != nil {
		Failed(c, err)
		return
	}
	Success(c, res)
}

func (s *Server) listPlugins(c *gin.Context) {
	res, err := s.ListPlugins(c.QueryArray("names")...)
	if err != nil {
		Failed(c, err)
		return
	}
	Success(c, res)
}

func (s *Server) addPlugin(c *gin.Context) {
	s.uploadPlugin(c, false)
}

func (s *Server) removePlugin(c *gin.Context) {
	var req proto.RemovePluginRequest
	if err := c.BindJSON(&req); err != nil {
		Failed(c, err)
		return
	}

	err := s.RemovePlugin(req.Names...)
	if err != nil {
		Failed(c, err)
		return
	}
	Success(c, nil)
}

func (s *Server) putPlugin(c *gin.Context) {
	s.uploadPlugin(c, true)
}

func (s *Server) uploadPlugin(c *gin.Context, isOverwrite bool) {
	pluginFile, err := c.FormFile("plugin")
	if err != nil {
		Failed(c, err)
		return
	}
	pluginName := filepath.Base(pluginFile.Filename)

	file, err := pluginFile.Open()
	if err != nil {
		Failed(c, err)
		return
	}

	body, err := ioutil.ReadAll(file)
	if err != nil {
		Failed(c, err)
		return
	}

	err = s.PutPlugin(pluginName, body, isOverwrite)
	if err != nil {
		Failed(c, err)
		return
	}
	Success(c, nil)
}
