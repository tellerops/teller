package providers

import (
	"errors"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/golang/mock/gomock"

	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/providers/mock_providers"
)

func TestKeeperSecretsManager(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_providers.NewMockKsmClient(ctrl)

	a := KeeperSecretsManager{
		client: client,
		logger: GetTestLogger(),
	}

	path := core.KeyPath{Path: "settings/prod/billing-svc/all", Field: "MG_KEY", Decrypt: true}
	key1 := core.KeyPath{Path: "settings/prod/billing-svc", Field: "MG_KEY", Decrypt: true}
	key2 := core.KeyPath{Path: "settings/prod/billing-svc", Env: "MG_KEY", Decrypt: true}

	returnSecrets := []core.EnvEntry{
		{Key: "MM_KEY", Value: "mailman"},
		{Key: "MG_KEY", Value: "shazam"},
	}
	client.EXPECT().GetSecret(key1).Return(&core.EnvEntry{Value: "shazam"}, nil).AnyTimes()
	client.EXPECT().GetSecret(key2).Return(&core.EnvEntry{Value: "shazam"}, nil).AnyTimes()
	client.EXPECT().GetSecrets(path).Return(returnSecrets, nil).AnyTimes()
	client.EXPECT().GetSecrets(nil).Return(returnSecrets, nil).AnyTimes()

	AssertProvider(t, &a, true)
}

func TestKeeperSecretsManagerFailures(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_providers.NewMockKsmClient(ctrl)

	a := KeeperSecretsManager{
		client: client,
		logger: GetTestLogger(),
	}

	path := core.KeyPath{Path: "settings/{{stage}}/billing-svc/all"}
	key1 := core.KeyPath{Env: "MG_KEY", Path: "settings/{{stage}}/billing-svc"}
	client.EXPECT().GetSecret(key1).Return(&core.EnvEntry{}, errors.New("error")).AnyTimes()
	client.EXPECT().GetSecrets(path).Return([]core.EnvEntry{}, errors.New("error")).AnyTimes()

	_, err := a.Get(core.KeyPath{Env: "MG_KEY", Path: "settings/{{stage}}/billing-svc"})
	assert.NotNil(t, err)

	_, err = a.GetMapping(core.KeyPath{Path: "settings/{{stage}}/billing-svc/all"})
	assert.NotNil(t, err)
}
