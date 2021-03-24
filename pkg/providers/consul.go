package providers

import (
	"sort"

	"github.com/hashicorp/consul/api"

	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/utils"
)

type ConsulClient interface {
	Get(key string, q *api.QueryOptions) (*api.KVPair, *api.QueryMeta, error)
	List(prefix string, q *api.QueryOptions) (api.KVPairs, *api.QueryMeta, error)
}
type Consul struct {
	client ConsulClient
}

func NewConsul() (core.Provider, error) {
	df := api.DefaultConfig()
	client, err := api.NewClient(df)
	if err != nil {
		return nil, err
	}
	kv := client.KV()
	return &Consul{client: kv}, nil
}

func (a *Consul) Name() string {
	return "consul"
}

func (a *Consul) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {
	kvs, err := a.getSecrets(p)
	if err != nil {
		return nil, err
	}
	entries := []core.EnvEntry{}
	for _, kv := range kvs {
		k := kv.Key
		v := string(kv.Value)
		seg := utils.LastSegment(k)
		entries = append(entries, core.EnvEntry{
			Key:          seg,
			Value:        v,
			ResolvedPath: p.Path,
			Provider:     a.Name(),
		})
	}
	sort.Sort(core.EntriesByKey(entries))
	return entries, nil
}

func (a *Consul) Get(p core.KeyPath) (*core.EnvEntry, error) {
	secret, err := a.getSecret(p)
	if err != nil {
		return nil, err
	}
	return &core.EnvEntry{
		Key:          p.Env,
		Value:        secret,
		ResolvedPath: p.Path,
		Provider:     a.Name(),
	}, nil
}

func (a *Consul) getSecrets(kp core.KeyPath) (api.KVPairs, error) {
	kvs, _, err := a.client.List(kp.Path, nil)
	return kvs, err
}

func (a *Consul) getSecret(kp core.KeyPath) (string, error) {
	kv, _, err := a.client.Get(kp.Path, nil)
	if err != nil {
		return "", err
	}
	if kv == nil {
		return "", err
	}
	return string(kv.Value), nil

}
