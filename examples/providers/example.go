package providers

import (
	"fmt"

	"github.com/spectralops/teller/pkg/core"

	"github.com/spectralops/teller/pkg/logging"
)

type Example struct {
	logger logging.Logger
}

//nolint

// func init() {
// 	metaInto := core.MetaInfo{
// 		Description:    "ProviderName",
// 		Name:           "provider_name",
// 		Authentication: "If you have the Consul CLI working and configured, there's no special action to take.\nConfiguration is environment based, as defined by client standard. See variables [here](https://github.com/hashicorp/consul/blob/master/api/api.go#L28).",
// 		ConfigTemplate: `
//   provider:
//     env:
//       KEY_EAXMPLE:
//         path: pathToKey
// `,
// 		Ops: core.OpMatrix{Get: true, GetMapping: true, Put: true, PutMapping: true},
// 	}
// 	RegisterProvider(metaInto, NewExample)
// }

// NewExample creates new provider instance
func NewExample(logger logging.Logger) (core.Provider, error) {

	return &Example{
		logger: logger,
	}, nil
}

// Name return the provider name
func (e *Example) Name() string {
	return "Example"
}

// Put will create a new single entry
func (e *Example) Put(p core.KeyPath, val string) error {
	return fmt.Errorf("provider %q does not implement write yet", e.Name())
}

// PutMapping will create a multiple entries
func (e *Example) PutMapping(p core.KeyPath, m map[string]string) error {
	return fmt.Errorf("provider %q does not implement write yet", e.Name())
}

// GetMapping returns a multiple entries
func (e *Example) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {

	return []core.EnvEntry{}, fmt.Errorf("provider %q does not implement write yet", e.Name())
}

// Get returns a single entry
func (e *Example) Get(p core.KeyPath) (*core.EnvEntry, error) {

	return &core.EnvEntry{}, fmt.Errorf("provider %q does not implement write yet", e.Name())
}

// Delete will delete entry
func (e *Example) Delete(kp core.KeyPath) error {
	return fmt.Errorf("provider %s does not implement delete yet", e.Name())
}

// DeleteMapping will delete the given path recessively
func (e *Example) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("provider %s does not implement delete yet", e.Name())
}
