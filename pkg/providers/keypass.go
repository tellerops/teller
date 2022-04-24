package providers

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spectralops/teller/pkg/core"

	"github.com/spectralops/teller/pkg/logging"
	"github.com/tobischo/gokeepasslib/v3"
)

var (
	// keyPathFields describe all the available fields in KeyPass entry
	keyPathFields = []string{"Notes", "Password", "URL", "UserName"}
)

type KeyPass struct {
	logger logging.Logger
	data   map[string]gokeepasslib.Entry
}

// NewKeyPass creates new provider instance
func NewKeyPass(logger logging.Logger) (core.Provider, error) {

	password := os.Getenv("KEYPASS_PASSWORD")
	if password == "" {
		return nil, errors.New("missing `KEYPASS_PASSWORD`")
	}
	dbPath := os.Getenv("KEYPASS_DB_PATH")
	if dbPath == "" {
		return nil, errors.New("missing `KEYPASS_DB_PATH`")
	}

	file, err := os.Open(dbPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	db := gokeepasslib.NewDatabase()
	db.Credentials = gokeepasslib.NewPasswordCredentials(password)
	err = gokeepasslib.NewDecoder(file).Decode(db)

	if err != nil {
		return nil, err
	}

	err = db.UnlockProtectedEntries()
	if err != nil {
		return nil, err
	}
	keyPass := &KeyPass{
		logger: logger,
	}
	keyPass.data = keyPass.prepareGroups("", db.Content.Root.Groups, nil)
	return keyPass, nil
}

// Name return the provider name
func (k *KeyPass) Name() string {
	return "KeyPass"
}

// Put will create a new single entry
func (k *KeyPass) Put(p core.KeyPath, val string) error {
	return fmt.Errorf("provider %q does not implement write yet", k.Name())
}

// PutMapping will create a multiple entries
func (k *KeyPass) PutMapping(p core.KeyPath, m map[string]string) error {
	return fmt.Errorf("provider %q does not implement write yet", k.Name())
}

// GetMapping returns a multiple entries
func (k *KeyPass) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {

	results := []core.EnvEntry{}
	for path, entry := range k.data { //nolint
		// get entries that start with the given path
		if strings.HasPrefix(path, p.Path) {
			if p.Source == "" {
				// getting all entries fields
				for _, field := range keyPathFields {
					val := entry.Get(field).Value.Content
					// skip on empty field
					if val == "" {
						k.logger.WithFields(map[string]interface{}{
							"field": field,
							"path":  path,
						}).Debug("empty field")
						continue
					}
					results = append(results, p.FoundWithKey(fmt.Sprintf("%s/%s", path, strings.ToLower(field)), val))
				}
			} else {
				fieldContent := entry.Get(p.Source)
				if fieldContent == nil {
					k.logger.WithFields(map[string]interface{}{
						"source": p.Source,
						"path":   path,
					}).Debug("field not found")
					continue
				}
				val := fieldContent.Value.Content
				if val != "" {
					results = append(results, p.FoundWithKey(path, val))
				}
			}
		}
	}
	return results, nil
}

// Get returns a single entry
func (k *KeyPass) Get(p core.KeyPath) (*core.EnvEntry, error) {
	ent := p.Missing()
	entry, found := k.data[p.Path]
	if !found {
		k.logger.WithField("path", p.Path).Debug("secret not found in path")
		return nil, fmt.Errorf("%v path: %s not exists", k.Name(), p.Path)
	}
	source := p.Source
	if source == "" {
		k.logger.WithField("path", p.Path).Debug("source attribute is empty, setting default field")
		source = "Password"
	}
	k.logger.WithFields(map[string]interface{}{
		"path":   p.Path,
		"source": source,
	}).Debug("get keypass field")
	ent = p.Found(entry.Get(source).Value.Content)

	return &ent, nil
}

// Delete will delete entry
func (k *KeyPass) Delete(kp core.KeyPath) error {
	return fmt.Errorf("provider %s does not implement delete yet", k.Name())
}

// DeleteMapping will delete the given path recessively
func (k *KeyPass) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("provider %s does not implement delete yet", k.Name())
}

// prepareGroups all KeyPass entries for easy seearch
func (k *KeyPass) prepareGroups(path string, groups []gokeepasslib.Group, mapData map[string]gokeepasslib.Entry) map[string]gokeepasslib.Entry {
	if mapData == nil {
		mapData = map[string]gokeepasslib.Entry{}
	}
	for _, group := range groups { //nolint
		// if entries found, adding the entry data fo the list
		if len(group.Entries) > 0 {
			for _, entry := range group.Entries { //nolint
				mapData[fmt.Sprintf("%s/%s/%s", path, group.Name, entry.GetTitle())] = entry
			}
		}
		if len(group.Groups) > 0 {
			// call recursively prepareGroups function get collect entries
			return k.prepareGroups(strings.TrimPrefix(fmt.Sprintf("%s/%s", path, group.Name), "/"), group.Groups, mapData)
		}
	}
	return mapData
}
