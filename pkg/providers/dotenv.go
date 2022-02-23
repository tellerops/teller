package providers

import (
	"fmt"
	"os"
	"path"
	"sort"

	"github.com/joho/godotenv"
	"github.com/mitchellh/go-homedir"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/utils"
)

const (
	filePerm = 0644
	dirPerm  = 0755
)

type DotEnvClient interface {
	Read(p string) (map[string]string, error)
	Write(p string, kvs map[string]string) error
}
type DotEnvReader struct {
}

func (d *DotEnvReader) Read(p string) (map[string]string, error) {
	content, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	return godotenv.Unmarshal(string(content))
}

func (d *DotEnvReader) Write(p string, kvs map[string]string) error {
	content, err := godotenv.Marshal(kvs)
	if err != nil {
		return err
	}
	return os.WriteFile(p, []byte(content), filePerm)
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
	k := p.EffectiveKey()
	return a.PutMapping(p, map[string]string{k: val})
}

func (a *Dotenv) PutMapping(kp core.KeyPath, m map[string]string) error {
	p, err := homedir.Expand(kp.Path)
	if err != nil {
		return err
	}

	// check if the file does exist
	_, err = os.Stat(p)
	switch {
	case err == nil:
		// get a fresh copy of a hash
		var into map[string]string
		into, err = a.client.Read(p)
		if err != nil {
			return err
		}
		utils.Merge(m, into)
		return a.client.Write(p, into)
	case os.IsNotExist(err):
		// ensure all subdirectories exist
		err = os.MkdirAll(path.Dir(p), dirPerm)
		if err != nil {
			return err
		}
		return a.client.Write(p, m)
	default:
		return err
	}
}

func (a *Dotenv) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {
	kvs, err := a.getSecrets(p)
	if err != nil {
		return nil, err
	}
	var entries []core.EnvEntry
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

	k := p.EffectiveKey()
	val, ok := kvs[k]
	if !ok {
		ent := p.Missing()
		return &ent, nil
	}

	ent := p.Found(val)
	return &ent, nil
}

func (a *Dotenv) Delete(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", a.Name())
}

func (a *Dotenv) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", a.Name())
}

func (a *Dotenv) getSecrets(kp core.KeyPath) (map[string]string, error) {
	p, err := homedir.Expand(kp.Path)
	if err != nil {
		return nil, err
	}
	return a.client.Read(p)
}
