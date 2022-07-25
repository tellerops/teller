package providers

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/keyvault/keyvault"
	kvauth "github.com/Azure/azure-sdk-for-go/services/keyvault/auth"
	"github.com/Azure/go-autorest/autorest"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
)

const AzureVaultDomain = "vault.azure.net"

type AzureKeyVaultClient interface {
	SetSecret(ctx context.Context, vaultBaseURL string, secretName string, parameters keyvault.SecretSetParameters) (result keyvault.SecretBundle, err error)
	GetSecret(ctx context.Context, vaultBaseURL string, secretName string, secretVersion string) (result keyvault.SecretBundle, err error)
	GetSecrets(ctx context.Context, vaultBaseURL string, maxresults *int32) (result keyvault.SecretListResultPage, err error)
	DeleteSecret(ctx context.Context, vaultBaseURL string, secretName string) (result keyvault.DeletedSecretBundle, err error)
}

type AzureKeyVault struct {
	client       AzureKeyVaultClient
	logger       logging.Logger
	vaultName    string
	vaultBaseURL string
}

const azureName = "azure_keyvault"

//nolint
func init() {
	metaInfo := core.MetaInfo{
		Description:    "Azure Key Vault",
		Name:           azureName,
		Authentication: "TODO(XXX)",
		ConfigTemplate: `
  # you can mix and match many files
  azure_keyvault:
    env_sync:
      path: azure
    env:
      FOO_BAR:
        path: foobar
		`,
		Ops: core.OpMatrix{Get: true, GetMapping: true, Put: true, PutMapping: true, Delete: true},
	}
	RegisterProvider(metaInfo, NewAzureKeyVault)
}

func NewAzureKeyVault(logger logging.Logger) (core.Provider, error) {
	vaultName := os.Getenv("KVAULT_NAME")
	if vaultName == "" {
		return nil, fmt.Errorf("cannot find KVAULT_NAME for azure key vault")
	}

	var authorizer autorest.Authorizer
	var err error

	if _, ok := os.LookupEnv("AZURE_CLI"); ok {
		authorizer, err = kvauth.NewAuthorizerFromCLI()
	} else {
		authorizer, err = kvauth.NewAuthorizerFromEnvironment()
	}

	if err != nil {
		return nil, err
	}

	basicClient := keyvault.New()
	basicClient.Authorizer = authorizer
	return &AzureKeyVault{client: &basicClient,
		vaultName:    vaultName,
		logger:       logger,
		vaultBaseURL: "https://" + vaultName + "." + AzureVaultDomain,
	}, nil
}

func (a *AzureKeyVault) Name() string {
	return "azure_keyvault"
}

func (a *AzureKeyVault) Put(p core.KeyPath, val string) error {
	a.logger.WithField("path", p.Path).Debug("set secret")
	_, err := a.client.SetSecret(context.TODO(), a.vaultBaseURL, p.Path, keyvault.SecretSetParameters{
		Value: &val,
	})
	return err
}

func (a *AzureKeyVault) PutMapping(p core.KeyPath, m map[string]string) error {
	for k, v := range m {
		ap := p.SwitchPath(k)
		err := a.Put(ap, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *AzureKeyVault) GetMapping(kp core.KeyPath) ([]core.EnvEntry, error) {
	r := []core.EnvEntry{}
	ctx := context.Background()
	a.logger.WithField("vault_base_url", a.vaultBaseURL).Debug("get secrets")
	secretList, err := a.client.GetSecrets(ctx, a.vaultBaseURL, nil)
	if err != nil {
		return nil, err
	}

	for secretList.NotDone() {
		for _, secret := range secretList.Values() {
			value, err := a.getSecret(core.KeyPath{Path: path.Base(*secret.ID)})
			if err != nil {
				return nil, err
			}
			if value.Value != nil {
				ent := kp.FoundWithKey(path.Base(*secret.ID), *value.Value)
				r = append(r, ent)
			}
		}

		err := secretList.NextWithContext(ctx)
		if err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (a *AzureKeyVault) Get(p core.KeyPath) (*core.EnvEntry, error) {
	secretResp, err := a.getSecret(p)
	if err != nil {
		return nil, err
	}
	if secretResp.Value == nil {
		a.logger.WithField("path", p.Path).Debug("secret is empty")
		ent := p.Missing()
		return &ent, nil
	}

	ent := p.Found(*secretResp.Value)
	return &ent, nil
}

func (a *AzureKeyVault) Delete(kp core.KeyPath) error {
	_, err := a.client.DeleteSecret(context.TODO(), a.vaultBaseURL, kp.Path)
	return err
}

func (a *AzureKeyVault) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", azureName)
}

func (a *AzureKeyVault) getSecret(kp core.KeyPath) (keyvault.SecretBundle, error) {
	a.logger.WithFields(map[string]interface{}{
		"vault_base_url": a.vaultBaseURL,
		"secret_name":    kp.Path,
	}).Debug("get secret")
	return a.client.GetSecret(context.Background(), a.vaultBaseURL, kp.Path, "")
}
