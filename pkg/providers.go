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
		"Heroku":                   "heroku",
		"Vault by Hashicorp":       "hashicorp_vault",
		"AWS SSM (aka paramstore)": "aws_ssm",
		"AWS Secrets Manager":      "aws_secretsmanager",
		"Google Secret Manager":    "google_secretmanager",
		"Etcd":                     "etcd",
		"Consul":                   "consul",
		".env":                     "dotenv",
		"Vercel":                   "vercel",
		"Azure Key Vault":          "azure_keyvault",
		"Doppler":                  "doppler",
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
	default:
		return nil, fmt.Errorf("provider %s does not exist", name)
	}
}
