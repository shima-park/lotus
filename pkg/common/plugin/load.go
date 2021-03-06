package plugin

import (
	"errors"
	"fmt"
	goplugin "plugin"
	"sync"
	"time"
)

var (
	rwlock        sync.RWMutex
	loadedPlugins []Plugin
)

type Plugin struct {
	Path     string
	Module   string
	Name     string
	OpenTime time.Time
}

func List() []Plugin {
	rwlock.RLock()
	defer rwlock.RUnlock()

	var snapshots []Plugin
	snapshots = append(snapshots, loadedPlugins...)
	return snapshots
}

func LoadPlugins(path string) error {
	return loadPlugins(path)
}

func loadPlugins(path string) error {
	rwlock.Lock()
	defer rwlock.Unlock()

	fmt.Printf("loading plugin bundle: %v\n", path)

	p, err := goplugin.Open(path)
	if err != nil {
		return err
	}

	sym, err := p.Lookup("Bundle")
	if err != nil {
		return err
	}

	ptr, ok := sym.(*map[string][]interface{})
	if !ok {
		return errors.New("invalid bundle type")
	}

	bundle := *ptr
	for name, plugins := range bundle {
		loader := registry[name]
		if loader == nil {
			continue
		}

		for _, plugin := range plugins {
			pluginName, err := loader(plugin)
			if err != nil {
				return err
			}

			loadedPlugins = append(loadedPlugins, Plugin{
				Path:     path,
				Module:   name,
				Name:     pluginName,
				OpenTime: time.Now(),
			})
		}

	}

	return nil
}
