package providers

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/alecthomas/assert"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/golang/mock/gomock"

	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/providers/mock_providers"
)

func TestAWSSSM(t *testing.T) {
	ctrl := gomock.NewController(t)
	// Assert that Bar() is invoked.
	defer ctrl.Finish()
	client := mock_providers.NewMockAWSSSMClient(ctrl)
	path := "settings/prod/billing-svc"
	val := "shazam"

	in := ssm.GetParameterInput{Name: &path, WithDecryption: aws.Bool(true)}
	out := ssm.GetParameterOutput{
		Parameter: &types.Parameter{
			Value: &val,
		},
	}
	client.EXPECT().GetParameter(gomock.Any(), gomock.Eq(&in)).Return(&out, nil).AnyTimes()
	s := AWSSSM{
		client: client,
		logger: GetTestLogger(),
	}
	AssertProvider(t, &s, false)
}

func TestAWSSSMFailures(t *testing.T) {
	ctrl := gomock.NewController(t)
	// Assert that Bar() is invoked.
	defer ctrl.Finish()
	client := mock_providers.NewMockAWSSSMClient(ctrl)
	client.EXPECT().GetParameter(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).AnyTimes()
	s := AWSSSM{
		client: client,
		logger: GetTestLogger(),
	}
	_, err := s.Get(core.KeyPath{Env: "MG_KEY", Path: "settings/{{stage}}/billing-svc"})
	assert.NotNil(t, err)
}
