package vault

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

const vaultHeader = "$ANSIBLE_VAULT;1.1;AES256"

type secret struct {
	salt []byte
	hmac []byte
	data []byte
}

func decodeSecret(input string) (*secret, error) {
	lines := strings.SplitN(input, "\n", 3)
	if len(lines) != 3 {
		return nil, errors.New("invalid secret")
	}

	salt, err := hex.DecodeString(lines[0])
	if err != nil {
		return nil, err
	}

	hmac, err := hex.DecodeString(lines[1])
	if err != nil {
		return nil, err
	}

	data, err := hex.DecodeString(lines[2])
	if err != nil {
		return nil, err
	}

	return &secret{salt, hmac, data}, nil
}

func encodeSecret(secret *secret, key *key) (string, error) {
	hmacEncrypt := hmac.New(sha256.New, key.hmacKey)
	hmacEncrypt.Write(secret.data)
	hexSalt := hex.EncodeToString(secret.salt)
	hexHmac := hmacEncrypt.Sum(nil)
	hexCipher := hex.EncodeToString(secret.data)

	combined := strings.Join([]string{
		string(hexSalt),
		hex.EncodeToString([]byte(hexHmac)),
		string(hexCipher),
	}, "\n")

	result := strings.Join([]string{
		vaultHeader,
		wrapText(hex.EncodeToString([]byte(combined))),
	}, "\n")

	return result, nil
}

func checkDigest(secret *secret, key *key) error {
	hash := hmac.New(sha256.New, key.hmacKey)
	hash.Write(secret.data)
	if !hmac.Equal(hash.Sum(nil), secret.hmac) {
		return errors.New("invalid password")
	}
	return nil
}

func wrapText(text string) string {
	src := []byte(text)
	result := []byte{}

	for i := 0; i < len(src); i++ {
		if i > 0 && i%80 == 0 {
			result = append(result, '\n')
		}
		result = append(result, src[i])
	}

	return string(result)
}

func hexDecode(input string) (string, error) {
	input = strings.TrimSpace(input)
	input = strings.Replace(input, "\r", "", -1)
	input = strings.Replace(input, "\n", "", -1)

	decoded, err := hex.DecodeString(input)
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}
