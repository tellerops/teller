package pkg

import (
	"fmt"

	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
	"github.com/spectralops/teller/pkg/providers"
)

type Providers interface {
	GetProvider(name string) (core.Provider, error)
	ProviderHumanToMachine() map[string]string
}

type BuiltinProviders struct {
}

func activeProviders(p *BuiltinProviders) []core.Provider {
	providers := []core.Provider{
		&providers.Heroku{},
		&providers.HashicorpVault{},
		&providers.AWSSSM{},
		&providers.AWSSecretsManager{},
		&providers.GoogleSecretManager{},
		&providers.Etcd{},
		&providers.Consul{},
		&providers.Dotenv{},
		&providers.Vercel{},
		&providers.AzureKeyVault{},
		&providers.Doppler{},
		&providers.CyberArkConjur{},
		&providers.Cloudflare{},
		&providers.CloudflareSecrets{},
		&providers.OnePassword{},
		&providers.Gopass{},
		&providers.LastPass{},
		&providers.GitHub{},
		&providers.KeyPass{},
		&providers.FileSystem{},
	}

	return providers
}

func (p *BuiltinProviders) ProviderHumanToMachine() map[string]string {
	activeProviders := activeProviders(p)
	m := make(map[string]string)
	for _, provider := range activeProviders {
		m[provider.Meta().Description] = provider.Name()
	}
	return m
}

func (p *BuiltinProviders) GetProvider(name string) (core.Provider, error) { //nolint
	logger := logging.GetRoot().WithField("provider_name", name)
	switch name {
	case "hashicorp_vault":
		return providers.NewHashicorpVault(logger)
	case "aws_ssm":
		return providers.NewAWSSSM(logger)
	case "aws_secretsmanager":
		return providers.NewAWSSecretsManager(logger)
	case "heroku":
		return providers.NewHeroku(logger)
	case "google_secretmanager":
		return providers.NewGoogleSecretManager(logger)
	case "etcd":
		return providers.NewEtcd(logger)
	case "consul":
		return providers.NewConsul(logger)
	case "dotenv":
		return providers.NewDotenv(logger)
	case "vercel":
		return providers.NewVercel(logger)
	case "azure_keyvault":
		return providers.NewAzureKeyVault(logger)
	case "doppler":
		return providers.NewDoppler(logger)
	case "cyberark_conjur":
		return providers.NewConjurClient(logger)
	case "cloudflare_workers_kv":
		return providers.NewCloudflareClient(logger)
	case "cloudflare_workers_secrets":
		return providers.NewCloudflareSecretsClient(logger)
	case "1password":
		return providers.NewOnePassword(logger)
	case "gopass":
		return providers.NewGopass(logger)
	case "lastpass":
		return providers.NewLastPass(logger)
	case "github":
		return providers.NewGitHub(logger)
	case "keypass":
		return providers.NewKeyPass(logger)
	case "filesystem":
		return providers.NewFileSystem(logger)
	default:
		return nil, fmt.Errorf("provider '%s' does not exist", name)
	}
}
