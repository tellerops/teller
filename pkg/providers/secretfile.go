package providers

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type SecretFileClient interface {
	Read(p string) (map[string]string, error)
	Exists(p string) (bool, error)
}
type SecretFileReader struct {
}

func (d *SecretFileReader) Read(p string) (map[string]string, error) {
	p, err := homedir.Expand(p)
	if err != nil {
		return nil, err
	}

	secrets := make(map[string]string)

	err = filepath.Walk(p,
		func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			if err != nil {
				return err
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			secrets[filepath.Base(path)] = strings.Trim(string(content), "\r\n")
			return nil
		})
	if err != nil {
		return nil, err
	}

	return secrets, nil
}

func (d *SecretFileReader) Exists(p string) (bool, error) {
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

func (d *SecretFileReader) Delete(p string) error {
	p, err := homedir.Expand(p)
	if err != nil {
		return err
	}

	return os.Remove(p)
}

type SecretFile struct {
	client SecretFileClient
	logger logging.Logger
}

func NewSecretFile(logger logging.Logger) (core.Provider, error) {
	return &SecretFile{
		client: &SecretFileReader{},
		logger: logger,
	}, nil
}

func (s *SecretFile) Name() string {
	return "secretfile"
}

func (s *SecretFile) Put(_ core.KeyPath, _ string) error {
	return fmt.Errorf("provider %q does not support write", s.Name())
}
func (s *SecretFile) PutMapping(_ core.KeyPath, _ map[string]string) error {
	return fmt.Errorf("provider %q does not support write", s.Name())
}

func (s *SecretFile) Delete(_ core.KeyPath) error {
	return fmt.Errorf("provider %q does not support delete", s.Name())
}

func (s *SecretFile) DeleteMapping(_ core.KeyPath) error {
	return fmt.Errorf("provider %q does not support delete", s.Name())
}

func (s *SecretFile) Get(p core.KeyPath) (*core.EnvEntry, error) {
	s.logger.WithField("path", p.Path).Debug("read secret file")
	kvs, err := s.client.Read(p.Path)
	if err != nil {
		return nil, err
	}

	k := filepath.Base(p.Path)
	val, ok := kvs[k]
	if !ok {
		s.logger.WithFields(map[string]interface{}{"path": p.Path, "key": k}).Debug("key not found")
		ent := p.Missing()
		return &ent, nil
	}

	ent := p.Found(val)
	return &ent, nil
}

func (s *SecretFile) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {
	s.logger.WithField("path", p.Path).Debug("read secret")
	kvs, err := s.client.Read(p.Path)
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
