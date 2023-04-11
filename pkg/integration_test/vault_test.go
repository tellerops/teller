//go:build integration
// +build integration

package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
	"github.com/spectralops/teller/pkg/providers"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type SecretData = map[string]interface{}

func TestGetVaultSecret(t *testing.T) {
	ctx := context.Background()
	const testToken = "vault-token"

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			AlwaysPullImage: false,
			Image:           "vault:1.6.3",
			ExposedPorts:    []string{"8200/tcp"},
			Env:             map[string]string{"VAULT_DEV_ROOT_TOKEN_ID": testToken},
			WaitingFor:      wait.ForLog("Vault server started!").WithStartupTimeout(20 * time.Second)},

		Started: true,
	}

	vaultContainer, err := testcontainers.GenericContainer(ctx, req)
	assert.NoError(t, err)
	defer vaultContainer.Terminate(ctx) //nolint

	ip, err := vaultContainer.Host(ctx)
	assert.NoError(t, err)
	port, err := vaultContainer.MappedPort(ctx, "8200")
	assert.NoError(t, err)
	host := fmt.Sprintf("http://%s:%s", ip, port.Port())

	//
	// pre-insert data w/API
	//
	config := &api.Config{Address: host}
	client, err := api.NewClient(config)
	assert.NoError(t, err)
	client.SetToken(testToken)
	secretData := SecretData{
		"MG_KEY": "value1",
	}
	_, err = client.Logical().Write("secret/data/settings/prod/billing-svc", SecretData{"data": secretData})
	assert.NoError(t, err)

	//
	// use provider to read data
	//
	t.Setenv("VAULT_ADDR", host)
	t.Setenv("VAULT_TOKEN", testToken)
	p, err := providers.NewHashicorpVault(logging.New())
	assert.NoError(t, err)
	kvp := core.KeyPath{Env: "MG_KEY", Path: "secret/data/settings/prod/billing-svc"}
	res, err := p.Get(kvp)

	assert.NoError(t, err)
	assert.Equal(t, "MG_KEY", res.Key)
	assert.Equal(t, "value1", res.Value)
	assert.Equal(t, "secret/data/settings/prod/billing-svc", res.ResolvedPath)

	err = (p.(*providers.HashicorpVault)).Put(kvp, "changed-secret")
	assert.NoError(t, err)

	res, err = p.Get(kvp)
	assert.NoError(t, err)
	assert.Equal(t, "MG_KEY", res.Key)
	assert.Equal(t, "changed-secret", res.Value)
	assert.Equal(t, "secret/data/settings/prod/billing-svc", res.ResolvedPath)
}
