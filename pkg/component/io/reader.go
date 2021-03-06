package io

import (
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/shima-park/lotus/pkg/component"
	"gopkg.in/yaml.v2"

	"github.com/pkg/errors"
	"github.com/shima-park/lotus/pkg/common/inject"
)

var (
	readerFactory       component.Factory   = NewReaderFactory()
	_                   component.Component = &Reader{}
	defaultReaderConfig                     = ReaderConfig{
		Name: "io_reader",
		Path: "stdin",
	}
	readerDescription = "file reader e.g.: stdin, stdout, stderr, /var/log/xxx.log"
)

func init() {
	if err := component.Register("io_reader", readerFactory); err != nil {
		panic(err)
	}
}

func NewReaderFactory() component.Factory {
	return component.NewFactory(
		defaultReaderConfig,
		readerDescription,
		inject.InterfaceOf((*io.Reader)(nil)),
		func(c string) (component.Component, error) {
			return NewReader(c)
		})
}

type ReaderConfig struct {
	Name string
	Path string
}

func (c ReaderConfig) Marshal() ([]byte, error) {
	return yaml.Marshal(c)
}

type Reader struct {
	path     string
	rc       io.ReadCloser
	instance component.Instance
}

func NewReader(rawConfig string) (*Reader, error) {
	conf := defaultReaderConfig
	err := yaml.Unmarshal([]byte(rawConfig), &conf)
	if err != nil {
		return nil, err
	}

	var f io.ReadCloser
	switch strings.TrimSpace(conf.Path) {
	case "stdin":
		f = os.Stdin
	case "stdout":
		f = os.Stdout
	case "stderr":
		f = os.Stderr
	default:
		file, err := os.Open(conf.Path)
		if err != nil {
			return nil, errors.Wrap(err, "io_reader")
		}
		f = file
	}

	return &Reader{
		path: conf.Path,
		rc:   f,
		instance: component.NewInstance(
			conf.Name,
			inject.InterfaceOf((*io.Reader)(nil)),
			reflect.ValueOf(f),
			f,
		),
	}, nil
}

func (r *Reader) Instance() component.Instance {
	return r.instance
}

func (r *Reader) Start() error {
	return nil
}

func (r *Reader) Stop() error {
	switch strings.TrimSpace(r.path) {
	case "stdin", "stdout", "stderr":
		return nil
	default:
		return r.rc.Close()
	}
}
