package connect

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/1Password/connect-sdk-go/onepassword"
)

const (
	vaultTag = "opvault"
	itemTag  = "opitem"
	fieldTag = "opfield"

	envVaultVar = "OP_VAULT"
)

type parsedItem struct {
	vaultUUID string
	itemTitle string
	fields    []*reflect.StructField
	values    []*reflect.Value
}

// Load Load configuration values based on strcut tag
func Load(client Client, i interface{}) error {
	configP := reflect.ValueOf(i)
	if configP.Kind() != reflect.Ptr {
		return fmt.Errorf("You must pass a pointer to Config struct")
	}

	config := configP.Elem()
	if config.Kind() != reflect.Struct {
		return fmt.Errorf("Config values can only be loaded into a struct")
	}

	t := config.Type()

	// Multiple fields may be from a single item so we will collect them
	items := map[string]parsedItem{}

	// Fetch the Vault from the environment
	vaultUUID, envVarFound := os.LookupEnv(envVaultVar)

	for i := 0; i < t.NumField(); i++ {
		value := config.Field(i)
		field := t.Field(i)
		tag := field.Tag.Get(itemTag)

		if tag == "" {
			continue
		}

		if !value.CanSet() {
			return fmt.Errorf("Cannot load config into private fields")
		}

		itemVault, err := vaultUUIDForField(&field, vaultUUID, envVarFound)
		if err != nil {
			return err
		}

		key := fmt.Sprintf("%s/%s", itemVault, tag)
		parsed := items[key]
		parsed.vaultUUID = itemVault
		parsed.itemTitle = tag
		parsed.fields = append(parsed.fields, &field)
		parsed.values = append(parsed.values, &value)
		items[key] = parsed
	}

	for _, item := range items {
		if err := setValuesForTag(client, &item); err != nil {
			return err
		}
	}

	return nil
}

func vaultUUIDForField(field *reflect.StructField, vaultUUID string, envVaultFound bool) (string, error) {
	// Check to see if a specific vault has been specified on the field
	// If the env vault id has not been found and item doesn't have a vault
	// return an error
	if vaultUUIDTag := field.Tag.Get(vaultTag); vaultUUIDTag == "" {
		if !envVaultFound {
			return "", fmt.Errorf("There is no vault for %q field", field.Name)
		}
	} else {
		return vaultUUIDTag, nil
	}

	return vaultUUID, nil
}

func setValuesForTag(client Client, parsedItem *parsedItem) error {
	item, err := client.GetItemByTitle(parsedItem.itemTitle, parsedItem.vaultUUID)
	if err != nil {
		return err
	}

	for i, field := range parsedItem.fields {
		value := parsedItem.values[i]
		path := field.Tag.Get(fieldTag)
		if path == "" {
			if field.Type == reflect.TypeOf(onepassword.Item{}) {
				value.Set(reflect.ValueOf(*item))
				return nil
			}
			return fmt.Errorf("There is no %q specified for %q", fieldTag, field.Name)
		}

		pathParts := strings.Split(path, ".")

		if len(pathParts) != 2 {
			return fmt.Errorf("Invalid field path format for %q", field.Name)
		}

		sectionID := sectionIDForName(pathParts[0], item.Sections)
		label := pathParts[1]

		for _, f := range item.Fields {
			fieldSectionID := ""
			if f.Section != nil {
				fieldSectionID = f.Section.ID
			}

			if fieldSectionID == sectionID && f.Label == label {
				if err := setValue(value, f.Value); err != nil {
					return err
				}
				break
			}
		}
	}

	return nil
}

func setValue(value *reflect.Value, toSet string) error {
	switch value.Kind() {
	case reflect.String:
		value.SetString(toSet)
	case reflect.Int:
		v, err := strconv.Atoi(toSet)
		if err != nil {
			return err
		}
		value.SetInt(int64(v))
	default:
		return fmt.Errorf("Unsupported type %q. Only string, int64, and onepassword.Item are supported", value.Kind())
	}

	return nil
}

func sectionIDForName(name string, sections []*onepassword.ItemSection) string {
	if sections == nil {
		return ""
	}

	for _, s := range sections {
		if name == strings.ToLower(s.Label) {
			return s.ID
		}
	}

	return ""
}
