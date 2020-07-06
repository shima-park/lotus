package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shima-park/lotus/pkg/rpc/proto"
)

func (s *Server) setRouter() {
	r := s.engine
	r.POST("/executor/generate-config", s.generateConfig)
	r.POST("/executor/add", s.addExecutor)
	r.POST("/executor/remove", s.removeExecutor)
	r.GET("/executor/ctrl", s.ctrlExecutor)
	r.GET("/executor/list", s.listExecutors)
	r.GET("/executor/visualize", s.visualizeExecutor)
	r.GET("/executor", s.findExecutor)

	r.GET("/component/list", s.listComponents)
	r.GET("/component", s.findComponent)

	r.GET("/processor/list", s.listProcessors)
	r.GET("/processor", s.findProcessor)

	r.GET("/plugin/list", s.listPlugins)
	r.POST("/plugin/upload", s.uploadPlugin)
	r.POST("/plugin/remove", s.removePlugin)

	r.GET("/metadata", s.getMetadata)
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
