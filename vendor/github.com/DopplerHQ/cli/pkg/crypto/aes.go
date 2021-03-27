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

// From https://gist.github.com/tscholl2/dc7dc15dc132ea70a98e8542fefffa28

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/DopplerHQ/cli/pkg/utils"
	"golang.org/x/crypto/pbkdf2"
)

func deriveKey(passphrase string, salt []byte) ([]byte, []byte, error) {
	if salt == nil {
		salt = make([]byte, 8)
		// http://www.ietf.org/rfc/rfc2898.txt
		// Salt.
		_, err := rand.Read(salt)
		if err != nil {
			return nil, nil, err
		}
	}

	return pbkdf2.Key([]byte(passphrase), salt, 50000, 32, sha256.New), salt, nil
}

// Encrypt plaintext with a passphrase; uses pbkdf2 for key deriv and aes-256-gcm for encryption
func Encrypt(passphrase string, plaintext []byte) (string, error) {
	now := time.Now()
	key, salt, err := deriveKey(passphrase, nil)
	if err != nil {
		return "", err
	}

	utils.LogDebug(fmt.Sprintf("PBKDF2 key derivation took %d ms", time.Now().Sub(now).Milliseconds()))

	iv := make([]byte, 12)
	// http://nvlpubs.nist.gov/nistpubs/Legacy/SP/nistspecialpublication800-38d.pdf
	// Section 8.2
	_, err = rand.Read(iv)
	if err != nil {
		return "", err
	}

	b, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(b)
	if err != nil {
		return "", err
	}

	data := aesgcm.Seal(nil, iv, plaintext, nil)
	return hex.EncodeToString(salt) + "-" + hex.EncodeToString(iv) + "-" + hex.EncodeToString(data), nil
}

// Decrypt ciphertext with a passphrase
func Decrypt(passphrase string, ciphertext []byte) (string, error) {
	arr := strings.Split(string(ciphertext), "-")
	salt, err := hex.DecodeString(arr[0])
	if err != nil {
		return "", err
	}

	iv, err := hex.DecodeString(arr[1])
	if err != nil {
		return "", err
	}

	data, err := hex.DecodeString(arr[2])
	if err != nil {
		return "", err
	}

	key, _, err := deriveKey(passphrase, salt)
	if err != nil {
		return "", err
	}

	b, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(b)
	if err != nil {
		return "", err
	}

	data, err = aesgcm.Open(nil, iv, data, nil)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
