package providers

import (
	"context"
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/keyvault/keyvault"
	"github.com/alecthomas/assert"
	"github.com/golang/mock/gomock"

	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/providers/mock_providers"
)

func String(v string) *string { return &v }

func TestAzureKeyVault(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_providers.NewMockAzureKeyVaultClient(ctrl)

	a := AzureKeyVault{
		client:       client,
		logger:       GetTestLogger(),
		vaultName:    "test",
		vaultBaseURL: "https://test/",
	}

	path := "settings/prod/billing-svc"
	shazam := "shazam"

	secretList := keyvault.SecretListResult{
		Value: &[]keyvault.SecretItem{
			{ID: String("all")},
			{ID: String("all-1")},
		},
	}

	stopNext := func(context.Context, keyvault.SecretListResult) (keyvault.SecretListResult, error) {
		return keyvault.SecretListResult{}, nil
	}
	returnSecrets := keyvault.NewSecretListResultPage(secretList, stopNext)
	client.EXPECT().GetSecret(gomock.Any(), a.vaultBaseURL, path, "").Return(keyvault.SecretBundle{Value: &shazam}, nil).AnyTimes()
	client.EXPECT().GetSecret(gomock.Any(), a.vaultBaseURL, "all", "").Return(keyvault.SecretBundle{Value: String("mailman")}, nil).AnyTimes()
	client.EXPECT().GetSecret(gomock.Any(), a.vaultBaseURL, "all-1", "").Return(keyvault.SecretBundle{Value: String("shazam")}, nil).AnyTimes()
	client.EXPECT().GetSecrets(gomock.Any(), a.vaultBaseURL, nil).Return(returnSecrets, nil).AnyTimes()

	AssertProvider(t, &a, true)

}

func TestAzureKeyVaultFailures(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_providers.NewMockAzureKeyVaultClient(ctrl)
	a := AzureKeyVault{
		client:       client,
		logger:       GetTestLogger(),
		vaultName:    "test",
		vaultBaseURL: "https://test/",
	}

	client.EXPECT().GetSecret(gomock.Any(), a.vaultBaseURL, "settings/{{stage}}/billing-svc", "").Return(keyvault.SecretBundle{}, errors.New("error")).AnyTimes()

	_, err := a.Get(core.KeyPath{Env: "MG_KEY", Path: "settings/{{stage}}/billing-svc"})
	assert.NotNil(t, err)
}
