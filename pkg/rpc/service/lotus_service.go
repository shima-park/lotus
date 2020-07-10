package service

import (
	"io/ioutil"

	"github.com/shima-park/lotus/pkg/rpc/proto"
	"github.com/shima-park/lotus/pkg/rpc/service/child_process/component"
	"github.com/shima-park/lotus/pkg/rpc/service/child_process/executor"
	_ "github.com/shima-park/lotus/pkg/rpc/service/child_process/include"

	"github.com/shima-park/lotus/pkg/rpc/service/child_process/plugin"
	"github.com/shima-park/lotus/pkg/rpc/service/child_process/processor"
)

type lotusService struct {
	metadata               *metadata
	executorProcessManager ExecutorProcessManager
}

func NewLotusService(metapath, masterAddr string) (LotusService, error) {
	metadata, err := newMetadata(metapath)
	if err != nil {
		return nil, err
	}

	s := &lotusService{
		metadata: metadata,
	}

	s.executorProcessManager = NewExecutorProcessManager(
		masterAddr,
		s.getPluginPaths,
	)

	return s, nil
}

func (s *lotusService) Start() error {
	for kind, path := range s.getExecutorConfigPaths() {
		err := s.executorProcessManager.Run(kind, path)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *lotusService) Stop() {
	s.executorProcessManager.Stop()
}

func (s *lotusService) getPluginPaths() []string {
	var paths []string
	s.metadata.read(func(r registry) error {
		for _, path := range r.PluginPaths {
			paths = append(paths, path)
		}
		return nil
	})
	return paths
}

func (s *lotusService) getExecutorConfigPaths() map[string]string {
	var kind2paths = map[string]string{}
	s.metadata.read(func(r registry) error {
		kind2paths = r.ExecutorConfigPaths
		return nil
	})
	return kind2paths
}

func (s *lotusService) GenerateExecutorConfig(req *proto.GenerateExecutorConfigRequest) (string, error) {
	return executor.GenerateExecutorConfig(
		req.Type, req.Components, req.Processors, s.getPluginPaths())
}

func (s *lotusService) RemoveExecutor(names ...string) error {
	return s.metadata.removeExecutorConfig(names...)
}

func (s *lotusService) ListExecutors(names ...string) ([]proto.ExecutorView, error) {
	var res []proto.ExecutorView
	s.metadata.read(func(r registry) error {
		for name, _ := range r.ExecutorConfigPaths {
			res = append(res, proto.ExecutorView{
				Name: name,
			})
		}
		return nil
	})
	return res, nil
}

func (s *lotusService) GetExecutorConfig(name string) ([]byte, error) {
	return s.metadata.loadExecutorConfig(name)
}

func (s *lotusService) PutExecutorConfig(edited []byte, isOverwrite bool) error {
	name, _, err := SniffNameAndKind(edited)
	if err != nil {
		return err
	}
	return s.metadata.putExecutorConfig(name, edited, isOverwrite)
}

func (s *lotusService) RunExecutorProcess(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	_, kind, err := SniffNameAndKind(data)
	if err != nil {
		return err
	}
	return s.executorProcessManager.Run(kind, path)
}

func (s *lotusService) RegisterExecutorProcess(name, id, slaveAddr string) error {
	return s.executorProcessManager.Register(name, id, slaveAddr)
}

func (s *lotusService) ListExecutorProcesses() ([]proto.ExecutorProcessView, error) {
	return s.executorProcessManager.List()
}

func (s *lotusService) RemoveExecutorProcess(names ...string) error {
	return s.RemoveExecutorProcess(names...)
}

func (s *lotusService) ListComponents(names ...string) ([]proto.ComponentView, error) {
	return component.ListComponents(s.getPluginPaths(), names)
}

func (s *lotusService) ListProcessors(names ...string) ([]proto.ProcessorView, error) {
	return processor.ListProcessors(s.getPluginPaths(), names)
}

func (s *lotusService) ListPlugins(names ...string) ([]proto.PluginView, error) {
	return plugin.ListPlugins(s.getPluginPaths(), names)
}

func (s *lotusService) PutPlugin(name string, binary []byte, isOverwrite bool) error {
	return s.metadata.putPlugin(name, binary, isOverwrite)
}

func (s *lotusService) RemovePlugin(names ...string) error {
	return s.metadata.removePluginConfig(names...)
}
