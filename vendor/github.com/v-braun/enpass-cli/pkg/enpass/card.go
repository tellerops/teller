package enpass

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"strings"

	"github.com/pkg/errors"
)

/*
2020/12/08 08:59:36 > ID
2020/12/08 08:59:36 > uuid
2020/12/08 08:59:36 > created_at
2020/12/08 08:59:36 > meta_updated_at
2020/12/08 08:59:36 > field_updated_at
2020/12/08 08:59:36 > title
2020/12/08 08:59:36 > subtitle
2020/12/08 08:59:36 > note
2020/12/08 08:59:36 > icon
2020/12/08 08:59:36 > favorite
2020/12/08 08:59:36 > trashed
2020/12/08 08:59:36 > archived
2020/12/08 08:59:36 > deleted
2020/12/08 08:59:36 > auto_submit
2020/12/08 08:59:36 > form_data
2020/12/08 08:59:36 > category
2020/12/08 08:59:36 > template
2020/12/08 08:59:36 > wearable
2020/12/08 08:59:36 > usage_count
2020/12/08 08:59:36 > last_used
2020/12/08 08:59:36 > key
2020/12/08 08:59:36 > extra
2020/12/08 08:59:36 > updated_at
2020/12/08 08:59:36 > ID
2020/12/08 08:59:36 > item_uuid
2020/12/08 08:59:36 > item_field_uid
2020/12/08 08:59:36 > label
2020/12/08 08:59:36 > value
2020/12/08 08:59:36 > deleted
2020/12/08 08:59:36 > sensitive
2020/12/08 08:59:36 > historical
2020/12/08 08:59:36 > type
2020/12/08 08:59:36 > form_id
2020/12/08 08:59:36 > updated_at
2020/12/08 08:59:36 > value_updated_at
2020/12/08 08:59:36 > orde
2020/12/08 08:59:36 > wearable
2020/12/08 08:59:36 > history
2020/12/08 08:59:36 > initial
2020/12/08 08:59:36 > hash
2020/12/08 08:59:36 > strength
2020/12/08 08:59:36 > algo_version
2020/12/08 08:59:36 > expiry
2020/12/08 08:59:36 > excluded
2020/12/08 08:59:36 > pwned_check_time
2020/12/08 08:59:36 > extra
*/

type Card struct {
	// plaintext
	UUID      string
	CreatedAt int64
	Type      string
	UpdatedAt int64
	Title     string
	Subtitle  string
	Note      string
	Trashed   int64
	Deleted   int64
	Category  string
	Label     string
	LastUsed  int64
	Sensitive bool
	Icon      string
	RawValue  string

	// encrypted
	value   string
	itemKey []byte
}

func (c *Card) IsTrashed() bool {
	return c.Trashed != 0
}

func (c *Card) IsDeleted() bool {
	return c.Deleted != 0
}

func (c *Card) Decrypt() (string, error) {
	// The key object is saved in binary from and actually consists of the
	// AES key (32 bytes) and a nonce (12 bytes) for GCM
	key := c.itemKey[:32]
	nonce := c.itemKey[32:]

	// If you deleted an item from Enpass, it stays in the database, but the
	// entries are cleared
	if len(nonce) == 0 {
		return "", errors.New("this item has been deleted")
	}

	// The value object holds the ciphertext (same length as plaintext) +
	// (authentication) tag (16 bytes) and is stored in hex
	ciphertextAndTag, err := hex.DecodeString(c.value)
	if err != nil {
		return "", errors.Wrap(err, "could not decode card hex cipherstring")
	}

	// As additional authenticated data (AAD) they use the UUID but without
	// the dashes: e.g. a2ec30c0aeed41f7aed7cc50e69ff506
	header, err := hex.DecodeString(strings.ReplaceAll(c.UUID, "-", ""))
	if err != nil {
		return "", errors.Wrap(err, "could not decode card hex AAD")
	}

	// Now we can initialize, decrypt the ciphertext and verify the AAD.
	// You can compare the SHA-1 output with the value stored in the db
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", errors.Wrap(err, "could not initialize card cipher")
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", errors.Wrap(err, "could not initialize GCM block")
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertextAndTag, header)
	if err != nil {
		return "", errors.Wrap(err, "could not decrypt data")
	}

	return string(plaintext), nil
}
