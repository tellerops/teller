package providers

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	ksm "github.com/keeper-security/secrets-manager-go/core"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
)

type KsmClient interface {
	GetSecret(p core.KeyPath) (*core.EnvEntry, error)
	GetSecrets(p core.KeyPath) ([]core.EnvEntry, error)
}

type KeeperSecretsManager struct {
	client KsmClient
	logger logging.Logger
}

const keeperName = "keeper_secretsmanager"

//nolint
func init() {
	metaInto := core.MetaInfo{
		Description:    "Keeper Secrets Manager",
		Name:           keeperName,
		Authentication: "You should populate `KSM_CONFIG=Base64ConfigString or KSM_CONFIG_FILE=ksm_config.json` in your environment.",
		ConfigTemplate: `
  keeper_secretsmanager:
  env_sync:
    path: record_uid
  env:
    FOO_BAR:
	  path: UID/custom_field/foo_bar
	  # use Keeper Notation to pull data from typed records
`,
		Ops: core.OpMatrix{Get: true, GetMapping: true},
	}
	RegisterProvider(metaInto, NewKeeperSecretsManager)
}

type SecretsManagerClient struct {
	sm *ksm.SecretsManager
}

func (c SecretsManagerClient) GetSecret(p core.KeyPath) (*core.EnvEntry, error) {
	nr, err := c.sm.GetNotationResults(p.Path)
	if err != nil {
		return nil, err
	}

	ent := core.EnvEntry{}
	if len(nr) > 0 {
		ent = p.Found(nr[0])
	} else {
		ent = p.Missing()
	}

	return &ent, nil
}

func (c SecretsManagerClient) GetSecrets(p core.KeyPath) ([]core.EnvEntry, error) {
	// p.Path must be a record UID
	// TODO: Add Path = folderUID... expect too many dulpicates, prefix key name with RUID?
	type KeyValuePair struct {
		key, value string
	}

	r := []core.EnvEntry{}

	recs, err := c.sm.GetSecrets([]string{p.Path})
	if err != nil {
		return nil, err
	}
	if len(recs) < 1 {
		return r, nil
	}

	entries := []KeyValuePair{}
	rec := recs[0]
	fields := rec.GetFieldsBySection(ksm.FieldSectionBoth)
	for _, field := range fields {
		fmap, ok := field.(map[string]interface{})
		if !ok {
			continue
		}

		iValues, ok := fmap["value"].([]interface{})
		if !ok || len(iValues) < 1 {
			continue
		}

		value := extractValue(iValues)
		if value == "" {
			continue
		}

		key := extractKey(fmap)
		entries = append(entries, KeyValuePair{key: key, value: value})
	}

	// avoid duplicate key names
	keymap := map[string]struct{}{}
	for _, e := range entries {
		key := e.key
		if _, found := keymap[e.key]; found {
			n := 1
			for {
				n++
				mkey := key + "_" + strconv.Itoa(n)
				if _, found := keymap[mkey]; !found {
					key = mkey
					break
				}
			}
		}
		keymap[key] = struct{}{}
		ent := p.FoundWithKey(key, e.value)
		r = append(r, ent)
	}

	return r, nil
}

func extractKey(fieldMap map[string]interface{}) string {
	key := ""
	if fLabel, ok := fieldMap["label"].(string); ok {
		key = strings.TrimSpace(fLabel)
	}
	if key == "" {
		if fType, ok := fieldMap["type"].(string); ok {
			key = strings.TrimSpace(fType)
		}
	}
	key = strings.ReplaceAll(key, " ", "_")
	return key
}

func extractValue(iValues []interface{}) string {
	value := ""
	_, isArray := iValues[0].([]interface{})
	_, isObject := iValues[0].(map[string]interface{})
	isJSON := len(iValues) > 1 || isArray || isObject
	if isJSON {
		if len(iValues) == 1 {
			if val, err := json.Marshal(iValues[0]); err == nil {
				value = string(val)
			}
		} else if val, err := json.Marshal(iValues); err == nil {
			value = string(val)
		}
	} else {
		val := iValues[0]
		// JavaScript number type, IEEE754 double precision float
		if fval, ok := val.(float64); ok && fval == float64(int(fval)) {
			val = int(fval) // convert to int
		}
		value = fmt.Sprintf("%v", val)
	}
	return value
}

func NewKsmClient() (KsmClient, error) {
	config := os.Getenv("KSM_CONFIG")
	configPath := os.Getenv("KSM_CONFIG_FILE")
	if config == "" && configPath == "" {
		return nil, fmt.Errorf("cannot find KSM_CONFIG or KSM_CONFIG_FILE for %s", keeperName)
	}

	// with both options present KSM_CONFIG overrides KSM_CONFIG_FILE
	var options *ksm.ClientOptions = nil
	if config != "" {
		options = &ksm.ClientOptions{Config: ksm.NewMemoryKeyValueStorage(config)}
	} else if stat, err := os.Stat(configPath); err == nil && stat.Size() > 2 {
		options = &ksm.ClientOptions{Config: ksm.NewFileKeyValueStorage(configPath)}
	}

	if options == nil {
		return nil, fmt.Errorf("failed to initialize KSM Client Options")
	}

	sm := ksm.NewSecretsManager(options)
	if sm == nil {
		return nil, fmt.Errorf("failed to initialize KSM Client")
	}

	return SecretsManagerClient{
		sm: sm,
	}, nil
}

func NewKeeperSecretsManager(logger logging.Logger) (core.Provider, error) {
	ksmClient, err := NewKsmClient()
	if err != nil {
		return nil, err
	}
	return &KeeperSecretsManager{
		client: ksmClient,
		logger: logger,
	}, nil
}

func (k *KeeperSecretsManager) Name() string {
	return keeperName
}

func (k *KeeperSecretsManager) Get(p core.KeyPath) (*core.EnvEntry, error) {
	return k.client.GetSecret(p)
}

func (k *KeeperSecretsManager) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {
	return k.client.GetSecrets(p)
}

func (k *KeeperSecretsManager) Put(p core.KeyPath, val string) error {
	return fmt.Errorf("provider %q does not implement write yet", k.Name())
}

func (k *KeeperSecretsManager) PutMapping(p core.KeyPath, m map[string]string) error {
	return fmt.Errorf("provider %q does not implement write mapping yet", k.Name())
}

func (k *KeeperSecretsManager) Delete(kp core.KeyPath) error {
	return fmt.Errorf("provider %s does not implement delete yet", k.Name())
}

func (k *KeeperSecretsManager) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("provider %s does not implement delete mapping yet", k.Name())
}
