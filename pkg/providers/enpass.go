package providers

import (
	"fmt"
	"os"
	"sort"

	"github.com/sirupsen/logrus"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
	"github.com/v-braun/enpass-cli/pkg/enpass"
)

type EnpassClient interface {
	GetEntry(cardType string, filters []string, unique bool) (*enpass.Card, error)
	GetEntries(cardType string, filters []string) ([]enpass.Card, error)
}

type Enpass struct {
	client EnpassClient
	logger logging.Logger
}

func NewEnpass(logger logging.Logger) (core.Provider, error) {
	vaultPath := os.Getenv("ENPASS_VAULT_PATH")
	masterPassword := os.Getenv("ENPASS_PASSWORD")
	vault, err := enpass.NewVault(vaultPath, logrus.ErrorLevel)
	if err != nil {
		return nil, err
	}
	creds := enpass.VaultCredentials{
		Password: masterPassword,
	}
	err = vault.Open(&creds)
	if err != nil {
		return nil, err
	}
	return &Enpass{client: vault, logger: logger}, nil
}

func (e *Enpass) Name() string {
	return "enpass"
}

func (e *Enpass) Put(p core.KeyPath, val string) error {
	return fmt.Errorf("provider %q does not implement write yet", e.Name())
}

func (e *Enpass) PutMapping(p core.KeyPath, m map[string]string) error {
	return fmt.Errorf("provider %q does not implement write yet", e.Name())
}

func (e *Enpass) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {
	res, err := e.getEntries(p)
	if err != nil {
		return nil, err
	}

	entries := []core.EnvEntry{}
	for key, val := range res {
		entries = append(entries, p.FoundWithKey(key, val))
	}
	sort.Sort(core.EntriesByKey(entries))

	return entries, nil
}

func (e *Enpass) Get(p core.KeyPath) (*core.EnvEntry, error) {
	c, err := e.getEntry(p)

	if err != nil {
		return nil, err
	}
	value, err := c.Decrypt()
	if err != nil {
		return nil, err
	}
	entry := p.Found(value)
	return &entry, nil
}

func (e *Enpass) Delete(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", e.Name())
}

func (e *Enpass) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", e.Name())
}

func (e *Enpass) getEntry(p core.KeyPath) (*enpass.Card, error) {
	entry, err := e.client.GetEntry(p.Path, []string{p.Env}, true)
	if err != nil {
		return nil, err
	}
	return entry, err
}

func (e *Enpass) getEntries(p core.KeyPath) (map[string]string, error) {
	res, err := e.client.GetEntries(p.Path, []string{})
	if err != nil {
		return nil, err
	}

	entries := map[string]string{}
	for _, entry := range res {
		if !entry.IsTrashed() && !entry.IsDeleted() {
			value, err := entry.Decrypt()
			if err != nil {
				return nil, err
			}
			entries[fmt.Sprintf("%v/%v", entry.UUID, entry.Title)] = value
		}
	}
	return entries, nil
}
