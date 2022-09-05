package pkg

import (
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
	providersMeta := providers.GetAllProvidersMeta()
	descriptionToNameMap := make(map[string]string)
	for _, meta := range providersMeta {
		descriptionToNameMap[meta.Description] = meta.Name
	}
	return descriptionToNameMap
}

func (p *BuiltinProviders) GetProvider(name string) (core.Provider, error) {
	return providers.ResolveProvider(name)
}
