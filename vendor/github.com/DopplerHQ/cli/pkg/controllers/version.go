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
package controllers

import (
	"time"

	"github.com/DopplerHQ/cli/pkg/http"
	"github.com/DopplerHQ/cli/pkg/models"
	"github.com/DopplerHQ/cli/pkg/utils"
	"github.com/DopplerHQ/cli/pkg/version"
)

// NewVersionAvailable checks whether a CLI version is available that's newer than this CLI
func NewVersionAvailable(prevVersionCheck models.VersionCheck) (bool, models.VersionCheck, error) {
	now := time.Now()
	check, err := http.GetLatestCLIVersion()
	if err != nil {
		utils.LogDebug("Unable to fetch latest CLI version")
		utils.LogDebugError(err)
		return false, models.VersionCheck{}, err
	}

	versionCheck := models.VersionCheck{CheckedAt: now, LatestVersion: version.Normalize(check.LatestVersion)}

	// skip if available version is unchanged from previous check
	if versionCheck.LatestVersion == prevVersionCheck.LatestVersion {
		utils.LogDebug("Previous version check is still latest version")
		return false, versionCheck, nil
	}

	newVersion, err := version.ParseVersion(versionCheck.LatestVersion)
	if err != nil {
		utils.LogDebug("Unable to parse new CLI version")
		return false, models.VersionCheck{}, err
	}

	currentVersion, err := version.ParseVersion(version.ProgramVersion)
	if err != nil {
		// if current version can't be parsed, consider an update available
		utils.LogDebug("Unable to parse current CLI version")
		utils.LogDebugError(err)
		return true, versionCheck, nil
	}

	compare := version.CompareVersions(currentVersion, newVersion)
	if compare == 1 {
		return true, versionCheck, nil
	}

	return false, versionCheck, nil
}
