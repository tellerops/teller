package providers

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
)

type AWSSSMClient interface {
	GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}
type AWSSSM struct {
	client AWSSSMClient
	logger logging.Logger
}

func NewAWSSSM(logger logging.Logger) (core.Provider, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	client := ssm.NewFromConfig(cfg)

	return &AWSSSM{client: client, logger: logger}, nil
}

func (a *AWSSSM) Name() string {
	return "aws_ssm"
}

func (a *AWSSSM) Put(p core.KeyPath, val string) error {
	return fmt.Errorf("provider %q does not implement write yet", a.Name())
}
func (a *AWSSSM) PutMapping(p core.KeyPath, m map[string]string) error {
	return fmt.Errorf("provider %q does not implement write yet", a.Name())
}

func (a *AWSSSM) GetMapping(kp core.KeyPath) ([]core.EnvEntry, error) {
	return nil, fmt.Errorf("does not support full env sync (path: %s)", kp.Path)
}

func (a *AWSSSM) Delete(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", a.Name())
}

func (a *AWSSSM) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", a.Name())
}

func (a *AWSSSM) Get(p core.KeyPath) (*core.EnvEntry, error) {
	secret, err := a.getSecret(p)
	if err != nil {
		return nil, err
	}

	if secret == nil {
		a.logger.WithField("path", p.Path).Debug("secret is empty")
		ent := p.Missing()
		return &ent, nil
	}

	ent := p.Found(*secret)
	return &ent, nil
}

func (a *AWSSSM) getSecret(kp core.KeyPath) (*string, error) {
	a.logger.WithField("path", kp.Path).Debug("get entry")
	res, err := a.client.GetParameter(context.TODO(), &ssm.GetParameterInput{Name: &kp.Path, WithDecryption: kp.Decrypt})
	if err != nil {
		return nil, err
	}

	if res == nil || res.Parameter.Value == nil {
		return nil, nil
	}

	return res.Parameter.Value, nil
}
