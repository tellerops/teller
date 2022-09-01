package providers

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/spectralops/teller/pkg/core"
)

func TestGenerateProvidersMetaJSON(t *testing.T) {
	var providersData = []core.MetaInfo{
		{
			Name:           "Provider_1",
			Description:    "Description of Provider 1",
			Authentication: "Provider 1 authentication instructions",
			ConfigTemplate: "Provider 1 config template",
			Ops:            core.OpMatrix{Get: true, GetMapping: true, Put: true, PutMapping: true},
		},
		{
			Name:           "Provider_2",
			Description:    "Description of Provider 2",
			Authentication: "Provider 2 authentication instructions",
			ConfigTemplate: "Provider 2 config template",
			Ops:            core.OpMatrix{Get: true, GetMapping: true, Put: true, PutMapping: true},
		},
	}

	providersMetaJSON, _ := GenerateProvidersMetaJSON("1.1", providersData)
	providersFileContent, _ := os.ReadFile("../../fixtures/providers-export/providers-meta.json")

	actualProvidersJSON, _ := json.Marshal(providersMetaJSON)
	expectedProvidersJSON, _ := json.Marshal(string(providersFileContent))

	assert.Equal(t, string(actualProvidersJSON), string(expectedProvidersJSON))
}
