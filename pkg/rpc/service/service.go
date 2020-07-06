package service

import (
	"io/ioutil"
	"os"

	"github.com/shima-park/lotus/pkg/common/log"
	"github.com/shima-park/lotus/pkg/rpc/proto"
)

type Service struct {
	metadata proto.Metadata
	proto.Executor
	proto.Component
	proto.Processor
	proto.Plugin
	proto.Server
}

func NewService(metadata proto.Metadata) *Service {
	s := &Service{
		metadata:  metadata,
		Executor:  NewExecutorService(metadata),
		Component: NewComponentService(),
		Processor: NewProcessorService(),
		Plugin:    NewPluginService(metadata),
	}

	return s
}

func (s *Service) Start() error {
	var errs ErrorGroup
	s.metadata.Snapshot(func(snapshot proto.Snapshot) {
		for _, path := range snapshot.PluginPaths {
			if isNotExists(path) {
				log.Warn("Plugin path(%s) is not exists", path)
				continue
			}
			err := TestPlugin(path)
			if err != nil {
				errs = append(errs, err)
			}
		}

		for _type, paths := range snapshot.ExecutorConfigPaths {
			for _, path := range paths {
				if isNotExists(path) {
					log.Warn("Plugin path(%s) is not exists", path)
					continue
				}

				log.Info("loading executor from: %s", path)
				data, err := ioutil.ReadFile(path)
				if err != nil {
					errs = append(errs, err)
				}

				err = s.Executor.Add(_type, data)
				if err != nil {
					errs = append(errs, err)
				}
			}
		}
	})

	return errs.Error()
}

func isNotExists(path string) bool {
	_, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return true
	}
	return false
}

func (s *Service) Stop() error {
	list, err := s.Executor.List()
	if err != nil {
		return err
	}

	for _, e := range list {
		_ = s.Executor.Control(proto.ControlCommandStop, e.Name)
	}
	return nil
}
