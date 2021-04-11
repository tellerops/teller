package core

type Severity string

const (
	High   Severity = "high"
	Medium Severity = "medium"
	Low    Severity = "low"
	None   Severity = "none"
)

type KeyPath struct {
	Env      string            `yaml:"env,omitempty"`
	Path     string            `yaml:"path"`
	Field    string            `yaml:"field,omitempty"`
	Remap    map[string]string `yaml:"remap,omitempty"`
	Decrypt  bool              `yaml:"decrypt,omitempty"`
	Optional bool              `yaml:"optional,omitempty"`
	Severity Severity          `yaml:"severity,omitempty" default:"high"`
}
type WizardAnswers struct {
	Project      string
	Providers    []string
	ProviderKeys map[string]bool
	Confirm      bool
}

func (k *KeyPath) WithEnv(env string) KeyPath {
	return KeyPath{
		Env:      env,
		Path:     k.Path,
		Field:    k.Field,
		Decrypt:  k.Decrypt,
		Optional: k.Optional,
	}
}
func (k *KeyPath) SwitchPath(path string) KeyPath {
	return KeyPath{
		Path:     path,
		Field:    k.Field,
		Env:      k.Env,
		Decrypt:  k.Decrypt,
		Optional: k.Optional,
	}
}

type EntriesByKey []EnvEntry

func (a EntriesByKey) Len() int           { return len(a) }
func (a EntriesByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a EntriesByKey) Less(i, j int) bool { return a[i].Key > a[j].Key }

type EnvEntry struct {
	Key          string
	Value        string
	Provider     string
	ResolvedPath string
	Severity     Severity
}
type EnvEntryLookup struct {
	Entries []EnvEntry
}

func (e *EnvEntryLookup) EnvBy(key, provider, path, dflt string) string {
	for _, e := range e.Entries {
		if e.Key == key && e.Provider == provider && e.ResolvedPath == path {
			return e.Value
		}

	}
	return dflt
}
func (e *EnvEntryLookup) EnvByKey(key, dflt string) string {
	for _, e := range e.Entries {
		if e.Key == key {
			return e.Value
		}

	}
	return dflt
}

func (e *EnvEntryLookup) EnvByKeyAndProvider(key, provider, dflt string) string {
	for _, e := range e.Entries {
		if e.Key == key && e.Provider == provider {
			return e.Value
		}

	}
	return dflt
}

type Provider interface {
	Name() string
	// in this case 'env' is empty, but EnvEntries are the value
	GetMapping(p KeyPath) ([]EnvEntry, error)

	// in this case env is filled
	Get(p KeyPath) (*EnvEntry, error)
}

type Match struct {
	Path       string
	Line       string
	LineNumber int
	MatchIndex int
	Entry      EnvEntry
}
