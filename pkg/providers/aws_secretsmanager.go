package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	smtypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/utils"
)

type AWSSecretsManagerClient interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
	CreateSecret(ctx context.Context, params *secretsmanager.CreateSecretInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.CreateSecretOutput, error)
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
	kvs, err := a.getSecret(p)
	if err != nil {
		return nil, err
	}

	var entries []core.EnvEntry
	for k, v := range kvs {
		entries = append(entries, p.FoundWithKey(k, v))
	}
	sort.Sort(core.EntriesByKey(entries))
	return entries, nil
}

func (a *AWSSecretsManager) Put(kp core.KeyPath, val string) error {
	k := kp.EffectiveKey()
	return a.PutMapping(kp, map[string]string{k: val})
}

func (a *AWSSecretsManager) PutMapping(kp core.KeyPath, m map[string]string) error {
	secrets, err := a.getSecret(kp)
	if err != nil {
		return err
	}

	secretAlreadyExist := len(secrets) != 0

	utils.Merge(m, secrets)
	secretBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}

	secretString := string(secretBytes)
	ctx := context.Background()
	if secretAlreadyExist {
		// secret already exist - put new value
		_, err = a.client.PutSecretValue(ctx, &secretsmanager.PutSecretValueInput{SecretId: &kp.Path, SecretString: &secretString})
		return err
	}

	// create secret
	_, err = a.client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{Name: &kp.Path, SecretString: &secretString})
	if err != nil {
		return err
	}

	return nil
}

func (a *AWSSecretsManager) Get(kp core.KeyPath) (*core.EnvEntry, error) {
	kvs, err := a.getSecret(kp)
	if err != nil {
		return nil, err
	}

	k := kp.EffectiveKey()
	val, ok := kvs[k]
	if !ok {
		ent := kp.Missing()
		return &ent, nil
	}

	ent := kp.Found(val)
	return &ent, nil
}

func (a *AWSSecretsManager) Delete(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", a.Name())
}

func (a *AWSSecretsManager) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", a.Name())
}

func (a *AWSSecretsManager) getSecret(kp core.KeyPath) (map[string]string, error) {
	res, err := a.client.GetSecretValue(context.Background(), &secretsmanager.GetSecretValueInput{SecretId: &kp.Path})
	if err != nil {
		var resNotFound *smtypes.ResourceNotFoundException
		if !errors.As(err, &resNotFound) {
			return nil, err
		}

		// doesn't exist - do not treat as an error
		return nil, nil
	}

	if res == nil || res.SecretString == nil {
		return nil, fmt.Errorf("data not found at %q", kp.Path)
	}

	var secret map[string]string
	err = json.Unmarshal([]byte(*res.SecretString), &secret)
	if err != nil {
		return nil, err
	}

	return secret, nil
}
