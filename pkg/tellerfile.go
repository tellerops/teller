package pkg

import (
	"os"

	"github.com/spectralops/teller/pkg/core"
	"gopkg.in/yaml.v2"
)

type ProvidersMap map[string]MappingConfig
type TellerFile struct {
	Opts       map[string]string `yaml:"opts,omitempty"`
	Confirm    string            `yaml:"confirm,omitempty"`
	Project    string            `yaml:"project,omitempty"`
	CarryEnv   bool              `yaml:"carry_env,omitempty"`
	Providers  ProvidersMap      `yaml:"providers,omitempty"`
	LoadedFrom string
}

type MappingConfig struct {
	Kind       string                   `yaml:"kind,omitempty"`
	EnvMapping *core.KeyPath            `yaml:"env_sync,omitempty"`
	Env        *map[string]core.KeyPath `yaml:"env,omitempty"`
}

func NewTellerFile(s string) (*TellerFile, error) {
	yamlFile, err := os.ReadFile(s)
	if err != nil {
		return nil, err
	}
	t := &TellerFile{}
	err = yaml.Unmarshal(yamlFile, t)
	if err != nil {
		return nil, err
	}
	t.LoadedFrom = s
	return t, nil
}
