package providers

import (
	"fmt"
	"sort"

	"github.com/alexsasharegan/dotenv"
	"github.com/mitchellh/go-homedir"
	"github.com/spectralops/teller/pkg/core"
)

type DotEnvClient interface {
	Read(p string) (map[string]string, error)
}
type DotEnvReader struct {
}

func (d *DotEnvReader) Read(p string) (map[string]string, error) {
	return dotenv.ReadFile(p)
}

type Dotenv struct {
	client DotEnvClient
}

func NewDotenv() (core.Provider, error) {
	return &Dotenv{
		client: &DotEnvReader{},
	}, nil
}

func (a *Dotenv) Name() string {
	return "dotenv"
}

func (a *Dotenv) Put(p core.KeyPath, val string) error {
	return fmt.Errorf("%v does not implement write yet", a.Name())
}

func (a *Dotenv) PutMapping(p core.KeyPath, m map[string]string) error {
	return fmt.Errorf("%v does not implement write yet", a.Name())
}

func (a *Dotenv) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {
	kvs, err := a.getSecrets(p)
	if err != nil {
		return nil, err
	}
	entries := []core.EnvEntry{}
	for k, v := range kvs {
		entries = append(entries, p.FoundWithKey(k, v))
	}
	sort.Sort(core.EntriesByKey(entries))
	return entries, nil
}

func (a *Dotenv) Get(p core.KeyPath) (*core.EnvEntry, error) {
	kvs, err := a.getSecrets(p)
	if err != nil {
		return nil, err
	}
	val, ok := kvs[p.Field]
	if val == "" {
		val, ok = kvs[p.Env]
	}

	if !ok {
		ent := p.Missing()
		return &ent, nil
	}

	ent := p.Found(val)
	return &ent, nil
}

func (a *Dotenv) getSecrets(kp core.KeyPath) (map[string]string, error) {
	p, err := homedir.Expand(kp.Path)
	if err != nil {
		return nil, err
	}
	return a.client.Read(p)
}
