package providers

import (
	"os"
	"sort"
	"strings"

	"github.com/spectralops/teller/pkg/core"

	"github.com/spectralops/teller/pkg/logging"
)

type ProcessEnv struct {
	logger logging.Logger
}

// NewProcessEnv creates new provider instance
func NewProcessEnv(logger logging.Logger) (core.Provider, error) {
	return &ProcessEnv{
		logger: logger,
	}, nil
}

// Name return the provider name
func (a *ProcessEnv) Name() string {
	return "process_env"
}

// GetMapping returns a multiple entries
func (a *ProcessEnv) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {
	a.logger.Debug("read secret")

	kvs := make(map[string]string)
	for _, envs := range os.Environ() {
		pair := strings.SplitN(envs, "=", 2)
		kvs[pair[0]] = pair[1]
	}
	var entries []core.EnvEntry
	for k, v := range kvs {
		entries = append(entries, p.FoundWithKey(k, v))
	}
	sort.Sort(core.EntriesByKey(entries))
	return entries, nil
}

// Get returns a single entry
func (a *ProcessEnv) Get(p core.KeyPath) (*core.EnvEntry, error) {
	a.logger.Debug("read secret")

	k := p.EffectiveKey()
	val, ok := os.LookupEnv(k)
	if !ok {
		a.logger.WithFields(map[string]interface{}{"key": k}).Debug("key not found")
		ent := p.Missing()
		return &ent, nil
	}

	ent := p.Found(val)
	return &ent, nil
}

// Delete will delete entry
func (a *ProcessEnv) Delete(kp core.KeyPath) error {
	return nil
}

// DeleteMapping will delete the given path recessively
func (a *ProcessEnv) DeleteMapping(kp core.KeyPath) error {
	return nil
}

// Put will create a new single entry
func (a *ProcessEnv) Put(p core.KeyPath, val string) error {
	return nil
}

// PutMapping will create a multiple entries
func (a *ProcessEnv) PutMapping(p core.KeyPath, m map[string]string) error {
	return nil
}
