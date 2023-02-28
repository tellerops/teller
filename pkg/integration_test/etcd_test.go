//go:build integration
// +build integration

package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
	"github.com/spectralops/teller/pkg/providers"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestGetEtcd(t *testing.T) {
	ctx := context.Background()
	const testToken = "vault-token"

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			AlwaysPullImage: false,
			Image:           "gcr.io/etcd-development/etcd:v3.3",
			ExposedPorts:    []string{"2379/tcp"},
			Cmd: []string{
				"etcd",
				"--name", "etcd0",
				"--advertise-client-urls", "http://0.0.0.0:2379",
				"--listen-client-urls", "http://0.0.0.0:2379",
			},
			Env:        map[string]string{},
			WaitingFor: wait.ForLog("ready to serve client requests").WithStartupTimeout(20 * time.Second)},
		Started: true,
	}

	vaultContainer, err := testcontainers.GenericContainer(ctx, req)
	assert.NoError(t, err)
	defer vaultContainer.Terminate(ctx) //nolint

	ip, err := vaultContainer.Host(ctx)
	assert.NoError(t, err)
	port, err := vaultContainer.MappedPort(ctx, "2379/tcp")
	assert.NoError(t, err)
	host := fmt.Sprintf("http://%s:%s", ip, port.Port())

	//
	// pre-insert data w/API
	//
	cfg := clientv3.Config{
		Endpoints: []string{host},
	}
	client, err := clientv3.New(cfg)
	assert.NoError(t, err)
	_, err = client.Put(context.TODO(), "path/to/svc/MG_KEY", "value1")
	assert.NoError(t, err)
	//
	// use provider to read data
	//
	t.Setenv("ETCDCTL_ENDPOINTS", host)
	p, err := providers.NewEtcd(logging.New())
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
