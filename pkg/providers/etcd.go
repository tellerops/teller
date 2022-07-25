package providers

import (
	"context"
	"crypto/tls"
	"fmt"
	"sort"

	"os"
	"strings"

	spb "go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"

	"go.etcd.io/etcd/pkg/v3/transport"

	"github.com/samber/lo"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
	"github.com/spectralops/teller/pkg/utils"
)

type EtcdClient interface {
	Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)
	Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error)
}
type Etcd struct {
	client EtcdClient
	logger logging.Logger
}

const EtcdName = "etcd"

//nolint
func init() {
	metaInfo := core.MetaInfo{
		Description:    "Etcd",
		Name:           EtcdName,
		Authentication: "These environment variables need to be populated\n* `ETCDCTL_ENDPOINTS`\nFor TLS:\n* `ETCDCTL_CA_FILE`\n* `ETCDCTL_CERT_FILE`\n* `ETCDCTL_KEY_FILE`",
		ConfigTemplate: `
  # Configure via environment:
  # ETCDCTL_ENDPOINTS
  # tls:
  # ETCDCTL_CA_FILE
  # ETCDCTL_CERT_FILE
  # ETCDCTL_KEY_FILE
  etcd:
    env_sync:
      path: /prod/foo
    env:
      ETC_DSN:
        path: /prod/foo/bar
`,
		Ops: core.OpMatrix{Get: true, GetMapping: true, Put: true, PutMapping: true},
	}
	RegisterProvider(metaInfo, NewEtcd)
}

func NewEtcd(logger logging.Logger) (core.Provider, error) {
	epstring := os.Getenv("ETCDCTL_ENDPOINTS")
	if epstring == "" {
		return nil, fmt.Errorf("cannot find ETCDCTL_ENDPOINTS for etcd")
	}

	eps := lo.Map(strings.Split(epstring, ","), func(s string, _ int) string { return strings.Trim(s, " ") })
	client, err := newClient(eps)
	if err != nil {
		return nil, err
	}
	return &Etcd{client: client, logger: logger}, nil
}

func (a *Etcd) Put(p core.KeyPath, val string) error {
	a.logger.WithField("path", p.Path).Debug("create key")
	_, err := a.client.Put(context.TODO(), p.Path, val)
	return err
}
func (a *Etcd) PutMapping(p core.KeyPath, m map[string]string) error {
	for k, v := range m {
		ap := p.SwitchPath(fmt.Sprintf("%v/%v", p.Path, k))
		err := a.Put(ap, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Etcd) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {
	kvs, err := a.getSecret(p, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	entries := []core.EnvEntry{}
	for _, kv := range kvs {
		k := string(kv.Key)
		v := string(kv.Value)
		seg := utils.LastSegment(k)
		entries = append(entries, p.FoundWithKey(seg, v))
	}
	sort.Sort(core.EntriesByKey(entries))
	return entries, nil
}

func (a *Etcd) Get(p core.KeyPath) (*core.EnvEntry, error) {
	kvs, err := a.getSecret(p)
	if err != nil {
		return nil, err
	}
	for _, kv := range kvs {
		k := string(kv.Key)
		v := string(kv.Value)
		if k == p.Path {
			ent := p.Found(v)
			return &ent, nil
		}
	}

	ent := p.Missing()
	return &ent, nil
}

func (a *Etcd) Delete(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", EtcdName)
}

func (a *Etcd) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", EtcdName)
}

func (a *Etcd) getSecret(kp core.KeyPath, opts ...clientv3.OpOption) ([]*spb.KeyValue, error) {
	a.logger.WithField("path", kp.Path).Debug("get key")
	res, err := a.client.Get(context.TODO(), kp.Path, opts...)
	if err != nil {
		return nil, err
	}
	return res.Kvs, nil
}

func newClient(eps []string) (*clientv3.Client, error) {
	tr, err := getTransport()
	if err != nil {
		return nil, err
	}

	cfg := clientv3.Config{
		TLS:       tr,
		Endpoints: eps,
	}
	return clientv3.New(cfg)
}
func getTransport() (*tls.Config, error) {
	cafile := os.Getenv("ETCDCTL_CA_FILE")
	certfile := os.Getenv("ETCDCTL_CERT_FILE")
	keyfile := os.Getenv("ETCDCTL_KEY_FILE")
	if cafile == "" || certfile == "" || keyfile == "" {
		return nil, nil
	}

	tlsinfo := &transport.TLSInfo{
		CertFile:      certfile,
		KeyFile:       keyfile,
		TrustedCAFile: cafile,
	}
	return tlsinfo.ClientConfig()
}
