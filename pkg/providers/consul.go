package providers

import (
	"fmt"
	"sort"

	"github.com/hashicorp/consul/api"

	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/utils"
)

type ConsulClient interface {
	Get(key string, q *api.QueryOptions) (*api.KVPair, *api.QueryMeta, error)
	List(prefix string, q *api.QueryOptions) (api.KVPairs, *api.QueryMeta, error)
	Put(p *api.KVPair, q *api.WriteOptions) (*api.WriteMeta, error)
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

func (a *Consul) Put(p core.KeyPath, val string) error {
	_, err := a.client.Put(&api.KVPair{
		Key:   p.Path,
		Value: []byte(val),
	}, &api.WriteOptions{})

	return err
}

func (a *Consul) PutMapping(p core.KeyPath, m map[string]string) error {
	for k, v := range m {
		ap := p.SwitchPath(fmt.Sprintf("%v/%v", p.Path, k))
		err := a.Put(ap, v)
		if err != nil {
			return err
		}
	}
	return nil
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
		entries = append(entries, p.FoundWithKey(seg, v))
	}
	sort.Sort(core.EntriesByKey(entries))
	return entries, nil
}

func (a *Consul) Get(p core.KeyPath) (*core.EnvEntry, error) {
	kv, err := a.getSecret(p)
	if err != nil {
		return nil, fmt.Errorf("%v cannot get value: %v", a.Name(), err)
	}

	if kv == nil {
		ent := p.Missing()
		return &ent, nil
	}

	ent := p.Found(string(kv.Value))
	return &ent, nil
}

func (a *Consul) getSecrets(kp core.KeyPath) (api.KVPairs, error) {
	kvs, _, err := a.client.List(kp.Path, nil)
	return kvs, err
}

func (a *Consul) getSecret(kp core.KeyPath) (*api.KVPair, error) {
	kv, _, err := a.client.Get(kp.Path, nil)
	if err != nil {
		return nil, err
	}
	return kv, nil

}
