package gokeepasslib

import (
	"crypto/aes"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"

	"github.com/aead/argon2"
)

// DBCredentials holds the key used to lock and unlock the database
type DBCredentials struct {
	Passphrase []byte // Passphrase if using one, stored in sha256 hash
	Key        []byte // Contents of the keyfile if using one, stored in sha256 hash
	Windows    []byte // Whatever is returned from windows user account auth, stored in sha256 hash
}

func (c *DBCredentials) buildCompositeKey() ([]byte, error) {
	hash := sha256.New()
	if c.Passphrase != nil { // If the hashed password is provided
		_, err := hash.Write(c.Passphrase)
		if err != nil {
			return nil, err
		}
	}
	if c.Key != nil { // If the hashed keyfile is provided
		_, err := hash.Write(c.Key)
		if err != nil {
			return nil, err
		}
	}
	if c.Windows != nil { // If the hashed password is provided
		_, err := hash.Write(c.Windows)
		if err != nil {
			return nil, err
		}
	}
	return hash.Sum(nil), nil
}

func (c *DBCredentials) buildTransformedKey(db *Database) ([]byte, error) {
	transformedKey, err := c.buildCompositeKey()
	if err != nil {
		return nil, err
	}

	if db.Header.IsKdbx4() {
		if reflect.DeepEqual(db.Header.FileHeaders.KdfParameters.UUID, KdfArgon2) {
			// Argon 2
			transformedKey = argon2.Key2d(
				transformedKey, // Master key
				db.Header.FileHeaders.KdfParameters.Salt[:],             // Salt
				uint32(db.Header.FileHeaders.KdfParameters.Iterations),  // Time cost
				uint32(db.Header.FileHeaders.KdfParameters.Memory)/1024, // Memory cost
				uint8(db.Header.FileHeaders.KdfParameters.Parallelism),  // Parallelism
				32, // Hash length
			)
		} else {
			// AES
			key, err := cryptAESKey(
				transformedKey,
				db.Header.FileHeaders.KdfParameters.Salt[:],
				db.Header.FileHeaders.KdfParameters.Rounds,
			)
			if err != nil {
				return nil, err
			}
			transformedKey = key[:]
		}
	} else {
		// AES
		key, err := cryptAESKey(
			transformedKey,
			db.Header.FileHeaders.TransformSeed,
			db.Header.FileHeaders.TransformRounds,
		)
		if err != nil {
			return nil, err
		}
		transformedKey = key[:]
	}
	return transformedKey, nil
}

func buildMasterKey(db *Database, transformedKey []byte) []byte {
	masterKey := sha256.New()
	masterKey.Write(db.Header.FileHeaders.MasterSeed)
	masterKey.Write(transformedKey)
	return masterKey.Sum(nil)
}

func buildHmacKey(db *Database, transformedKey []byte) []byte {
	masterKey := sha512.New()
	masterKey.Write(db.Header.FileHeaders.MasterSeed)
	masterKey.Write(transformedKey)
	masterKey.Write([]byte{0x01})
	hmacKey := sha512.New()
	hmacKey.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
	hmacKey.Write(masterKey.Sum(nil))
	return hmacKey.Sum(nil)
}

func cryptAESKey(masterKey []byte, seed []byte, rounds uint64) ([]byte, error) {
	block, err := aes.NewCipher(seed)
	if err != nil {
		return nil, err
	}

	newKey := make([]byte, len(masterKey))
	copy(newKey, masterKey)

	for i := uint64(0); i < rounds; i++ {
		block.Encrypt(newKey, newKey)
		block.Encrypt(newKey[16:], newKey[16:])
	}

	hash := sha256.Sum256(newKey)
	return hash[:], nil
}

// NewPasswordCredentials builds a new DBCredentials from a Password string
func NewPasswordCredentials(password string) *DBCredentials {
	hashedpw := sha256.Sum256([]byte(password))
	return &DBCredentials{Passphrase: hashedpw[:]}
}

// ParseKeyFile returns the hashed key from a key file at the path specified by location, parsing xml if needed
func ParseKeyFile(location string) ([]byte, error) {
	file, err := os.Open(location)
	if err != nil {
		return nil, err
	}

	var data []byte
	if data, err = ioutil.ReadAll(file); err != nil {
		return nil, err
	}

	return ParseKeyData(data)
}

var keyDataPattern = regexp.MustCompile(`<Data>(.+)</Data>`)

// ParseKeyData returns the hashed key from a key file in bytes, parsing xml if needed
func ParseKeyData(data []byte) ([]byte, error) {
	if keyDataPattern.Match(data) { //If keyfile is in xml form, extract key data
		base := keyDataPattern.FindSubmatch(data)[1]
		data = make([]byte, base64.StdEncoding.DecodedLen(len(base)))
		if _, err := base64.StdEncoding.Decode(data, base); err != nil {
			return nil, err
		}
	}

	if len(data) < 32 {
		return data, nil
	}

	// Slice necessary due to padding at the end of the hash
	return data[:32], nil
}

// NewKeyCredentials builds a new DBCredentials from a key file at the path specified by location
func NewKeyCredentials(location string) (*DBCredentials, error) {
	key, err := ParseKeyFile(location)
	if err != nil {
		return nil, err
	}

	return &DBCredentials{Key: key}, nil
}

// NewKeyDataCredentials builds a new DBCredentials from a key file in bytes
func NewKeyDataCredentials(data []byte) (*DBCredentials, error) {
	key, err := ParseKeyData(data)
	if err != nil {
		return nil, err
	}

	return &DBCredentials{Key: key}, nil
}

// NewPasswordAndKeyCredentials builds a new DBCredentials from a password and the key file at the path specified by location
func NewPasswordAndKeyCredentials(password, location string) (*DBCredentials, error) {
	key, err := ParseKeyFile(location)
	if err != nil {
		return nil, err
	}

	hashedpw := sha256.Sum256([]byte(password))

	return &DBCredentials{
		Passphrase: hashedpw[:],
		Key:        key,
	}, nil
}

// NewPasswordAndKeyDataCredentials builds a new DBCredentials from a password and the key file in bytes
func NewPasswordAndKeyDataCredentials(password string, data []byte) (*DBCredentials, error) {
	key, err := ParseKeyData(data)
	if err != nil {
		return nil, err
	}

	hashedpw := sha256.Sum256([]byte(password))

	return &DBCredentials{
		Passphrase: hashedpw[:],
		Key:        key,
	}, nil
}

func (c *DBCredentials) String() string {
	return fmt.Sprintf(
		"Hashed Passphrase: %x\nHashed Key: %x\nHashed Windows Auth: %x",
		c.Passphrase,
		c.Key,
		c.Windows,
	)
}
