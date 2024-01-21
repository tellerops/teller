package storage

import (
	"errors"

	"github.com/cyberark/conjur-api-go/conjurapi/logging"
	"github.com/zalando/go-keyring"
)

type KeyringStorageProvider struct {
	machineName string
}

var keyring_keys = []string{"login", "password", "authn_token"}
var ErrWritingCredentials = errors.New("unable to write credentials to keyring")
var ErrReadingCredentials = errors.New("unable to read credentials from keyring")

func NewKeyringStorageProvider(machineName string) *KeyringStorageProvider {
	return &KeyringStorageProvider{
		machineName: machineName,
	}
}

// IsKeyringAvailable returns true if the keyring is available on the system
func IsKeyringAvailable() bool {
	// Try to get a value. If there's an error other than "not found", then the
	// keyring is not available.
	_, err := keyring.Get("test", "test")
	return err == keyring.ErrNotFound
}

func (k *KeyringStorageProvider) StoreCredentials(login string, password string) error {
	err := keyring.Set(k.machineName, "login", login)
	if err != nil {
		logging.ApiLog.Debug(err)
		return ErrWritingCredentials
	}

	err = keyring.Set(k.machineName, "password", password)
	if err != nil {
		logging.ApiLog.Debug(err)
		return ErrWritingCredentials
	}

	return nil
}

func (k *KeyringStorageProvider) ReadCredentials() (string, string, error) {
	login, err := keyring.Get(k.machineName, "login")
	if err != nil && err != keyring.ErrNotFound {
		logging.ApiLog.Debug(err)
		return "", "", ErrReadingCredentials
	}
	password, err := keyring.Get(k.machineName, "password")
	if err != nil && err != keyring.ErrNotFound {
		logging.ApiLog.Debug(err)
		return "", "", ErrReadingCredentials
	}
	return login, password, nil
}

func (k *KeyringStorageProvider) ReadAuthnToken() ([]byte, error) {
	token, err := keyring.Get(k.machineName, "authn_token")
	if err != nil && err != keyring.ErrNotFound {
		logging.ApiLog.Debug(err)
		return nil, ErrReadingCredentials
	}
	return []byte(token), nil
}

func (k *KeyringStorageProvider) StoreAuthnToken(token []byte) error {
	err := keyring.Set(k.machineName, "authn_token", string(token))
	if err != nil {
		logging.ApiLog.Debug(err)
		return ErrWritingCredentials
	}
	return nil
}

func (k *KeyringStorageProvider) PurgeCredentials() error {
	for _, key := range keyring_keys {
		err := keyring.Delete(k.machineName, key)
		if err != nil {
			logging.ApiLog.Debugf("Error when deleting %s from keyring: %s", key, err)
		}
	}
	return nil
}
