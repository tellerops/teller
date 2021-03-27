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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/DopplerHQ/cli/pkg/crypto"
	"github.com/DopplerHQ/cli/pkg/models"
	"github.com/DopplerHQ/cli/pkg/utils"
	"gopkg.in/yaml.v3"
)

// DefaultMetadataDir the directory containing metadata files
var DefaultMetadataDir string

// MetadataFilePath calculates the name of the metadata file
func MetadataFilePath(token string, project string, config string) string {
	var name string
	if project == "" && config == "" {
		name = fmt.Sprintf("%s", token)
	} else {
		name = fmt.Sprintf("%s:%s:%s", token, project, config)
	}

	fileName := fmt.Sprintf(".metadata-%s.json", crypto.Hash(name))
	path := filepath.Join(DefaultMetadataDir, fileName)
	if absPath, err := filepath.Abs(path); err == nil {
		return absPath
	}
	return path
}

// MetadataFile reads the contents of the metadata file
func MetadataFile(path string) (models.SecretsFileMetadata, Error) {
	utils.LogDebug(fmt.Sprintf("Using metadata file %s", path))

	if _, err := os.Stat(path); err != nil {
		var e Error
		e.Err = err
		if os.IsNotExist(err) {
			e.Message = "Metadata file does not exist"
		} else {
			e.Message = "Unable to read metadata file"
		}
		return models.SecretsFileMetadata{}, e
	}

	utils.LogDebug(fmt.Sprintf("Reading metadata file %s", path))
	response, err := ioutil.ReadFile(path) // #nosec G304
	if err != nil {
		return models.SecretsFileMetadata{}, Error{Err: err, Message: "Unable to read metadata file"}
	}

	var metadata models.SecretsFileMetadata
	if err := yaml.Unmarshal(response, &metadata); err != nil {
		return models.SecretsFileMetadata{}, Error{Err: err, Message: "Unable to parse metadata file"}
	}

	return metadata, Error{}
}

// WriteMetadataFile writes the contents of the metadata file
func WriteMetadataFile(path string, etag string, hash string) Error {
	utils.LogDebug(fmt.Sprintf("Writing ETag to metadata file %s", path))

	metadata := models.SecretsFileMetadata{
		Version: "1",
		ETag:    etag,
		Hash:    hash,
	}

	metadataBytes, err := yaml.Marshal(metadata)
	if err != nil {
		return Error{Err: err, Message: "Unable to marshal metadata to YAML"}
	}

	if err := utils.WriteFile(path, []byte(metadataBytes), utils.RestrictedFilePerms()); err != nil {
		return Error{Err: err, Message: "Unable to write metadata file"}
	}

	return Error{}
}

// SecretsCacheFile reads the contents of the cache file
func SecretsCacheFile(path string, passphrase string) (map[string]string, Error) {
	utils.LogDebug(fmt.Sprintf("Using fallback file for cache %s", path))

	if _, err := os.Stat(path); err != nil {
		return nil, Error{Err: err, Message: "Unable to stat cache file"}
	}

	response, err := ioutil.ReadFile(path) // #nosec G304
	if err != nil {
		return nil, Error{Err: err, Message: "Unable to read cache file"}
	}

	utils.LogDebug("Decrypting cache file")
	decryptedSecrets, err := crypto.Decrypt(passphrase, response)
	if err != nil {
		return nil, Error{Err: err, Message: "Unable to decrypt cache file"}
	}

	secrets := map[string]string{}
	err = json.Unmarshal([]byte(decryptedSecrets), &secrets)
	if err != nil {
		return nil, Error{Err: err, Message: "Unable to parse cache file"}
	}

	return secrets, Error{}
}
