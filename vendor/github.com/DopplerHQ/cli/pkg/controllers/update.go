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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/DopplerHQ/cli/pkg/http"
	"github.com/DopplerHQ/cli/pkg/models"
	"github.com/DopplerHQ/cli/pkg/utils"
	"github.com/DopplerHQ/cli/pkg/version"
)

// Error controller errors
type Error struct {
	Err     error
	Message string
}

// Unwrap get the original error
func (e *Error) Unwrap() error { return e.Err }

// IsNil whether the error is nil
func (e *Error) IsNil() bool { return e.Err == nil && e.Message == "" }

// RunInstallScript downloads and executes the CLI install scriptm, returning true if an update was installed
func RunInstallScript() (bool, string, Error) {
	// download script
	script, apiErr := http.GetCLIInstallScript()
	if !apiErr.IsNil() {
		return false, "", Error{Err: apiErr.Unwrap(), Message: apiErr.Message}
	}

	// write script to temp file
	tmpFile, err := utils.WriteTempFile("install.sh", script, 0555)
	// clean up temp file once we're done with it
	defer os.Remove(tmpFile)

	// execute script
	utils.LogDebug("Executing install script")
	command := []string{tmpFile, "--debug"}
	out, err := exec.Command(command[0], command[1:]...).CombinedOutput() // #nosec G204
	strOut := string(out)
	// log output before checking error
	utils.LogDebug(fmt.Sprintf("Executing \"%s\"", strings.Join(command, " ")))
	if utils.Debug {
		fmt.Println(strOut)
	}
	if err != nil {
		return false, "", Error{Err: err, Message: "Unable to install the latest Doppler CLI"}
	}

	// find installed version within script output
	// Ex: `Installed Doppler CLI v3.7.1`
	re := regexp.MustCompile(`Installed Doppler CLI v(\d+\.\d+\.\d+)`)
	matches := re.FindStringSubmatch(strOut)
	if matches == nil || len(matches) != 2 {
		return false, "", Error{Err: errors.New("Unable to determine new CLI version")}
	}
	// parse latest version string
	newVersion, err := version.ParseVersion(matches[1])
	if err != nil {
		return false, "", Error{Err: err, Message: "Unable to parse new CLI version"}
	}

	wasUpdated := false
	// parse current version string
	currentVersion, currVersionErr := version.ParseVersion(version.ProgramVersion)
	if currVersionErr != nil {
		// unexpected error; just consider it an update and continue executing
		wasUpdated = true
		utils.LogDebug("Unable to parse current CLI version")
		utils.LogDebugError(currVersionErr)
	}

	if !wasUpdated {
		wasUpdated = version.CompareVersions(currentVersion, newVersion) == 1
	}

	return wasUpdated, newVersion.String(), Error{}
}

// CLIChangeLog fetches the latest changelog
func CLIChangeLog() (map[string]models.ChangeLog, http.Error) {
	response, apiError := http.GetChangelog()
	if !apiError.IsNil() {
		return nil, apiError

	}

	changes := models.ParseChangeLog(response)
	return changes, http.Error{}
}
