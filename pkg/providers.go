package pkg

import (
	"fmt"

	"github.com/spectralops/teller/pkg/core"
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
	switch name {
	case "hashicorp_vault":
		return providers.NewHashicorpVault()
	case "aws_ssm":
		return providers.NewAWSSSM()
	case "aws_secretsmanager":
		return providers.NewAWSSecretsManager()
	case "heroku":
		return providers.NewHeroku()
	case "google_secretmanager":
		return providers.NewGoogleSecretManager()
	case "etcd":
		return providers.NewEtcd()
	case "consul":
		return providers.NewConsul()
	case "dotenv":
		return providers.NewDotenv()
	case "vercel":
		return providers.NewVercel()
	case "azure_keyvault":
		return providers.NewAzureKeyVault()
	case "doppler":
		return providers.NewDoppler()
	case "cyberark_conjur":
		return providers.NewConjurClient()
	case "cloudflare_workers_kv":
		return providers.NewCloudflareClient()
	case "cloudflare_workers_secrets":
		return providers.NewCloudflareSecretsClient()
	case "1password":
		return providers.NewOnePassword()
	case "gopass":
		return providers.NewGopass()
	case "lastpass":
		return providers.NewLastPass()
	case "github":
		return providers.NewGitHub()
	default:
		return nil, fmt.Errorf("provider '%s' does not exist", name)
	}
}
