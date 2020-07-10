package executor

import (
	"io"

	"github.com/shima-park/lotus/pkg/core/common/monitor"
	utilhttp "github.com/shima-park/lotus/pkg/util/http"
	"github.com/shima-park/seed/executor"
)

type ExecutorClient struct {
	addr string
}

func NewExecutorClient(addr string) *ExecutorClient {
	return &ExecutorClient{
		addr: utilhttp.NormalizeURL(addr),
	}
}

func (c *ExecutorClient) api(path string) string {
	return c.addr + path
}

func (c *ExecutorClient) Start() error {
	return utilhttp.PostJSON(c.api("/start"), nil, nil)
}

func (c *ExecutorClient) Stop() {
	utilhttp.PostJSON(c.api("/stop"), nil, nil)
}

func (c *ExecutorClient) State() executor.State {
	var st executor.State
	utilhttp.GetJSON(c.api("/state"), &st)
	return st
}

func (c *ExecutorClient) Monitor() monitor.Monitor {
	return nil
}

func (c *ExecutorClient) Visualize(w io.Writer, format string) error {
	return nil
}
