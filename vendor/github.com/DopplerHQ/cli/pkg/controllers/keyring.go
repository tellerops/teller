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
	"fmt"
	"strings"

	"github.com/zalando/go-keyring"
)

const keyringService = "doppler-cli"
const keyringSecretPrefix = "secret"

// IsKeyringSecret checks whether the secret is stored in keyring
func IsKeyringSecret(value string) bool {
	return strings.HasPrefix(value, fmt.Sprintf("%s-", keyringSecretPrefix))
}

// GenerateKeyringID generates a keyring-compliant key
func GenerateKeyringID(id string) string {
	return fmt.Sprintf("%s-%s", keyringSecretPrefix, id)
}

// GetKeyring fetches a secret from the keyring
func GetKeyring(id string) (string, Error) {
	value, err := keyring.Get(keyringService, id)
	if err != nil {
		if err == keyring.ErrUnsupportedPlatform {
			return "", Error{Err: err, Message: "Your OS does not support keyring"}
		} else if err == keyring.ErrNotFound {
			return "", Error{Err: err, Message: "Token not found in system keyring"}
		} else {
			return "", Error{Err: err, Message: "Unable to retrieve value from system keyring"}
		}
	}

	return value, Error{}
}

// SetKeyring saves a value to the keyring
func SetKeyring(key string, value string) Error {
	if err := keyring.Set(keyringService, key, value); err != nil {
		if err == keyring.ErrUnsupportedPlatform {
			return Error{Err: err, Message: "Your OS does not support keyring"}
		} else {
			return Error{Err: err, Message: "Unable to access system keyring for secure storage"}
		}
	}

	return Error{}
}

// DeleteKeyring removes a value from the keyring
func DeleteKeyring(key string) Error {
	if err := keyring.Delete(keyringService, key); err != nil {
		return Error{Err: err, Message: "Unable to remove value from keyring"}
	}

	return Error{}
}
