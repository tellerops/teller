package providers

import (
	"encoding/json"
	"fmt"

	"github.com/keybase/go-keychain"
	"github.com/spectralops/teller/pkg/core"

	"github.com/spectralops/teller/pkg/logging"
)

type keychainQuery struct {
	Service     string `json:"service"`
	Account     string `json:"account"`
	Label       string `json:"label"`
	AccessGroup string `json:"accessGroup"`
}

type MacKeychain struct {
	logger logging.Logger
}

// MacKeychain creates new provider instance
func NewMacKeychain(logger logging.Logger) (core.Provider, error) {
	return &MacKeychain{
		logger: logger,
	}, nil
}

// Name return the provider name
func (mc *MacKeychain) Name() string {
	return "Mac_Keychain"
}

// Put will create a new single entry
func (mc *MacKeychain) Put(p core.KeyPath, val string) error {

	queryItemData, err := mc.toKeychainQuery(p.Path)
	if err != nil {
		return err
	}

	itemQuery := keychain.NewItem()

	mc.setItemData(&itemQuery, queryItemData)
	itemQuery.SetData([]byte(val))
	itemQuery.SetAccessible(keychain.AccessibleDefault)
	itemQuery.SetSecClass(keychain.SecClassGenericPassword)

	return keychain.AddItem(itemQuery)
}

// PutMapping will create a multiple entries
func (mc *MacKeychain) PutMapping(p core.KeyPath, m map[string]string) error {

	// Creating a put mapping can be a bit tricky. Some attributes need to be a unique values as an account.
	// To creates a put mapping functionality, we need to get multiple different attributes for a single record which can be understandable to the users.
	return fmt.Errorf("provider %q does not implement write mapping yet", mc.Name())
}

// GetMapping returns a multiple entries
func (mc *MacKeychain) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {

	queryItemData, err := mc.toKeychainQuery(p.Path)
	if err != nil {
		return nil, err
	}

	itemQuery := keychain.NewItem()
	mc.setItemData(&itemQuery, queryItemData)
	itemQuery.SetSecClass(keychain.SecClassGenericPassword)

	itemQuery.SetMatchLimit(keychain.MatchLimitAll)
	itemQuery.SetReturnAttributes(true)
	results, err := keychain.QueryItem(itemQuery)
	if err != nil {
		return nil, err
	}

	entries := []core.EnvEntry{}
	for i := range results {
		password, err := keychain.GetGenericPassword(results[i].Service, results[i].Account, results[i].Label, results[i].AccessGroup)
		if err != nil {
			return nil, err
		}
		entries = append(entries, p.FoundWithKey(fmt.Sprintf("%s_%s", results[i].Service, results[i].Label), string(password)))
	}
	return entries, nil
}

// Get returns a single entry
func (mc *MacKeychain) Get(p core.KeyPath) (*core.EnvEntry, error) {

	queryItemData, err := mc.toKeychainQuery(p.Path)
	if err != nil {
		return nil, err
	}

	password, err := keychain.GetGenericPassword(queryItemData.Service, queryItemData.Account, queryItemData.Label, queryItemData.AccessGroup)
	var ent = p.Missing()
	if err != nil {
		return nil, err
	}
	if len(password) == 0 {
		return nil, keychain.ErrorItemNotFound
	}
	ent = p.Found(string(password))

	return &ent, nil
}

// Delete will delete entry
func (mc *MacKeychain) Delete(kp core.KeyPath) error {

	queryItemData, err := mc.toKeychainQuery(kp.Path)
	if err != nil {
		return err
	}

	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetService(queryItemData.Service)
	item.SetAccount(queryItemData.Account)
	item.SetLabel(queryItemData.Label)
	item.SetAccessGroup(queryItemData.AccessGroup)
	item.SetMatchLimit(keychain.MatchLimitOne)
	return keychain.DeleteItem(item)

}

// DeleteMapping will delete the given path recessively
func (mc *MacKeychain) DeleteMapping(kp core.KeyPath) error {

	queryItemData, err := mc.toKeychainQuery(kp.Path)
	if err != nil {
		return err
	}

	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetService(queryItemData.Service)
	item.SetAccount(queryItemData.Account)
	item.SetLabel(queryItemData.Label)
	item.SetAccessGroup(queryItemData.AccessGroup)
	item.SetMatchLimit(keychain.MatchLimitAll)
	return keychain.DeleteItem(item)
}

func (mc *MacKeychain) toKeychainQuery(jsonQuery string) (*keychainQuery, error) {

	keychainQueryData := &keychainQuery{}
	err := json.Unmarshal([]byte(jsonQuery), keychainQueryData)
	if err != nil {
		mc.logger.WithField("path", jsonQuery).Debug("invalid item JSON configuration")
		return nil, err
	}
	return keychainQueryData, nil
}

func (mc *MacKeychain) setItemData(item *keychain.Item, query *keychainQuery) {

	if query.Service != "" {
		mc.logger.WithField("value", query.Service).Debug("set service to keychain query")
		item.SetService(query.Service)
	}
	if query.Account != "" {
		mc.logger.WithField("value", query.Account).Debug("set account to keychain query")
		item.SetAccount(query.Account)
	}
	if query.Label != "" {
		mc.logger.WithField("value", query.Label).Debug("set label to keychain query")
		item.SetLabel(query.Label)
	}
	if query.AccessGroup != "" {
		mc.logger.WithField("value", query.AccessGroup).Debug("set accessGroup to keychain query")
		item.SetAccessGroup(query.AccessGroup)
	}

}
