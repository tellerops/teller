package pkg

import (
	"bytes"
	"errors"
	"sort"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/spectralops/teller/pkg/core"
)

// implements both Providers and Provider interface, for testing return only itself.
type InMemProvider struct {
	inmem       map[string]string
	alwaysError bool
}

func (im *InMemProvider) GetProvider(name string) (core.Provider, error) {
	return im, nil //hardcode to return self
}
func (im *InMemProvider) ProviderHumanToMachine() map[string]string {
	return map[string]string{
		"Inmem": "inmem",
	}
}

func (im *InMemProvider) Name() string {
	return "inmem"
}
func NewInMemProvider(alwaysError bool) (Providers, error) {
	return &InMemProvider{
		inmem: map[string]string{
			"prod/billing/FOO":    "foo_shazam",
			"prod/billing/MG_KEY": "mg_shazam",
		},
		alwaysError: alwaysError,
	}, nil

}
func (im *InMemProvider) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {
	if im.alwaysError {
		return nil, errors.New("error")
	}

	var entries []core.EnvEntry

	for k, v := range im.inmem {
		entries = append(entries, core.EnvEntry{
			Key:          k,
			Value:        v,
			ResolvedPath: p.Path,
			Provider:     im.Name(),
		})
	}
	sort.Sort(core.EntriesByKey(entries))
	return entries, nil
}
func (im *InMemProvider) Get(p core.KeyPath) (*core.EnvEntry, error) {
	if im.alwaysError {
		return nil, errors.New("error")
	}
	s := im.inmem[p.Path]
	return &core.EnvEntry{
		Key:          p.Env,
		Value:        s,
		ResolvedPath: p.Path,
		Provider:     im.Name(),
	}, nil
}

func TestTellerExports(t *testing.T) {
	tl := Teller{
		Entries:   []core.EnvEntry{},
		Providers: &BuiltinProviders{},
	}

	b := tl.ExportEnv()
	assert.Equal(t, b, "#!/bin/sh\n")

	tl = Teller{
		Entries: []core.EnvEntry{
			{Key: "k", Value: "v", Provider: "test-provider", ResolvedPath: "path/kv"},
		},
	}

	b = tl.ExportEnv()
	assert.Equal(t, b, "#!/bin/sh\nexport k=v\n")
}

func TestTellerCollect(t *testing.T) {
	var b bytes.Buffer
	p, _ := NewInMemProvider(false)
	tl := Teller{
		Providers: p,
		Porcelain: &Porcelain{
			Out: &b,
		},
		Populate: core.NewPopulate(map[string]string{"stage": "prod"}),
		Config: &TellerFile{
			Project:    "test-project",
			LoadedFrom: "nowhere",
			Providers: map[string]MappingConfig{
				"inmem": {
					Env: &map[string]core.KeyPath{
						"MG_KEY": {
							Path: "{{stage}}/billing/MG_KEY",
						},
						"FOO_BAR": {
							Path: "{{stage}}/billing/FOO",
						},
					},
				},
			},
		},
	}
	err := tl.Collect()
	assert.Nil(t, err)
	assert.Equal(t, len(tl.Entries), 2)
	assert.Equal(t, tl.Entries[0].Key, "MG_KEY")
	assert.Equal(t, tl.Entries[0].Value, "mg_shazam")
	assert.Equal(t, tl.Entries[0].ResolvedPath, "prod/billing/MG_KEY")
	assert.Equal(t, tl.Entries[0].Provider, "inmem")

	assert.Equal(t, tl.Entries[1].Key, "FOO_BAR")
	assert.Equal(t, tl.Entries[1].Value, "foo_shazam")
	assert.Equal(t, tl.Entries[1].ResolvedPath, "prod/billing/FOO")
	assert.Equal(t, tl.Entries[1].Provider, "inmem")
}

func TestTellerCollectWithSync(t *testing.T) {
	var b bytes.Buffer
	p, _ := NewInMemProvider(false)
	tl := Teller{
		Providers: p,
		Porcelain: &Porcelain{
			Out: &b,
		},
		Populate: core.NewPopulate(map[string]string{"stage": "prod"}),
		Config: &TellerFile{
			Project:    "test-project",
			LoadedFrom: "nowhere",
			Providers: map[string]MappingConfig{
				"inmem": {
					EnvMapping: &core.KeyPath{
						Path: "{{stage}}/billing",
					},
				},
			},
		},
	}
	err := tl.Collect()
	assert.Nil(t, err)
	assert.Equal(t, len(tl.Entries), 2)
	assert.Equal(t, tl.Entries[0].Key, "prod/billing/MG_KEY")
	assert.Equal(t, tl.Entries[0].Value, "mg_shazam")
	assert.Equal(t, tl.Entries[0].ResolvedPath, "prod/billing")
	assert.Equal(t, tl.Entries[0].Provider, "inmem")

	assert.Equal(t, tl.Entries[1].Key, "prod/billing/FOO")
	assert.Equal(t, tl.Entries[1].Value, "foo_shazam")
	assert.Equal(t, tl.Entries[1].ResolvedPath, "prod/billing")
	assert.Equal(t, tl.Entries[1].Provider, "inmem")
}
func TestTellerCollectWithErrors(t *testing.T) {
	var b bytes.Buffer
	p, _ := NewInMemProvider(true)
	tl := Teller{
		Providers: p,
		Porcelain: &Porcelain{
			Out: &b,
		},
		Populate: core.NewPopulate(map[string]string{"stage": "prod"}),
		Config: &TellerFile{
			Project:    "test-project",
			LoadedFrom: "nowhere",
			Providers: map[string]MappingConfig{
				"inmem": {
					EnvMapping: &core.KeyPath{
						Path: "{{stage}}/billing",
					},
				},
			},
		},
	}
	err := tl.Collect()
	assert.NotNil(t, err)
}
func TestTellerPorcelainNonInteractive(t *testing.T) {
	var b bytes.Buffer

	entries := []core.EnvEntry{}

	tl := Teller{
		Entries: entries,
		Porcelain: &Porcelain{
			Out: &b,
		},
		Config: &TellerFile{
			Project:    "test-project",
			LoadedFrom: "nowhere",
		},
	}

	tl.PrintEnvKeys()
	assert.Equal(t, b.String(), "-*- teller: loaded variables for test-project using nowhere -*-\n\n")
	b.Reset()

	tl.Entries = append(tl.Entries, core.EnvEntry{
		Key: "k", Value: "v", Provider: "test-provider", ResolvedPath: "path/kv",
	})

	tl.PrintEnvKeys()
	assert.Equal(t, b.String(), "-*- teller: loaded variables for test-project using nowhere -*-\n\n[test-provider path/kv] k = v*****\n")

}
