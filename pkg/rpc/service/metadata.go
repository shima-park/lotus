package service

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/shima-park/lotus/pkg/common/log"
	"github.com/shima-park/lotus/pkg/rpc/proto"
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
	PluginPaths         []string             `yaml:"plugin_paths"`
	ExecutorConfigPaths map[string]*[]string `yaml:"executor_config_paths"`
}

func NewMetadata(metapath string) (proto.Metadata, error) {
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
		registry: &registry{
			ExecutorConfigPaths: map[string]*[]string{},
		},
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

func (m *metadata) PutPlugin(name string, bin []byte) (path string, err error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	filename := m.GetPath(proto.FileTypePlugin, name)

	err = os.MkdirAll(filepath.Dir(filename), 0777)
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(filename, bin, 0666)
	if err != nil {
		return "", err
	}

	err = m.addPluginPath(filename)
	if err != nil {
		return "", err
	}

	return filename, m.save()
}

func (m *metadata) PutExecutorRawConfig(_type, name string, raw []byte) (path string, err error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	filename := m.GetPath(proto.FileTypeExecutorConfig, name)

	err = os.MkdirAll(filepath.Dir(name), 0777)
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(filename, raw, 0666)
	if err != nil {
		return "", err
	}

	err = m.addExecutorConfigPath(_type, filename)
	if err != nil {
		return "", err
	}

	return filename, m.save()
}

func (m *metadata) RemoveExecutorConfigPath(_type, path string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	paths := *m.registry.ExecutorConfigPaths[_type]
	for i, s := range paths {
		if s == path {
			paths = append(paths[:i], paths[i+1:]...)
			break
		}
	}
	return nil
}

func (m *metadata) AddPluginPath(path string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.addPluginPath(path)
}

func (m *metadata) addPluginPath(path string) error {
	return m.addPath(path, "*.so", &m.registry.PluginPaths)
}

func (m *metadata) AddExecutorConfigPath(_type, path string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.addExecutorConfigPath(_type, path)
}

func (m *metadata) addExecutorConfigPath(_type, path string) error {
	paths, ok := m.registry.ExecutorConfigPaths[_type]
	if !ok {
		m.registry.ExecutorConfigPaths[_type] = &[]string{path}
	}

	return m.addPath(path, "*.yaml", paths)
}

func (m *metadata) GetPath(ft proto.FileType, filename string) string {
	switch ft {
	case proto.FileTypePlugin:
		if filepath.Ext(filename) == "" {
			filename += ".so"
		}
		return filepath.Join(m.metapath, string(ft), filename)
	case proto.FileTypeExecutorConfig:
		if filepath.Ext(filename) == "" {
			filename += ".yaml"
		}
		return filepath.Join(m.metapath, string(ft), filename)
	default:
		panic(fmt.Sprintf("Unknown file type: %s", ft))
	}
}

func (m *metadata) RemovePluginPath(path string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	for i, s := range m.registry.PluginPaths {
		if s == path {
			m.registry.PluginPaths = append(m.registry.PluginPaths[:i], m.registry.PluginPaths[i+1:]...)
			break
		}
	}

	_, err := os.Stat(path)
	if !os.IsNotExist(err) {
		_ = os.Remove(path)
	}

	return m.save()
}

func (m *metadata) addPath(path string, pattern string, target *[]string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}

	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	var paths []string
	if fi.IsDir() {
		var err error
		paths, err = filepath.Glob(filepath.Join(path, pattern))
		if err != nil {
			return err
		}

		if len(paths) == 0 {
			return errors.New("not match any " + pattern)
		}
	} else {
		paths = []string{path}
	}

	for _, p := range paths {
		if stringInSlice(p, m.registry.PluginPaths) {
			continue
		}

		*target = append(*target, p)
	}

	return m.save()
}

func (m *metadata) Overwrite(ft proto.FileType, path string, data []byte) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	return ioutil.WriteFile(path, data, 0644)
}

func (m *metadata) Snapshot(do func(proto.Snapshot)) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	var s proto.Snapshot
	for _, pp := range m.registry.PluginPaths {
		s.PluginPaths = append(s.PluginPaths, pp)
	}

	for name, paths := range m.registry.ExecutorConfigPaths {
		for _, path := range *paths {
			s.ExecutorConfigPaths[name] = append(s.ExecutorConfigPaths[name], path)
		}
	}

	do(s)
}

func stringInSlice(t string, ss []string) bool {
	for _, s := range ss {
		if t == s {
			return true
		}
	}
	return false
}
