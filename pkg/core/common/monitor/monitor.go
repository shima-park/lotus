package monitor

import (
	"expvar"
	"sync"
	"time"
)

type Var = expvar.Var

type KeyValue = expvar.KeyValue

type Elapsed time.Duration

func (e Elapsed) String() string {
	d := time.Duration(e)
	d = d.Truncate(time.Second)
	return d.String()
}

type Time time.Time

func (t Time) String() string {
	return time.Time(t).Format("2006-01-02 15:04:05")
}

type String string

func (s String) String() string {
	return string(s)
}

type Monitor interface {
	With(namespace string) Monitor
	Add(key string, delta int64)
	AddFloat(key string, delta float64)
	Delete(key string)
	Do(f func(root, namespace string, kv KeyValue))
	Get(key string) Var
	Set(key string, av Var)
	String() string
}

type monitor struct {
	root      string
	namespace string
	vars      *expvar.Map
	lock      *sync.RWMutex
	childs    map[string]*monitor
}

func NewMonitor(namespace string) Monitor {
	return newMonitor(namespace, namespace)
}

func newMonitor(root, namespace string) *monitor {
	return &monitor{
		root:      namespace,
		namespace: namespace,
		lock:      &sync.RWMutex{},
		vars:      new(expvar.Map).Init(),
		childs:    map[string]*monitor{},
	}
}

func (m *monitor) With(namespace string) Monitor {
	var nm *monitor
	m.lock.Lock()
	var ok bool
	if nm, ok = m.childs[namespace]; !ok {
		nm = newMonitor(m.root, namespace)
		m.childs[namespace] = nm
	}
	m.lock.Unlock()

	return nm
}

func (m *monitor) Add(key string, delta int64) {
	m.vars.Add(key, delta)
}

func (m *monitor) AddFloat(key string, delta float64) {
	m.vars.AddFloat(key, delta)
}

func (m *monitor) Delete(key string) {
	m.vars.Delete(key)
}

func (m *monitor) Do(f func(root, namespace string, kv KeyValue)) {
	m.vars.Do(func(kv KeyValue) {
		f(m.root, m.namespace, kv)
	})

	m.lock.RLock()
	for _, c := range m.childs {
		c.Do(f)
	}
	m.lock.RUnlock()
}

func (m *monitor) Get(key string) Var {
	if v := m.vars.Get(key); v != nil {
		return v
	}

	m.lock.RLock()
	for _, c := range m.childs {
		if v := c.Get(key); v != nil {
			return v
		}
	}
	m.lock.RUnlock()

	return String("")
}

func (m *monitor) Set(key string, av Var) {
	m.vars.Set(key, av)
}

func (m *monitor) String() string {
	return m.vars.String()
}
