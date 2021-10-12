package providers

import (
	"context"
	"fmt"
	"os"
	"sort"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/spectralops/teller/pkg/core"
)

type CloudflareClient interface {
	WriteWorkersKV(ctx context.Context, namespaceID, key string, value []byte) (cloudflare.Response, error)
	WriteWorkersKVBulk(ctx context.Context, namespaceID string, kvs cloudflare.WorkersKVBulkWriteRequest) (cloudflare.Response, error)
	ReadWorkersKV(ctx context.Context, namespaceID string, key string) ([]byte, error)
	ListWorkersKVs(ctx context.Context, namespaceID string) (cloudflare.ListStorageKeysResponse, error)
}

type Cloudflare struct {
	client CloudflareClient
}

func NewCloudflareClient() (core.Provider, error) {
	api, err := cloudflare.New(
		os.Getenv("CLOUDFLARE_API_KEY"),
		os.Getenv("CLOUDFLARE_API_EMAIL"),
	)

	cloudflare.UsingAccount(os.Getenv("CLOUDFLARE_ACCOUNT_ID"))(api)

	if err != nil {
		return nil, err
	}

	return &Cloudflare{client: api}, err
}

func (c *Cloudflare) Name() string {
	return "cloudflare_workers_kv"
}

func (c *Cloudflare) Put(p core.KeyPath, val string) error {
	return fmt.Errorf("%v does not implement write yet", c.Name())
}

func (c *Cloudflare) PutMapping(p core.KeyPath, m map[string]string) error {
	return fmt.Errorf("%v does not implement write yet", c.Name())
}

func (c *Cloudflare) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {
	entries, err := c.getSecrets(p)
	if err != nil {
		return nil, err
	}
	sort.Sort(core.EntriesByKey(entries))
	return entries, nil
}

func (c *Cloudflare) Get(p core.KeyPath) (*core.EnvEntry, error) {
	secret, err := c.getSecret(p)
	if err != nil {
		return nil, err
	}
	ent := p.Found(string(secret))
	return &ent, nil
}

func (c *Cloudflare) getSecrets(p core.KeyPath) ([]core.EnvEntry, error) {
	secrets, err := c.client.ListWorkersKVs(context.TODO(), p.Path)
	if err != nil {
		return nil, err
	}

	entries := []core.EnvEntry{}
	for _, k := range secrets.Result {
		p.Field = k.Name
		secret, err := c.getSecret(p)
		if err != nil {
			entries = append(entries, p.Missing())
		}
		entries = append(entries, p.FoundWithKey(k.Name, string(secret)))
	}

	return entries, nil
}

func (c *Cloudflare) getSecret(p core.KeyPath) ([]byte, error) {
	k := p.Field
	if k == "" {
		k = p.Env
	}
	if k == "" {
		return nil, fmt.Errorf("Key required for fetching secrets. Received \"\"")
	}
	return c.client.ReadWorkersKV(context.TODO(), p.Path, k)
}
