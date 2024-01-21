package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	klog "github.com/keeper-security/secrets-manager-go/core/logger"
)

const (
	DEFAULT_CONFIG_PATH string = "client-config.json"
)

type IKeyValueStorage interface {
	ReadStorage() map[string]interface{}
	SaveStorage(updatedConfig map[string]interface{})
	Get(key ConfigKey) string
	Set(key ConfigKey, value interface{}) map[string]interface{}
	Delete(key ConfigKey) map[string]interface{}
	DeleteAll() map[string]interface{}
	Contains(key ConfigKey) bool
	IsEmpty() bool
}

// File based implementation of the key value storage
type fileKeyValueStorage struct {
	ConfigPath string
}

func (f *fileKeyValueStorage) ReadStorage() map[string]interface{} {
	f.createConfigFileIfMissing()
	content, err := ioutil.ReadFile(f.ConfigPath)
	if err != nil {
		klog.Error("Unable to open file: " + f.ConfigPath + " Error: " + err.Error())
		return map[string]interface{}{}
	}

	// RFC3629 - having the BOM in JSON string is forbidden
	// Implementations MUST NOT add a byte order mark (U+FEFF)
	// Implementations that parse JSON texts MAY ignore BOM rather than treating it as an error.
	// In JSON BOM is useless and will always appear as the octet sequence EF BB BF.
	if len(content) >= 3 && content[0] == 0xef && content[1] == 0xbb && content[2] == 0xbf {
		content = content[3:]
	}

	// check for valid UTF-8, check for UTF-16 BE/LE separately
	// all 128 ASCII chars are also valid UTF-8 but 0x00 must be escaped in JSON
	if !utf8.Valid(content) ||
		(len(content) > 1 && (content[0] == 0 || content[1] == 0)) {
		klog.Error("Config file is not utf-8 encoded JSON.")
		return map[string]interface{}{}
	}

	// If it was an empty file, overwrite with the JSON config
	content = bytes.TrimSpace(content)
	if len(content) == 0 {
		klog.Warning("Looks like config file is empty.")
		content = []byte("{}")
	}

	var payload map[string]interface{}
	err = json.Unmarshal(content, &payload)
	if err != nil {
		klog.Error("Error parsing JSON configuration file: " + err.Error())
		return map[string]interface{}{}
	}
	return payload
}

func (f *fileKeyValueStorage) SaveStorage(updatedConfig map[string]interface{}) {
	f.createConfigFileIfMissing()
	content, err := json.MarshalIndent(updatedConfig, "", "    ")
	if err != nil {
		klog.Error("Error writing JSON: " + err.Error())
		return
	}
	if err := ioutil.WriteFile(f.ConfigPath, content, 0666); err != nil {
		klog.Error("Error writing JSON configuration file: " + err.Error())
	}
}

func (f *fileKeyValueStorage) Get(key ConfigKey) string {
	config := f.ReadStorage()
	if value, found := config[string(key)]; found {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return ""
}

func (f *fileKeyValueStorage) Set(key ConfigKey, value interface{}) map[string]interface{} {
	config := f.ReadStorage()
	config[string(key)] = value
	f.SaveStorage(config)
	return config
}

func (f *fileKeyValueStorage) Delete(key ConfigKey) map[string]interface{} {
	config := f.ReadStorage()
	kv := string(key)
	if _, found := config[kv]; found {
		delete(config, kv)
		klog.Debug("Removed key: " + kv)
	} else {
		klog.Warning(fmt.Sprintf("No key '%s' was found in config", kv))
	}

	f.SaveStorage(config)
	return config
}

func (f *fileKeyValueStorage) DeleteAll() map[string]interface{} {
	config := f.ReadStorage()

	for k := range config {
		delete(config, k)
	}

	f.SaveStorage(config)

	return config
}

func (f *fileKeyValueStorage) Contains(key ConfigKey) bool {
	config := f.ReadStorage()
	_, found := config[string(key)]
	return found
}

func (f *fileKeyValueStorage) IsEmpty() bool {
	config := f.ReadStorage()

	return len(config) == 0
}

func (f *fileKeyValueStorage) createConfigFileIfMissing() {
	if ok, err := PathExists(f.ConfigPath); !ok {
		if err != nil {
			klog.Error("Error accessing config file: " + err.Error())
		}

		if err := os.MkdirAll(filepath.Dir(f.ConfigPath), 0755); err != nil {
			klog.Error("Error creating folders: " + err.Error())
		}

		if c, err := os.Create(f.ConfigPath); err == nil {
			defer c.Close()
			if _, err := c.WriteString("{}"); err != nil {
				klog.Error("Failed to write config content: " + err.Error())
			}
		} else {
			klog.Error("Unable to create file: " + err.Error())
		}
	}
}

func NewFileKeyValueStorage(filePath ...interface{}) *fileKeyValueStorage {
	configPath := DEFAULT_CONFIG_PATH

	if len(filePath) > 0 {
		switch t := filePath[0].(type) {
		case string:
			configPath = t
		default:
			klog.Warning("Incorrect config file path - switching to default config path.")
		}
	} else if envKeeperConfigFile := strings.TrimSpace(os.Getenv("KSM_CONFIG_FILE")); envKeeperConfigFile != "" {
		configPath = envKeeperConfigFile
	}

	return &fileKeyValueStorage{
		ConfigPath: configPath,
	}
}

// Memory based implementation of the key value storage
type memoryKeyValueStorage struct {
	Config map[ConfigKey]string
}

func NewMemoryKeyValueStorage(config ...interface{}) *memoryKeyValueStorage {
	iConfig := make(map[string]interface{})
	sConfig := make(map[string]string)
	if len(config) > 0 {
		switch t := config[0].(type) {
		case string:
			jsonStr := t
			// Decode if config json was provided as base64 string
			oldWriter := klog.Writer()
			klog.SetOutput(io.Discard)
			if decodedJson := Base64ToString(jsonStr); len(decodedJson) > 2 {
				jsonStr = decodedJson
			}
			klog.SetOutput(oldWriter)

			iConfig = JsonToDict(jsonStr)
			if len(iConfig) == 0 {
				strJson := fmt.Sprintf("%.16s", t)
				if len(t) > len(strJson) {
					strJson += "..."
				}
				klog.Error(fmt.Sprintf("Could not load config data. Text size: %d  Text: '%s'", len(t), strJson))
			}
		case map[string]interface{}:
			iConfig = t
		case map[string]string:
			sConfig = t
		default:
			klog.Error(fmt.Sprintf("skipping unsupported config type '%v'", t))
		}
	}
	if len(iConfig) > 0 {
		for key, value := range iConfig {
			switch value := value.(type) {
			case string:
				sConfig[key] = value
			default:
				klog.Error(fmt.Sprintf("skipping Config['%s'] - unsupported value type '%v'", key, value))
			}
		}
	}

	newConfig := make(map[ConfigKey]string)
	for key, value := range sConfig {
		if k := GetConfigKey(key); k != "" {
			newConfig[k] = value
		} else {
			klog.Error("skipping unknown config key value: " + key)
		}
	}

	return &memoryKeyValueStorage{
		Config: newConfig,
	}
}

func (m *memoryKeyValueStorage) ReadStorage() map[string]interface{} {
	// To match what FileKeyValueStorage does, we need to return the enum values as keys
	// instead of the enum keys
	dictConfig := map[string]interface{}{}
	for key, value := range m.Config {
		dictConfig[string(key)] = value
	}

	return dictConfig
}

func (m *memoryKeyValueStorage) SaveStorage(updatedConfig map[string]interface{}) {}

func (m *memoryKeyValueStorage) Get(key ConfigKey) string {
	if val, ok := m.Config[key]; ok {
		return val
	}
	return ""
}

func (m *memoryKeyValueStorage) Set(key ConfigKey, value interface{}) map[string]interface{} {
	switch v := value.(type) {
	case string:
		m.Config[key] = v
		return m.ReadStorage()
	default:
		klog.Error(fmt.Sprintf("Unknown value for ConfigKey: %s, Value: %v", string(key), v))
	}
	return nil
}

func (m *memoryKeyValueStorage) Delete(key ConfigKey) map[string]interface{} {
	if _, found := m.Config[key]; found {
		delete(m.Config, key)
		klog.Debug("Removed key: " + key)
	} else {
		klog.Warning(fmt.Sprintf("No key '%s' was found in config", string(key)))
	}
	return m.ReadStorage()
}

func (m *memoryKeyValueStorage) DeleteAll() map[string]interface{} {
	m.Config = map[ConfigKey]string{}
	return m.ReadStorage()
}

func (m *memoryKeyValueStorage) Contains(key ConfigKey) bool {
	_, found := m.Config[key]
	return found
}

func (m *memoryKeyValueStorage) IsEmpty() bool {
	return len(m.Config) == 0
}
