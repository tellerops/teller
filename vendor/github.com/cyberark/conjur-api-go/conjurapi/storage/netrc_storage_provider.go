package storage

import (
	"errors"
	"fmt"
	"os"

	"github.com/bgentry/go-netrc/netrc"
)

type NetrcStorageProvider struct {
	netRCPath   string
	machineName string
}

func NewNetrcStorageProvider(netRCPath, machineName string) *NetrcStorageProvider {
	return &NetrcStorageProvider{
		netRCPath:   netRCPath,
		machineName: machineName,
	}
}

// StoreCredentials stores credentials to the specified .netrc file
func (s *NetrcStorageProvider) StoreCredentials(login string, password string) error {
	err := s.ensureNetrcFileExists()
	if err != nil {
		return err
	}

	nrc, err := netrc.ParseFile(s.netRCPath)
	if err != nil {
		return err
	}

	m := nrc.FindMachine(s.machineName)
	if m == nil || m.IsDefault() {
		_ = nrc.NewMachine(s.machineName, login, password, "")
	} else {
		m.UpdateLogin(login)
		m.UpdatePassword(password)
	}

	data, err := nrc.MarshalText()
	if err != nil {
		return err
	}

	data = ensureEndsWithNewline(data)

	return os.WriteFile(s.netRCPath, data, 0600)
}

func (s *NetrcStorageProvider) ReadCredentials() (string, string, error) {
	nrc, err := netrc.ParseFile(s.netRCPath)
	if err != nil {
		return "", "", err
	}

	m := nrc.FindMachine(s.machineName)
	if m == nil {
		return "", "", fmt.Errorf("No credentials found in NetRCPath")
	}

	return m.Login, m.Password, nil
}

// ReadAuthnToken fetches the cached conjur access token. We only do this for OIDC
// since we don't have access to the Conjur API key and this is the only credential we can save.
func (s *NetrcStorageProvider) ReadAuthnToken() ([]byte, error) {
	_, tokenStr, err := s.ReadCredentials()
	if err != nil {
		return nil, err
	}

	return []byte(tokenStr), nil
}

// StoreAuthnToken stores the conjur access token. We only do this for OIDC
// since we don't have access to the Conjur API key and this is the only credential we can save.
func (s *NetrcStorageProvider) StoreAuthnToken(token []byte) error {
	// We should be able to use an empty string for username, but unfortunately
	// this causes panics later on. Instead use a dummy value.
	return s.StoreCredentials("[oidc]", string(token))
}

// PurgeCredentials purges credentials from the specified .netrc file
func (s *NetrcStorageProvider) PurgeCredentials() error {
	// Remove cached credentials (username, api key) from .netrc
	nrc, err := netrc.ParseFile(s.netRCPath)
	if err != nil {
		// If the .netrc file doesn't exist, we don't need to do anything
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		// Any other error should be returned
		return err
	}

	nrc.RemoveMachine(s.machineName)

	data, err := nrc.MarshalText()
	if err != nil {
		return err
	}

	return os.WriteFile(s.netRCPath, data, 0600)
}

func (s *NetrcStorageProvider) ensureNetrcFileExists() error {
	_, err := os.Stat(s.netRCPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.WriteFile(s.netRCPath, []byte{}, 0600)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func ensureEndsWithNewline(data []byte) []byte {
	if data[len(data)-1] != byte('\n') {
		data = append(data, byte('\n'))
	}
	return data
}
