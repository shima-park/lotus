package service

import (
	"errors"

	"gopkg.in/yaml.v2"
)

func SniffNameAndKind(raw []byte) (string, string, error) {
	var sniff struct {
		Name string `yaml:"name"`
		Kind string `yaml:"kind"`
	}

	err := yaml.Unmarshal(raw, &sniff)
	if err != nil {
		return "", "", err
	}

	if sniff.Name == "" {
		return "", "", errors.New("No name field or the name is empty")
	}

	if sniff.Kind == "" {
		return "", "", errors.New("No kind field or the kind is empty")
	}

	return sniff.Name, sniff.Kind, nil
}
