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
	descriptionToNameMap := make(map[string]string)
	for _, provider := range activeProviders {
		descriptionToNameMap[provider.Meta().Description] = provider.Name()
	}
	return descriptionToNameMap
}

func (p *BuiltinProviders) GetProvider(name string) (core.Provider, error) { //nolint
	providerByName := make(map[string]core.Provider)
	activeProviders := activeProviders(p)
	for _, provider := range activeProviders {
		providerByName[provider.Name()] = provider
	}
	if _, ok := providerByName[name]; ok {
		logger := logging.GetRoot().WithField("provider_name", name)
		return providerByName[name].Init(logger)
	}

	return nil, fmt.Errorf("provider '%s' does not exist", name)

}
