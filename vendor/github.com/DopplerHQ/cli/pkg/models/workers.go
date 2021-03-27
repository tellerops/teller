/*
Copyright © 2020 Doppler <support@doppler.com>

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
package models

import (
	"encoding/json"

	"github.com/DopplerHQ/cli/pkg/utils"
)

// ChangeLog lists changes
type ChangeLog struct {
	Changes []string `json:"changes"`
}

// ParseChangeLog parse change log
func ParseChangeLog(response []byte) map[string]ChangeLog {
	var releaseMap []map[string]interface{}
	if err := json.Unmarshal(response, &releaseMap); err != nil {
		utils.HandleError(err, "Unable to parse changelog")
	}

	changes := map[string]ChangeLog{}

	for _, release := range releaseMap {
		v := release["version"].(string)
		var list []string
		for _, change := range release["changes"].([]interface{}) {
			list = append(list, change.(string))
		}

		changes[v] = ChangeLog{Changes: list}
	}

	return changes
}
