package providers

import (
	"fmt"
	"strings"

	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
)

var providersMap = map[string]core.RegisteredProvider{}

func RegisterProvider(metaInfo core.MetaInfo, builder func(logger logging.Logger) (core.Provider, error)) {
	loweredProviderName := strings.ToLower(metaInfo.Name)
	if _, ok := providersMap[loweredProviderName]; ok {
		panic(fmt.Sprintf("provider '%s' already exists", loweredProviderName))
	}
	providersMap[loweredProviderName] = core.RegisteredProvider{Meta: metaInfo, Builder: builder}
}

func ResolveProvider(providerName string) (core.Provider, error) {
	loweredProviderName := strings.ToLower(providerName)
	if registeredProvider, ok := providersMap[loweredProviderName]; ok {
		logger := logging.GetRoot().WithField("provider_name", loweredProviderName)
		return registeredProvider.Builder(logger)
	}
	return nil, fmt.Errorf("provider '%s' does not exist", providerName)

}

func ResolveProviderMeta(providerName string) (core.MetaInfo, error) {
	loweredProviderName := strings.ToLower(providerName)
	if registeredProvider, ok := providersMap[loweredProviderName]; ok {
		return registeredProvider.Meta, nil
	}
	return core.MetaInfo{}, fmt.Errorf("provider '%s' does not exist", providerName)
}

func GetAllProvidersMeta() []core.MetaInfo {
	metaInfoList := []core.MetaInfo{}
	for _, value := range providersMap {
		metaInfoList = append(metaInfoList, value.Meta)
	}
	return metaInfoList
}
