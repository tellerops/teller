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

func (p *BuiltinProviders) ProviderHumanToMachine() map[string]string {
	return map[string]string{
		"Heroku":                      "heroku",
		"Vault by Hashicorp":          "hashicorp_vault",
		"AWS SSM (aka paramstore)":    "aws_ssm",
		"AWS Secrets Manager":         "aws_secretsmanager",
		"Google Secret Manager":       "google_secretmanager",
		"Etcd":                        "etcd",
		"Consul":                      "consul",
		".env":                        "dotenv",
		"Vercel":                      "vercel",
		"Azure Key Vault":             "azure_keyvault",
		"Doppler":                     "doppler",
		"CyberArk Conjur":             "cyberark_conjur",
		"Cloudlflare Workers KV":      "cloudflare_workers_kv",
		"Cloudlflare Workers Secrets": "cloudflare_workers_secrets",
		"1Password":                   "1password",
		"Gopass":                      "gopass",
		"LastPass":                    "lastpass",
		"GitHub":                      "github",
	}
}

func (p *BuiltinProviders) GetProvider(name string) (core.Provider, error) {
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
	default:
		return nil, fmt.Errorf("provider '%s' does not exist", name)
	}
}
