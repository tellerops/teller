package providers

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spectralops/teller/pkg/core"

	"github.com/mattn/lastpass-go"
	"github.com/spectralops/teller/pkg/logging"
)

const (
	findingNoteCount = 2
)

type LastPass struct {
	accounts map[string]*lastpass.Account
	logger   logging.Logger
}

const LastPassName = "lastpass"

//nolint
func init() {
	metaInfo := core.MetaInfo{
		Description:    "Lastpass",
		Authentication: "TODO(XXX)",
		Name:           LastPassName,
		ConfigTemplate: `
  # Configure via environment variables:
  # LASTPASS_USERNAME
  # LASTPASS_PASSWORD

  lastpass:
    env_sync:
      path: # LastPass item ID
    env:
      ETC_DSN:
        path: # Lastpass item ID
        # field: by default taking password property. in case you want other property un-mark this line and set the lastpass property name.
`,
		Ops: core.OpMatrix{GetMapping: true, Get: true},
	}

	RegisterProvider(metaInfo, NewLastPass)
}

func NewLastPass(logger logging.Logger) (core.Provider, error) {
	username := os.Getenv("LASTPASS_USERNAME")
	masterPassword := os.Getenv("LASTPASS_PASSWORD")

	vault, err := lastpass.CreateVault(username, masterPassword)
	if err != nil {
		return nil, err
	}

	accountsMap := map[string]*lastpass.Account{}
	for _, account := range vault.Accounts {
		accountsMap[account.Id] = account
	}

	return &LastPass{accounts: accountsMap, logger: logger}, nil
}

func (l *LastPass) Put(p core.KeyPath, val string) error {
	return fmt.Errorf("provider %q does not implement write yet", LastPassName)
}

func (l *LastPass) PutMapping(p core.KeyPath, m map[string]string) error {
	return fmt.Errorf("provider %q does not implement write yet", LastPassName)
}

func (l *LastPass) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {

	item, err := l.getSecretByID(p.Path)
	if err != nil {
		return nil, err
	}

	entries := []core.EnvEntry{}
	entries = append(entries, p.FoundWithKey("Name", item.Name), p.FoundWithKey("Password", item.Password), p.FoundWithKey("Url", item.Url))

	for k, v := range l.notesToMap(item.Notes) {
		entries = append(entries, p.FoundWithKey(strings.ReplaceAll(k, " ", "_"), v))
	}

	return entries, nil
}

func (l *LastPass) Get(p core.KeyPath) (*core.EnvEntry, error) {

	item, err := l.getSecretByID(p.Path)
	if err != nil {
		return nil, err
	}

	var ent = p.Missing()
	// if field not defined, password field returned
	if p.Field == "" {
		l.logger.WithField("path", p.Path).Debug("field attribute is empty, return item password attribute")
		ent = p.Found(item.Password)
	} else {
		l.logger.WithFields(map[string]interface{}{
			"path":  p.Path,
			"field": p.Field,
		}).Debug("field attribute present, search filed name in notes list")
		key, err := l.getNodeByKeyName(p.Field, item.Notes)
		if err == nil {
			ent = p.Found(key)
		}
	}

	return &ent, nil
}

func (l *LastPass) Delete(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", LastPassName)
}

func (l *LastPass) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", LastPassName)
}

func (l *LastPass) getSecretByID(id string) (*lastpass.Account, error) {

	if item, found := l.accounts[id]; found {
		return item, nil
	}
	return nil, errors.New("item ID not found")

}

// notesToMap parse LastPass note convention to map string
//
// Example:
// `
// card:a
// Type:b
// `
// TO:
// {"card": "a", "Type": "b"}
func (l *LastPass) notesToMap(notes string) map[string]string {

	results := map[string]string{}
	scanner := bufio.NewScanner(strings.NewReader(notes))
	for scanner.Scan() {
		findings := strings.SplitN(scanner.Text(), ":", 2) //nolint: gomnd
		if len(findings) == findingNoteCount {
			results[strings.TrimSpace(findings[0])] = strings.TrimSpace(findings[1])
		}
	}
	return results
}

// getNodeByKeyName parse LastPass note convention and search if one of the note equal to the given key
func (l *LastPass) getNodeByKeyName(key, notes string) (string, error) {

	scanner := bufio.NewScanner(strings.NewReader(notes))
	for scanner.Scan() {
		findings := strings.SplitN(scanner.Text(), ":", 2) //nolint: gomnd
		if len(findings) == findingNoteCount && findings[0] == key {
			return strings.TrimSpace(findings[1]), nil
		}
	}
	l.logger.WithFields(map[string]interface{}{
		"field": key,
	}).Debug("field attribute not found in notes list")
	return "", errors.New("key not found")
}
