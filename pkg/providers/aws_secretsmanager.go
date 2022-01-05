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
	PutSecretValue(ctx context.Context, params *secretsmanager.PutSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.PutSecretValueOutput, error)
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
		entries = append(entries, p.FoundWithKey(k, v))
	}
	sort.Sort(core.EntriesByKey(entries))
	return entries, nil
}

func (a *AWSSecretsManager) Put(p core.KeyPath, val string) error {
	secrets, _ := a.getSecret(p)
	k := p.EffectiveKey()
	if secrets == nil {
		secrets = map[string]string{k: val}
	} else {
		secrets[k] = val
	}
	return a.putSecret(p, map[string]string{k: val})
}

func (a *AWSSecretsManager) PutMapping(p core.KeyPath, m map[string]string) error {
	secrets, _ := a.getSecret(p)
	if secrets == nil {
		return a.putSecret(p, m)
	}

	for k, v := range m {
		secrets[k] = v
	}

	return a.putSecret(p, secrets)
}

func (a *AWSSecretsManager) Get(p core.KeyPath) (*core.EnvEntry, error) {
	secret, err := a.getSecret(p)
	if err != nil {
		return nil, err
	}

	data := secret
	k, ok := data[p.Env]
	if p.Field != "" {
		k, ok = data[p.Field]
	}

	if !ok {
		ent := p.Missing()
		return &ent, nil
	}

	ent := p.Found(k)
	return &ent, nil
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

func (a *AWSSecretsManager) putSecret(kp core.KeyPath, m map[string]string) error {
	encodedSecret, err := json.Marshal(m)
	if err != nil {
		return err
	}

	encodedSecretString := string(encodedSecret)

	_ , err = a.client.PutSecretValue(context.TODO(), &secretsmanager.PutSecretValueInput{SecretId: &kp.Path, SecretString: &encodedSecretString})
	if err != nil {
		return err
	}
	return nil
}


