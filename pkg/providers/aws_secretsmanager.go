package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	smtypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
	"github.com/spectralops/teller/pkg/utils"
)

type AWSSecretsManagerClient interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
	CreateSecret(ctx context.Context, params *secretsmanager.CreateSecretInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.CreateSecretOutput, error)
	PutSecretValue(ctx context.Context, params *secretsmanager.PutSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.PutSecretValueOutput, error)
	DescribeSecret(ctx context.Context, params *secretsmanager.DescribeSecretInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.DescribeSecretOutput, error)
	DeleteSecret(ctx context.Context, params *secretsmanager.DeleteSecretInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.DeleteSecretOutput, error)
}

type AWSSecretsManager struct {
	client                                    AWSSecretsManagerClient
	logger                                    logging.Logger
	deletionDisableRecoveryWindow             bool
	treatSecretMarkedForDeletionAsNonExisting bool
	deletionRecoveryWindowInDays              int64
}

var defaultDeletionRecoveryWindowInDays int64 = 7

const versionSplit = ","

//nolint
func init() {
	metaInfo := core.MetaInfo{
		Name:           "aws_secretsmanager",
		Description:    "AWS Secrets Manager",
		Authentication: "Your standard `AWS_DEFAULT_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` need to be populated in your environment. `AWS_ENDPOINT` is used to allow usage of localstack",
		ConfigTemplate: `
  # configure only from environment
  aws_secretsmanager:
    env_sync:
      path: prod/foo/bar
    env:
      FOO_BAR:
        path: prod/foo/bar
        field: SOME_KEY
`,
		Ops: core.OpMatrix{Get: true, GetMapping: true, Put: true, PutMapping: true, Delete: true, DeleteMapping: true},
	}
	RegisterProvider(metaInfo, NewAWSSecretsManager)
}

func NewAWSSecretsManager(logger logging.Logger) (core.Provider, error) {
	customResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		awsEndpointOverride := os.Getenv("AWS_ENDPOINT")
		if awsEndpointOverride != "" {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           awsEndpointOverride,
				SigningRegion: region,
			}, nil
		}

		// returning EndpointNotFoundError will allow the service to fallback to its default resolution
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithEndpointResolver(customResolver))
	if err != nil {
		return nil, err
	}

	client := secretsmanager.NewFromConfig(cfg)

	return &AWSSecretsManager{
		client:                        client,
		logger:                        logger,
		deletionRecoveryWindowInDays:  defaultDeletionRecoveryWindowInDays,
		deletionDisableRecoveryWindow: false,
		treatSecretMarkedForDeletionAsNonExisting: true,
	}, nil
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

	secrets = utils.Merge(secrets, m)
	secretBytes, err := json.Marshal(secrets)
	if err != nil {
		return err
	}

	secretString := string(secretBytes)
	ctx := context.Background()
	if secretAlreadyExist {
		// secret already exist - put new value
		a.logger.WithField("path", kp.Path).Debug("secret already exists, update the existing one")
		_, err = a.client.PutSecretValue(ctx, &secretsmanager.PutSecretValueInput{SecretId: &kp.Path, SecretString: &secretString})
		return err
	}

	// create secret
	a.logger.WithField("path", kp.Path).Debug("create secret")
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
		a.logger.WithField("key", k).Debug("key not found in kvs secrets")
		ent := kp.Missing()
		return &ent, nil
	}

	ent := kp.Found(val)
	return &ent, nil
}

func (a *AWSSecretsManager) Delete(kp core.KeyPath) error {
	kvs, err := a.getSecret(kp)
	if err != nil {
		return err
	}

	k := kp.EffectiveKey()
	delete(kvs, k)

	if len(kvs) == 0 {
		return a.DeleteMapping(kp)
	}

	secretBytes, err := json.Marshal(kvs)
	if err != nil {
		return err
	}

	secretString := string(secretBytes)
	ctx := context.Background()
	a.logger.WithField("path", kp.Path).Debug("put secret value")
	_, err = a.client.PutSecretValue(ctx, &secretsmanager.PutSecretValueInput{SecretId: &kp.Path, SecretString: &secretString})
	return err
}

func (a *AWSSecretsManager) DeleteMapping(kp core.KeyPath) error {
	kvs, err := a.getSecret(kp)
	if err != nil {
		return err
	}

	if kvs == nil {
		// already deleted
		a.logger.WithField("path", kp.Path).Debug("already deleted")
		return nil
	}

	ctx := context.Background()
	a.logger.WithField("path", kp.Path).Debug("delete secret")
	_, err = a.client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
		SecretId:                   &kp.Path,
		RecoveryWindowInDays:       &a.deletionRecoveryWindowInDays,
		ForceDeleteWithoutRecovery: &a.deletionDisableRecoveryWindow,
	})

	return err
}

func (a *AWSSecretsManager) getSecret(kp core.KeyPath) (map[string]string, error) {
	a.logger.WithField("path", kp.Path).Debug("get secret value")
	valueInput := secretsmanager.GetSecretValueInput{SecretId: &kp.Path}

	splitVersion := strings.Split(kp.Path, versionSplit)
	//nolint:gomnd
	if len(splitVersion) == 2 {
		a.logger.WithFields(map[string]interface{}{
			"path":    splitVersion[0],
			"version": splitVersion[1],
		}).Debug("add version")
		valueInput.SecretId = &splitVersion[0]
		valueInput.VersionId = &splitVersion[1]
	}

	res, err := a.client.GetSecretValue(context.Background(), &valueInput)

	var (
		resNotFoundErr *smtypes.ResourceNotFoundException
		invalidReqErr  *smtypes.InvalidRequestException
	)

	switch {
	case err == nil:
		if res == nil || res.SecretString == nil {
			return nil, fmt.Errorf("data not found at %q", kp.Path)
		}

		var secret map[string]interface{}
		err = json.Unmarshal([]byte(*res.SecretString), &secret)
		if err != nil {
			return nil, err
		}

		stringParse := map[string]string{}
		for k, v := range secret {
			stringParse[k] = fmt.Sprintf("%v", v)
		}
		return stringParse, nil
	case errors.As(err, &resNotFoundErr):
		// doesn't exist - do not treat as an error
		return nil, nil
	case a.treatSecretMarkedForDeletionAsNonExisting && errors.As(err, &invalidReqErr):
		// see whether it is marked for deletion
		markedForDeletion, markedForDeletionErr := a.isSecretMarkedForDeletion(kp)
		if err != nil {
			return nil, markedForDeletionErr
		}

		if markedForDeletion {
			// doesn't exist anymore - do not treat as an error
			return nil, nil
		}

		return nil, nil
	}

	return nil, err

}

func (a *AWSSecretsManager) isSecretMarkedForDeletion(kp core.KeyPath) (bool, error) {
	data, err := a.client.DescribeSecret(context.Background(), &secretsmanager.DescribeSecretInput{SecretId: &kp.Path})
	if err != nil {
		return false, err
	}

	return data.DeletedDate != nil, nil
}
