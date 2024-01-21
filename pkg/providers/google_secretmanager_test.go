package providers

import (
	"errors"
	"testing"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
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
	outDelete := &secretmanagerpb.SecretVersion{
		Name: string(sec.Data),
	}
	outAdd := &secretmanagerpb.SecretVersion{
		Name: string(sec.Data),
	}
	outList := &secretmanager.SecretIterator{
		Response: string(sec.Data),
	}
	in := secretmanagerpb.AccessSecretVersionRequest{
		Name: path,
	}
	inDelete := secretmanagerpb.DestroySecretVersionRequest{
		Name: path,
	}
	inList := secretmanagerpb.ListSecretsRequest{
		Parent: path,
	}
	inAdd := secretmanagerpb.AddSecretVersionRequest{
		Parent: path,
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte("some value"),
		},
	}
	client.EXPECT().AccessSecretVersion(gomock.Any(), gomock.Eq(&in)).Return(out, nil).AnyTimes()
	client.EXPECT().DestroySecretVersion(gomock.Any(), gomock.Eq(&inDelete)).Return(outDelete, nil).AnyTimes()
	client.EXPECT().ListSecrets(gomock.Any(), gomock.Eq(&inList)).Return(outList).AnyTimes()
	client.EXPECT().AddSecretVersion(gomock.Any(), gomock.Eq(&inAdd)).Return(outAdd, nil).AnyTimes()
	s := GoogleSecretManager{
		client: client,
		logger: GetTestLogger(),
	}
	ConfigurableAssertProvider(t, &s, false, false)
}

func TestGoogleSMWithField(t *testing.T) {
	ctrl := gomock.NewController(t)
	// Assert that Bar() is invoked.
	defer ctrl.Finish()
	client := mock_providers.NewMockGoogleSMClient(ctrl)
	path := "settings/prod/billing-svc"

	data := `{"MG_KEY":"shazam", "SMTP_PASS":"mailman"}`
	sec := &secretmanagerpb.SecretPayload{
		Data: []byte(data),
	}
	out := &secretmanagerpb.AccessSecretVersionResponse{
		Payload: sec,
	}
	outDelete := &secretmanagerpb.SecretVersion{
		Name: string(sec.Data),
	}
	outAdd := &secretmanagerpb.SecretVersion{
		Name: string(sec.Data),
	}
	outList := &secretmanager.SecretIterator{
		Response: string(sec.Data),
	}
	in := secretmanagerpb.AccessSecretVersionRequest{
		Name: path,
	}
	inDelete := secretmanagerpb.DestroySecretVersionRequest{
		Name: path,
	}
	inList := secretmanagerpb.ListSecretsRequest{
		Parent: path,
	}
	inAdd := secretmanagerpb.AddSecretVersionRequest{
		Parent: path,
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte("some value"),
		},
	}
	client.EXPECT().AccessSecretVersion(gomock.Any(), gomock.Eq(&in)).Return(out, nil).AnyTimes()
	client.EXPECT().DestroySecretVersion(gomock.Any(), gomock.Eq(&inDelete)).Return(outDelete, nil).AnyTimes()
	client.EXPECT().ListSecrets(gomock.Any(), gomock.Eq(&inList)).Return(outList).AnyTimes()
	client.EXPECT().AddSecretVersion(gomock.Any(), gomock.Eq(&inAdd)).Return(outAdd, nil).AnyTimes()
	s := GoogleSecretManager{
		client: client,
		logger: GetTestLogger(),
	}
	ConfigurableAssertProvider(t, &s, false, true)
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
