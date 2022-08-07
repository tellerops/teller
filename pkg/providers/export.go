package providers

import (
	"encoding/json"

	"github.com/spectralops/teller/pkg/core"
)

type TellerExport struct {
	Version string `json:"version"`
}

type ProvidersMetaRoot struct {
	Teller    TellerExport    `json:"teller"`
	Providers []core.MetaInfo `json:"providers"`
}

func GetProvidersMetaJSON(version string) (string, error) {
	providersMetaList := GetAllProvidersMeta()

	tellerObject := TellerExport{
		Version: version,
	}

	result := ProvidersMetaRoot{
		Teller:    tellerObject,
		Providers: providersMetaList,
	}

	jsonOutput, err := json.MarshalIndent(result, "", "  ")

	if err != nil {
		return "", err
	}

	return string(jsonOutput), nil
}
