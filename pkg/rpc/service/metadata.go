package service

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"

	"github.com/shima-park/lotus/pkg/core/common/log"
	"gopkg.in/yaml.v2"
)

const (
	METADATA_PATH     = "meta"
	METADATA_FILENAME = "meta.yaml"
)

type metadata struct {
	metapath string
	metafile string

	lock     sync.RWMutex
	registry *registry
}

type registry struct {
	PluginPaths         map[string]string `yaml:"plugin_configs"`
	ExecutorConfigPaths map[string]string `yaml:"executor_configs"` // key: executor type
}

func newMetadata(metapath string) (*metadata, error) {
	if metapath == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		metapath = filepath.Join(pwd, METADATA_PATH)
	}

	m := &metadata{
		metapath: metapath,
		metafile: filepath.Join(metapath, METADATA_FILENAME),
		registry: &registry{},
	}

	err := os.MkdirAll(metapath, 0750)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create data path %s", metapath)
	}

	_, err = os.Stat(m.metafile)
	if os.IsNotExist(err) {
		log.Info("No metadata file found under: %s. Creating a new metadata file.", m.metafile)
		if err = m.save(); err != nil {
			return nil, err
		}
	} else {
		if err != nil {
			return nil, err
		}

		data, err := ioutil.ReadFile(m.metafile)
		if err != nil {
			return nil, err
		}

		if err = yaml.Unmarshal(data, m.registry); err != nil {
			return nil, err
		}
	}

	return m, nil
}

func (m *metadata) save() error {
	data, err := yaml.Marshal(m.registry)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(m.metafile, data, 0644)
}

func (m *metadata) read(do func(r registry) error) error {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return do(*m.registry)
}

func (m *metadata) write(do func(r *registry) error) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	err := do(m.registry)
	if err != nil {
		return err
	}

	return m.save()
}

//
//func (m *metadata) addPath(path string, pattern string, target *[]string) error {
//	path = strings.TrimSpace(path)
//	if path == "" {
//		return nil
//	}
//
//	fi, err := os.Stat(path)
//	if err != nil {
//		return err
//	}
//
//	var paths []string
//	if fi.IsDir() {
//		var err error
//		paths, err = filepath.Glob(filepath.Join(path, pattern))
//		if err != nil {
//			return err
//		}
//
//		if len(paths) == 0 {
//			return errors.New("not match any " + pattern)
//		}
//	} else {
//		paths = []string{path}
//	}
//
//	for _, p := range paths {
//		if stringInSlice(p, *target) {
//			continue
//		}
//
//		*target = append(*target, p)
//	}
//
//	return m.save()
//}
//

//
//func stringInSlice(t string, ss []string) bool {
//	for _, s := range ss {
//		if t == s {
//			return true
//		}
//	}
//	return false
//}
//
//func (m *metadata) GetPluginConfigList(names ...string) ([]proto.PluginConfig, error) {
//	var list []proto.PluginConfig
//	m.read(func(r *registry) error {
//		for _, conf := range r.PluginConfigs {
//			if len(names) > 0 && !stringInSlice(conf.Name, names) {
//				continue
//			}
//			list = append(list, conf)
//		}
//		return nil
//	})
//	return list, nil
//}
//
//func (m *metadata) GetExecutorConfigList(names ...string) ([]proto.ExecutorConfig, error) {
//	var list []proto.ExecutorConfig
//	m.read(func(r *registry) error {
//		for _, conf := range r.ExecutorConfigs {
//			if len(names) > 0 && !stringInSlice(conf.Name, names) {
//				continue
//			}
//			list = append(list, conf)
//		}
//		return nil
//	})
//	return list, nil
//}
//
func (m *metadata) putExecutorConfig(name string, raw []byte, isOverwrite bool) (err error) {
	err = m.write(func(r *registry) error {
		if !isOverwrite {
			_, ok := r.ExecutorConfigPaths[name]
			if ok {
				return fmt.Errorf("%s executor config is exists", name)
			}
		}

		path, err := m.writeFile("executors", name, ".yaml", raw)
		if err != nil {
			return err
		}

		r.ExecutorConfigPaths[name] = path
		return nil
	})

	return
}

func (m *metadata) loadExecutorConfig(name string) ([]byte, error) {
	var data []byte
	err := m.read(func(r registry) error {
		path, ok := r.ExecutorConfigPaths[name]
		if !ok {
			return errors.New("Not found executor config " + name)
		}

		var err error
		data, err = ioutil.ReadFile(path)
		return err
	})

	return data, err
}

func (m *metadata) writeFile(subPath string, name, defExt string, data []byte) (string, error) {
	filename := name
	if filepath.Ext(name) == "" {
		filename += defExt
	}

	path := filepath.Join(m.metapath, subPath, filename)

	err := os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(path, data, 0644)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (m *metadata) putPlugin(name string, bin []byte, isOverwrite bool) (err error) {
	err = m.write(func(r *registry) error {
		if !isOverwrite {
			_, ok := r.PluginPaths[name]
			if ok {
				return fmt.Errorf("%s plugin path is exists", name)
			}
		}

		path, err := m.writeFile("plugins", name, ".so", bin)
		if err != nil {
			return err
		}

		r.PluginPaths[name] = path
		return nil
	})

	return
}

func (m *metadata) removeExecutorConfig(names ...string) error {
	err := m.write(func(r *registry) error {
		for _, name := range names {
			_, ok := r.ExecutorConfigPaths[name]
			if ok {
				delete(r.ExecutorConfigPaths, name)
			}
		}

		return nil
	})
	return err
}

func (m *metadata) removePluginConfig(names ...string) error {
	err := m.write(func(r *registry) error {
		for _, name := range names {
			_, ok := r.PluginPaths[name]
			if ok {
				delete(r.PluginPaths, name)
			}
		}

		return nil
	})
	return err
}
