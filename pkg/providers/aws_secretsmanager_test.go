package providers

import (
	"errors"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/golang/mock/gomock"

	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/providers/mock_providers"
)

func TestAWSSecretsManager(t *testing.T) {
	ctrl := gomock.NewController(t)
	// Assert that Bar() is invoked.
	defer ctrl.Finish()
	client := mock_providers.NewMockAWSSecretsManagerClient(ctrl)
	path := "settings/prod/billing-svc"
	pathmap := "settings/prod/billing-svc/all"
	in := secretsmanager.GetSecretValueInput{SecretId: &path}
	inmap := secretsmanager.GetSecretValueInput{SecretId: &pathmap}
	data := `{"MG_KEY":"shazam", "SMTP_PASS":"mailman"}`
	out := secretsmanager.GetSecretValueOutput{
		SecretString: &data,
	}
	client.EXPECT().GetSecretValue(gomock.Any(), gomock.Eq(&in)).Return(&out, nil).AnyTimes()
	client.EXPECT().GetSecretValue(gomock.Any(), gomock.Eq(&inmap)).Return(&out, nil).AnyTimes()
	s := AWSSecretsManager{
		client: client,
		logger: GetTestLogger(),
	}
	AssertProvider(t, &s, true)
}

func TestAWSSecretsManagerFailures(t *testing.T) {
	ctrl := gomock.NewController(t)
	// Assert that Bar() is invoked.
	defer ctrl.Finish()
	client := mock_providers.NewMockAWSSecretsManagerClient(ctrl)
	client.EXPECT().GetSecretValue(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).AnyTimes()
	s := AWSSecretsManager{
		client: client,
		logger: GetTestLogger(),
	}
	_, err := s.Get(core.KeyPath{Env: "MG_KEY", Path: "settings/{{stage}}/billing-svc"})
	assert.NotNil(t, err)
}
