package service

import "github.com/shima-park/lotus/pkg/rpc/proto"

type LotusService interface {
	Start() error
	Stop()

	GenerateExecutorConfig(*proto.GenerateExecutorConfigRequest) (string, error)
	PutExecutorConfig(edited []byte, isOverwrite bool) error
	RemoveExecutor(names ...string) error
	ListExecutors(names ...string) ([]proto.ExecutorView, error)
	GetExecutorConfig(name string) ([]byte, error)

	RunExecutorProcess(path string) error
	RegisterExecutorProcess(name, id, slaveAddr string) error
	ListExecutorProcesses() ([]proto.ExecutorProcessView, error)
	RemoveExecutorProcess(names ...string) error

	ListComponents(names ...string) ([]proto.ComponentView, error)

	ListProcessors(names ...string) ([]proto.ProcessorView, error)

	ListPlugins(names ...string) ([]proto.PluginView, error)
	PutPlugin(name string, binary []byte, isOverwrite bool) error
	RemovePlugin(names ...string) error
}

type ExecutorProcessManager interface {
	Register(name, id string, slaveAddr string) error
	Deregister(name, id string) error
	List() ([]proto.ExecutorProcessView, error)
	Remove(names ...string) error
	Run(kind, path string) error
	Stop()
}
