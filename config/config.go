package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/google/go-github/v50/github"
	"github.com/k1LoW/go-github-client/v50/factory"
)

type Config struct {
	Requests []*Request `yaml:"requests"`
}

type Request struct {
	Condition string                 `yaml:"condition"`
	Transform map[string]interface{} `yaml:"transform"`
}

func Load(p string) (*Config, error) {
	ctx := context.Background()
	var (
		b   []byte
		err error
	)
	switch {
	case strings.HasPrefix(p, "github://"):
		c, err := factory.NewGithubClient()
		if err != nil {
			return nil, err
		}
		splitted := strings.SplitN(strings.TrimPrefix(p, "github://"), "/", 3)
		if len(splitted) != 3 {
			return nil, fmt.Errorf("invalid config url: %s", p)
		}
		f, _, _, err := c.Repositories.GetContents(ctx, splitted[0], splitted[1], splitted[2], &github.RepositoryContentGetOptions{})
		if err != nil {
			return nil, err
		}
		if f == nil {
			return nil, fmt.Errorf("invalid config url: %s", p)
		}
		cc, err := f.GetContent()
		if err != nil {
			return nil, err
		}
		b = []byte(cc)
	default:
		b, err = os.ReadFile(p)
		if err != nil {
			return nil, err
		}
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return nil, err
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (cfg *Config) validate() error {
	if len(cfg.Requests) == 0 {
		return errors.New("no settings")
	}
	return nil
}
