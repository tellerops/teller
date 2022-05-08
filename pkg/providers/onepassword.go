package providers

import (
	"fmt"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/logging"
)

type OnePasswordClient interface {
	GetItemByTitle(title, vaultUUID string) (*onepassword.Item, error)
	UpdateItem(item *onepassword.Item, vaultUUID string) (*onepassword.Item, error)
}

type OnePassword struct {
	client OnePasswordClient
	logger logging.Logger
}

func NewOnePassword(logger logging.Logger) (core.Provider, error) {

	client, err := connect.NewClientFromEnvironment()
	if err != nil {
		return nil, err
	}
	return &OnePassword{client: client, logger: logger}, nil
}

func (o *OnePassword) Name() string {
	return "1password"
}

func (o *OnePassword) Meta() core.MetaInfo {
	return core.MetaInfo{}
}

func (o *OnePassword) Put(p core.KeyPath, val string) error {

	item, err := o.getItemByTitle(p)
	if err != nil {
		return err
	}

	for _, field := range item.Fields {
		if field.Label == p.Field {
			field.Value = val
			o.logger.WithFields(map[string]interface{}{
				"item_id":  item.ID,
				"vault_id": p.Source,
			}).Debug("update item")
			_, err := o.client.UpdateItem(item, p.Source)
			return err
		}
	}

	return fmt.Errorf("label %v not found", p.Field)
}

func (o *OnePassword) PutMapping(p core.KeyPath, m map[string]string) error {
	return fmt.Errorf("provider %q does not implement write multiple keys", o.Name())
}

func (o *OnePassword) GetMapping(p core.KeyPath) ([]core.EnvEntry, error) {

	item, err := o.getItemByTitle(p)
	if err != nil {
		return nil, err
	}

	entries := []core.EnvEntry{}
	for _, field := range item.Fields {
		entries = append(entries, p.FoundWithKey(field.Label, field.Value))
	}

	return entries, nil
}

func (o *OnePassword) Get(p core.KeyPath) (*core.EnvEntry, error) {

	item, err := o.getItemByTitle(p)
	if err != nil {
		return nil, err
	}

	var ent = p.Missing()
	for _, field := range item.Fields {
		if field.Label == p.Field || field.Label == p.Env {
			ent = p.Found(field.Value)
			break
		}
		o.logger.WithFields(map[string]interface{}{
			"field": p.Field,
			"env":   p.Env,
			"label": field.Label,
		}).Debug("item not found from list")
	}

	return &ent, nil
}

func (o *OnePassword) Delete(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", o.Name())
}

func (o *OnePassword) DeleteMapping(kp core.KeyPath) error {
	return fmt.Errorf("%s does not implement delete yet", o.Name())
}

func (o *OnePassword) getItemByTitle(p core.KeyPath) (*onepassword.Item, error) {

	o.logger.WithFields(map[string]interface{}{
		"item_id":  p.Path,
		"vault_id": p.Source,
	}).Debug("get item by title")
	item, err := o.client.GetItemByTitle(p.Path, p.Source)
	if err != nil {
		return nil, fmt.Errorf("key %s not found in vaultUUID %s, error: %v", p.Path, p.Source, err)
	}

	return item, nil
}
