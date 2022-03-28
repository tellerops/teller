package providers

import (
	"errors"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/golang/mock/gomock"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/providers/mock_providers"
)

func TestGoogleSM(t *testing.T) {
	ctrl := gomock.NewController(t)
	// Assert that Bar() is invoked.
	defer ctrl.Finish()
	client := mock_providers.NewMockGoogleSMClient(ctrl)
	path := "settings/prod/billing-svc"
	sec := &secretmanagerpb.SecretPayload{
		Data: []byte("shazam"),
	}
	out := &secretmanagerpb.AccessSecretVersionResponse{
		Payload: sec,
	}
	in := secretmanagerpb.AccessSecretVersionRequest{
		Name: path,
	}
	client.EXPECT().AccessSecretVersion(gomock.Any(), gomock.Eq(&in)).Return(out, nil).AnyTimes()
	s := GoogleSecretManager{
		client: client,
		logger: GetTestLogger(),
	}
	AssertProvider(t, &s, false)
}

func TestGoogleSMFailures(t *testing.T) {
	ctrl := gomock.NewController(t)
	// Assert that Bar() is invoked.
	defer ctrl.Finish()
	client := mock_providers.NewMockGoogleSMClient(ctrl)
	client.EXPECT().AccessSecretVersion(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).AnyTimes()
	s := GoogleSecretManager{
		client: client,
		logger: GetTestLogger(),
	}
	_, err := s.Get(core.KeyPath{Env: "MG_KEY", Path: "settings/{{stage}}/billing-svc"})
	assert.NotNil(t, err)
}
