package providers

import (
	"fmt"
	"sort"

	"github.com/hashicorp/vault/api"
	"github.com/spectralops/teller/pkg/core"
)

type HashicorpClient interface {
	Read(path string) (*api.Secret, error)
	Write(path string, data map[string]interface{}) (*api.Secret, error)
}
type HashicorpVault struct {
	client HashicorpClient
}

func NewHashicorpVault() (core.Provider, error) {
	conf := api.DefaultConfig()
	err := conf.ReadEnvironment()
	if err != nil {
		return nil, err
	}

	client, err := api.NewClient(conf)

	if err != nil {
		return nil, err
	}

	return &HashicorpVault{client: client.Logical()}, nil
}

func (h *HashicorpVault) Name() string {
	return "hashicorp_vault"
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
		entries = append(entries, core.EnvEntry{Key: k, Value: v.(string), Provider: h.Name(), ResolvedPath: p.Path})
	}
	sort.Sort(core.EntriesByKey(entries))
	return entries, nil
}

func (h *HashicorpVault) Get(p core.KeyPath) (*core.EnvEntry, error) {
	secret, err := h.getSecret(p)
	if err != nil {
		return nil, err
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
		k = data[p.Field]
	}

	if k == nil {
		return nil, fmt.Errorf("field at '%s' does not exist", p.Path)
	}

	return &core.EnvEntry{
		Key:          p.Env,
		Value:        k.(string),
		ResolvedPath: p.Path,
		Provider:     h.Name(),
	}, nil
}

func (h *HashicorpVault) Put(p core.KeyPath, val string) error {
	k := p.Env
	if p.Field != "" {
		k = p.Field
	}
	m := map[string]string{k: val}
	_, err := h.client.Write(p.Path, map[string]interface{}{"data": m})
	return err
}

func (h *HashicorpVault) getSecret(kp core.KeyPath) (*api.Secret, error) {
	secret, err := h.client.Read(kp.Path)
	if err != nil {
		return nil, err
	}

	if secret == nil || len(secret.Data) == 0 {
		return nil, fmt.Errorf("data not found at '%s'", kp.Path)
	}

	if len(secret.Warnings) > 0 {
		fmt.Println(secret.Warnings)
	}

	return secret, nil
}
