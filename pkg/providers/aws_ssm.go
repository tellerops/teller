package providers

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/spectralops/teller/pkg/core"
)

type AWSSSMClient interface {
	GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}
type AWSSSM struct {
	client AWSSSMClient
}

func NewAWSSSM() (core.Provider, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	client := ssm.NewFromConfig(cfg)

	return &AWSSSM{client: client}, nil
}

func (a *AWSSSM) Name() string {
	return "aws_ssm"
}

func (a *AWSSSM) Put(p core.KeyPath, val string) error {
	return fmt.Errorf("%v does not implement write yet", a.Name())
}

func (a *AWSSSM) GetMapping(kp core.KeyPath) ([]core.EnvEntry, error) {
	return nil, fmt.Errorf("does not support full env sync (path: %s)", kp.Path)
}

func (a *AWSSSM) Get(p core.KeyPath) (*core.EnvEntry, error) {
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

func (a *AWSSSM) getSecret(kp core.KeyPath) (string, error) {
	res, err := a.client.GetParameter(context.TODO(), &ssm.GetParameterInput{Name: &kp.Path, WithDecryption: kp.Decrypt})
	if err != nil {
		return "", err
	}

	if res == nil || res.Parameter.Value == nil {
		return "", fmt.Errorf("data not found at '%s'", kp.Path)
	}

	return *res.Parameter.Value, nil
}
