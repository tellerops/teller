package enpass

import (
	"bufio"
	"crypto/sha512"
	"os"

	"github.com/pkg/errors"
	"golang.org/x/crypto/pbkdf2"
)

const (
	// current key derivation algo
	keyDerivationAlgo = "pbkdf2"
	// current database encryption algo
	dbEncryptionAlgo = "aes-256-cbc"
	// database key salt length
	saltLength = 16
	// length of the database master key (capped)
	masterKeyLength = 64
)

// generateMasterPassword : generates the master password to decrypt the vault database
func (v *Vault) generateMasterPassword(password []byte, keyfilePath string) ([]byte, error) {
	if keyfilePath == "" {
		v.logger.Debug("not using keyfile")

		if password == nil {
			return nil, errors.New("empty master password provided")
		}

		return password, nil
	}

	v.logger.Debug("using keyfile")

	keyfileBytes, err := loadKeyFilePassword(keyfilePath)
	if err != nil {
		return nil, err
	}

	return append(password, keyfileBytes...), nil
}

// extractSalt : extract the encryption salt stored in the database
func (v *Vault) extractSalt() ([]byte, error) {
	f, err := os.OpenFile(v.databaseFilename, os.O_RDONLY, 0)
	if err != nil {
		return []byte{}, errors.Wrap(err, "could not open database")
	}
	defer func() { _ = f.Close() }()

	bytesSalt, err := bufio.NewReader(f).Peek(saltLength)
	if err != nil {
		return []byte{}, errors.Wrap(err, "could not read database salt")
	}

	return bytesSalt, nil
}

// deriveKey : generate the SQLCipher crypto key, possibly with the 64-bit Keyfile
func (v *Vault) deriveKey(masterPassword []byte, salt []byte) ([]byte, error) {
	if v.vaultInfo.KDFAlgo != keyDerivationAlgo {
		return nil, errors.New("key derivation algo has changed, open up a github issue")
	}

	if v.vaultInfo.EncryptionAlgo != dbEncryptionAlgo {
		return nil, errors.New("database encryption algo has changed, open up a github issue")
	}

	// The database key is derived from the master password
	// and the database salt with 100k iterations of PBKDF2-HMAC-SHA512
	return pbkdf2.Key(masterPassword, salt, v.vaultInfo.KDFIterations, sha512.Size, sha512.New), nil
}
