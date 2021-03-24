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
		{Key: "k", Value: "v", Provider: "test-provider", ResolvedPath: "path/kv"},
	})
	assert.Equal(t, b.String(), "[test-provider path/kv] k = v*****\n")
	b.Reset()
}
