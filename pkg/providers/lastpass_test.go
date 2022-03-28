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

	// collect all the fields that returned from GetMapping
	allValues := []string{}
	for _, f := range ents {
		allValues = append(allValues, f.Value)
	}

	assert.Contains(t, allValues, "secret-name-2")
	assert.Contains(t, allValues, "username-2")
	assert.Contains(t, allValues, "http://test.com")
	assert.Contains(t, allValues, "shazam")
	assert.Contains(t, allValues, "secret-from-note")

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
