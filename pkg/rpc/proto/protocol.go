package proto

import "time"

type API interface {
	GenerateExecutorConfig(*GenerateExecutorConfigRequest) ([]byte, error)
	AddExecutor(*AddExecutorRequest) error
	PutExecutor(*PutExecutorRequest) error
	GetExecutor(*GetExecutorRequest) (*GetExecutorResponse, error)
	RemoveExecutor(*RemoveExecutorRequest) error
	RegisterExecutorProcess(*RegisterExecutorProcessRequest) error
	RemoveExecutorProcess(*RemoveExecutorProcessRequest) error
	ListExecutor() ([]ExecutorView, error)
	ListExecutorProcess() ([]ExecutorProcessView, error)
	ListComponent() ([]ComponentView, error)
	ListProcessor() ([]ProcessorView, error)
	ListPlugin() ([]PluginView, error)
	AddPlugin(*AddPluginRequest) error
	PutPlugin(*PutPluginRequest) error
	RemovePlugin(*RemovePluginRequest) error
}

type GenerateExecutorConfigRequest struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Components []string `json:"components"`
	Processors []string `json:"processors"`
}

type AddExecutorRequest struct {
	Config []byte `json:"config"`
}

type PutExecutorRequest struct {
	Config []byte `json:"config"`
}

type GetExecutorRequest struct {
	Name string `json:"name"`
}

type GetExecutorResponse struct {
	Config []byte `json:"config"`
}

type RemoveExecutorRequest struct {
	Names []string `json:"names"`
}

type RegisterExecutorProcessRequest struct {
	Name      string `json:"name"`
	ID        string `json:"id"`
	ServeAddr string `json:"serve_addr"`
}

type RemoveExecutorProcessRequest struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type AddPluginRequest struct {
	Name              string `json:"name"`
	ShareObjectBinary []byte `json:"share_object_binary"`
}

type PutPluginRequest struct {
	Name              string `json:"name"`
	ShareObjectBinary []byte `json:"share_object_binary"`
}

type RemovePluginRequest struct {
	Names []string `json:"names"`
}

type VisualizeFormat string

const (
	VisualizeFormatSVG  VisualizeFormat = "svg"
	VisualizeFormatPng  VisualizeFormat = "png"
	VisualizeFormatDot  VisualizeFormat = "dot"
	VisualizeFormatTerm VisualizeFormat = "term"
)

type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type ComponentView struct {
	Name         string `json:"name"`
	SampleConfig string `json:"sample_config"`
	Description  string `json:"description"`
	InjectName   string `json:"inject_name,omitempty"`
	ReflectType  string `json:"reflect_type,omitempty"`
	ReflectValue string `json:"reflect_value,omitempty"`
}

type ProcessorView struct {
	Name         string `json:"name"`
	SampleConfig string `json:"sample_config"`
	Description  string `json:"description"`
}

type PluginView struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Module   string    `json:"module"`
	OpenTime time.Time `json:"open_time"`
	Error    string    `json:"error"`
}

type ExecutorProcessView struct {
	Name      string `json:"name"`
	ID        string `json:"id"`
	ServeAddr string `json:"serve_addr"`
}

type ExecutorView struct {
	Name string `json:"name"`
}
