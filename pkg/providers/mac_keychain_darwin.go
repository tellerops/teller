//go:build darwin
// +build darwin

package providers

import (
	"github.com/99designs/keyring"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
)

// pathPrefix is the prefix for the keychain path
const pathPrefix = "teller-"

type KeychainClient interface {
	Set(item keyring.Item) error
	Get(key string) (keyring.Item, error)
	Keys() ([]string, error)
	Remove(key string) error
}

// MacKeychain is the provider for Mac Keychain
type MacKeychain struct {
	client KeychainClient
	logger logging.Logger
}

// nolint
func init() {
	metaInfo := core.MetaInfo{
		Description:    "Mac_Keychain",
		Authentication: "",
		Name:           "mac_keychain",
		ConfigTemplate: `
  # you can mix and match many files
  mac_keychain:
    env_sync:
      path: 'myApp'
    env:
      ETC_DSN:
        path: 'myApp'
        field: 'dsn-etc'
`,
		Ops: core.OpMatrix{Get: true, GetMapping: true, Put: true, PutMapping: true, Delete: true, DeleteMapping: true},
	}
	RegisterProvider(metaInfo, NewMacKeychain)
}

// NewMacKeychain creates new provider instance
func NewMacKeychain(logger logging.Logger) (core.Provider, error) {
	return &MacKeychain{
		logger: logger,
	}, nil
}

// Name return the provider name
func (mc *MacKeychain) Name() string {
	return "Mac_Keychain"
}

// newKeyring creates a new keyring instance
func (mc *MacKeychain) newKeyring(serviceName string) (KeychainClient, error) {
	var err error
	if mc.client == nil {
		// Use the best keyring implementation for your operating system
		mc.client, err = keyring.Open(keyring.Config{
			ServiceName:              pathPrefix + serviceName,
			KeychainTrustApplication: true, // trust the application to access the keychain without prompting the user
		})
	}

	return mc.client, err
}

// Put will create a new single entry
func (mc *MacKeychain) Put(p core.KeyPath, val string) error {
	kr, err := mc.newKeyring(p.Path)
	if err != nil {
		return err
	}

	if p.Field == "" {
		p.Field = p.Env
	}

	_ = kr.Set(keyring.Item{
		Key:                         p.Field,
		Data:                        []byte(val),
		KeychainNotTrustApplication: false,
	})

	return nil
}

// PutMapping will create a multiple entries
func (mc *MacKeychain) PutMapping(p core.KeyPath, m map[string]string) error {

	// Creating a put mapping can be a bit tricky. Some attributes need to be a unique values as an account.
	// To creates a put mapping functionality, we need to get multiple different attributes for a single record which can be understandable to the users.
	ring, err := mc.newKeyring(p.Path)
	if err != nil {
		return err
	}

	for k, v := range m {
		_ = ring.Set(keyring.Item{
			Key:                         k,
			Data:                        []byte(v),
			KeychainNotTrustApplication: false,
		})
	}

	return nil
}

// GetMapping returns a multiple entries
func (mc *MacKeychain) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {
	entries := []core.EnvEntry{}
	ring, err := mc.newKeyring(p.Path)
	if err != nil {
		return entries, err
	}

	keys, err := ring.Keys()
	if err != nil {
		return entries, err
	}

	for _, keyName := range keys {
		item, err := ring.Get(keyName)
		if err != nil {
			return nil, err
		}
		entries = append(entries, p.FoundWithKey(item.Key, string(item.Data)))
	}
	return entries, nil
}

// Get returns a single entry
func (mc *MacKeychain) Get(p core.KeyPath) (*core.EnvEntry, error) {

	ring, err := mc.newKeyring(p.Path)
	if err != nil {
		return nil, err
	}
	if p.Field == "" {
		p.Field = p.Env
	}

	item, err := ring.Get(p.Field)
	var ent = p.Missing()
	if err != nil {
		if err == keyring.ErrKeyNotFound {
			return &ent, nil
		}
		return nil, err
	}

	ent = p.Found(string(item.Data))

	return &ent, nil
}

// Delete will delete entry
func (mc *MacKeychain) Delete(p core.KeyPath) error {
	ring, err := mc.newKeyring(p.Path)
	if err != nil {
		return err
	}

	if p.Field == "" {
		p.Field = p.Env
	}

	return ring.Remove(p.Field)
}

// DeleteMapping will delete the given path recessively
func (mc *MacKeychain) DeleteMapping(p core.KeyPath) error {
	ring, err := mc.newKeyring(p.Path)
	if err != nil {
		return err
	}

	keys, err := ring.Keys()
	if err != nil {
		return err
	}

	for _, keyName := range keys {
		_ = ring.Remove(keyName)
	}

	return nil
}
