package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
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

const GoogleSecretManagerName = "google_secretmanager"

//nolint
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
		Ops: core.OpMatrix{Get: true, Put: true, Delete: true},
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
	reg := regexp.MustCompile(`(?i)/versions/\d+$`)
	res := reg.ReplaceAllString(p.Path, "")
	return a.addSecret(res, val)
}
func (a *GoogleSecretManager) PutMapping(p core.KeyPath, m map[string]string) error {
	for k, v := range m {
		path := fmt.Sprintf("%v/secrets/%v", p.Path, k)
		err := a.addSecret(path, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *GoogleSecretManager) GetMapping(kp core.KeyPath) ([]core.EnvEntry, error) {
	secrets, err := a.getSecrets(kp.Path)
	if err != nil {
		return nil, err
	}

	entries := []core.EnvEntry{}

	for _, val := range secrets {
		path := fmt.Sprintf("%s/%s", val, "versions/latest")
		secretVal, err := a.getSecret(path)
		if err != nil {
			return nil, err
		}
		key := strings.TrimPrefix(val, kp.Path)
		entries = append(entries, kp.FoundWithKey(key, secretVal))
	}
	sort.Sort(core.EntriesByKey(entries))

	return entries, nil
}

func (a *GoogleSecretManager) Get(p core.KeyPath) (*core.EnvEntry, error) {
	secret, err := a.getSecret(p.Path)
	if err != nil {
		return nil, err
	}

	ent := p.Found(secret)

	if ent.Field != "" {
		var valueSecrets map[string]interface{}
		err = json.Unmarshal([]byte(ent.Value), &valueSecrets)
		if err != nil {
			return nil, err
		}
		fieldValue := valueSecrets[ent.Field].(string)
		ent.Value = fieldValue
	}
	return &ent, nil
}

func (a *GoogleSecretManager) Delete(kp core.KeyPath) error {
	return a.deleteSecret(kp.Path)
}

func (a *GoogleSecretManager) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", GoogleSecretManagerName)
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

func (a *GoogleSecretManager) deleteSecret(path string) error {
	req := &secretmanagerpb.DestroySecretVersionRequest{
		Name: path,
	}
	_, err := a.client.DestroySecretVersion(context.TODO(), req)
	return err
}

func (a *GoogleSecretManager) addSecret(path, val string) error {
	req := &secretmanagerpb.AddSecretVersionRequest{
		Parent: path,
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(val),
		},
	}

	_, err := a.client.AddSecretVersion(context.TODO(), req)
	return err
}

func (a *GoogleSecretManager) getSecrets(path string) ([]string, error) {
	req := &secretmanagerpb.ListSecretsRequest{
		Parent: path,
	}
	entries := []string{}

	it := a.client.ListSecrets(context.TODO(), req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, err
		}
		entries = append(entries, resp.Name)
	}
	return entries, nil
}
