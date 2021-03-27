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

// SecretsFileMetadata contains metadata about a secrets file
type SecretsFileMetadata struct {
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	ETag    string `json:"etag,omitempty" yaml:"etag,omitempty"`
	Hash    string `json:"hash,omitempty" yaml:"hash,omitempty"`
}

// ParseSecretsFileMetadata parse secrets file metadata
func ParseSecretsFileMetadata(data map[string]interface{}) SecretsFileMetadata {
	var parsedMetadata SecretsFileMetadata

	if data["version"] != nil {
		parsedMetadata.Version = data["version"].(string)
	}
	if data["etag"] != nil {
		parsedMetadata.ETag = data["etag"].(string)
	}
	if data["hash"] != nil {
		parsedMetadata.Hash = data["hash"].(string)
	}

	return parsedMetadata
}
