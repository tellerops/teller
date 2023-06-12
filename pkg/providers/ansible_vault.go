package providers

import (
	"fmt"
	"os"
	"sort"

	"github.com/joho/godotenv"
	vault "github.com/sosedoff/ansible-vault-go"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
)

type AnsibleVaultClient interface {
	Read(p string) (map[string]string, error)
}

type AnsibleVaultReader struct {
	passPhrase string
}

func (a AnsibleVaultReader) Read(p string) (map[string]string, error) {
	content, err := vault.DecryptFile(p, a.passPhrase)
	if err != nil {
		return nil, err
	}
	return godotenv.Unmarshal(content)
}

type AnsibleVault struct {
	logger logging.Logger
	client AnsibleVaultClient
}

//nolint
func init() {
	metaInto := core.MetaInfo{
		Description:    "Ansible Vault",
		Name:           "ansible_vault",
		Authentication: "ANSIBLE_VAULT_PASSPHRASE.",
		ConfigTemplate: `
  # Configure via environment variables for integration:
  # ANSIBLE_VAULT_PASSPHRASE: Ansible Vault Password

  ansible_vault:
    env_sync:
       path: ansible/vars/vault_{{stage}}.yml

    env:
      KEY1:
        path: ansible/vars/vault_{{stage}}.yml
      NONEXIST_KEY:
        path: ansible/vars/vault_{{stage}}.yml
`,
		Ops: core.OpMatrix{Get: true, GetMapping: true, Put: false, PutMapping: false},
	}
	RegisterProvider(metaInto, NewAnsibleVault)
}

// NewAnsibleVault creates new provider instance
func NewAnsibleVault(logger logging.Logger) (core.Provider, error) {
	ansibleVaultPassphrase := os.Getenv("ANSIBLE_VAULT_PASSPHRASE")
	return &AnsibleVault{
		logger: logger,
		client: &AnsibleVaultReader{
			passPhrase: ansibleVaultPassphrase,
		},
	}, nil
}

// Name return the provider name
func (a *AnsibleVault) Name() string {
	return "AnsibleVault"
}

// Put will create a new single entry
func (a *AnsibleVault) Put(p core.KeyPath, val string) error {
	return fmt.Errorf("provider %q does not implement write yet", a.Name())
}

// PutMapping will create a multiple entries
func (a *AnsibleVault) PutMapping(p core.KeyPath, m map[string]string) error {
	return fmt.Errorf("provider %q does not implement write yet", a.Name())
}

// GetMapping returns a multiple entries
func (a *AnsibleVault) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {
	// Read existing secret
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

// Get returns a single entry
func (a *AnsibleVault) Get(p core.KeyPath) (*core.EnvEntry, error) { //nolint:dupl
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

// Delete will delete entry
func (a *AnsibleVault) Delete(kp core.KeyPath) error {
	return fmt.Errorf("provider %s does not implement delete yet", a.Name())
}

// DeleteMapping will delete the given path recessively
func (a *AnsibleVault) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("provider %s does not implement delete yet", a.Name())
}
