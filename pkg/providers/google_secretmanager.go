package providers

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/googleapis/gax-go/v2"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
)

type GoogleSMClient interface {
	AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error)
}
type GoogleSecretManager struct {
	client GoogleSMClient
	logger logging.Logger
}

const GoogleSecretManagerName = "google_secretmanager"

func init() {
	metaInfo := core.MetaInfo{
		Description:    "Google Secret Manager",
		Name:           GoogleSecretManagerName,
		Authentication: "You should populate `GOOGLE_APPLICATION_CREDENTIALS=account.json` in your environment to your relevant `account.json` that you get from Google.",
		ConfigTemplate: `
  # GOOGLE_APPLICATION_CREDENTIALS=foobar.json
  # https://cloud.google.com/secret-manager/docs/reference/libraries#setting_up_authentication
  google_secretmanager:
    env:
      FOO_GOOG:
        # need to supply the relevant version (versions/1)
        path: projects/123/secrets/FOO_GOOG/versions/1
`,
		Ops: core.OpMatrix{Get: true},
	}

	RegisterProvider(metaInfo, NewGoogleSecretManager)
}

func NewGoogleSecretManager(logger logging.Logger) (core.Provider, error) {
	client, err := secretmanager.NewClient(context.TODO())
	if err != nil {
		return nil, err
	}
	return &GoogleSecretManager{client: client, logger: logger}, nil
}

func (a *GoogleSecretManager) Put(p core.KeyPath, val string) error {
	return fmt.Errorf("provider %q does not implement write yet", GoogleSecretManagerName)
}
func (a *GoogleSecretManager) PutMapping(p core.KeyPath, m map[string]string) error {
	return fmt.Errorf("provider %q does not implement write yet", GoogleSecretManagerName)
}

func (a *GoogleSecretManager) GetMapping(kp core.KeyPath) ([]core.EnvEntry, error) {
	return nil, fmt.Errorf("does not support full env sync (path: %s)", kp.Path)
}

func (a *GoogleSecretManager) Get(p core.KeyPath) (*core.EnvEntry, error) {
	secret, err := a.getSecret(p)
	if err != nil {
		return nil, err
	}

	ent := p.Found(secret)
	return &ent, nil
}

func (a *GoogleSecretManager) Delete(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", GoogleSecretManagerName)
}

func (a *GoogleSecretManager) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", GoogleSecretManagerName)
}

func (a *GoogleSecretManager) getSecret(kp core.KeyPath) (string, error) {
	r := secretmanagerpb.AccessSecretVersionRequest{
		Name: kp.Path,
	}
	a.logger.WithField("path", r.Name).Debug("get secret")

	secret, err := a.client.AccessSecretVersion(context.TODO(), &r)
	if err != nil {
		return "", err
	}
	return string(secret.Payload.Data), nil
}
