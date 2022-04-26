package crypto

import (
	"crypto/cipher"
	"crypto/sha512"
	"encoding/base64"

	"github.com/aead/chacha20"
)

// ChaChaStream is a ChaCha20 cipher that implements Stream and Encrypter interface
type ChaChaStream struct {
	cipher cipher.Stream
}

// NewChaChaEncrypter initialize a new ChaChaStream interfaced with Encrypter
func NewChaChaEncrypter(key []byte, iv []byte) (*ChaChaStream, error) {
	cipher, err := chacha20.NewCipher(iv, key)
	if err != nil {
		return nil, err
	}

	c := ChaChaStream{
		cipher: cipher,
	}
	return &c, nil
}

// NewChaChaStream initialize a new ChaChaStream interfaced with Stream
func NewChaChaStream(key []byte) (*ChaChaStream, error) {
	hash := sha512.Sum512(key)

	cipher, err := chacha20.NewCipher(hash[32:44], hash[:32])
	if err != nil {
		return nil, err
	}

	c := ChaChaStream{
		cipher: cipher,
	}
	return &c, nil
}

// Decrypt returns the decrypted data
func (cs *ChaChaStream) Decrypt(data []byte) []byte {
	ret := make([]byte, len(data))
	cs.cipher.XORKeyStream(ret, data)
	return ret
}

// Encrypt returns the encrypted data
func (cs *ChaChaStream) Encrypt(data []byte) []byte {
	return cs.Decrypt(data)
}

// Unpack returns the payload as unencrypted byte array
func (cs *ChaChaStream) Unpack(payload string) []byte {
	decoded, _ := base64.StdEncoding.DecodeString(payload)

	data := make([]byte, len(decoded))
	cs.cipher.XORKeyStream(data, decoded)
	return data
}

// Pack returns the payload as encrypted string
func (cs *ChaChaStream) Pack(payload []byte) string {
	data := make([]byte, len(payload))

	cs.cipher.XORKeyStream(data, payload)
	str := base64.StdEncoding.EncodeToString(data)
	return str
}
