package providers

import (
	"os"
	"path"
	"sort"

	"github.com/joho/godotenv"
	"github.com/mitchellh/go-homedir"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
	"github.com/spectralops/teller/pkg/utils"
)

const (
	filePerm = 0644
	dirPerm  = 0755
)

type DotEnvClient interface {
	Read(p string) (map[string]string, error)
	Write(p string, kvs map[string]string) error
	Exists(p string) (bool, error)
	Delete(p string) error
}
type DotEnvReader struct {
}

func (d *DotEnvReader) Read(p string) (map[string]string, error) {
	p, err := homedir.Expand(p)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}

	return godotenv.Unmarshal(string(content))
}

func (d *DotEnvReader) Write(p string, kvs map[string]string) error {
	content, err := godotenv.Marshal(kvs)
	if err != nil {
		return err
	}

	p, err = homedir.Expand(p)
	if err != nil {
		return err
	}

	// ensure all subdirectories exist
	err = os.MkdirAll(path.Dir(p), dirPerm)
	if err != nil {
		return err
	}

	return os.WriteFile(p, []byte(content), filePerm)
}

func (d *DotEnvReader) Exists(p string) (bool, error) {
	p, err := homedir.Expand(p)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (d *DotEnvReader) Delete(p string) error {
	p, err := homedir.Expand(p)
	if err != nil {
		return err
	}

	return os.Remove(p)
}

type Dotenv struct {
	client DotEnvClient
	logger logging.Logger
}

//nolint
func init() {
	metaInfo := core.MetaInfo{
		Description:    ".env",
		Authentication: "",
		Name:           "dotenv",
		ConfigTemplate: `
  # you can mix and match many files
  dotenv:
    env_sync:
      path: ~/my-dot-env.env
    env:
      FOO_BAR:
        path: ~/my-dot-env.env
`,
		Ops: core.OpMatrix{Get: true, GetMapping: true, Put: true, PutMapping: true, Delete: true, DeleteMapping: true},
	}

	RegisterProvider(metaInfo, NewDotenv)
}

func NewDotenv(logger logging.Logger) (core.Provider, error) {
	return &Dotenv{
		client: &DotEnvReader{},
		logger: logger,
	}, nil
}

func (a *Dotenv) Put(p core.KeyPath, val string) error {
	k := p.EffectiveKey()
	return a.PutMapping(p, map[string]string{k: val})
}

func (a *Dotenv) PutMapping(kp core.KeyPath, m map[string]string) error {
	exists, err := a.client.Exists(kp.Path)
	if err != nil {
		a.logger.WithField("path", kp.Path).Debug("secret path not exists")
		return err
	}

	if !exists {
		a.logger.WithField("path", kp.Path).Debug("set secret")
		return a.client.Write(kp.Path, m)
	}

	a.logger.WithField("path", kp.Path).Debug("read secret")
	// get a fresh copy of a hash
	secrets, err := a.client.Read(kp.Path)
	if err != nil {
		return err
	}

	secrets = utils.Merge(secrets, m)
	a.logger.WithField("path", kp.Path).Debug("merge and write secrets to path")
	return a.client.Write(kp.Path, secrets)
}

func (a *Dotenv) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {
	a.logger.WithField("path", p.Path).Debug("read secret")
	kvs, err := a.client.Read(p.Path)
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

func (a *Dotenv) Get(p core.KeyPath) (*core.EnvEntry, error) { //nolint:dupl
	a.logger.WithField("path", p.Path).Debug("read secret")
	kvs, err := a.client.Read(p.Path)
	if err != nil {
		return nil, err
	}

	k := p.EffectiveKey()
	val, ok := kvs[k]
	if !ok {
		a.logger.WithFields(map[string]interface{}{"path": p.Path, "key": k}).Debug("key not found")
		ent := p.Missing()
		return &ent, nil
	}

	ent := p.Found(val)
	return &ent, nil
}

func (a *Dotenv) Delete(kp core.KeyPath) error {
	a.logger.WithField("path", kp.Path).Debug("read secret")
	kvs, err := a.client.Read(kp.Path)
	if err != nil {
		return err
	}

	k := kp.EffectiveKey()
	delete(kvs, k)

	if len(kvs) == 0 {
		return a.DeleteMapping(kp)
	}

	p, err := homedir.Expand(kp.Path)
	if err != nil {
		return err
	}

	return a.client.Write(p, kvs)
}

func (a *Dotenv) DeleteMapping(kp core.KeyPath) error {
	exists, err := a.client.Exists(kp.Path)
	if err != nil {
		a.logger.WithField("path", kp.Path).Debug("secret path not exists")
		return err
	}

	if !exists {
		// already deleted
		a.logger.WithField("path", kp.Path).Debug("secret already deleted")
		return nil
	}

	a.logger.WithField("path", kp.Path).Debug("delete key")
	return a.client.Delete(kp.Path)
}
