package pkg

import (
	"bytes"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/spectralops/teller/pkg/core"
)

func TestPorcelainNonInteractive(t *testing.T) {
	var b bytes.Buffer
	p := Porcelain{
		Out: &b,
	}
	p.DidCreateNewFile("myfile.yaml")
	assert.Equal(t, b.String(), "Created file: myfile.yaml\n")
	b.Reset()

	p.PrintContext("project", "place")
	assert.Equal(t, b.String(), "-*- teller: loaded variables for project using place -*-\n")
	b.Reset()

	p.PrintEntries([]core.EnvEntry{
		{IsFound: true, Key: "k", Value: "v", ProviderName: "test-provider", ResolvedPath: "path/kv"},
	})
	assert.Equal(t, b.String(), "[test-provider path/kv] k = v*****\n")
	b.Reset()
}

func TestPorcelainPrintDrift(t *testing.T) {
	var b bytes.Buffer
	p := Porcelain{
		Out: &b,
	}
	p.PrintDrift([]core.DriftedEntry{
		{
			Diff: "changed",
			Source: core.EnvEntry{

				Source: "s1", Key: "k", Value: "v", ProviderName: "test-provider", ResolvedPath: "path/kv",
			},

			Target: core.EnvEntry{

				Sink: "s1", Key: "k", Value: "x", ProviderName: "test-provider", ResolvedPath: "path/kv",
			},
		},
		{
			Diff: "changed",
			Source: core.EnvEntry{
				Source: "s2", Key: "k2", Value: "1", ProviderName: "test-provider", ResolvedPath: "path/kv",
			},

			Target: core.EnvEntry{
				Sink: "s2", Key: "k2", Value: "2", ProviderName: "test-provider", ResolvedPath: "path/kv",
			},
		},
	})
	assert.Equal(t, b.String(), `Drifts detected: 2

changed [s1] test-provider k v***** != test-provider k x*****
changed [s2] test-provider k2 1***** != test-provider k2 2*****
`)
}
