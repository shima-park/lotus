package service

import (
	"io/ioutil"
	"os"

	"github.com/shima-park/lotus/pkg/common/log"
	"github.com/shima-park/lotus/pkg/common/plugin"
	"github.com/shima-park/lotus/pkg/pipeline"
	"github.com/shima-park/lotus/pkg/rpc/proto"
	"gopkg.in/yaml.v2"
)

type Service struct {
	metadata        proto.Metadata
	pipelineManager PipelinerManager
	proto.Pipeline
	proto.Component
	proto.Processor
	proto.Plugin
	proto.Server
}

func NewService(metadata proto.Metadata) *Service {
	s := &Service{
		metadata:        metadata,
		pipelineManager: NewPipelinerManager(),
	}

	s.Pipeline = NewPipelineService(metadata, s.pipelineManager)
	s.Component = NewComponentService()
	s.Processor = NewProcessorService()
	s.Plugin = NewPluginService(metadata)

	return s
}

func (s *Service) Start() error {
	plugins := s.metadata.ListPaths(proto.FileTypePlugin)
	for _, path := range plugins {
		if s.cleanIfNotExists(proto.FileTypePlugin, path) {
			continue
		}
		err := plugin.LoadPlugins(path)
		if err != nil {
			return err
		}
	}

	pipes := s.metadata.ListPaths(proto.FileTypePipelineConfig)
	for _, path := range pipes {
		if s.cleanIfNotExists(proto.FileTypePipelineConfig, path) {
			continue
		}
		err := loadPipelineFromFile(path, s.pipelineManager)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) cleanIfNotExists(ft proto.FileType, path string) bool {
	_, err := os.Lstat(path)
	if os.IsNotExist(err) {
		log.Info("Remove %s %s because it's not exists", ft, path)
		rerr := s.metadata.RemovePath(ft, path)
		if rerr != nil {
			log.Error("Remove %s %s error: %v", ft, path, rerr)
		}
		return true
	}
	return false
}

func loadPipelineFromFile(path string, pm PipelinerManager) error {
	log.Info("loading pipeline from: %s", path)

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	var conf pipeline.Config
	err = yaml.Unmarshal(content, &conf)
	if err != nil {
		return err
	}

	_, err = pm.AddPipeline(conf)
	if err != nil {
		return err
	}

	if conf.Bootstrap {
		_ = pm.Start(conf.Name)
	}

	return nil
}

func (s *Service) Stop() error {
	for _, p := range s.pipelineManager.List() {
		p.Stop()
	}
	return nil
}
