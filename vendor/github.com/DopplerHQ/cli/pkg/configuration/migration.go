/*
Copyright © 2019 Doppler <support@doppler.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package configuration

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/DopplerHQ/cli/pkg/models"
	"github.com/DopplerHQ/cli/pkg/utils"
)

type oldConfig struct {
	Pipeline    string
	Environment string
	Key         string
}

var jsonFile = filepath.Join(utils.HomeDir(), ".doppler.json")

func jsonExists() bool {
	return utils.Exists(jsonFile)
}

// migrateJSONToYaml migrate ~/.doppler.json to yaml config
func migrateJSONToYaml() {
	jsonConfig := parseJSONConfig()
	newConfig := convertOldConfig(jsonConfig)
	writeConfig(newConfig)
}

func convertOldConfig(oldConfig map[string]oldConfig) models.ConfigFile {
	config := map[string]models.FileScopedOptions{}

	for key, val := range oldConfig {
		var err error
		// skip items that fail to parse
		if key, err = NormalizeScope(key); err == nil {
			config[key] = models.FileScopedOptions{EnclaveProject: val.Pipeline, EnclaveConfig: val.Environment, Token: val.Key}
		}
	}

	return models.ConfigFile{Scoped: config}
}

func parseJSONConfig() map[string]oldConfig {
	fileContents, err := ioutil.ReadFile(jsonFile) // #nosec G304
	if err != nil {
		utils.HandleError(err)
	}

	var config map[string]oldConfig
	err = json.Unmarshal(fileContents, &config)
	if err != nil {
		utils.HandleError(err)
	}

	return config
}
