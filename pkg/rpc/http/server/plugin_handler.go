package server

//import (
//	"io/ioutil"
//	"path/filepath"
//
//	"github.com/gin-gonic/gin"
//)
//
//func (s *Server) listPlugins(c *gin.Context) {
//	res, err := s.Plugin.List(c.QueryArray("names")...)
//	if err != nil {
//		Failed(c, err)
//		return
//	}
//
//	Success(c, res)
//}
//
//func (s *Server) addPlugin(c *gin.Context) {
//	pluginFile, err := c.FormFile("plugin")
//	if err != nil {
//		Failed(c, err)
//		return
//	}
//	pluginName := filepath.Base(pluginFile.Filename)
//
//	file, err := pluginFile.Open()
//	if err != nil {
//		Failed(c, err)
//		return
//	}
//
//	body, err := ioutil.ReadAll(file)
//	if err != nil {
//		Failed(c, err)
//		return
//	}
//
//	err = s.Plugin.Add(pluginName, body)
//	if err != nil {
//		Failed(c, err)
//		return
//	}
//
//	Success(c, nil)
//}
//
//func (s *Server) removePlugin(c *gin.Context) {
//	var paths []string
//	err := c.BindJSON(&paths)
//	if err != nil {
//		Failed(c, err)
//		return
//	}
//
//	err = s.Plugin.Remove(paths...)
//	if err != nil {
//		Failed(c, err)
//		return
//	}
//
//	Success(c, nil)
//}
