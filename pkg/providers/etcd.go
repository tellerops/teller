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

	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/utils"
	"github.com/thoas/go-funk"
)

type EtcdClient interface {
	Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)
}
type Etcd struct {
	client EtcdClient
}

func NewEtcd() (core.Provider, error) {
	epstring := os.Getenv("ETCDCTL_ENDPOINTS")
	if epstring == "" {
		return nil, fmt.Errorf("cannot find ETCDCTL_ENDPOINTS for etcd")
	}

	eps := funk.Map(strings.Split(epstring, ","), func(s string) string { return strings.Trim(s, " ") }).([]string)
	client, err := newClient(eps)
	if err != nil {
		return nil, err
	}
	return &Etcd{client: client}, nil
}

func (a *Etcd) Name() string {
	return "etcd"
}

func (a *Etcd) Put(p core.KeyPath, val string) error {
	return fmt.Errorf("%v does not implement write yet", a.Name())
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

func (a *Etcd) Get(p core.KeyPath) (*core.EnvEntry, error) {

	kvs, err := a.getSecret(p)
	if err != nil {
		return nil, err
	}
	for _, kv := range kvs {
		k := string(kv.Key)
		v := string(kv.Value)
		if k == p.Path {
			return &core.EnvEntry{
				Key:          p.Env,
				Value:        v,
				ResolvedPath: p.Path,
				Provider:     a.Name(),
			}, nil
		}
	}

	return nil, fmt.Errorf("key %s not found", p.Path)
}

func (a *Etcd) getSecret(kp core.KeyPath, opts ...clientv3.OpOption) ([]*spb.KeyValue, error) {
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
