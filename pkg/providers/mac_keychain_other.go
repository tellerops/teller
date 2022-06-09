//go:build dragonfly || freebsd || windows || linux || netbsd || openbsd || solaris
// +build dragonfly freebsd windows linux netbsd openbsd solaris

package providers

import (
	"fmt"

	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
)

type MacKeychain struct {
}

// MacKeychain creates new provider instance
func NewMacKeychain(logger logging.Logger) (core.Provider, error) { //nolint
	return &MacKeychain{}, nil
}

// Name return the provider name
func (mc *MacKeychain) Name() string {
	return "Mac_Keychain"
}

// Put will create a new single entry
func (mc *MacKeychain) Put(p core.KeyPath, val string) error { //nolint
	return fmt.Errorf("provider %q not supported in Linux", mc.Name())
}

// PutMapping will create a multiple entries
func (mc *MacKeychain) PutMapping(p core.KeyPath, m map[string]string) error { //nolint
	return fmt.Errorf("provider %q not supported in Linux", mc.Name())
}

// GetMapping returns a multiple entries
func (mc *MacKeychain) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) { //nolint

	return nil, fmt.Errorf("provider %q not supported in Linux", mc.Name())
}

// Get returns a single entry
func (mc *MacKeychain) Get(p core.KeyPath) (*core.EnvEntry, error) { //nolint
	return nil, fmt.Errorf("provider %q not supported in Linux", mc.Name())
}

// Delete will delete entry
func (mc *MacKeychain) Delete(kp core.KeyPath) error { //nolint

	return fmt.Errorf("provider %q not supported in Linux", mc.Name())

}

// DeleteMapping will delete the given path recessively
func (mc *MacKeychain) DeleteMapping(kp core.KeyPath) error { //nolint
	return fmt.Errorf("provider %q not supported in Linux", mc.Name())
}
