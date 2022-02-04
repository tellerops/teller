package pkg

import (
	"bytes"
	"errors"
	"fmt"
	"os"
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

func (im *InMemProvider) Put(p core.KeyPath, val string) error {
	return fmt.Errorf("%v does not implement write yet", im.Name())
}
func (im *InMemProvider) PutMapping(p core.KeyPath, m map[string]string) error {
	return fmt.Errorf("%v does not implement write yet", im.Name())
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
			"prod/billing/FOO":          "foo_shazam",
			"prod/billing/MG_KEY":       "mg_shazam",
			"prod/billing/BEFORE_REMAP": "test_env_remap",
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
			ProviderName: im.Name(),
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
		ProviderName: im.Name(),
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
			{Key: "k", Value: "v", ProviderName: "test-provider", ResolvedPath: "path/kv"},
		},
	}

	b = tl.ExportEnv()
	assert.Equal(t, b, "#!/bin/sh\nexport k=v\n")

	b, err := tl.ExportYAML()
	assert.NoError(t, err)
	assert.Equal(t, b, "k: v\n")
	b, err = tl.ExportJSON()
	assert.NoError(t, err)
	assert.Equal(t, b, "{\n  \"k\": \"v\"\n}")
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
	assert.Equal(t, tl.Entries[0].ProviderName, "inmem")

	assert.Equal(t, tl.Entries[1].Key, "FOO_BAR")
	assert.Equal(t, tl.Entries[1].Value, "foo_shazam")
	assert.Equal(t, tl.Entries[1].ResolvedPath, "prod/billing/FOO")
	assert.Equal(t, tl.Entries[1].ProviderName, "inmem")
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
						Remap: map[string]string{
							"prod/billing/BEFORE_REMAP": "prod/billing/REMAPED",
						},
					},
				},
			},
		},
	}
	err := tl.Collect()
	assert.Nil(t, err)
	assert.Equal(t, len(tl.Entries), 3)
	assert.Equal(t, tl.Entries[0].Key, "prod/billing/REMAPED")
	assert.Equal(t, tl.Entries[0].Value, "test_env_remap")
	assert.Equal(t, tl.Entries[0].ResolvedPath, "prod/billing")
	assert.Equal(t, tl.Entries[0].ProviderName, "inmem")

	assert.Equal(t, tl.Entries[1].Key, "prod/billing/MG_KEY")
	assert.Equal(t, tl.Entries[1].Value, "mg_shazam")
	assert.Equal(t, tl.Entries[1].ResolvedPath, "prod/billing")
	assert.Equal(t, tl.Entries[1].ProviderName, "inmem")

	assert.Equal(t, tl.Entries[2].Key, "prod/billing/FOO")
	assert.Equal(t, tl.Entries[2].Value, "foo_shazam")
	assert.Equal(t, tl.Entries[2].ResolvedPath, "prod/billing")
	assert.Equal(t, tl.Entries[2].ProviderName, "inmem")
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
		IsFound: true,
		Key:     "k", Value: "v", ProviderName: "test-provider", ResolvedPath: "path/kv",
	})

	tl.PrintEnvKeys()
	assert.Equal(t, b.String(), "-*- teller: loaded variables for test-project using nowhere -*-\n\n[test-provider path/kv] k = v*****\n")

}

func TestTellerDrift(t *testing.T) {
	tl := Teller{
		Entries: []core.EnvEntry{
			{Key: "k", Value: "v", Source: "s1", ProviderName: "test-provider", ResolvedPath: "path/kv"},
			{Key: "k", Value: "v", Sink: "s1", ProviderName: "test-provider2", ResolvedPath: "path/kv"},
			{Key: "kX", Value: "vx", Source: "s1", ProviderName: "test-provider", ResolvedPath: "path/kv"},
			{Key: "kX", Value: "CHANGED", Sink: "s1", ProviderName: "test-provider2", ResolvedPath: "path/kv"},

			// these do not have sink/source
			{Key: "k--", Value: "00", ProviderName: "test-provider", ResolvedPath: "path/kv"},
			{Key: "k--", Value: "11", ProviderName: "test-provider2", ResolvedPath: "path/kv"},
		},
	}

	drifts := tl.Drift([]string{})

	assert.Equal(t, len(drifts), 1)
	d := drifts[0]
	assert.Equal(t, d.Source.Value, "vx")
	assert.Equal(t, d.Target.Value, "CHANGED")
}

func TestTellerMirrorDrift(t *testing.T) {
	tlrfile, err := NewTellerFile("../fixtures/mirror-drift/teller.yml")
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	tl := NewTeller(tlrfile, []string{}, false)

	drifts, err := tl.MirrorDrift("source", "target")
	assert.NoError(t, err)

	assert.Equal(t, len(drifts), 2)
	d := drifts[0]
	assert.Equal(t, d.Source.Key, "THREE")
	assert.Equal(t, d.Source.Value, "3")
	assert.Equal(t, d.Diff, "missing")
	assert.Equal(t, d.Target.Value, "")

	d = drifts[1]
	assert.Equal(t, d.Source.Key, "ONE")
	assert.Equal(t, d.Source.Value, "1")
	assert.Equal(t, d.Diff, "changed")
	assert.Equal(t, d.Target.Value, "5")
}

func TestTellerSync(t *testing.T) {
	tlrfile, err := NewTellerFile("../fixtures/sync/teller.yml")
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	tl := NewTeller(tlrfile, []string{}, false)

	//nolint
	err = os.WriteFile("../fixtures/sync/target.env", []byte(`
FOO=1
`), 0644)
	assert.NoError(t, err)

	//nolint
	err = os.WriteFile("../fixtures/sync/target2.env", []byte(`
FOO=2
`), 0644)

	assert.NoError(t, err)

	err = tl.Sync("source", []string{"target", "target2"}, true)

	assert.NoError(t, err)

	content, err := os.ReadFile("../fixtures/sync/target.env")
	assert.NoError(t, err)

	assert.Equal(t, string(content), `FOO="1"
ONE="1"
THREE="3"
TWO="2"`)

	content, err = os.ReadFile("../fixtures/sync/target2.env")
	assert.NoError(t, err)

	assert.Equal(t, string(content), `FOO="2"
ONE="1"
THREE="3"
TWO="2"`)
}
