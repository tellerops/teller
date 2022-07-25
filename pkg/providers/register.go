package providers

import (
	"fmt"

	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
)

var providersMap = map[string]core.RegisteredProvider{}

func RegisterProvider(metaInfo core.MetaInfo, builder func(logger logging.Logger) (core.Provider, error)) {
	if _, ok := providersMap[metaInfo.Name]; ok {
		panic(fmt.Sprintf("provider '%s' already exists", metaInfo.Name))
	}
	providersMap[metaInfo.Name] = core.RegisteredProvider{Meta: metaInfo, Builder: builder}
}

func ResolveProvider(providerName string) (core.Provider, error) {
	if registeredProvider, ok := providersMap[providerName]; ok {
		logger := logging.GetRoot().WithField("provider_name", providerName)
		return registeredProvider.Builder(logger)
	}
	return nil, fmt.Errorf("provider '%s' does not exist", providerName)

}

func ResolveProviderMeta(providerName string) (core.MetaInfo, error) {
	if registeredProvider, ok := providersMap[providerName]; ok {
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
