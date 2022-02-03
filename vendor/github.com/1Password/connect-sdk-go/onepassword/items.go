package onepassword

import (
	"encoding/json"
	"strings"
	"time"
)

// ItemCategory Represents the template of the Item
type ItemCategory string

const (
	Login                ItemCategory = "LOGIN"
	Password             ItemCategory = "PASSWORD"
	Server               ItemCategory = "SERVER"
	Database             ItemCategory = "DATABASE"
	CreditCard           ItemCategory = "CREDIT_CARD"
	Membership           ItemCategory = "MEMBERSHIP"
	Passport             ItemCategory = "PASSPORT"
	SoftwareLicense      ItemCategory = "SOFTWARE_LICENSE"
	OutdoorLicense       ItemCategory = "OUTDOOR_LICENSE"
	SecureNote           ItemCategory = "SECURE_NOTE"
	WirelessRouter       ItemCategory = "WIRELESS_ROUTER"
	BankAccount          ItemCategory = "BANK_ACCOUNT"
	DriverLicense        ItemCategory = "DRIVER_LICENSE"
	Identity             ItemCategory = "IDENTITY"
	RewardProgram        ItemCategory = "REWARD_PROGRAM"
	Document             ItemCategory = "DOCUMENT"
	EmailAccount         ItemCategory = "EMAIL_ACCOUNT"
	SocialSecurityNumber ItemCategory = "SOCIAL_SECURITY_NUMBER"
	ApiCredential        ItemCategory = "API_CREDENTIAL"
	Custom               ItemCategory = "CUSTOM"
)

// UnmarshalJSON Unmarshall Item Category enum strings to Go string enums
func (ic *ItemCategory) UnmarshalJSON(b []byte) error {
	var s string
	json.Unmarshal(b, &s)
	category := ItemCategory(s)
	switch category {
	case Login, Password, Server, Database, CreditCard, Membership, Passport, SoftwareLicense,
		OutdoorLicense, SecureNote, WirelessRouter, BankAccount, DriverLicense, Identity, RewardProgram,
		Document, EmailAccount, SocialSecurityNumber, ApiCredential:
		*ic = category
	default:
		*ic = Custom
	}

	return nil
}

// Item represents an item returned to the consumer
type Item struct {
	ID    string `json:"id"`
	Title string `json:"title"`

	URLs     []ItemURL `json:"urls,omitempty"`
	Favorite bool      `json:"favorite,omitempty"`
	Tags     []string  `json:"tags,omitempty"`
	Version  int       `json:"version,omitempty"`
	Trashed  bool      `json:"trashed,omitempty"`

	Vault    ItemVault    `json:"vault"`
	Category ItemCategory `json:"category,omitempty"` // TODO: switch this to `category`

	Sections []*ItemSection `json:"sections,omitempty"`
	Fields   []*ItemField   `json:"fields,omitempty"`
	Files    []*File        `json:"files,omitempty"`

	LastEditedBy string    `json:"lastEditedBy,omitempty"`
	CreatedAt    time.Time `json:"createdAt,omitempty"`
	UpdatedAt    time.Time `json:"updatedAt,omitempty"`
}

// ItemVault represents the Vault the Item is found in
type ItemVault struct {
	ID string `json:"id"`
}

// ItemURL is a simplified item URL
type ItemURL struct {
	Primary bool   `json:"primary,omitempty"`
	URL     string `json:"href"`
}

// ItemSection Representation of a Section on an item
type ItemSection struct {
	ID    string `json:"id,omitempty"`
	Label string `json:"label,omitempty"`
}

// GeneratorRecipe Representation of a "recipe" used to generate a field
type GeneratorRecipe struct {
	Length        int      `json:"length,omitempty"`
	CharacterSets []string `json:"characterSets,omitempty"`
}

// ItemField Representation of a single field on an Item
type ItemField struct {
	ID       string           `json:"id"`
	Section  *ItemSection     `json:"section,omitempty"`
	Type     string           `json:"type"`
	Purpose  string           `json:"purpose,omitempty"`
	Label    string           `json:"label,omitempty"`
	Value    string           `json:"value,omitempty"`
	Generate bool             `json:"generate,omitempty"`
	Recipe   *GeneratorRecipe `json:"recipe,omitempty"`
	Entropy  float64          `json:"entropy,omitempty"`
}

// Get Retrieve the value of a field on the item by its label. To specify a
// field from a specific section pass in <section label>.<field label>. If
// no field matching the selector is found return "".
func (i *Item) GetValue(field string) string {
	if i == nil || len(i.Fields) == 0 {
		return ""
	}

	sectionFilter := false
	sectionLabel := ""
	fieldLabel := field
	if strings.Contains(field, ".") {
		parts := strings.Split(field, ".")

		// Test to make sure the . isn't the last character
		if len(parts) == 2 {
			sectionFilter = true
			sectionLabel = parts[0]
			fieldLabel = parts[1]
		}
	}

	for _, f := range i.Fields {
		if sectionFilter {
			if f.Section != nil {
				if sectionLabel != i.SectionLabelForID(f.Section.ID) {
					continue
				}
			}
		}

		if fieldLabel == f.Label {
			return f.Value
		}
	}

	return ""
}

func (i *Item) SectionLabelForID(id string) string {
	if i != nil || len(i.Sections) > 0 {
		for _, s := range i.Sections {
			if s.ID == id {
				return s.Label
			}
		}
	}

	return ""
}
