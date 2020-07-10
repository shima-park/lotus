package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/shima-park/lotus/pkg/rpc/proto"
	cp "github.com/shima-park/lotus/pkg/rpc/service/child_process/executor"
)

type executorProcessManager struct {
	masterAddr string

	getPluginPathsFunc GetPluginPathsFunc

	lock sync.RWMutex
	pgs  map[string]*processGroup
}

type processGroup struct {
	kind       string
	configPath string
	instances  []*process
}

type process struct {
	process      *cp.ExecutorProcess
	id           string
	serveAddr    string
	createTime   time.Time
	registerTime time.Time
}

func NewExecutorProcessManager(masterAddr string,
	getPluginPathsFunc GetPluginPathsFunc) ExecutorProcessManager {

	return &executorProcessManager{
		masterAddr:         masterAddr,
		getPluginPathsFunc: getPluginPathsFunc,
		pgs:                map[string]*processGroup{},
	}
}

type GetPluginPathsFunc func() []string

func (m *executorProcessManager) Run(kind, path string) error {
	conf := cp.ExecutorProcessConfig{
		Kind:        kind,
		ConfigPath:  path,
		ExitTimeout: time.Second * 30,
		PluginPaths: m.getPluginPathsFunc(),
		MasterAddr:  m.masterAddr,
		Callback: func(name string, err error) {
			if err != nil {
				fmt.Println("========receive error:", err, "delete executor", name)
			}
			m.lock.Lock()
			delete(m.pgs, name)
			m.lock.Unlock()
		},
	}

	ep, err := cp.NewExecutorProcess(conf)
	if err != nil {
		return err
	}

	err = ep.Run()
	if err != nil {
		return err
	}

	m.lock.Lock()
	_, ok := m.pgs[conf.Name]
	if ok {
		m.lock.Unlock()
		return fmt.Errorf("Executor process %s is exists", conf.Name)
	}
	m.pgs[conf.Name] = &processGroup{
		kind:       kind,
		configPath: path,
	}
	m.lock.Unlock()

	return nil
}

func (m *executorProcessManager) Register(name string, id, slaveAddr string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	pg, ok := m.pgs[name]
	if !ok {
		return fmt.Errorf("Not found executor name: %s", name)
	}

	pg.instances = append(pg.instances, &process{
		id:           id,
		serveAddr:    slaveAddr,
		registerTime: time.Now(),
	})
	return nil
}

func (m *executorProcessManager) Deregister(name, id string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	pg, ok := m.pgs[name]
	if !ok {
		return fmt.Errorf("Not found executor name: %s", name)
	}

	for i, p := range pg.instances {
		if p.id == id {
			pg.instances = append(pg.instances[:i], pg.instances[i+1:]...)
		}
	}

	return nil
}

func (m *executorProcessManager) List() ([]proto.ExecutorProcessView, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	var res []proto.ExecutorProcessView
	for _, pg := range m.pgs {
		for _, i := range pg.instances {
			res = append(res, proto.ExecutorProcessView{
				Name:      pg.kind,
				ID:        i.id,
				ServeAddr: i.serveAddr,
			})
		}
	}
	return res, nil
}

func (m *executorProcessManager) Remove(names ...string) error {
	return nil
}

func (m *executorProcessManager) Stop() {
	m.lock.Lock()
	defer m.lock.Unlock()

	for _, pg := range m.pgs {
		for _, i := range pg.instances {
			fmt.Println("stopping", i.id, i.serveAddr)
			i.process.Stop()
			fmt.Println("stopped", i.id, i.serveAddr)
		}
	}
}
