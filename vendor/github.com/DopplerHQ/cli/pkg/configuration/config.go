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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/DopplerHQ/cli/pkg/controllers"
	"github.com/DopplerHQ/cli/pkg/models"
	"github.com/DopplerHQ/cli/pkg/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// baseConfigDir (e.g. /home/user/)
var baseConfigDir string

// UserConfigDir (e.g. /home/user/.doppler)
var UserConfigDir string

// UserConfigFile (e.g. /home/user/doppler/.doppler.yaml)
var UserConfigFile string

// Scope to use for config file
var Scope = "."

var configFileName = ".doppler.yaml"
var configContents models.ConfigFile

func init() {
	baseConfigDir = utils.HomeDir()
	UserConfigDir = filepath.Join(baseConfigDir, ".doppler")
	UserConfigFile = filepath.Join(UserConfigDir, configFileName)
}

// Setup the config directory and config file
func Setup() {
	utils.LogDebug(fmt.Sprintf("Using config file %s", UserConfigFile))

	if !utils.Exists(UserConfigDir) {
		utils.LogDebug(fmt.Sprintf("Creating the config directory %s", UserConfigDir))
		err := os.Mkdir(UserConfigDir, 0700)
		if err != nil {
			utils.HandleError(err, fmt.Sprintf("Unable to create config directory %s", UserConfigDir))
		}
	}

	// This may be different from `UserConfigDir` if `--configuration` was provided
	configDir := filepath.Dir(UserConfigFile)
	if !utils.Exists(configDir) {
		utils.HandleError(fmt.Errorf("Configuration file directory does not exist %s", configDir))
	}

	if !utils.Exists(UserConfigFile) {
		v1ConfigA := filepath.Join(utils.ConfigDir(), configFileName)
		v1ConfigB := filepath.Join(utils.HomeDir(), configFileName)
		if utils.Exists(v1ConfigA) {
			utils.LogDebug("Migrating the config from CLI v1")
			err := os.Rename(v1ConfigA, UserConfigFile)
			if err != nil {
				utils.HandleError(err, "Unable to migrate config from CLI v1")
			}
		} else if utils.Exists(v1ConfigB) {
			utils.LogDebug("Migrating the config from CLI v1")
			err := os.Rename(v1ConfigB, UserConfigFile)
			if err != nil {
				utils.HandleError(err, "Unable to migrate config from CLI v1")
			}
		} else if jsonExists() {
			utils.LogDebug("Migrating the config from the Node CLI")
			migrateJSONToYaml()
		} else {
			utils.LogDebug("Creating a new config file")
			var blankConfig models.ConfigFile
			writeConfig(blankConfig)
		}
	}
}

// LoadConfig load the configuration file
func LoadConfig() {
	configContents = readConfig()
}

// VersionCheck the last version check
func VersionCheck() models.VersionCheck {
	return configContents.VersionCheck
}

// SetVersionCheck the last version check
func SetVersionCheck(version models.VersionCheck) {
	configContents.VersionCheck = version
	writeConfig(configContents)
}

// Get the config at the specified scope
func Get(scope string) models.ScopedOptions {
	var normalizedScope string
	var err error
	if normalizedScope, err = NormalizeScope(scope); err != nil {
		utils.HandleError(err, fmt.Sprintf("Invalid scope: %s", scope))
	}
	if !strings.HasSuffix(normalizedScope, string(filepath.Separator)) {
		normalizedScope = normalizedScope + string(filepath.Separator)
	}
	var scopedConfig models.ScopedOptions

	for confScope, conf := range configContents.Scoped {
		confScopePath := confScope
		// both paths must end in / to prevent partial match (e.g. /test matching /test123)
		if !strings.HasSuffix(confScopePath, string(filepath.Separator)) {
			confScopePath = confScopePath + string(filepath.Separator)
		}

		if !strings.HasPrefix(normalizedScope, confScopePath) {
			continue
		}

		pairs := models.Pairs(conf)
		scopedPairs := models.ScopedPairs(&scopedConfig)
		for name, pair := range pairs {
			if pair != "" {
				scopedPair := scopedPairs[name]
				if *scopedPair == (models.ScopedOption{}) || len(confScope) > len(scopedPair.Scope) {
					scopedPair.Value = pair
					scopedPair.Scope = confScope
					scopedPair.Source = models.ConfigFileSource.String()
				}
			}
		}
	}

	if controllers.IsKeyringSecret(scopedConfig.Token.Value) {
		utils.LogDebug(fmt.Sprintf("Retrieving %s from system keyring", models.ConfigToken.String()))
		token, err := controllers.GetKeyring(scopedConfig.Token.Value)
		if !err.IsNil() {
			utils.HandleError(err.Unwrap(), err.Message)
		}

		scopedConfig.Token.Value = token
	}

	return scopedConfig
}

// LocalConfig retrieves the config for the scoped directory
func LocalConfig(cmd *cobra.Command) models.ScopedOptions {
	// config file (lowest priority)
	localConfig := Get(Scope)

	// environment variables
	if !utils.GetBoolFlag(cmd, "no-read-env") {
		pairs := models.EnvPairs(&localConfig)
		envVars := []string{}
		for envVar := range pairs {
			envVars = append(envVars, envVar)
		}

		// sort variables so that they are processed in a deterministic order
		// this also ensures ENCLAVE_ variables are given precedence over (i.e. read after) DOPPLER_ variables,
		// which is necessary for backwards compatibility until we drop support for ENCLAVE_ variables
		sort.Strings(envVars)

		for _, envVar := range envVars {
			envValue := os.Getenv(envVar)
			if envValue != "" {
				pair := pairs[envVar]
				pair.Value = envValue
				pair.Scope = "/"
				pair.Source = models.EnvironmentSource.String()
			}
		}
	}

	// individual flags (highest priority)
	flagSet := cmd.Flags().Changed("token")
	if flagSet || localConfig.Token.Value == "" {
		localConfig.Token.Value = cmd.Flag("token").Value.String()
		localConfig.Token.Scope = "/"

		if flagSet {
			localConfig.Token.Source = models.FlagSource.String()
		} else {
			localConfig.Token.Source = models.DefaultValueSource.String()
		}
	}

	flagSet = cmd.Flags().Changed("api-host")
	if flagSet || localConfig.APIHost.Value == "" {
		localConfig.APIHost.Value = cmd.Flag("api-host").Value.String()
		localConfig.APIHost.Scope = "/"

		if flagSet {
			localConfig.APIHost.Source = models.FlagSource.String()
		} else {
			localConfig.APIHost.Source = models.DefaultValueSource.String()
		}
	}

	flagSet = cmd.Flags().Changed("dashboard-host")
	if flagSet || localConfig.DashboardHost.Value == "" {
		localConfig.DashboardHost.Value = cmd.Flag("dashboard-host").Value.String()
		localConfig.DashboardHost.Scope = "/"

		if flagSet {
			localConfig.DashboardHost.Source = models.FlagSource.String()
		} else {
			localConfig.DashboardHost.Source = models.DefaultValueSource.String()
		}
	}

	flagSet = cmd.Flags().Changed("no-verify-tls")
	if flagSet || localConfig.VerifyTLS.Value == "" {
		noVerifyTLS := cmd.Flag("no-verify-tls").Value.String()
		localConfig.VerifyTLS.Value = strconv.FormatBool(!utils.GetBool(noVerifyTLS, false))
		localConfig.VerifyTLS.Scope = "/"

		if flagSet {
			localConfig.VerifyTLS.Source = models.FlagSource.String()
		} else {
			localConfig.VerifyTLS.Source = models.DefaultValueSource.String()
		}
	}

	// these flags below do not have a default value and should only be used if specified by the user (or will cause invalid memory access)
	flagSet = cmd.Flags().Changed("project")
	if flagSet {
		localConfig.EnclaveProject.Value = cmd.Flag("project").Value.String()
		localConfig.EnclaveProject.Scope = "/"

		if flagSet {
			localConfig.EnclaveProject.Source = models.FlagSource.String()
		} else {
			localConfig.EnclaveProject.Source = models.DefaultValueSource.String()
		}
	}

	flagSet = cmd.Flags().Changed("config")
	if flagSet {
		localConfig.EnclaveConfig.Value = cmd.Flag("config").Value.String()
		localConfig.EnclaveConfig.Scope = "/"

		if flagSet {
			localConfig.EnclaveConfig.Source = models.FlagSource.String()
		} else {
			localConfig.EnclaveConfig.Source = models.DefaultValueSource.String()
		}
	}

	return localConfig
}

// AllConfigs get all configs we know about
func AllConfigs() map[string]models.FileScopedOptions {
	all := map[string]models.FileScopedOptions{}
	for scope, scopedOptions := range configContents.Scoped {
		options := scopedOptions

		if controllers.IsKeyringSecret(options.Token) {
			utils.LogDebug(fmt.Sprintf("Retrieving %s from system keyring", models.ConfigToken.String()))
			token, err := controllers.GetKeyring(options.Token)
			if !err.IsNil() {
				utils.HandleError(err.Unwrap(), err.Message)
			}

			options.Token = token
		}

		all[scope] = options
	}
	return all
}

// Set properties on a scoped config
func Set(scope string, options map[string]string) {
	var normalizedScope string
	var err error
	if normalizedScope, err = NormalizeScope(scope); err != nil {
		utils.HandleError(err, fmt.Sprintf("Invalid scope: %s", scope))
	}

	config := configContents.Scoped[normalizedScope]
	previousToken := config.Token

	for key, value := range options {
		if !IsValidConfigOption(key) {
			utils.HandleError(errors.New("invalid option "+key), "")
		}

		if key == models.ConfigToken.String() {
			utils.LogDebug(fmt.Sprintf("Saving %s to system keyring", key))
			uuid, err := utils.UUID()
			if err != nil {
				utils.HandleError(err, "Unable to generate UUID for keyring")
			}
			id := controllers.GenerateKeyringID(uuid)

			if controllerError := controllers.SetKeyring(id, value); !controllerError.IsNil() {
				utils.LogDebugError(controllerError.Unwrap())
				utils.LogDebug(controllerError.Message)
			} else {
				value = id

				// remove old token from keyring
				if controllers.IsKeyringSecret(previousToken) {
					utils.LogDebug("Removing previous token from system keyring")
					if controllerError := controllers.DeleteKeyring(previousToken); !controllerError.IsNil() {
						utils.LogDebugError(controllerError.Unwrap())
						utils.LogDebug(controllerError.Message)
					}
				}
			}
		}

		SetConfigValue(&config, key, value)
		configContents.Scoped[normalizedScope] = config
	}

	writeConfig(configContents)
}

// Unset a local config
func Unset(scope string, options []string) {
	var normalizedScope string
	var err error
	if normalizedScope, err = NormalizeScope(scope); err != nil {
		utils.HandleError(err, fmt.Sprintf("Invalid scope: %s", scope))
	}

	if configContents.Scoped[normalizedScope] == (models.FileScopedOptions{}) {
		return
	}

	for _, key := range options {
		if !IsValidConfigOption(key) {
			utils.HandleError(errors.New("invalid option "+key), "")
		}

		config := configContents.Scoped[normalizedScope]

		if key == models.ConfigToken.String() {
			previousToken := config.Token
			// remove old token from keyring
			if controllers.IsKeyringSecret(previousToken) {
				if controllerError := controllers.DeleteKeyring(previousToken); !controllerError.IsNil() {
					utils.LogDebugError(controllerError.Unwrap())
					utils.LogDebug(controllerError.Message)
				}
			}
		}

		SetConfigValue(&config, key, "")
		configContents.Scoped[normalizedScope] = config
	}

	if configContents.Scoped[normalizedScope] == (models.FileScopedOptions{}) {
		delete(configContents.Scoped, normalizedScope)
	}

	writeConfig(configContents)
}

// Write config to filesystem
func writeConfig(config models.ConfigFile) {
	bytes, err := yaml.Marshal(config)
	if err != nil {
		utils.HandleError(err)
	}

	utils.LogDebug(fmt.Sprintf("Writing user config to %s", UserConfigFile))
	if err := utils.WriteFile(UserConfigFile, bytes, os.FileMode(0600)); err != nil {
		utils.HandleError(err)
	}
}

func readConfig() models.ConfigFile {
	utils.LogDebug("Reading config file")

	fileContents, err := ioutil.ReadFile(UserConfigFile) // #nosec G304
	if err != nil {
		utils.HandleError(err, "Unable to read user config file")
	}

	var config models.ConfigFile
	err = yaml.Unmarshal(fileContents, &config)
	if err != nil {
		utils.HandleError(err, "Unable to parse user config file")
	}

	// sort scopes before normalizing so that if multiple scopes normalize to
	// the same value (like '/' and '*') they'll apply in a deterministic order
	var sorted []string
	for scope := range config.Scoped {
		sorted = append(sorted, scope)
	}
	sort.Strings(sorted)

	// normalize config scope and merge options from conflicting scopes
	normalizedOptions := map[string]models.FileScopedOptions{}
	for _, scope := range sorted {
		var normalizedScope string
		if normalizedScope, err = NormalizeScope(scope); err != nil {
			utils.HandleError(err, fmt.Sprintf("Invalid scope: %s", scope))
		}
		scopedOption := normalizedOptions[normalizedScope]

		options := config.Scoped[scope]
		if options.APIHost != "" {
			scopedOption.APIHost = options.APIHost
		}
		if options.DashboardHost != "" {
			scopedOption.DashboardHost = options.DashboardHost
		}
		if options.EnclaveConfig != "" {
			scopedOption.EnclaveConfig = options.EnclaveConfig
		}
		if options.EnclaveProject != "" {
			scopedOption.EnclaveProject = options.EnclaveProject
		}
		if options.Token != "" {
			scopedOption.Token = options.Token
		}
		if options.VerifyTLS != "" {
			scopedOption.VerifyTLS = options.VerifyTLS
		}

		normalizedOptions[normalizedScope] = scopedOption
	}

	config.Scoped = normalizedOptions
	return config
}

// IsValidConfigOption whether the specified key is a valid config option
func IsValidConfigOption(key string) bool {
	configOptions := map[string]interface{}{
		models.ConfigToken.String():          nil,
		models.ConfigAPIHost.String():        nil,
		models.ConfigDashboardHost.String():  nil,
		models.ConfigVerifyTLS.String():      nil,
		models.ConfigEnclaveProject.String(): nil,
		models.ConfigEnclaveConfig.String():  nil,
	}

	_, exists := configOptions[key]
	return exists
}

// IsTranslatableConfigOption checks whether the key can be translated to a valid config option
func IsTranslatableConfigOption(key string) bool {
	// TODO remove this function when releasing CLI v4 (DPLR-435)
	if key == "config" || key == "project" {
		return true
	}

	return false
}

// TranslateFriendlyOption to its config option name
func TranslateFriendlyOption(key string) string {
	// TODO remove this function when releasing CLI v4 (DPLR-435)
	if key == "config" {
		return models.ConfigEnclaveConfig.String()
	}
	if key == "project" {
		return models.ConfigEnclaveProject.String()
	}
	return key
}

// TranslateConfigOption to its friendly name
func TranslateConfigOption(key string) string {
	// TODO remove this function when releasing CLI v4 (DPLR-435)
	if key == models.ConfigEnclaveConfig.String() {
		return "config"
	}
	if key == models.ConfigEnclaveProject.String() {
		return "project"
	}
	return key
}

// SetConfigValue set the value for the specified key in the config
func SetConfigValue(conf *models.FileScopedOptions, key string, value string) {
	if key == models.ConfigToken.String() {
		(*conf).Token = value
	} else if key == models.ConfigAPIHost.String() {
		(*conf).APIHost = value
	} else if key == models.ConfigDashboardHost.String() {
		(*conf).DashboardHost = value
	} else if key == models.ConfigVerifyTLS.String() {
		(*conf).VerifyTLS = value
	} else if key == models.ConfigEnclaveProject.String() {
		(*conf).EnclaveProject = value
	} else if key == models.ConfigEnclaveConfig.String() {
		(*conf).EnclaveConfig = value
	}
}

// NormalizeScope from legacy '*' to '/'
func NormalizeScope(scope string) (string, error) {
	if scope == "*" {
		return "/", nil
	}

	return utils.ParsePath(scope)
}
