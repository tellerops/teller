//go:build integration
// +build integration

package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
	"github.com/spectralops/teller/pkg/providers"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type ConsulData = map[string]interface{}

func TestGetConsul(t *testing.T) {
	ctx := context.Background()
	const testToken = "vault-token"

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			AlwaysPullImage: false,
			Image:           "consul:1.9.5",
			ExposedPorts:    []string{"8500/tcp"},
			Env:             map[string]string{},
			WaitingFor:      wait.ForLog("Started gRPC server").WithStartupTimeout(20 * time.Second)},

		Started: true,
	}

	vaultContainer, err := testcontainers.GenericContainer(ctx, req)
	assert.NoError(t, err)
	defer vaultContainer.Terminate(ctx) //nolint

	ip, err := vaultContainer.Host(ctx)
	assert.NoError(t, err)
	port, err := vaultContainer.MappedPort(ctx, "8500")
	assert.NoError(t, err)
	host := fmt.Sprintf("http://%s:%s", ip, port.Port())

	//
	// pre-insert data w/API
	//
	config := &api.Config{Address: host}
	client, err := api.NewClient(config)
	assert.NoError(t, err)
	_, err = client.KV().Put(&api.KVPair{Key: "path/to/svc/MG_KEY", Value: []byte("value1")}, &api.WriteOptions{})
	assert.NoError(t, err)

	//
	// use provider to read data
	//
	t.Setenv("CONSUL_HTTP_ADDR", host)
	p, err := providers.NewConsul(logging.New())
	assert.NoError(t, err)
	kvp := core.KeyPath{Env: "MG_KEY", Path: "path/to/svc/MG_KEY"}
	res, err := p.Get(kvp)

	assert.NoError(t, err)
	assert.Equal(t, "MG_KEY", res.Key)
	assert.Equal(t, "value1", res.Value)
	assert.Equal(t, "path/to/svc/MG_KEY", res.ResolvedPath)

	err = p.Put(kvp, "changed-secret")
	assert.NoError(t, err)

	res, err = p.Get(kvp)
	assert.NoError(t, err)
	assert.Equal(t, "MG_KEY", res.Key)
	assert.Equal(t, "changed-secret", res.Value)
	assert.Equal(t, "path/to/svc/MG_KEY", res.ResolvedPath)

	err = p.PutMapping(kvp.SwitchPath("path/to/allmap"), map[string]string{"K1": "v1", "K2": "v2"})
	assert.NoError(t, err)
	ents, err := p.GetMapping(kvp.SwitchPath("path/to/allmap"))
	assert.NoError(t, err)
	assert.Equal(t, 2, len(ents))
	assert.Equal(t, "K1", ents[1].Key)
	assert.Equal(t, "K2", ents[0].Key)
	assert.Equal(t, "v1", ents[1].Value)
	assert.Equal(t, "v2", ents[0].Value)
}
