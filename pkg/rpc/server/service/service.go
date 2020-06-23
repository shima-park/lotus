package service

import (
	"io/ioutil"

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
	for _, path := range s.metadata.ListPaths(proto.FileTypePlugin) {
		err := plugin.LoadPlugins(path)
		if err != nil {
			return err
		}
	}

	for _, path := range s.metadata.ListPaths(proto.FileTypePipelineConfig) {
		err := loadPipelineFromFile(path, s.pipelineManager)
		if err != nil {
			return err
		}
	}
	return nil
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

	return nil
}

func (s *Service) Stop() error {
	for _, p := range s.pipelineManager.List() {
		p.Stop()
	}
	return nil
}
