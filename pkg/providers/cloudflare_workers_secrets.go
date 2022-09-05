package providers

import (
	"context"
	"errors"
	"fmt"
	"os"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
)

var (
	ErrCloudFlareSourceFieldIsMissing = errors.New("`source` field is missing")
)

type CloudflareSecretsClient interface {
	SetWorkersSecret(ctx context.Context, script string, req *cloudflare.WorkersPutSecretRequest) (cloudflare.WorkersPutSecretResponse, error)
	DeleteWorkersSecret(ctx context.Context, script, secretName string) (cloudflare.Response, error)
}

type CloudflareSecrets struct {
	client CloudflareSecretsClient
	logger logging.Logger
}

const CloudflareWorkersSecretName = "cloudflare_workers_secret"

//nolint
func init() {
	metaInfo := core.MetaInfo{
		Description:    "Cloudflare Workers Secrets",
		Name:           CloudflareWorkersSecretName,
		Authentication: "requires the following environment variables to be set:\n`CLOUDFLARE_API_KEY`: Your Cloudflare api key.\n`CLOUDFLARE_API_EMAIL`: Your email associated with the api key.\n`CLOUDFLARE_ACCOUNT_ID`: Your account ID.\n",
		ConfigTemplate: `
  # Configure via environment variables for integration:
  # CLOUDFLARE_API_KEY: Your Cloudflare api key.
  # CLOUDFLARE_API_EMAIL: Your email associated with the api key.
  # CLOUDFLARE_ACCOUNT_ID: Your account ID.

  cloudflare_workers_secrets:
    env_sync:
      source: # Mandatory: script field
    env:
      script-value:
        path: foo-secret
        source: # Mandatory: script field
		`,
		Ops: core.OpMatrix{Put: true, PutMapping: true, Delete: true},
	}
	RegisterProvider(metaInfo, NewCloudflareSecretsClient)
}

func NewCloudflareSecretsClient(logger logging.Logger) (core.Provider, error) {
	api, err := cloudflare.New(
		os.Getenv("CLOUDFLARE_API_KEY"),
		os.Getenv("CLOUDFLARE_API_EMAIL"),
	)

	if err != nil {
		return nil, err
	}

	cloudflare.UsingAccount(os.Getenv("CLOUDFLARE_ACCOUNT_ID"))(api) //nolint
	return &CloudflareSecrets{client: api, logger: logger}, nil
}

func (c *CloudflareSecrets) Put(p core.KeyPath, val string) error {

	if p.Source == "" {
		return ErrCloudFlareSourceFieldIsMissing
	}

	secretName, err := c.getSecretName(p)
	if err != nil {
		return err
	}

	secretRequest := cloudflare.WorkersPutSecretRequest{
		Name: secretName,
		Text: val,
		Type: cloudflare.WorkerSecretTextBindingType,
	}

	c.logger.WithFields(map[string]interface{}{
		"script": p.Source,
		"name":   secretRequest.Name,
	}).Debug("set workers secret")
	_, err = c.client.SetWorkersSecret(context.TODO(), p.Source, &secretRequest)

	return err
}

func (c *CloudflareSecrets) PutMapping(p core.KeyPath, m map[string]string) error {
	if p.Source == "" {
		return ErrCloudFlareSourceFieldIsMissing
	}

	for k, v := range m {
		ap := p.WithEnv(fmt.Sprintf("%v/%v", p.Path, k))

		err := c.Put(ap, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CloudflareSecrets) Delete(p core.KeyPath) error {

	if p.Source == "" {
		return ErrCloudFlareSourceFieldIsMissing
	}

	secretName, err := c.getSecretName(p)
	if err != nil {
		return err
	}

	c.logger.WithFields(map[string]interface{}{
		"script": p.Source,
		"name":   secretName,
	}).Debug("delete workers secret")
	_, err = c.client.DeleteWorkersSecret(context.TODO(), p.Source, secretName)
	return err
}

func (c *CloudflareSecrets) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {
	return nil, fmt.Errorf("%s does not support read functionality", CloudflareWorkersSecretName)
}

func (c *CloudflareSecrets) Get(p core.KeyPath) (*core.EnvEntry, error) {
	return nil, fmt.Errorf("%s does not support read functionality", CloudflareWorkersSecretName)
}

func (c *CloudflareSecrets) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement deleteMapping yet", CloudflareWorkersSecretName)
}

func (c *CloudflareSecrets) getSecretName(p core.KeyPath) (string, error) {

	k := p.Field
	if k == "" {
		c.logger.WithField("field", p.Field).Debug("`field` attribute not configured. trying to get `env` attribute")
		k = p.Env
	}
	if k == "" {
		return "", fmt.Errorf("key required for fetching secrets. Received \"\"")
	}
	return k, nil

}
