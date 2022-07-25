package providers

import (
	"fmt"
	"sort"

	"github.com/hashicorp/vault/api"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
)

type HashicorpClient interface {
	Read(path string) (*api.Secret, error)
	Write(path string, data map[string]interface{}) (*api.Secret, error)
}
type HashicorpVault struct {
	client HashicorpClient
	logger logging.Logger
}

const HashicorpVaultName = "hashicorp_vault"

//nolint
func init() {
	metaInfo := core.MetaInfo{
		Description:    "Hashicorp Vault",
		Name:           HashicorpVaultName,
		Authentication: "Configuration is environment based, as defined by client standard. See variables [here](https://github.com/hashicorp/vault/blob/api/v1.0.4/api/client.go#L28).",
		ConfigTemplate: `
  # configure only from environment
  # https://github.com/hashicorp/vault/blob/api/v1.0.4/api/client.go#L28
  # this vars should not go through to the executing cmd
  hashicorp_vault:
    env_sync:
      path: secret/data/{{"{{stage}}"}}/billing/web/env
    env:
      SMTP_PASS:
        path: secret/data/{{"{{stage}}"}}/wordpress
        field: smtp
`,
		Ops: core.OpMatrix{Get: true, GetMapping: true, Put: true, PutMapping: true},
	}

	RegisterProvider(metaInfo, NewHashicorpVault)
}

func NewHashicorpVault(logger logging.Logger) (core.Provider, error) {
	conf := api.DefaultConfig()
	err := conf.ReadEnvironment()
	if err != nil {
		return nil, err
	}

	client, err := api.NewClient(conf)

	if err != nil {
		return nil, err
	}

	return &HashicorpVault{client: client.Logical(), logger: logger}, nil
}

func (h *HashicorpVault) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {
	secret, err := h.getSecret(p)
	if err != nil {
		return nil, err
	}

	// vault returns a secret kv struct as either data{} or data.data{} depending on engine
	var k map[string]interface{}
	if val, ok := secret.Data["data"]; ok {
		k = val.(map[string]interface{})
	} else {
		k = secret.Data
	}

	entries := []core.EnvEntry{}
	for k, v := range k {
		entries = append(entries, p.FoundWithKey(k, v.(string)))
	}
	sort.Sort(core.EntriesByKey(entries))
	return entries, nil
}

func (h *HashicorpVault) Get(p core.KeyPath) (*core.EnvEntry, error) {
	secret, err := h.getSecret(p)
	if err != nil {
		return nil, err
	}

	if secret == nil {
		h.logger.WithField("path", p.Path).Debug("secret is empty")
		ent := p.Missing()
		return &ent, nil
	}

	// vault returns a secret kv struct as either data{} or data.data{} depending on engine
	var data map[string]interface{}
	if val, ok := secret.Data["data"]; ok {
		data = val.(map[string]interface{})
	} else {
		data = secret.Data
	}

	k := data[p.Env]
	if p.Field != "" {
		h.logger.WithField("path", p.Path).Debug("`env` attribute not found in returned data. take `field` attribute")
		k = data[p.Field]
	}

	if k == nil {
		h.logger.WithField("path", p.Path).Debug("key not found")
		ent := p.Missing()
		return &ent, nil
	}

	ent := p.Found(k.(string))
	return &ent, nil
}

func (h *HashicorpVault) Put(p core.KeyPath, val string) error {
	k := p.Env
	if p.Field != "" {
		h.logger.WithField("path", p.Path).Debug("`env` attribute not configured. take `field` attribute")
		k = p.Field
	}
	m := map[string]string{k: val}
	h.logger.WithField("path", p.Path).Debug("write secret")
	_, err := h.client.Write(p.Path, map[string]interface{}{"data": m})
	return err
}
func (h *HashicorpVault) PutMapping(p core.KeyPath, m map[string]string) error {
	h.logger.WithField("path", p.Path).Debug("write secret")
	_, err := h.client.Write(p.Path, map[string]interface{}{"data": m})
	return err
}

func (h *HashicorpVault) Delete(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", HashicorpVaultName)
}

func (h *HashicorpVault) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", HashicorpVaultName)
}

func (h *HashicorpVault) getSecret(kp core.KeyPath) (*api.Secret, error) {
	h.logger.WithField("path", kp.Path).Debug("read secret")
	secret, err := h.client.Read(kp.Path)
	if err != nil {
		return nil, err
	}

	if secret == nil || len(secret.Data) == 0 {
		return nil, fmt.Errorf("secret not found in path: %s", kp.Path)
	}

	if len(secret.Warnings) > 0 {
		fmt.Println(secret.Warnings)
	}

	return secret, nil
}
