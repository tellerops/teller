//go:build integration
// +build integration

package integration_test

import (
	"os"
	"testing"

	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
	"github.com/spectralops/teller/pkg/providers"
	"github.com/stretchr/testify/assert"
)

func TestGetDotEnv(t *testing.T) {
	//
	// pre-insert data
	//
	f, err := os.CreateTemp(t.TempDir(), "dotenv-*")
	assert.NoError(t, err)
	f.WriteString("MG_KEY=123\n")
	f.Close()

	//
	// use provider to read data
	//
	p, err := providers.NewDotenv(logging.New())
	assert.NoError(t, err)
	kvp := core.KeyPath{Env: "MG_KEY", Path: f.Name()}
	res, err := p.Get(kvp)

	assert.NoError(t, err)
	assert.Equal(t, "MG_KEY", res.Key)
	assert.Equal(t, "123", res.Value)
	assert.Equal(t, f.Name(), res.ResolvedPath)

	err = p.Put(kvp, "changed-secret")
	assert.NoError(t, err)

	res, err = p.Get(kvp)
	assert.NoError(t, err)
	assert.Equal(t, "MG_KEY", res.Key)
	assert.Equal(t, "changed-secret", res.Value)
	assert.Equal(t, f.Name(), res.ResolvedPath)

	err = p.PutMapping(kvp, map[string]string{"K1": "val1", "K2": "val2"})
	assert.NoError(t, err)

	res, err = p.Get(core.KeyPath{Env: "K1", Path: f.Name()})
	assert.NoError(t, err)
	assert.Equal(t, "K1", res.Key)
	assert.Equal(t, "val1", res.Value)

	err = p.Delete(core.KeyPath{Env: "MG_KEY", Path: f.Name()})
	assert.NoError(t, err)

	entries, err := p.GetMapping(core.KeyPath{Path: f.Name()})
	for _, entry := range entries {
		assert.NotEqual(t, "MG_KEY", entry.Key)
	}
	assert.NoError(t, err)

	err = p.DeleteMapping(kvp)
	assert.NoError(t, err)

	_, err = os.Stat(f.Name())
	assert.True(t, os.IsNotExist(err))
}
