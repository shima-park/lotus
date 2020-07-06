package service

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"

	"github.com/docker/docker/pkg/reexec"
	"github.com/gin-gonic/gin"
	"github.com/shima-park/lotus/pkg/common/monitor"
	"github.com/shima-park/lotus/pkg/executor"
	"github.com/shima-park/lotus/pkg/rpc/proto"

	utilhttp "github.com/shima-park/lotus/pkg/util/http"
)

func init() {
	reexec.Register("executor", startExecutor)
	if reexec.Init() {
		os.Exit(0)
	}
}

func startExecutor() {
	var name = flag.String("name", "", "executor name")
	var configPath = flag.String("config", "", "config path")
	var serverAddr = flag.String("master", "", "master server address")

	flag.Parse()

	body, err := ioutil.ReadFile(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	f, err := executor.GetFactory(*name)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	exec, err := f.New(string(body))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	srv, err := NewExecutorServer(exec)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	_ = serverAddr // TODO

	if err = srv.Start(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func StartExecutorChildProcess(configPath, masterAddr string) (executor.Executor, error) {
	cmd := reexec.Command("executor", "--config", configPath, "--master", masterAddr)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	if len(output) > 0 {
		return nil, errors.New(string(output))
	}
	return nil, nil
}

type ExecutorServer struct {
	engine   *gin.Engine
	listener net.Listener
	exec     executor.Executor
}

func NewExecutorServer(exec executor.Executor) (*ExecutorServer, error) {
	p := &ExecutorServer{
		engine: gin.Default(),
		exec:   exec,
	}

	var err error
	p.listener, err = net.Listen("tcp", ":2000")
	if err != nil {
		return nil, err
	}

	p.setRouter()

	return p, nil
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, proto.Result{
		Data: data,
	})
}

func Failed(c *gin.Context, err error) {
	c.JSON(http.StatusOK, proto.Result{
		Code: http.StatusInternalServerError,
		Msg:  err.Error(),
	})
}

func (e *ExecutorServer) setRouter() {
	r := e.engine
	r.GET("/name", func(c *gin.Context) {
		Success(c, e.exec.Name())
	})
	r.POST("/start", func(c *gin.Context) {
		err := e.exec.Start()
		if err != nil {
			Failed(c, err)
			return
		}
		Success(c, e.exec.Name())
	})
	r.POST("/stop", func(c *gin.Context) {
		e.exec.Stop()
		Success(c, nil)
	})
	r.GET("/state", func(c *gin.Context) {
		Success(c, e.exec.State())
	})
	r.GET("/components", func(c *gin.Context) {
		Success(c, e.exec.ListComponents())
	})
	r.GET("/processors", func(c *gin.Context) {
		Success(c, e.exec.ListProcessors())
	})
	r.GET("/metrics", func(c *gin.Context) {
		// TODO
	})
	r.GET("/config", func(c *gin.Context) {
		Success(c, e.exec.Config())
	})
	r.GET("/visualize", func(c *gin.Context) {
		// TODO
	})
	r.GET("/check", func(c *gin.Context) {
		// TODO
	})
	r.GET("/error", func(c *gin.Context) {
		Success(c, e.exec.Error())
	})
}

func (p *ExecutorServer) Start() error {
	return http.Serve(p.listener, p.engine)
}

func (p *ExecutorServer) Addr() string {
	return p.listener.Addr().String()
}

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

func (c *ExecutorClient) Name() string {
	var name string
	utilhttp.GetJSON(c.api("/name"), &name)
	return name
}

func (c *ExecutorClient) Start() error {
	return utilhttp.PostJSON(c.api("/start"), nil, nil)
}

func (c *ExecutorClient) Stop() {
	utilhttp.PostJSON(c.api("/stop"), nil, nil)
}

func (c *ExecutorClient) State() executor.State {
	var st executor.State
	utilhttp.GetJSON(c.api("/stop"), &st)
	return st
}

func (c *ExecutorClient) ListComponents() []executor.Component {
	var res []executor.Component
	utilhttp.GetJSON(c.api("/stop"), &res)
	return res
}
func (c *ExecutorClient) ListProcessors() []executor.Processor {
	return nil
}
func (c *ExecutorClient) Monitor() monitor.Monitor {
	return nil
}
func (c *ExecutorClient) Config() string {
	return ""
}
func (c *ExecutorClient) Visualize(w io.Writer, format string) error {
	return nil
}
func (c *ExecutorClient) CheckDependence() []error {
	return nil
}
func (c *ExecutorClient) Error() error {
	return nil
}
