package providers

import (
	"context"
	"fmt"
	"sort"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/googleapis/gax-go/v2"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
	"google.golang.org/api/iterator"
)

type GoogleSMClient interface {
	AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error)
	DestroySecretVersion(ctx context.Context, req *secretmanagerpb.DestroySecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.SecretVersion, error)
	ListSecrets(ctx context.Context, in *secretmanagerpb.ListSecretsRequest, opts ...gax.CallOption) *secretmanager.SecretIterator
	AddSecretVersion(ctx context.Context, req *secretmanagerpb.AddSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.SecretVersion, error)
}
type GoogleSecretManager struct {
	client GoogleSMClient
	logger logging.Logger
}

func NewGoogleSecretManager(logger logging.Logger) (core.Provider, error) {
	client, err := secretmanager.NewClient(context.TODO())
	if err != nil {
		return nil, err
	}
	return &GoogleSecretManager{client: client, logger: logger}, nil
}

func (a *GoogleSecretManager) Name() string {
	return "google_secretmanager"
}

func (a *GoogleSecretManager) Put(p core.KeyPath, val string) error {
	i := strings.LastIndex(p.Path, "/versions/")
	req := &secretmanagerpb.AddSecretVersionRequest{
		Parent: p.Path[:i],
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(val),
		},
	}

	_, err := a.client.AddSecretVersion(context.TODO(), req)
	return err
}
func (a *GoogleSecretManager) PutMapping(p core.KeyPath, m map[string]string) error {
	return fmt.Errorf("provider %q does not implement write yet", a.Name())
}

func (a *GoogleSecretManager) GetMapping(kp core.KeyPath) ([]core.EnvEntry, error) {
	return a.getSecrets(kp)
}

func (a *GoogleSecretManager) Get(p core.KeyPath) (*core.EnvEntry, error) {
	secret, err := a.getSecret(p.Path)
	if err != nil {
		return nil, err
	}

	ent := p.Found(secret)
	return &ent, nil
}

func (a *GoogleSecretManager) Delete(kp core.KeyPath) error {
	return a.deleteSecret(kp)
}

func (a *GoogleSecretManager) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", a.Name())
}

func (a *GoogleSecretManager) getSecret(path string) (string, error) {
	r := secretmanagerpb.AccessSecretVersionRequest{
		Name: path,
	}
	a.logger.WithField("path", r.Name).Debug("get secret")

	secret, err := a.client.AccessSecretVersion(context.TODO(), &r)
	if err != nil {
		return "", err
	}
	return string(secret.Payload.Data), nil
}

func (a *GoogleSecretManager) deleteSecret(kp core.KeyPath) error {
	req := &secretmanagerpb.DestroySecretVersionRequest{
		Name: kp.Path,
	}
	_, err := a.client.DestroySecretVersion(context.TODO(), req)
	return err
}

func (a *GoogleSecretManager) getSecrets(kp core.KeyPath) ([]core.EnvEntry, error) {
	req := &secretmanagerpb.ListSecretsRequest{
		Parent: kp.Path,
	}
	entries := []core.EnvEntry{}
	it := a.client.ListSecrets(context.TODO(), req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, err
		}
		path := resp.Name + "/versions/latest"
		secret, err := a.getSecret(path)
		if err != nil {
			return nil, err
		}
		entries = append(entries, kp.FoundWithKey(strings.TrimPrefix(resp.Name, kp.Path), secret))
	}
	sort.Sort(core.EntriesByKey(entries))

	return entries, nil
}
