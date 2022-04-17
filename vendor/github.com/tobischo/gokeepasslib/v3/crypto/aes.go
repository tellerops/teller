package crypto

import (
	"crypto/aes"
	"crypto/cipher"
)

// AESEncrypter is an AES cipher that implements Encrypter interface
type AESEncrypter struct {
	block        cipher.Block
	encryptionIV []byte
}

// NewAESEncrypter initialize a new AESEncrypter interfaced with Encrypter
func NewAESEncrypter(key []byte, iv []byte) (*AESEncrypter, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	e := AESEncrypter{
		block:        block,
		encryptionIV: iv,
	}
	return &e, nil
}

// Decrypt returns the decrypted data
func (ae *AESEncrypter) Decrypt(data []byte) []byte {
	ret := make([]byte, len(data))
	mode := cipher.NewCBCDecrypter(ae.block, ae.encryptionIV)
	mode.CryptBlocks(ret, data)
	return ret
}

// Encrypt returns the encrypted data
func (ae *AESEncrypter) Encrypt(data []byte) []byte {
	ret := make([]byte, len(data))
	mode := cipher.NewCBCEncrypter(ae.block, ae.encryptionIV)
	mode.CryptBlocks(ret, data)
	return ret
}
