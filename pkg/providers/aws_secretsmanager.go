package providers

import (
	"context"
	"encoding/json"
	"fmt"

	"sort"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/spectralops/teller/pkg/core"
)

/*
build interface for the client,
replace with this,
in tests use literal constructor
rig the mock inside
*/
type AWSSecretsManagerClient interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}
type AWSSecretsManager struct {
	client AWSSecretsManagerClient
}

func NewAWSSecretsManager() (core.Provider, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	client := secretsmanager.NewFromConfig(cfg)

	return &AWSSecretsManager{client: client}, nil
}

func (a *AWSSecretsManager) Name() string {
	return "aws_secretsmanager"
}

func (a *AWSSecretsManager) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {
	secret, err := a.getSecret(p)
	if err != nil {
		return nil, err
	}

	k := secret

	entries := []core.EnvEntry{}
	for k, v := range k {
		entries = append(entries, core.EnvEntry{Key: k, Value: v, Provider: a.Name(), ResolvedPath: p.Path})
	}
	sort.Sort(core.EntriesByKey(entries))
	return entries, nil
}

func (a *AWSSecretsManager) Put(p core.KeyPath, val string) error {
	return fmt.Errorf("%v does not implement write yet", a.Name())
}

func (a *AWSSecretsManager) Get(p core.KeyPath) (*core.EnvEntry, error) {
	secret, err := a.getSecret(p)
	if err != nil {
		return nil, err
	}

	data := secret
	k := data[p.Env]
	if p.Field != "" {
		k = data[p.Field]
	}

	return &core.EnvEntry{
		Key:          p.Env,
		Value:        k,
		ResolvedPath: p.Path,
		Provider:     a.Name(),
	}, nil
}

func (a *AWSSecretsManager) getSecret(kp core.KeyPath) (map[string]string, error) {
	res, err := a.client.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{SecretId: &kp.Path})
	if err != nil {
		return nil, err
	}

	if res == nil || res.SecretString == nil {
		return nil, fmt.Errorf("data not found at '%s'", kp.Path)
	}
	m := map[string]string{}
	err = json.Unmarshal([]byte(*res.SecretString), &m)

	if err != nil {
		return nil, err
	}

	return m, nil
}
