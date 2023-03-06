package vault

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

func encrypt(data []byte, salt []byte, key *key) ([]byte, error) {
	aesCipher, err := aes.NewCipher(key.cipherKey)
	if err != nil {
		return nil, err
	}

	plaintext := pad(data)
	ciphertext := make([]byte, len(plaintext))

	aesBlock := cipher.NewCTR(aesCipher, key.iv)
	aesBlock.XORKeyStream(ciphertext, plaintext)

	return ciphertext, nil
}

func decrypt(secret *secret, key *key) (string, error) {
	aesCipher, err := aes.NewCipher(key.cipherKey)
	if err != nil {
		return "", err
	}

	plainText := make([]byte, len(secret.data))

	aesBlock := cipher.NewCTR(aesCipher, key.iv)
	aesBlock.XORKeyStream(plainText, secret.data)

	result, err := unpad(plainText)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

func pad(src []byte) []byte {
	padlen := aes.BlockSize - len(src)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padlen)}, padlen)
	return append(src, padtext...)
}

func unpad(src []byte) ([]byte, error) {
	length := len(src)
	padlen := int(src[length-1])
	if padlen > length {
		return nil, ErrInvalidPadding
	}
	return src[:(length - padlen)], nil
}
