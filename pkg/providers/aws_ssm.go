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
	PutParameter(ctx context.Context, params *ssm.PutParameterInput, optFns ...func(*ssm.Options)) (*ssm.PutParameterOutput, error)
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
	return a.putSecret(p, &val)
}
func (a *AWSSSM) PutMapping(p core.KeyPath, m map[string]string) error {
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

	if secret == nil {
		ent := p.Missing()
		return &ent, nil
	}

	ent := p.Found(*secret)
	return &ent, nil
}

func (a *AWSSSM) getSecret(kp core.KeyPath) (*string, error) {
	res, err := a.client.GetParameter(context.TODO(), &ssm.GetParameterInput{Name: &kp.Path, WithDecryption: kp.Decrypt})
	if err != nil {
		return nil, err
	}

	if res == nil || res.Parameter.Value == nil {
		return nil, nil
	}

	return res.Parameter.Value, nil
}

func (a *AWSSSM) putSecret(kp core.KeyPath, val *string) error {
	_, err := a.client.PutParameter(context.TODO(), &ssm.PutParameterInput{Name: &kp.Path, Value: val})
	return err
}