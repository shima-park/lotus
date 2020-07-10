package executor

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/shima-park/lotus/pkg/rpc/proto"
	utilhttp "github.com/shima-park/lotus/pkg/util/http"
	"github.com/shima-park/seed/executor"
)

type ExecutorServer struct {
	engine   *gin.Engine
	listener net.Listener
	name     string
	exec     executor.Executor
}

func NewExecutorServer(name string, exec executor.Executor) (*ExecutorServer, error) {
	p := &ExecutorServer{
		engine: gin.Default(),
		exec:   exec,
		name:   name,
	}

	var err error
	p.listener, err = net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}
	fmt.Println("=========address:", p.listener.Addr().String())
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
	r.POST("/start", func(c *gin.Context) {
		err := e.exec.Start()
		if err != nil {
			Failed(c, err)
			return
		}
		Success(c, nil)
	})
	r.POST("/stop", func(c *gin.Context) {
		e.exec.Stop()
		Success(c, nil)
	})
	r.GET("/state", func(c *gin.Context) {
		Success(c, e.exec.State())
	})
	r.GET("/metrics", func(c *gin.Context) {
		// TODO
	})
	r.GET("/visualize", func(c *gin.Context) {
		// TODO
	})
}

func (p *ExecutorServer) Register(serverAddr string) error {
	req := &proto.RegisterExecutorProcessRequest{
		Name:      p.name,
		ID:        fmt.Sprint(os.Getpid()),
		ServeAddr: p.listener.Addr().String(),
	}

	url := fmt.Sprintf("%s/executor/register", serverAddr)
	fmt.Println("name", p.name, "addr", req.ServeAddr, "url", url)
	return utilhttp.PostJSON(url, req, nil)
}

func (p *ExecutorServer) Start() error {
	return http.Serve(p.listener, p.engine)
}

func (p *ExecutorServer) Addr() string {
	return p.listener.Addr().String()
}
