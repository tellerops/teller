package onepassword

import (
	"encoding/json"
	"time"
)

// Vault represents a 1password Vault
type Vault struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`

	AttrVersion    int       `json:"attributeVersion,omitempty"`
	ContentVersoin int       `json:"contentVersion,omitempty"`
	Items          int       `json:"items,omitempty"`
	Type           VaultType `json:"type,omitempty"`

	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

// VaultType Representation of what the Vault Type is
type VaultType string

const (
	PersonalVault    VaultType = "PERSONAL"
	EveryoneVault    VaultType = "EVERYONE"
	TransferVault    VaultType = "TRANSFER"
	UserCreatedVault VaultType = "USER_CREATED"
	UnknownVault     VaultType = "UNKNOWN"
)

// UnmarshalJSON Unmarshall Vault Type enum strings to Go string enums
func (vt *VaultType) UnmarshalJSON(b []byte) error {
	var s string
	json.Unmarshal(b, &s)
	vaultType := VaultType(s)
	switch vaultType {
	case PersonalVault, EveryoneVault, TransferVault, UserCreatedVault:
		*vt = vaultType
	default:
		*vt = UnknownVault
	}
	return nil
}
