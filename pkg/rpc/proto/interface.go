package proto

type Executor interface {
	Add(_type string, config []byte) error
	Remove(executorInstanceIDs ...string) error
	Recreate(_type string, config []byte) error
	List() ([]ExecutorView, error)
	Find(executorInstanceID string) (*ExecutorView, error)
	Control(cmd ControlCommand, executorInstanceIDs ...string) error
	Visualize(format VisualizeFormat, executorInstanceID string) ([]byte, error)
}

type Component interface {
	List() ([]ComponentView, error)
	Find(name string) (*ComponentView, error)
}

type Processor interface {
	List() ([]ProcessorView, error)
	Find(name string) (*ProcessorView, error)
}

type Plugin interface {
	Add(name string, bin []byte) error
	AddPath(path string) error
	Remove(names ...string) error
	List() ([]PluginView, error)
}

type Server interface {
	Metadata() (MetadataView, error)
}

type FileType string

const (
	FileTypePlugin         FileType = "plugins"
	FileTypeExecutorConfig FileType = "executors"
)

type Snapshot struct {
	PluginPaths         []string
	ExecutorConfigPaths map[string][]string // key: executor type
}

type Metadata interface {
	PutPlugin(name string, bin []byte) (path string, err error)
	PutExecutorRawConfig(_type, name string, raw []byte) (path string, err error)
	AddPluginPath(path string) error
	AddExecutorConfigPath(_type, path string) error
	RemovePluginPath(path string) error
	RemoveExecutorConfigPath(_type, path string) error
	Snapshot(do func(Snapshot))
}
