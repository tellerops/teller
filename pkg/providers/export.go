package providers

import (
	"encoding/json"

	"github.com/spectralops/teller/pkg/core"
)

type TellerExport struct {
	Version   string                   `json:"version"`
	Providers map[string]core.MetaInfo `json:"providers"`
}

func GenerateProvidersMetaJSON(version string, providersMetaList []core.MetaInfo) (string, error) {
	providersMetaMap := make(map[string]core.MetaInfo)
	for _, provider := range providersMetaList {
		providersMetaMap[provider.Name] = provider
	}

	tellerObject := TellerExport{
		Version:   version,
		Providers: providersMetaMap,
	}

	jsonOutput, err := json.MarshalIndent(tellerObject, "", "  ")

	if err != nil {
		return "", err
	}

	return string(jsonOutput), nil
}
