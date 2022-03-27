package providers

import (
	"testing"

	"github.com/alecthomas/assert"
	"github.com/mattn/lastpass-go"
	"github.com/spectralops/teller/pkg/core"
)

func TestLastPass(t *testing.T) {
	path := "settings/prod/billing-svc"
	pathmap := "settings/prod/billing-svc/all"

	notes := `MG_KEY: shazam
Sec key: secret-from-note
`
	lastPassProvider := LastPass{
		accounts: map[string]*lastpass.Account{
			path: {
				Id:       "id",
				Name:     "secret-name",
				Username: "username",
				Password: "shazam",
				Url:      "http://test.com",
				Group:    "",
				Notes:    notes,
			},
			pathmap: {
				Id:       "id-2",
				Name:     "secret-name-2",
				Username: "shazam",
				Password: "username-2",
				Url:      "http://test.com",
				Group:    "",
				Notes:    notes,
			},
		},
		logger: GetTestLogger(),
	}
	AssertProvider(t, &lastPassProvider, false)

	p := core.NewPopulate(map[string]string{"stage": "prod"})
	kpmap := p.KeyPath(core.KeyPath{Field: "MG_KEY", Path: "settings/{{stage}}/billing-svc/all", Decrypt: true})

	ents, err := lastPassProvider.GetMapping(kpmap)
	assert.Nil(t, err)
	assert.Equal(t, len(ents), 5)
	assert.Equal(t, ents[0].Value, "secret-name-2")
	assert.Equal(t, ents[1].Value, "username-2")
	assert.Equal(t, ents[2].Value, "http://test.com")
	assert.Equal(t, ents[3].Value, "shazam")
	assert.Equal(t, ents[4].Value, "secret-from-note")
	assert.Equal(t, ents[4].Key, "Sec_key")
}

func TestLastPassFailures(t *testing.T) {
	lastPassProvider := LastPass{
		accounts: map[string]*lastpass.Account{},
		logger:   GetTestLogger(),
	}

	_, err := lastPassProvider.Get(core.KeyPath{Env: "MG_KEY", Path: "settings/{{stage}}/billing-svc"})
	assert.NotNil(t, err)
	_, err = lastPassProvider.GetMapping(core.KeyPath{Env: "MG_KEY", Path: "settings/{{stage}}/billing-svc"})
	assert.NotNil(t, err)
}
