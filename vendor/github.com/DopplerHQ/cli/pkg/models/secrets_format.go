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

// SecretsFormat the format secrets should use
type SecretsFormat int

// the source of the value
const (
	JSON SecretsFormat = iota
	ENV
	YAML
	DOCKER
	ENV_NO_FILE
)

var SecretFormats = []string{"json", "env", "yaml", "docker", "env-no-quotes"}

func (s SecretsFormat) String() string {
	return SecretFormats[s]
}

// OutputFile the default secrets file name
func (s SecretsFormat) OutputFile() string {
	return [...]string{"doppler.json", "doppler.env", "secrets.yaml", "doppler.env", "doppler.env"}[s]
}

// SecretsFormatList list of supported secrets formats
var SecretsFormatList []SecretsFormat

func init() {
	SecretsFormatList = append(SecretsFormatList, JSON)
	SecretsFormatList = append(SecretsFormatList, ENV)
	SecretsFormatList = append(SecretsFormatList, YAML)
	SecretsFormatList = append(SecretsFormatList, DOCKER)
	SecretsFormatList = append(SecretsFormatList, ENV_NO_FILE)
}
