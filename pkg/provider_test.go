package pkg

import (
	"fmt"
	"testing"

	"github.com/alecthomas/assert"
)

func TestGetProvider(t *testing.T) {
	providers := &BuiltinProviders{}
	p, err := providers.GetProvider("missing")
	assert.Nil(t, p)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "provider 'missing' does not exist")

	for _, v := range providers.ProviderHumanToMachine() {
		_, err := providers.GetProvider(v)
		if err != nil {
			assert.NotContains(t, err.Error(), fmt.Sprintf("provider %s does not exist", v))
		}
	}
}
