package providers

import (
	"fmt"
	"sort"

	"github.com/hashicorp/consul/api"

	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
	"github.com/spectralops/teller/pkg/utils"
)

type ConsulClient interface {
	Get(key string, q *api.QueryOptions) (*api.KVPair, *api.QueryMeta, error)
	List(prefix string, q *api.QueryOptions) (api.KVPairs, *api.QueryMeta, error)
	Put(p *api.KVPair, q *api.WriteOptions) (*api.WriteMeta, error)
}

type Consul struct {
	client ConsulClient
	logger logging.Logger
}

const consulName = "consul"

//nolint
func init() {
	metaInto := core.MetaInfo{
		Description:    "Consul",
		Name:           consulName,
		Authentication: "If you have the Consul CLI working and configured, there's no special action to take.\nConfiguration is environment based, as defined by client standard. See variables [here](https://github.com/hashicorp/consul/blob/master/api/api.go#L28).",
		ConfigTemplate: `
  # Configure via environment:
  # CONSUL_HTTP_ADDR
  consul:
    env_sync:
      path: redis/config
    env:
      ETC_DSN:
        path: redis/config/foobar
`,
		Ops: core.OpMatrix{Get: true, GetMapping: true, Put: true, PutMapping: true},
	}
	RegisterProvider(metaInto, NewConsul)
}

func NewConsul(logger logging.Logger) (core.Provider, error) {
	df := api.DefaultConfig()
	client, err := api.NewClient(df)
	if err != nil {
		return nil, err
	}
	kv := client.KV()
	return &Consul{client: kv, logger: logger}, nil
}

func (a *Consul) Put(p core.KeyPath, val string) error {
	a.logger.WithField("path", p.Path).Debug("put value")
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
		return nil, fmt.Errorf("%v cannot get value: %v", consulName, err)
	}

	if kv == nil {
		a.logger.WithField("path", p.Path).Debug("kv is empty")
		ent := p.Missing()
		return &ent, nil
	}

	ent := p.Found(string(kv.Value))
	return &ent, nil
}

func (a *Consul) Delete(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", consulName)
}

func (a *Consul) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", consulName)
}

func (a *Consul) getSecrets(kp core.KeyPath) (api.KVPairs, error) {
	a.logger.WithField("path", kp.Path).Debug("get all keys under a prefix")
	kvs, _, err := a.client.List(kp.Path, nil)
	return kvs, err
}

func (a *Consul) getSecret(kp core.KeyPath) (*api.KVPair, error) {
	a.logger.WithField("path", kp.Path).Debug("get value")
	kv, _, err := a.client.Get(kp.Path, nil)
	if err != nil {
		return nil, err
	}
	return kv, nil
}
