package providers

import (
	"context"
	"fmt"
	"os"
	"sort"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
)

type CloudflareClient interface {
	WriteWorkersKV(ctx context.Context, namespaceID, key string, value []byte) (cloudflare.Response, error)
	WriteWorkersKVBulk(ctx context.Context, namespaceID string, kvs cloudflare.WorkersKVBulkWriteRequest) (cloudflare.Response, error)
	ReadWorkersKV(ctx context.Context, namespaceID string, key string) ([]byte, error)
	ListWorkersKVs(ctx context.Context, namespaceID string) (cloudflare.ListStorageKeysResponse, error)
}

type Cloudflare struct {
	client CloudflareClient
	logger logging.Logger
}

const cloudFlareWorkersKVName = "cloudflare_workers_kv"

//nolint
func init() {
	metaInfo := core.MetaInfo{
		Description:    "Cloudflare Workers K/V",
		Name:           cloudFlareWorkersKVName,
		Authentication: "requires the following environment variables to be set:\n`CLOUDFLARE_API_KEY`: Your Cloudflare api key.\n`CLOUDFLARE_API_EMAIL`: Your email associated with the api key.\n`CLOUDFLARE_ACCOUNT_ID`: Your account ID.\n",
		ConfigTemplate: `
		TODO(XXX): Missing
`,
		Ops: core.OpMatrix{Get: true, GetMapping: true},
	}
	RegisterProvider(metaInfo, NewCloudflareClient)
}

func NewCloudflareClient(logger logging.Logger) (core.Provider, error) {
	api, err := cloudflare.New(
		os.Getenv("CLOUDFLARE_API_KEY"),
		os.Getenv("CLOUDFLARE_API_EMAIL"),
	)

	if err != nil {
		return nil, err
	}

	cloudflare.UsingAccount(os.Getenv("CLOUDFLARE_ACCOUNT_ID"))(api) //nolint

	return &Cloudflare{client: api, logger: logger}, nil
}

func (c *Cloudflare) Put(p core.KeyPath, val string) error {
	return fmt.Errorf("provider %q does not implement write yet", cloudFlareWorkersKVName)
}

func (c *Cloudflare) PutMapping(p core.KeyPath, m map[string]string) error {
	return fmt.Errorf("provider %q does not implement write yet", cloudFlareWorkersKVName)
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

func (c *Cloudflare) Delete(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", cloudFlareWorkersKVName)
}

func (c *Cloudflare) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", cloudFlareWorkersKVName)
}

func (c *Cloudflare) getSecrets(p core.KeyPath) ([]core.EnvEntry, error) {
	c.logger.WithField("namespace_id", p.Path).Debug("get workers KVs")
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
		c.logger.WithField("field", p.Field).Debug("`field` attribute not configured. trying to get `env` attribute")
		k = p.Env
	}
	if k == "" {
		return nil, fmt.Errorf("Key required for fetching secrets. Received \"\"") //nolint
	}
	c.logger.WithFields(map[string]interface{}{
		"namespace_id": p.Path,
		"name":         k,
	}).Debug("read worker kv")
	return c.client.ReadWorkersKV(context.TODO(), p.Path, k)
}
