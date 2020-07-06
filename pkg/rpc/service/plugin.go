package service

import (
	"io/ioutil"
	"os"

	"github.com/shima-park/lotus/pkg/common/plugin"
	"github.com/shima-park/lotus/pkg/rpc/proto"
)

type pluginService struct {
	metadata proto.Metadata
}

func NewPluginService(metadata proto.Metadata) proto.Plugin {
	return &pluginService{
		metadata: metadata,
	}
}

func (s *pluginService) List() ([]proto.PluginView, error) {
	var res []proto.PluginView
	for _, p := range plugin.List() {
		res = append(res, proto.PluginView{
			Name:     p.Name,
			Path:     p.Path,
			Module:   p.Module,
			OpenTime: p.OpenTime.Format("2006-01-02 15:04:05"),
		})
	}
	return res, nil
}

func (s *pluginService) AddPath(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	fi, err := file.Stat()
	if err != nil {
		return err
	}

	return s.Add(fi.Name(), data)
}

func (s *pluginService) Add(pluginName string, bin []byte) error {
	path, err := s.metadata.PutPlugin(pluginName, bin)
	if err != nil {
		return err
	}

	err = TestPlugin(path)
	if err != nil {
		//rerr := s.metadata.RemovePluginPath(path)
		//if rerr != nil {
		//	log.Error("Failed to plugin path: %s error: %s", path, rerr)
		//}
		return err
	}
	return nil
}

func (s *pluginService) Remove(names ...string) error {
	var eg ErrorGroup
	for _, p := range plugin.List() {
		err := s.metadata.RemovePluginPath(p.Path)
		if err != nil {
			eg = append(eg, err)
			continue
		}
	}

	return eg.Error()
}
