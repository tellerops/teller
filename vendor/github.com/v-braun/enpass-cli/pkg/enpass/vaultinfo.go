package enpass

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pkg/errors"
)

type VaultInfo struct {
	EncryptionAlgo string `json:"encryption_algo"`
	HasKeyfile     int    `json:"have_keyfile"`
	KDFAlgo        string `json:"kdf_algo"`
	KDFIterations  int    `json:"kdf_iter"`
	VaultNumItems  int    `json:"vault_items_count"`
	VaultName      string `json:"vault_name"`
	VaultVersion   int    `json:"version"`
}

// loadVaultInfo : the vault info file dictates how we should decrypt the vault database
func (v *Vault) loadVaultInfo() (VaultInfo, error) {
	vaultInfoBytes, err := ioutil.ReadFile(v.vaultInfoFilename)
	if err != nil {
		return VaultInfo{}, errors.Wrap(err, "could not read vault info")
	}

	var vaultInfo VaultInfo
	if err := json.Unmarshal(vaultInfoBytes, &vaultInfo); err != nil {
		return VaultInfo{}, errors.Wrap(err, "could not parse vault info")
	}

	v.logger.
		WithField("vault_name", vaultInfo.VaultName).
		WithField("vault_version", vaultInfo.VaultVersion).
		Debug("vault info loaded")

	return vaultInfo, nil
}
