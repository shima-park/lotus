package kafka

import (
	"reflect"

	"github.com/shima-park/lotus/pkg/core/common/log"
	"github.com/shima-park/seed/component"
	"gopkg.in/yaml.v2"

	"github.com/Shopify/sarama"
	"github.com/shima-park/lotus/pkg/core/common/inject"
)

var (
	producerFactory       component.Factory   = NewProducerFactory()
	_                     component.Component = &Producer{}
	defaultProducerConfig                     = ProducerConfig{
		Name:  "kafka_producer",
		Addrs: []string{"localhost:9092"},
	}
	producerDescription = "kafka producer factory"
)

func init() {
	if err := component.Register("kafka_producer", producerFactory); err != nil {
		panic(err)
	}
}

func NewProducerFactory() component.Factory {
	return component.NewFactory(
		defaultProducerConfig,
		producerDescription,
		inject.InterfaceOf((*sarama.SyncProducer)(nil)),
		func(c string) (component.Component, error) {
			return NewProducer(c)
		})
}

type ProducerConfig struct {
	Name  string   `yaml:"name"`
	Addrs []string `yaml:"addrs"`
}

func (c ProducerConfig) Marshal() ([]byte, error) {
	return yaml.Marshal(c)
}

type Producer struct {
	config   ProducerConfig
	producer sarama.SyncProducer
	instance component.Instance
}

func NewProducer(rawConfig string) (*Producer, error) {
	conf := defaultProducerConfig
	err := yaml.Unmarshal([]byte(rawConfig), &conf)
	if err != nil {
		return nil, err
	}

	log.Info("Kafka producer config: %+v", conf)

	producer, err := sarama.NewSyncProducer(conf.Addrs, nil)
	if err != nil {
		return nil, err
	}

	return &Producer{
		config:   conf,
		producer: producer,
		instance: component.NewInstance(
			conf.Name,
			inject.InterfaceOf((*sarama.SyncProducer)(nil)),
			reflect.ValueOf(producer),
			producer,
		),
	}, nil
}

func (c *Producer) Instance() component.Instance {
	return c.instance
}

func (c *Producer) Start() error {
	return nil
}

func (c *Producer) Stop() error {
	return c.producer.Close()
}
