package executor

import (
	"fmt"

	"github.com/shima-park/lotus/pkg/component"
	"github.com/shima-park/lotus/pkg/processor"
)

type Executor interface {
	Name() string
	Config() string
	Start() error
	Stop()
	State() State
	ListComponents() []Component
	ListProcessors() []Processor
	Error() error
}

var (
	Idle    State = 0
	Running State = 1
	Exited  State = 3
)

type State int32

func (s State) String() string {
	switch s {
	case Idle:
		return "idle"
	case Running:
		return "running"
	case Exited:
		return "exited"
	}
	return fmt.Sprintf("unknown(%d)", s)
}

type Component struct {
	Name      string
	RawConfig string
	Component component.Component
	Factory   component.Factory
}

type Processor struct {
	Name      string
	RawConfig string
	Processor processor.Processor
	Factory   processor.Factory
}
