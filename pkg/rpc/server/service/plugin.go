package service

import (
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
		if !s.metadata.ExistsPath(proto.FileTypePlugin, p.Path) {
			continue
		}
		res = append(res, proto.PluginView{
			Name:     p.Name,
			Path:     p.Path,
			Module:   p.Module,
			OpenTime: p.OpenTime.Format("2006-01-02 15:04:05"),
		})
	}
	return res, nil
}

func (s *pluginService) Open(path string) error {
	return plugin.LoadPlugins(path)
}

func (s *pluginService) Add(path string) error {
	err := plugin.LoadPlugins(path)
	if err != nil {
		return err
	}
	return s.metadata.AddPath(proto.FileTypePlugin, path)
}

func (s *pluginService) Remove(names ...string) error {
	m := map[string]string{} // name:path
	for _, p := range plugin.List() {
		m[p.Name] = p.Path
	}

	var errs []string
	for _, name := range names {
		path, ok := m[name]
		if ok {
			err := s.metadata.RemovePath(proto.FileTypePlugin, path)
			if err != nil {
				errs = append(errs, err.Error())
				continue
			}
		}
	}

	return ErrorGroup(errs).Error()
}
