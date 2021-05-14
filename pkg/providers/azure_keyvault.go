package providers

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/keyvault/keyvault"
	kvauth "github.com/Azure/azure-sdk-for-go/services/keyvault/auth"
	"github.com/spectralops/teller/pkg/core"
)

const AzureVaultDomain = "vault.azure.net"

type AzureKeyVault struct {
	client    *keyvault.BaseClient
	vaultName string
}

func NewAzureKeyVault() (core.Provider, error) {
	vaultName := os.Getenv("KVAULT_NAME")
	if vaultName == "" {
		return nil, fmt.Errorf("cannot find KVAULT_NAME for azure key vault")
	}

	authorizer, err := kvauth.NewAuthorizerFromEnvironment()
	if err != nil {
		return nil, err
	}

	basicClient := keyvault.New()
	basicClient.Authorizer = authorizer
	return &AzureKeyVault{client: &basicClient, vaultName: vaultName}, nil
}

func (a *AzureKeyVault) Name() string {
	return "azure_keyvault"
}
func (a *AzureKeyVault) Put(p core.KeyPath, val string) error {
	return fmt.Errorf("%v does not implement write yet", a.Name())
}
func (a *AzureKeyVault) GetMapping(kp core.KeyPath) ([]core.EnvEntry, error) {
	r := []core.EnvEntry{}
	ctx := context.Background()
	secretList, err := a.client.GetSecrets(ctx, "https://"+a.vaultName+"."+AzureVaultDomain, nil)
	if err != nil {
		return nil, err
	}

	for secretList.NotDone() {
		for _, secret := range secretList.Values() {
			value, err := a.getSecret(core.KeyPath{Path: path.Base(*secret.ID)})
			if err != nil {
				return nil, err
			}

			r = append(r, core.EnvEntry{Key: path.Base(*secret.ID), Value: value, Provider: a.Name()})
		}
		err := secretList.NextWithContext(ctx)
		if err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (a *AzureKeyVault) Get(p core.KeyPath) (*core.EnvEntry, error) {
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

func (a *AzureKeyVault) getSecret(kp core.KeyPath) (string, error) {
	// assuming latest version
	secretResp, err := a.client.GetSecret(context.Background(), "https://"+a.vaultName+"."+AzureVaultDomain, kp.Path, "")
	if err != nil {
		return "", err
	}

	return *secretResp.Value, nil
}
