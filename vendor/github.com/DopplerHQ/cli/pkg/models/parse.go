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
package models

import (
	"encoding/json"
)

// ParseWorkplaceSettings parse workplace settings
func ParseWorkplaceSettings(info map[string]interface{}) WorkplaceSettings {
	var workplaceInfo WorkplaceSettings

	if info["id"] != nil {
		workplaceInfo.ID = info["id"].(string)
	}
	if info["name"] != nil {
		workplaceInfo.Name = info["name"].(string)
	}
	if info["billing_email"] != nil {
		workplaceInfo.BillingEmail = info["billing_email"].(string)
	}

	return workplaceInfo
}

// ParseProjectInfo parse project info
func ParseProjectInfo(info map[string]interface{}) ProjectInfo {
	var projectInfo ProjectInfo

	if info["id"] != nil {
		projectInfo.ID = info["id"].(string)
	}
	if info["name"] != nil {
		projectInfo.Name = info["name"].(string)
	}
	if info["description"] != nil {
		projectInfo.Description = info["description"].(string)
	}
	if info["created_at"] != nil {
		projectInfo.CreatedAt = info["created_at"].(string)
	}

	return projectInfo
}

// ParseEnvironmentInfo parse environment info
func ParseEnvironmentInfo(info map[string]interface{}) EnvironmentInfo {
	var environmentInfo EnvironmentInfo

	if info["id"] != nil {
		environmentInfo.ID = info["id"].(string)
	}
	if info["name"] != nil {
		environmentInfo.Name = info["name"].(string)
	}
	if info["created_at"] != nil {
		environmentInfo.CreatedAt = info["created_at"].(string)
	}
	if info["initial_fetch_at"] != nil {
		environmentInfo.InitialFetchAt = info["initial_fetch_at"].(string)
	}
	if info["project"] != nil {
		environmentInfo.Project = info["project"].(string)
	}

	return environmentInfo
}

// ParseConfigInfo parse config info
func ParseConfigInfo(info map[string]interface{}) ConfigInfo {
	var configInfo ConfigInfo

	if info["name"] != nil {
		configInfo.Name = info["name"].(string)
	}
	if info["root"] != nil {
		configInfo.Root = info["root"].(bool)
	}
	if info["locked"] != nil {
		configInfo.Locked = info["locked"].(bool)
	}
	if info["environment"] != nil {
		configInfo.Environment = info["environment"].(string)
	}
	if info["project"] != nil {
		configInfo.Project = info["project"].(string)
	}
	if info["created_at"] != nil {
		configInfo.CreatedAt = info["created_at"].(string)
	}
	if info["initial_fetch_at"] != nil {
		configInfo.InitialFetchAt = info["initial_fetch_at"].(string)
	}
	if info["last_fetch_at"] != nil {
		configInfo.LastFetchAt = info["last_fetch_at"].(string)
	}

	return configInfo
}

// ParseConfigLog parse config log
func ParseConfigLog(log map[string]interface{}) ConfigLog {
	var parsedLog ConfigLog

	if log["id"] != nil {
		parsedLog.ID = log["id"].(string)
	}
	if log["text"] != nil {
		parsedLog.Text = log["text"].(string)
	}
	if log["html"] != nil {
		parsedLog.HTML = log["html"].(string)
	}
	if log["created_at"] != nil {
		parsedLog.CreatedAt = log["created_at"].(string)
	}
	if log["config"] != nil {
		parsedLog.Config = log["config"].(string)
	}
	if log["environment"] != nil {
		parsedLog.Environment = log["environment"].(string)
	}
	if log["project"] != nil {
		parsedLog.Project = log["project"].(string)
	}
	if log["user"] != nil {
		user := log["user"].(map[string]interface{})
		parsedLog.User.Email = user["email"].(string)
		parsedLog.User.Name = user["name"].(string)
		parsedLog.User.Username = user["username"].(string)
		parsedLog.User.ProfileImage = user["profile_image_url"].(string)
	}
	if log["diff"] != nil {
		for _, diff := range log["diff"].([]interface{}) {
			diffMap := diff.(map[string]interface{})
			d := LogDiff{}
			if diffMap["name"] != nil {
				d.Name = diffMap["name"].(string)
			}
			if diffMap["added"] != nil {
				d.Added = diffMap["added"].(string)
			}
			if diffMap["removed"] != nil {
				d.Removed = diffMap["removed"].(string)
			}
			parsedLog.Diff = append(parsedLog.Diff, d)
		}
	}

	return parsedLog
}

// ParseActivityLog parse activity log
func ParseActivityLog(log map[string]interface{}) ActivityLog {
	var parsedLog ActivityLog

	if log["id"] != nil {
		parsedLog.ID = log["id"].(string)
	}
	if log["text"] != nil {
		parsedLog.Text = log["text"].(string)
	}
	if log["html"] != nil {
		parsedLog.HTML = log["html"].(string)
	}
	if log["created_at"] != nil {
		parsedLog.CreatedAt = log["created_at"].(string)
	}
	if log["enclave_config"] != nil {
		parsedLog.EnclaveConfig = log["enclave_config"].(string)
	}
	if log["enclave_environment"] != nil {
		parsedLog.EnclaveEnvironment = log["enclave_environment"].(string)
	}
	if log["enclave_project"] != nil {
		parsedLog.EnclaveProject = log["enclave_project"].(string)
	}
	if log["user"] != nil {
		user := log["user"].(map[string]interface{})
		if user["email"] != nil {
			parsedLog.User.Email = user["email"].(string)
		}
		if user["name"] != nil {
			parsedLog.User.Name = user["name"].(string)
		}
		if user["username"] != nil {
			parsedLog.User.Username = user["username"].(string)
		}
		if user["profile_image_url"] != nil {
			parsedLog.User.ProfileImage = user["profile_image_url"].(string)
		}
	}

	return parsedLog
}

// ParseSecrets parse secrets
func ParseSecrets(response []byte) (map[string]ComputedSecret, error) {
	var result map[string]interface{}
	err := json.Unmarshal(response, &result)
	if err != nil {
		return nil, err
	}

	computed := map[string]ComputedSecret{}
	secrets := result["secrets"].(map[string]interface{})
	for key, secret := range secrets {
		computedSecret := ComputedSecret{Name: key}
		val := secret.(map[string]interface{})
		if val["raw"] != nil {
			computedSecret.RawValue = val["raw"].(string)
		}
		if val["computed"] != nil {
			computedSecret.ComputedValue = val["computed"].(string)
		}
		computed[key] = computedSecret
	}

	return computed, nil
}

// ParseConfigServiceToken parse config service token
func ParseConfigServiceToken(token map[string]interface{}) ConfigServiceToken {
	var parsedToken ConfigServiceToken

	if token["name"] != nil {
		parsedToken.Name = token["name"].(string)
	}
	if token["key"] != nil {
		parsedToken.Token = token["key"].(string)
	}
	if token["slug"] != nil {
		parsedToken.Slug = token["slug"].(string)
	}
	if token["project"] != nil {
		parsedToken.Project = token["project"].(string)
	}
	if token["environment"] != nil {
		parsedToken.Environment = token["environment"].(string)
	}
	if token["config"] != nil {
		parsedToken.Config = token["config"].(string)
	}
	if token["created_at"] != nil {
		parsedToken.CreatedAt = token["created_at"].(string)
	}

	return parsedToken
}
