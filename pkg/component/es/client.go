package es

import (
	"reflect"
	"strings"

	"github.com/shima-park/lotus/pkg/common/log"
	"github.com/shima-park/lotus/pkg/component"
	"gopkg.in/yaml.v2"

	"github.com/olivere/elastic/v7"
)

var (
	factory       component.Factory   = NewFactory()
	_             component.Component = &Client{}
	defaultConfig                     = Config{
		Name: "es_client",
		Addr: "127.0.0.1:9200",
	}
	description = "es client factory"
)

func init() {
	if err := component.Register("es_client", factory); err != nil {
		panic(err)
	}
}

func NewFactory() component.Factory {
	return component.NewFactory(
		defaultConfig,
		description,
		reflect.TypeOf(&elastic.Client{}),
		func(c string) (component.Component, error) {
			return NewClient(c)
		})
}

type Config struct {
	Name string `yaml:"name"`
	Addr string `yaml:"addr"`
}

func (c Config) Marshal() ([]byte, error) {
	return yaml.Marshal(c)
}

type Client struct {
	c        *elastic.Client
	instance component.Instance
}

func NewClient(rawConfig string) (*Client, error) {
	conf := defaultConfig
	err := yaml.Unmarshal([]byte(rawConfig), &conf)
	if err != nil {
		return nil, err
	}

	log.Info("ES config: %+v", conf)

	var options []elastic.ClientOptionFunc
	if conf.Addr != "" {
		url := conf.Addr
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			url = "http://" + url
		}

		options = append(options, elastic.SetURL(url))
	}

	c, err := elastic.NewClient(options...)
	if err != nil {
		return nil, err
	}

	return &Client{
		c: c,
		instance: component.NewInstance(
			conf.Name,
			reflect.TypeOf(c),
			reflect.ValueOf(c),
			c,
		),
	}, nil
}

func (c *Client) Instance() component.Instance {
	return c.instance
}

func (c *Client) Start() error {
	return nil
}

func (c *Client) Stop() error {
	return nil
}
