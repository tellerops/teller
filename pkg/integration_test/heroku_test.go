//go:build integration_api
// +build integration_api

package integration_test

import (
	"context"
	"os"
	"testing"

	heroku "github.com/heroku/heroku-go/v5"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
	"github.com/spectralops/teller/pkg/providers"
	"github.com/stretchr/testify/assert"
)

func TestGetHeroku(t *testing.T) {

	//
	// TEST PRECONDITION:
	//
	// HEROKU_API_KEY populated
	// 'teller-heroku-integration' app exists
	//

	//
	// pre-insert data w/API
	//
	heroku.DefaultTransport.BearerToken = os.Getenv("HEROKU_API_KEY")
	svc := heroku.NewService(heroku.DefaultClient)

	v := "value1"
	_, err := svc.ConfigVarUpdate(context.TODO(), "teller-heroku-integration", map[string]*string{"MG_KEY": &v, "K1": nil, "K2": nil})
	assert.NoError(t, err)

	//
	// use provider to read data
	//
	p, err := providers.NewHeroku(logging.New())
	assert.NoError(t, err)
	kvp := core.KeyPath{Env: "MG_KEY", Path: "teller-heroku-integration"}
	res, err := p.Get(kvp)

	assert.NoError(t, err)
	assert.Equal(t, "MG_KEY", res.Key)
	assert.Equal(t, "value1", res.Value)
	assert.Equal(t, "teller-heroku-integration", res.ResolvedPath)

	err = p.Put(kvp, "changed-secret")
	assert.NoError(t, err)

	res, err = p.Get(kvp)
	assert.NoError(t, err)
	assert.Equal(t, "MG_KEY", res.Key)
	assert.Equal(t, "changed-secret", res.Value)
	assert.Equal(t, "teller-heroku-integration", res.ResolvedPath)

	err = p.PutMapping(kvp, map[string]string{"K1": "v1", "K2": "v2"})
	assert.NoError(t, err)
	ents, err := p.GetMapping(kvp)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(ents))
	assert.Equal(t, "MG_KEY", ents[0].Key)
	assert.Equal(t, "changed-secret", ents[0].Value)
	assert.Equal(t, "K1", ents[2].Key)
	assert.Equal(t, "v1", ents[2].Value)
	assert.Equal(t, "K2", ents[1].Key)
	assert.Equal(t, "v2", ents[1].Value)
}
