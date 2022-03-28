package providers

import (
	"testing"

	"github.com/alecthomas/assert"
	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/golang/mock/gomock"

	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/providers/mock_providers"
)

func TestCloudflareWorkersSecretsPut(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_providers.NewMockCloudflareSecretsClient(ctrl)

	expectedWorkerPutRequest := cloudflare.WorkersPutSecretRequest{
		Name: "MG_KEY",
		Text: "put-secret",
		Type: "secret_text",
	}
	client.EXPECT().SetWorkersSecret(gomock.Any(), "script-key", &expectedWorkerPutRequest).Return(cloudflare.WorkersPutSecretResponse{}, nil).AnyTimes()

	c := CloudflareSecrets{
		client: client,
		logger: GetTestLogger(),
	}
	assert.Nil(t, c.Put(core.KeyPath{Field: "MG_KEY", Source: "script-key"}, "put-secret"))
	assert.Nil(t, c.Put(core.KeyPath{Env: "MG_KEY", Source: "script-key"}, "put-secret"))
	assert.NotNil(t, c.Put(core.KeyPath{Path: "script-key"}, "put-secret"))
	assert.EqualError(t, c.Delete(core.KeyPath{Field: "MG_KEY"}), ErrCloudFlareSourceFieldIsMissing.Error())
}

func TestCloudflareWorkersSecretsDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_providers.NewMockCloudflareSecretsClient(ctrl)

	client.EXPECT().DeleteWorkersSecret(gomock.Any(), "script-key", "MG_KEY").Return(cloudflare.Response{}, nil).AnyTimes()

	c := CloudflareSecrets{
		client: client,
		logger: GetTestLogger(),
	}

	assert.Nil(t, c.Delete(core.KeyPath{Field: "MG_KEY", Source: "script-key"}))
	assert.EqualError(t, c.Delete(core.KeyPath{Field: "MG_KEY"}), ErrCloudFlareSourceFieldIsMissing.Error())
	assert.NotNil(t, c.Delete(core.KeyPath{Path: "script-key"}))

}
