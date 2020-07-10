package executor

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/pkg/reexec"
	"github.com/google/uuid"
	"github.com/natefinch/lumberjack"
	cp "github.com/shima-park/lotus/pkg/rpc/service/child_process"
	"github.com/shima-park/seed/executor"
	"github.com/shima-park/seed/plugin"
)

const (
	CHILD_PROCESS_EXECUTOR = "executor"
)

func init() {
	reexec.Register(CHILD_PROCESS_EXECUTOR, startExecutor)
	if reexec.Init() {
		os.Exit(0)
	}
}

func startExecutor() {
	var kind = flag.String("kind", "", "kind of executor")
	var configPath = flag.String("config-path", "", "config path")
	var pluginPaths = flag.String("plugins", "", "plugin paths")
	var serverAddr = flag.String("master", "", "master server address")

	flag.Parse()

	f, err := executor.GetFactory(*kind)
	cp.Failed(err)
	fmt.Println("=======plugins:", *pluginPaths)
	// TODO fix me do not need to open all plugin
	for _, path := range strings.Split(*pluginPaths, ",") {
		if path == "" {
			continue
		}
		_, err := plugin.LoadPlugins(path)
		cp.Failed(err)
	}

	body, err := ioutil.ReadFile(*configPath)
	cp.Failed(err)

	exec, err := f.New(string(body))
	cp.Failed(err)

	srv, err := NewExecutorServer(*name, exec)
	cp.Failed(err)

	go func() {
		err = srv.Start()
		cp.Failed(err)
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	err = srv.Register(*serverAddr)
	cp.Failed(err)

	<-signals
}

type ExecutorProcessConfig struct {
	Kind        string
	ConfigPath  string
	PluginPaths []string
	MasterAddr  string
	ExitTimeout time.Duration
	Callback    func(name string, err error)
}

type ExecutorProcess struct {
	Config    ExecutorProcessConfig
	LogPath   string
	Command   *exec.Cmd
	errOut    bytes.Buffer
	Error     error
	ServeAddr string

	done chan struct{}
	wg   sync.WaitGroup
}

func NewExecutorProcess(conf ExecutorProcessConfig) (*ExecutorProcess, error) {
	ep := &ExecutorProcess{
		Config:  conf,
		LogPath: fmt.Sprintf("logs/%s.log", conf.Name),
		done:    make(chan struct{}),
	}

	if ep.Config.Name == "" {
		ep.Config.Name = fmt.Sprintf("%s-%s", conf.Name, uuid.New().String())
	}

	err := os.MkdirAll(filepath.Dir(ep.LogPath), 0777)
	if err != nil {
		return nil, err
	}

	if conf.ExitTimeout < time.Duration(0) {
		conf.ExitTimeout = time.Second * 30
	}

	ep.Command = reexec.Command(
		CHILD_PROCESS_EXECUTOR,
		"--kind", conf.Kind,
		"--config-path", conf.ConfigPath,
		"--plugins", strings.Join(conf.PluginPaths, ","),
		"--master", conf.MasterAddr,
	)

	output := &lumberjack.Logger{
		Filename:   ep.LogPath,
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	}

	ep.Command.Stderr = &ep.errOut
	ep.Command.Stdout = output

	return ep, nil
}

func (p *ExecutorProcess) Run() error {
	if err := p.Command.Start(); err != nil {
		return err
	}

	waitDone := make(chan struct{})
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()

		err := p.Command.Wait()
		close(waitDone)
		if err != nil {
			p.Error = fmt.Errorf("%s %s", err, p.errOut.String())
		} else {
			if p.errOut.Len() > 0 {
				p.Error = errors.New(p.errOut.String())
			}
		}
		if p.Config.Callback != nil {
			p.Config.Callback(p.Config.Name, p.Error)
		}
	}()

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		select {
		case <-p.done:
			p.Command.Process.Signal(os.Interrupt)
			select {
			case <-waitDone:
				return
			case <-time.After(p.Config.ExitTimeout):
				p.Command.Process.Kill()
			}
		case <-waitDone:
			return
		}
	}()

	return nil
}

func (p *ExecutorProcess) Stop() {
	close(p.done)
	p.wg.Wait()
}
