package service

import (
	"errors"

	"gopkg.in/yaml.v2"
)

func SniffName(raw []byte) (string, error) {
	var sniff struct {
		Name string `yaml:"name"`
	}

	err := yaml.Unmarshal(raw, &sniff)
	if err != nil {
		return "", err
	}

	if sniff.Name == "" {
		return "", errors.New("No name field or the name is empty")
	}

	return sniff.Name, nil
}
