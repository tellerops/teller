package pkg

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
	"github.com/spectralops/teller/pkg/providers"
)

// implements both Providers and Provider interface, for testing return only itself.
type InMemProvider struct {
	inmem       map[string]string
	alwaysError bool
}

func (im *InMemProvider) Put(p core.KeyPath, val string) error {
	return fmt.Errorf("provider %q does not implement write yet", im.Name())
}
func (im *InMemProvider) PutMapping(p core.KeyPath, m map[string]string) error {
	return fmt.Errorf("provider %q does not implement write yet", im.Name())
}

func (im *InMemProvider) Delete(kp core.KeyPath) error {
	if im.alwaysError {
		return errors.New("error")
	}

	k := kp.EffectiveKey()

	delete(im.inmem, fmt.Sprintf("%s/%s", kp.Path, k))
	return nil
}

func (im *InMemProvider) DeleteMapping(kp core.KeyPath) error {
	if im.alwaysError {
		return errors.New("error")
	}

	for key := range im.inmem {
		if !strings.HasPrefix(key, kp.Path) {
			continue
		}

		delete(im.inmem, key)
	}

	return nil
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

func (im *InMemProvider) Meta() core.MetaInfo {
	return core.MetaInfo{}
}

// nolint
func init() {
	inmemProviderMeta := core.MetaInfo{
		Name:        "inmem-provider",
		Description: "test-provider",
	}

	inmemProviderErrorMeta := core.MetaInfo{
		Name:        "inmem-provider-error",
		Description: "test-provider-error",
	}

	providers.RegisterProvider(inmemProviderMeta, NewInMemProvider)
	providers.RegisterProvider(inmemProviderErrorMeta, NewInMemErrorProvider)
}

func NewInMemProvider(logger logging.Logger) (core.Provider, error) {
	return &InMemProvider{
		inmem: map[string]string{
			"prod/billing/FOO":          "foo_shazam",
			"prod/billing/MG_KEY":       "mg_shazam",
			"prod/billing/BEFORE_REMAP": "test_env_remap",
		},
		alwaysError: false,
	}, nil

}

func NewInMemErrorProvider(logger logging.Logger) (core.Provider, error) {
	return &InMemProvider{
		inmem: map[string]string{
			"prod/billing/FOO":          "foo_shazam",
			"prod/billing/MG_KEY":       "mg_shazam",
			"prod/billing/BEFORE_REMAP": "test_env_remap",
		},
		alwaysError: true,
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

func getLogger() logging.Logger {
	logger := logging.New()
	logger.SetLevel("null")
	return logger
}

func TestTellerExports(t *testing.T) {
	tl := Teller{
		Logger:    getLogger(),
		Entries:   []core.EnvEntry{},
		Providers: &BuiltinProviders{},
	}

	b := tl.ExportEnv()
	assert.Equal(t, b, "#!/bin/sh\n")

	tl = Teller{
		Logger: getLogger(),
		Entries: []core.EnvEntry{
			{Key: "k", Value: "v", ProviderName: "test-provider", ResolvedPath: "path/kv"},
		},
	}

	b = tl.ExportEnv()
	assert.Equal(t, b, "#!/bin/sh\nexport k='v'\n")

	b, err := tl.ExportYAML()
	assert.NoError(t, err)
	assert.Equal(t, b, "k: v\n")
	b, err = tl.ExportJSON()
	assert.NoError(t, err)
	assert.Equal(t, b, "{\n  \"k\": \"v\"\n}")
}

func TestTellerShExportEscaped(t *testing.T) {
	tl := Teller{
		Logger:    getLogger(),
		Entries:   []core.EnvEntry{},
		Providers: &BuiltinProviders{},
	}

	b := tl.ExportEnv()
	assert.Equal(t, b, "#!/bin/sh\n")

	tl = Teller{
		Logger: getLogger(),
		Entries: []core.EnvEntry{
			{Key: "k", Value: `()"';@  \(\)\"\'\;\@`, ProviderName: "test-provider", ResolvedPath: "path/kv"},
		},
	}

	b = tl.ExportEnv()
	assert.Equal(t, b, "#!/bin/sh\nexport k='()\"'\"'\"';@  \\(\\)\\\"\\'\"'\"'\\;\\@'\n")
}

func TestTellerCollect(t *testing.T) {
	var b bytes.Buffer
	tl := Teller{
		Logger:    getLogger(),
		Providers: &BuiltinProviders{},
		Porcelain: &Porcelain{
			Out: &b,
		},
		Populate: core.NewPopulate(map[string]string{"stage": "prod"}),
		Config: &TellerFile{
			Project:    "test-project",
			LoadedFrom: "nowhere",
			Providers: map[string]MappingConfig{
				"inmem-provider": {
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
	assert.Equal(t, tl.Entries[0].ProviderName, "inmem-provider")

	assert.Equal(t, tl.Entries[1].Key, "FOO_BAR")
	assert.Equal(t, tl.Entries[1].Value, "foo_shazam")
	assert.Equal(t, tl.Entries[1].ResolvedPath, "prod/billing/FOO")
	assert.Equal(t, tl.Entries[1].ProviderName, "inmem-provider")
}

func TestTellerCollectWithSync(t *testing.T) {
	var b bytes.Buffer
	tl := Teller{
		Logger:    getLogger(),
		Providers: &BuiltinProviders{},
		Porcelain: &Porcelain{
			Out: &b,
		},
		Populate: core.NewPopulate(map[string]string{"stage": "prod"}),
		Config: &TellerFile{
			Project:    "test-project",
			LoadedFrom: "nowhere",
			Providers: map[string]MappingConfig{
				"inmem-provider": {
					EnvMapping: &core.KeyPath{
						Path: "{{stage}}/billing",
						Remap: &map[string]string{
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
	assert.Equal(t, tl.Entries[0].ProviderName, "inmem-provider")

	assert.Equal(t, tl.Entries[1].Key, "prod/billing/MG_KEY")
	assert.Equal(t, tl.Entries[1].Value, "mg_shazam")
	assert.Equal(t, tl.Entries[1].ResolvedPath, "prod/billing")
	assert.Equal(t, tl.Entries[1].ProviderName, "inmem-provider")

	assert.Equal(t, tl.Entries[2].Key, "prod/billing/FOO")
	assert.Equal(t, tl.Entries[2].Value, "foo_shazam")
	assert.Equal(t, tl.Entries[2].ResolvedPath, "prod/billing")
	assert.Equal(t, tl.Entries[2].ProviderName, "inmem-provider")
}

func TestTellerCollectWithSyncRemapWith(t *testing.T) {
	var b bytes.Buffer
	tl := Teller{
		Logger:    getLogger(),
		Providers: &BuiltinProviders{},
		Porcelain: &Porcelain{
			Out: &b,
		},
		Populate: core.NewPopulate(map[string]string{"stage": "prod"}),
		Config: &TellerFile{
			Project:    "test-project",
			LoadedFrom: "nowhere",
			Providers: map[string]MappingConfig{
				"inmem-provider": {
					EnvMapping: &core.KeyPath{
						Path: "{{stage}}/billing",
						RemapWith: &map[string]core.RemapKeyPath{
							"prod/billing/BEFORE_REMAP": {
								Field:      "prod/billing/REMAPED",
								Severity:   core.None,
								RedactWith: "-",
							},
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
	assert.Equal(t, tl.Entries[0].Severity, core.None)
	assert.Equal(t, tl.Entries[0].RedactWith, "-")
	assert.Equal(t, tl.Entries[0].Value, "test_env_remap")
	assert.Equal(t, tl.Entries[0].ResolvedPath, "prod/billing")
	assert.Equal(t, tl.Entries[0].ProviderName, "inmem-provider")

	assert.Equal(t, tl.Entries[1].Key, "prod/billing/MG_KEY")
	assert.Equal(t, tl.Entries[1].Value, "mg_shazam")
	assert.Equal(t, tl.Entries[1].ResolvedPath, "prod/billing")
	assert.Equal(t, tl.Entries[1].ProviderName, "inmem-provider")

	assert.Equal(t, tl.Entries[2].Key, "prod/billing/FOO")
	assert.Equal(t, tl.Entries[2].Value, "foo_shazam")
	assert.Equal(t, tl.Entries[2].ResolvedPath, "prod/billing")
	assert.Equal(t, tl.Entries[2].ProviderName, "inmem-provider")
}

func TestTellerCollectWithErrors(t *testing.T) {
	var b bytes.Buffer
	tl := Teller{
		Logger:    getLogger(),
		Providers: &BuiltinProviders{},
		Porcelain: &Porcelain{
			Out: &b,
		},
		Populate: core.NewPopulate(map[string]string{"stage": "prod"}),
		Config: &TellerFile{
			Project:    "test-project",
			LoadedFrom: "nowhere",
			Providers: map[string]MappingConfig{
				"inmem-provider-error": {
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
		Logger:  getLogger(),
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

func TestTellerEntriesOutputSort(t *testing.T) {
	var b bytes.Buffer

	entries := []core.EnvEntry{}

	tl := Teller{
		Logger:  getLogger(),
		Entries: entries,
		Porcelain: &Porcelain{
			Out: &b,
		},
		Config: &TellerFile{
			Project:    "test-project",
			LoadedFrom: "nowhere",
		},
	}

	tl.Entries = append(tl.Entries, core.EnvEntry{
		IsFound: true,
		Key:     "c", Value: "c", ProviderName: "test-provider", ResolvedPath: "path/kv",
	})
	tl.Entries = append(tl.Entries, core.EnvEntry{
		IsFound: true,
		Key:     "a", Value: "a", ProviderName: "test-provider", ResolvedPath: "path/kv",
	})
	tl.Entries = append(tl.Entries, core.EnvEntry{
		IsFound: true,
		Key:     "b", Value: "b", ProviderName: "test-provider", ResolvedPath: "path/kv",
	})
	tl.Entries = append(tl.Entries, core.EnvEntry{
		IsFound: true,
		Key:     "k", Value: "v", ProviderName: "alpha", ResolvedPath: "path/kv",
	})
	tl.Entries = append(tl.Entries, core.EnvEntry{
		IsFound: true,
		Key:     "k", Value: "v", ProviderName: "BETA", ResolvedPath: "path/kv",
	})

	tl.PrintEnvKeys()
	assert.Equal(t, b.String(), "-*- teller: loaded variables for test-project using nowhere -*-\n\n[alpha path/kv] k = v*****\n[BETA path/kv] k = v*****\n[test-provider path/kv] a = a*****\n[test-provider path/kv] b = b*****\n[test-provider path/kv] c = c*****\n")
}

func TestTellerDrift(t *testing.T) {
	tl := Teller{
		Logger: getLogger(),
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

	tl := NewTeller(tlrfile, []string{}, false, getLogger())

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

	tl := NewTeller(tlrfile, []string{}, false, getLogger())

	err = os.WriteFile("../fixtures/sync/target.env", []byte(`
FOO=1
`), 0644)
	assert.NoError(t, err)

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

func TestTemplateFile(t *testing.T) {

	tlrfile, err := NewTellerFile("../fixtures/sync/teller.yml")
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	tl := NewTeller(tlrfile, []string{}, false, getLogger())
	tl.Entries = append(tl.Entries, core.EnvEntry{Key: "TEST-PLACEHOLDER", Value: "secret-here"})

	tempFolder, _ := os.MkdirTemp(os.TempDir(), "test-template")
	defer os.RemoveAll(tempFolder)

	templatePath := filepath.Join(tempFolder, "target.tpl")      // prepare template file path
	destinationPath := filepath.Join(tempFolder, "starget.envs") // prepare destination file path

	err = os.WriteFile(templatePath, []byte(`Hello, {{.Teller.EnvByKey "TEST-PLACEHOLDER" "default-value" }}!`), 0644)
	assert.NoError(t, err)

	err = tl.templateFile(templatePath, destinationPath)
	assert.NoError(t, err)

	txt, err := ioutil.ReadFile(destinationPath)
	assert.NoError(t, err)
	assert.Equal(t, string(txt), "Hello, secret-here!")

}

func TestTemplateFolder(t *testing.T) {

	tlrfile, err := NewTellerFile("../fixtures/sync/teller.yml")
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	tl := NewTeller(tlrfile, []string{}, false, getLogger())
	tl.Entries = append(tl.Entries, core.EnvEntry{Key: "TEST-PLACEHOLDER", Value: "secret-here"})
	tl.Entries = append(tl.Entries, core.EnvEntry{Key: "TEST-PLACEHOLDER-2", Value: "secret2-here"})

	rootTempDir := os.TempDir()
	tempFolder, _ := os.MkdirTemp(rootTempDir, "test-template") // create temp root folder
	// Create template folders structure
	templateFolder := filepath.Join(tempFolder, "from")
	err = os.MkdirAll(templateFolder, os.ModePerm)
	assert.NoError(t, err)
	err = os.MkdirAll(filepath.Join(templateFolder, "folder1", "folder2"), os.ModePerm)
	assert.NoError(t, err)

	// copy to:
	copyToFolder := filepath.Join(tempFolder, "to")

	err = os.MkdirAll(copyToFolder, os.ModePerm)
	assert.NoError(t, err)

	defer os.RemoveAll(tempFolder)

	err = os.WriteFile(filepath.Join(templateFolder, "target.tpl"), []byte(`Hello, {{.Teller.EnvByKey "TEST-PLACEHOLDER" "default-value" }}!`), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(templateFolder, "folder1", "folder2", "target2.tpl"), []byte(`Hello, {{.Teller.EnvByKey "TEST-PLACEHOLDER-2" "default-value" }}!`), 0644)
	assert.NoError(t, err)

	err = tl.templateFolder(templateFolder, copyToFolder)
	assert.NoError(t, err)
	fmt.Println(copyToFolder)

	txt, err := ioutil.ReadFile(filepath.Join(copyToFolder, "target.tpl"))
	assert.NoError(t, err)
	assert.Equal(t, string(txt), "Hello, secret-here!")

	txt, err = ioutil.ReadFile(filepath.Join(copyToFolder, "folder1", "folder2", "target2.tpl"))
	assert.NoError(t, err)
	assert.Equal(t, string(txt), "Hello, secret2-here!")

}

func TestTellerDelete(t *testing.T) {
	fooPath := "/sample/path/FOO"
	p := &InMemProvider{
		inmem: map[string]string{
			fooPath:            "foo",
			"/sample/path/BAR": "bar",
		},
		alwaysError: false,
	}
	tl := Teller{
		Logger:    getLogger(),
		Providers: p,
		Porcelain: &Porcelain{
			Out: ioutil.Discard,
		},
		Populate: core.NewPopulate(map[string]string{"stage": "prod"}),
		Config: &TellerFile{
			Project:    "test-project",
			LoadedFrom: "nowhere",
			Providers: map[string]MappingConfig{
				"inmem-provider": {
					Env: &map[string]core.KeyPath{
						"FOO": {
							Path: "/sample/path",
							Env:  "FOO",
						},
						"BAR": {
							Path: "/sample/path",
							Env:  "BAR",
						},
					},
				},
			},
		},
	}

	keysToDelete := []string{"FOO"}
	err := tl.Delete(keysToDelete, []string{"inmem-provider"}, "", false)
	assert.NoError(t, err)

	assert.Equal(t, len(p.inmem), 1)
	_, ok := p.inmem[fooPath]
	assert.False(t, ok)

	keysToDelete = []string{"BAR"}
	err = tl.Delete(keysToDelete, []string{"inmem-provider"}, "/sample/path", false)
	assert.NoError(t, err)

	assert.Equal(t, len(p.inmem), 0)
}

func TestTellerDeleteAll(t *testing.T) {
	p := &InMemProvider{
		inmem: map[string]string{
			"/sample/path/FOO": "foo",
			"/sample/path/BAR": "bar",
		},
		alwaysError: false,
	}
	tl := Teller{
		Logger:    getLogger(),
		Providers: p,
		Porcelain: &Porcelain{
			Out: ioutil.Discard,
		},
		Populate: core.NewPopulate(map[string]string{"stage": "prod"}),
		Config: &TellerFile{
			Project:    "test-project",
			LoadedFrom: "nowhere",
			Providers: map[string]MappingConfig{
				"inmem-provider": {
					Env: &map[string]core.KeyPath{
						"FOO": {
							Path: "/sample/path",
							Env:  "FOO",
						},
						"BAR": {
							Path: "/sample/path",
							Env:  "BAR",
						},
					},
				},
			},
		},
	}

	err := tl.Delete([]string{}, []string{"inmem-provider"}, "/sample/path", true)
	assert.NoError(t, err)

	assert.Equal(t, len(p.inmem), 0)
}
