package config

import (
	"os"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Requests []*Request `yaml:"requests"`
}

type Request struct {
	Condition string                 `yaml:"condition"`
	Transform map[string]interface{} `yaml:"transform"`
}

func Load(p string) (*Config, error) {
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
