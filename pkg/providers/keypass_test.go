package providers

import (
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/spectralops/teller/pkg/core"
)

func TestKetPass(t *testing.T) {

	_, filename, _, _ := runtime.Caller(0) //nolint
	os.Setenv("KEYPASS_PASSWORD", "1234")

	os.Setenv("KEYPASS_DB_PATH", path.Join(path.Dir(filename), "mock_providers", "keypass.kdbx"))

	k, err := NewKeyPass(GetTestLogger())
	assert.Nil(t, err)
	AssertProvider(t, k, false)
	p := core.NewPopulate(map[string]string{"stage": "prod"})
	kpmap := p.KeyPath(core.KeyPath{Field: "MG_KEY", Path: "settings/{{stage}}/billing-svc/all", Decrypt: true})
	ents, err := k.GetMapping(kpmap)
	assert.Nil(t, err)
	assert.Equal(t, len(ents), 4)
}

func TestKeypassFailures(t *testing.T) {

	_, filename, _, _ := runtime.Caller(0) //nolint
	os.Setenv("KEYPASS_PASSWORD", "1234")

	os.Setenv("KEYPASS_DB_PATH", path.Join(path.Dir(filename), "mock_providers", "keypass.kdbx"))

	k, _ := NewKeyPass(GetTestLogger())
	_, err := k.Get(core.KeyPath{Env: "NOT_EXISTS", Path: "settings"})
	assert.NotNil(t, err)

}
