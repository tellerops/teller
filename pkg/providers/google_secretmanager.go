package providers

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/googleapis/gax-go/v2"
	"github.com/spectralops/teller/pkg/core"
)

type GoogleSMClient interface {
	AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error)
}
type GoogleSecretManager struct {
	client GoogleSMClient
}

func NewGoogleSecretManager() (core.Provider, error) {
	client, err := secretmanager.NewClient(context.TODO())
	if err != nil {
		return nil, err
	}
	return &GoogleSecretManager{client: client}, nil
}

func (a *GoogleSecretManager) Name() string {
	return "google_secretmanager"
}

func (a *GoogleSecretManager) Put(p core.KeyPath, val string) error {
	return fmt.Errorf("%v does not implement write yet", a.Name())
}

func (a *GoogleSecretManager) GetMapping(kp core.KeyPath) ([]core.EnvEntry, error) {
	return nil, fmt.Errorf("does not support full env sync (path: %s)", kp.Path)
}

func (a *GoogleSecretManager) Get(p core.KeyPath) (*core.EnvEntry, error) {
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

func (a *GoogleSecretManager) getSecret(kp core.KeyPath) (string, error) {
	r := secretmanagerpb.AccessSecretVersionRequest{
		Name: kp.Path,
	}

	secret, err := a.client.AccessSecretVersion(context.TODO(), &r)
	if err != nil {
		return "", err
	}
	return string(secret.Payload.Data), nil
}
