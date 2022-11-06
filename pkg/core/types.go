package core

import (
	"strings"

	"github.com/spectralops/teller/pkg/logging"
)

type Severity string

const (
	High   Severity = "high"
	Medium Severity = "medium"
	Low    Severity = "low"
	None   Severity = "none"
)

type RemapKeyPath struct {
	Field      string   `yaml:"field,omitempty"`
	Severity   Severity `yaml:"severity,omitempty"`
	RedactWith string   `yaml:"redact_with,omitempty"`
}

type KeyPath struct {
	Env        string                   `yaml:"env,omitempty"`
	Path       string                   `yaml:"path"`
	Field      string                   `yaml:"field,omitempty"`
	Remap      *map[string]string       `yaml:"remap,omitempty"`
	RemapWith  *map[string]RemapKeyPath `yaml:"remap_with,omitempty"`
	Decrypt    bool                     `yaml:"decrypt,omitempty"`
	Optional   bool                     `yaml:"optional,omitempty"`
	Severity   Severity                 `yaml:"severity,omitempty" default:"high"`
	RedactWith string                   `yaml:"redact_with,omitempty" default:"**REDACTED**"`
	Source     string                   `yaml:"source,omitempty"`
	Sink       string                   `yaml:"sink,omitempty"`
}

type WizardAnswers struct {
	Project      string
	Providers    []string
	ProviderKeys map[string]bool
	Confirm      bool
}

func (k *KeyPath) EffectiveKey() string {
	key := k.Env
	if k.Field != "" {
		key = k.Field
	}
	return key
}

func (k *KeyPath) EffectiveRemap() map[string]RemapKeyPath {
	remap := make(map[string]RemapKeyPath)
	if k.Remap != nil {
		for k, v := range *k.Remap {
			remap[k] = RemapKeyPath{Field: v}
		}
	} else if k.RemapWith != nil {
		remap = *k.RemapWith
	}
	return remap
}

func (k *KeyPath) Missing() EnvEntry {
	return EnvEntry{
		IsFound:      false,
		Key:          k.Env,
		Field:        k.Field,
		ResolvedPath: k.Path,
	}
}

func (k *KeyPath) Found(v string) EnvEntry {
	return EnvEntry{
		IsFound:      true,
		Key:          k.Env,
		Field:        k.Field,
		Value:        v,
		ResolvedPath: k.Path,
	}
}

// NOTE: consider doing what 'updateParams' does in these builders
func (k *KeyPath) FoundWithKey(key, v string) EnvEntry {
	return EnvEntry{
		IsFound:      true,
		Key:          key,
		Field:        k.Field,
		Value:        v,
		ResolvedPath: k.Path,
	}
}

func (k *KeyPath) WithEnv(env string) KeyPath {
	return KeyPath{
		Env:      env,
		Path:     k.Path,
		Field:    k.Field,
		Decrypt:  k.Decrypt,
		Optional: k.Optional,
		Source:   k.Source,
		Sink:     k.Sink,
	}
}
func (k *KeyPath) SwitchPath(path string) KeyPath {
	return KeyPath{
		Path:     path,
		Field:    k.Field,
		Env:      k.Env,
		Decrypt:  k.Decrypt,
		Optional: k.Optional,
		Source:   k.Source,
		Sink:     k.Sink,
	}
}

type DriftedEntriesBySource []DriftedEntry

func (a DriftedEntriesBySource) Len() int           { return len(a) }
func (a DriftedEntriesBySource) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a DriftedEntriesBySource) Less(i, j int) bool { return a[i].Source.Source < a[j].Source.Source }

type EntriesByProvider []EnvEntry

func (a EntriesByProvider) Len() int      { return len(a) }
func (a EntriesByProvider) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a EntriesByProvider) Less(i, j int) bool {
	firstProviderName := strings.ToLower(a[i].ProviderName)
	secondProviderName := strings.ToLower(a[j].ProviderName)
	if firstProviderName != secondProviderName {
		return firstProviderName < secondProviderName
	}
	return a[i].Key < a[j].Key
}

type EntriesByKey []EnvEntry

func (a EntriesByKey) Len() int           { return len(a) }
func (a EntriesByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a EntriesByKey) Less(i, j int) bool { return a[i].Key > a[j].Key }

type EntriesByValueSize []EnvEntry

func (a EntriesByValueSize) Len() int           { return len(a) }
func (a EntriesByValueSize) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a EntriesByValueSize) Less(i, j int) bool { return len(a[i].Value) > len(a[j].Value) }

type EnvEntry struct {
	Key          string
	Field        string
	Value        string
	ProviderName string
	ResolvedPath string
	Severity     Severity
	RedactWith   string
	Source       string
	Sink         string
	IsFound      bool
}

func (ee *EnvEntry) AddressingKeyPath() *KeyPath {
	return &KeyPath{
		Env:   ee.Key,
		Field: ee.Field,
		Path:  ee.ResolvedPath,
	}
}

type DriftedEntry struct {
	Diff   string
	Source EnvEntry
	Target EnvEntry
}
type EnvEntryLookup struct {
	Entries []EnvEntry
}

func (ee *EnvEntryLookup) EnvBy(key, provider, path, dflt string) string {
	for i := range ee.Entries {
		e := ee.Entries[i]
		if e.Key == key && e.ProviderName == provider && e.ResolvedPath == path {
			return e.Value
		}
	}
	return dflt
}
func (ee *EnvEntryLookup) EnvByKey(key, dflt string) string {
	for i := range ee.Entries {
		e := ee.Entries[i]
		if e.Key == key {
			return e.Value
		}

	}
	return dflt
}

func (ee *EnvEntryLookup) EnvByKeyAndProvider(key, provider, dflt string) string {
	for i := range ee.Entries {
		e := ee.Entries[i]
		if e.Key == key && e.ProviderName == provider {
			return e.Value
		}

	}
	return dflt
}

type Provider interface {
	// in this case 'env' is empty, but EnvEntries are the value
	GetMapping(p KeyPath) ([]EnvEntry, error)

	// in this case env is filled
	Get(p KeyPath) (*EnvEntry, error)

	Put(p KeyPath, val string) error
	PutMapping(p KeyPath, m map[string]string) error

	Delete(p KeyPath) error
	DeleteMapping(p KeyPath) error
}

type MetaInfo struct {
	Description    string
	Name           string
	Authentication string
	ConfigTemplate string
	Ops            OpMatrix
}
type OpMatrix struct {
	Delete        bool
	DeleteMapping bool
	Put           bool
	PutMapping    bool
	Get           bool
	GetMapping    bool
}

type Match struct {
	Path       string
	Line       string
	LineNumber int
	MatchIndex int
	Entry      EnvEntry
}

type RegisteredProvider struct {
	Meta    MetaInfo
	Builder func(logger logging.Logger) (Provider, error)
}
