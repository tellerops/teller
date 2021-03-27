// Copyright 2013 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package keyring

import (
	"encoding/hex"
	"fmt"
	"os/exec"
	"strings"
)

const (
	execPathKeychain = "/usr/bin/security"

	// encodingPrefix is a well-known prefix added to strings encoded by Set.
	encodingPrefix = "go-keyring-encoded:"
)

type macOSXKeychain struct{}

// func (*MacOSXKeychain) IsAvailable() bool {
// 	return exec.Command(execPathKeychain).Run() != exec.ErrNotFound
// }

// Set stores stores user and pass in the keyring under the defined service
// name.
func (k macOSXKeychain) Get(service, username string) (string, error) {
	out, err := exec.Command(
		execPathKeychain,
		"find-generic-password",
		"-s", service,
		"-wa", username).CombinedOutput()
	if err != nil {
		if strings.Contains(fmt.Sprintf("%s", out), "could not be found") {
			err = ErrNotFound
		}
		return "", err
	}

	trimStr := strings.TrimSpace(string(out[:]))
	// if the string has the well-known prefix, assume it's encoded
	if strings.HasPrefix(trimStr, encodingPrefix) {
		dec, err := hex.DecodeString(trimStr[len(encodingPrefix):])
		return string(dec), err
	}

	return trimStr, nil
}

// Set stores a secret in the keyring given a service name and a user.
func (k macOSXKeychain) Set(service, username, password string) error {
	// if the added secret has multiple lines, osx will hex encode it
	// identify this with a well-known prefix.
	if strings.ContainsRune(password, '\n') {
		password = encodingPrefix + hex.EncodeToString([]byte(password))
	}

	return exec.Command(
		execPathKeychain,
		"add-generic-password",
		"-U", //update if exists
		"-s", service,
		"-a", username,
		"-w", password).Run()
}

// Delete deletes a secret, identified by service & user, from the keyring.
func (k macOSXKeychain) Delete(service, username string) error {
	out, err := exec.Command(
		execPathKeychain,
		"delete-generic-password",
		"-s", service,
		"-a", username).CombinedOutput()
	if strings.Contains(fmt.Sprintf("%s", out), "could not be found") {
		err = ErrNotFound
	}
	return err
}

func init() {
	provider = macOSXKeychain{}
}
