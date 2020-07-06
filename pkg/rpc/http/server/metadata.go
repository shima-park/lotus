package server

import (
	"github.com/gin-gonic/gin"
	"github.com/shima-park/lotus/pkg/rpc/proto"
)

func (s *Server) getMetadata(c *gin.Context) {
	var snapshot proto.Snapshot
	s.metadata.Snapshot(func(s proto.Snapshot) {
		snapshot = s
	})
	Success(c, proto.MetadataView{
		PluginPaths:         snapshot.PluginPaths,
		ExecutorConfigPaths: snapshot.ExecutorConfigPaths,
		HTTPAddr:            s.options.HTTPAddr,
		Version:             s.options.Version,
		Branch:              s.options.Branch,
		Commit:              s.options.Commit,
		Built:               s.options.Built,
	})
}
