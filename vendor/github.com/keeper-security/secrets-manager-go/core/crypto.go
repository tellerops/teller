package core

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"errors"
	"io"
	"math/big"
	"strings"
)

const (
	Aes256KeySize    = 32
	AesGcmNonceSize  = 12
	AesCbcNonceSize  = 16
	DefaultBlockSize = 16
)

type PublicKey ecdsa.PublicKey
type PrivateKey ecdsa.PrivateKey

func PadBinary(s []byte) []byte {
	return pkcs7Pad(s)
}

func UnpadBinary(s []byte) []byte {
	return pkcs7Unpad(s)
}

// Bytes concatenates public key x and y values
func (pub *PublicKey) Bytes() (buf []byte) {
	x := pub.X.Bytes()
	y := pub.Y.Bytes()
	buf = append(x, y...)
	return
}

// SetBytes decodes buf and stores the values in pub X and Y
func (pub *PublicKey) SetBytes(buf []byte) *PublicKey {
	bigX := new(big.Int)
	bigY := new(big.Int)
	bigX.SetBytes(buf[:32])
	bigY.SetBytes(buf[32:64])

	pub.X = bigX
	pub.Y = bigY
	pub.Curve = elliptic.P256()
	return pub
}

// Check if public key is valid for the curve
func (pub *PublicKey) Check(curve elliptic.Curve) bool {
	if pub.Curve != curve {
		return false
	}
	if !curve.IsOnCurve(pub.X, pub.Y) {
		return false
	}
	return true
}

// Bytes returns private key D value
func (priv *PrivateKey) Bytes() []byte {
	return priv.D.Bytes()
}

// SetBytes reconstructs the private key from D bytes
func (priv *PrivateKey) SetBytes(d []byte) *PrivateKey {
	bigD := new(big.Int)
	bigD.SetBytes(d)
	priv.D = bigD
	priv.Curve = elliptic.P256()
	if priv.PublicKey.X == nil {
		priv.PublicKey.Curve = elliptic.P256()
		priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(priv.D.Bytes())
	}
	return priv
}

// GetPublicKey returns the associated PublicKey for this privatekey,
// If the key is missing then one is generated.
func (priv *PrivateKey) GetPublicKey() *PublicKey {
	if priv.PublicKey.X == nil {
		priv.PublicKey.Curve = elliptic.P256()
		priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(priv.D.Bytes())
	}
	return (*PublicKey)(&priv.PublicKey)
	//return PublicKey(priv.PublicKey)
}

// Hex returns private key bytes as a hex string
func (priv *PrivateKey) Hex() string {
	return hex.EncodeToString(priv.Bytes())
}

// Equals compares two private keys with constant time (to resist timing attacks)
func (priv *PrivateKey) Equals(k *PrivateKey) bool {
	return subtle.ConstantTimeCompare(priv.D.Bytes(), k.D.Bytes()) == 1
}

// Sign signs digest with priv, reading randomness from rand.
//
//	The opts argument is not currently used but, in keeping with the crypto.Signer interface,
//	should be the hash function used to digest the message.
func (priv *PrivateKey) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error) {
	return (*ecdsa.PrivateKey)(priv).Sign(rand, digest, opts)
}

func GenerateP256Keys() (PrivateKey, error) {
	return GenerateKeys(elliptic.P256()) // golang suppors only SECP256R1
}

func GenerateKeys(curve elliptic.Curve) (PrivateKey, error) {
	k, err := ecdsa.GenerateKey(curve, rand.Reader)
	return PrivateKey(*k), err
}

func GeneratePrivateKeyEcc() (PrivateKey, error) {
	return GenerateP256Keys()
}

func GeneratePrivateKeyDer() ([]byte, error) {
	privateKey, err := GeneratePrivateKeyEcc()
	if err != nil {
		return []byte{}, err
	}
	// Export to DER - PKCS #8 ASN.1 DER form with NoEncryption
	if privateKeyDer, err := x509.MarshalPKCS8PrivateKey((*ecdsa.PrivateKey)(&privateKey)); err != nil {
		return []byte{}, err
	} else {
		return privateKeyDer, nil
	}
}

func GenerateNewEccKey() (PrivateKey, error) {
	return GenerateP256Keys()
}

func EcPublicKeyFromEncodedPoint(publicKey []byte) (crypto.PublicKey, error) {
	// see https://tools.ietf.org/html/rfc6637#section-6
	if x, y := elliptic.Unmarshal(elliptic.P256(), publicKey); x != nil {
		return PublicKey{Curve: elliptic.P256(), X: x, Y: y}, nil
	} else {
		return PublicKey{}, errors.New("bad ECC public key")
	}
}

func EcPublicKeyToEncodedPoint(pub *ecdsa.PublicKey) ([]byte, error) {
	// see https://tools.ietf.org/html/rfc6637#section-6
	if pub.Curve != elliptic.P256() {
		return nil, errors.New("unsupported ECC curve type")
	}
	return elliptic.Marshal(pub.Curve, pub.X, pub.Y), nil
}

// Encrypt a message using AES-GCM.
func EncryptAesGcm(data []byte, key []byte) ([]byte, error) {
	return EncryptAesGcmFull(data, key, nil)
}

// Encrypt a message using AES-GCM with custom nonce.
func EncryptAesGcmFull(data, key, nonce []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	if len(nonce) == 0 {
		nonce, err = GetRandomBytes(AesGcmNonceSize)
		if err != nil {
			return nil, err
		}
	}
	if len(nonce) != AesGcmNonceSize {
		return nil, errors.New("incorrect nonce size")
	}

	result := gcm.Seal(nonce, nonce, data, nil)
	return result, nil
}

// Decrypt AES-GCM encrypted message
func Decrypt(data, key []byte) ([]byte, error) {
	if len(data) <= AesGcmNonceSize {
		return nil, errors.New("error decrypting AES-GCM - message is too short")
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, AesGcmNonceSize)
	copy(nonce, data)

	result, err := gcm.Open(nil, nonce, data[AesGcmNonceSize:], nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Encrypt a message using AES-CBC.
func EncryptAesCbc(data []byte, key []byte) ([]byte, error) {
	return EncryptAesCbcFull(data, key, nil)
}

// Encrypt a message using AES-CBC with custom nonce.
func EncryptAesCbcFull(data, key, nonce []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(nonce) == 0 {
		nonce, err = GetRandomBytes(AesCbcNonceSize)
		if err != nil {
			return nil, err
		}
	}
	if len(nonce) != AesCbcNonceSize {
		return nil, errors.New("incorrect nonce size")
	}

	cbc := cipher.NewCBCEncrypter(c, nonce)
	data = pkcs7Pad(data)
	encrypted := make([]byte, len(data))
	cbc.CryptBlocks(encrypted, data)

	result := append(nonce, encrypted...)
	return result, nil
}

// Decrypt AES-CBC encrypted message
func DecryptAesCbc(data, key []byte) ([]byte, error) {
	if len(data) <= AesCbcNonceSize {
		return nil, errors.New("error decrypting AES-CBC - message is too short")
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, AesCbcNonceSize)
	copy(nonce, data)
	cbc := cipher.NewCBCDecrypter(c, nonce)

	data = data[AesCbcNonceSize:]
	decrypted := make([]byte, len(data))
	cbc.CryptBlocks(decrypted, data)

	result := pkcs7Unpad(decrypted)
	return result, nil
}

func PublicEncrypt(data []byte, serverPublicRawKeyBytes []byte, idz []byte) (encrypted []byte, err error) {
	ephemeralKey2, err := GenerateNewEccKey()
	if err != nil {
		return nil, err
	}
	ephemeralKey2PublicKey := (*ecdsa.PublicKey)(ephemeralKey2.GetPublicKey())

	ephemeralPublicKey, err := EcPublicKeyFromEncodedPoint(serverPublicRawKeyBytes)
	if err != nil {
		return nil, err
	}

	epk, ok := ephemeralPublicKey.(PublicKey)
	if !ok {
		return nil, errors.New("bad format for ECC public key")
	}

	sharedKey, err := ECDH(ephemeralKey2, epk)
	if err != nil {
		return nil, err
	}

	encryptedData, err := EncryptAesGcm(data, sharedKey)
	if err != nil {
		return nil, err
	}

	ephPublicKey, err := EcPublicKeyToEncodedPoint(ephemeralKey2PublicKey)
	if err != nil {
		return nil, err
	}
	encrypted = append(ephPublicKey, encryptedData...)

	return encrypted, nil
}

func DecryptRecord(data, secretKey []byte) (string, error) {
	if record, err := Decrypt(data, secretKey); err == nil {
		recordJson := BytesToString(record)
		return recordJson, nil
	} else {
		return "", err
	}
}

func LoadDerPrivateKeyDer(data []byte) (*PrivateKey, error) {
	if len(data) < 1 {
		return nil, errors.New("private key data is empty")
	}
	// Import private key - PKCS #8 ASN.1 DER form with NoEncryption
	if key, err := x509.ParsePKCS8PrivateKey(data); err == nil {
		switch k := key.(type) {
		case *ecdsa.PrivateKey:
			return (*PrivateKey)(k), nil
		case *rsa.PrivateKey:
			return nil, errors.New("private key is in an unsupported format: RSA Private Key")
		case ed25519.PrivateKey:
			return nil, errors.New("private key is in an unsupported format: Ed25519 Private Key")
		default:
			return nil, errors.New("private key is in an unsupported format")
		}
	} else {
		return nil, errors.New("private key data parsing error: " + err.Error())
	}
}

func DerBase64PrivateKeyToPrivateKey(privateKeyDerBase64 string) (*PrivateKey, error) {
	if strings.TrimSpace(privateKeyDerBase64) != "" {
		privateKeyDerBase64Bytes := Base64ToBytes(privateKeyDerBase64)
		return LoadDerPrivateKeyDer(privateKeyDerBase64Bytes)
	}
	return nil, errors.New("private key data is empty")
}

func extractPublicKeyBytes(privateKeyDerBase64 interface{}) ([]byte, error) {
	pkDerBase64 := ""
	switch v := privateKeyDerBase64.(type) {
	case string:
		pkDerBase64 = v
	case []byte:
		pkDerBase64 = BytesToBase64(v)
	default:
		return nil, errors.New("extracting public key DER bytes failed - PK must be string or byte slice")
	}

	if ecPrivateKey, err := DerBase64PrivateKeyToPrivateKey(pkDerBase64); err == nil {
		pubKey := ecPrivateKey.GetPublicKey()
		if pubKeyBytes, err := EcPublicKeyToEncodedPoint((*ecdsa.PublicKey)(pubKey)); err == nil {
			return pubKeyBytes, nil
		} else {
			return nil, errors.New("error extracting public key from DER: " + err.Error())
		}
	} else {
		return nil, errors.New("error extracting private key from DER: " + err.Error())
	}
}

func Sign(data []byte, privateKey *PrivateKey) ([]byte, error) {
	msgHash := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand.Reader, (*ecdsa.PrivateKey)(privateKey), msgHash[:])
	if err != nil {
		return []byte{}, errors.New("signature generation failed: " + err.Error())
	}
	ecdsaSig := ECDSASignature{R: r, S: s}
	if signature, err := asn1.Marshal(ecdsaSig); err == nil {
		return signature, nil
	} else {
		return []byte{}, errors.New("signature serialization failed: " + err.Error())
	}
}

// Verify validates decrypted message against the given public key.
// On success, returns nil, on failure returns a relevant error.
func Verify(data []byte, signature []byte, publicKey *PublicKey) error {
	sig := &ECDSASignature{}
	_, err := asn1.Unmarshal(signature, sig)
	if err != nil {
		return err
	}
	h := sha256.Sum256(data)
	valid := ecdsa.Verify(
		(*ecdsa.PublicKey)(publicKey),
		h[:],
		sig.R,
		sig.S,
	)
	if !valid {
		return errors.New("signature validation failed")
	}
	// signature is valid
	return nil
}

// ErrKeyExchange is returned if the key exchange fails.
var ErrKeyExchange = errors.New("key exchange failed")

// ECDH computes a shared key from a private key and a peer's public key.
func ECDH(priv PrivateKey, pub PublicKey) ([]byte, error) {
	privKey := (*ecdsa.PrivateKey)(&priv)
	pubKey := (*ecdsa.PublicKey)(&pub)
	return ECDH_Ecdsa(privKey, pubKey)
}

// ECDH computes a shared key from a private key and a peer's public key.
func ECDH_Ecdsa(priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey) ([]byte, error) {
	if pub == nil || priv == nil {
		return nil, ErrKeyExchange
	} else if priv.Curve != pub.Curve {
		return nil, ErrKeyExchange
	} else if !priv.Curve.IsOnCurve(pub.X, pub.Y) {
		return nil, ErrKeyExchange
	}

	x, _ := pub.Curve.ScalarMult(pub.X, pub.Y, priv.D.Bytes())
	if x == nil {
		return nil, ErrKeyExchange
	}

	// x.Bytes() may return less than 32 bytes - pad with leading 0
	buf := x.Bytes()
	if len(buf) < 32 {
		buf = x.FillBytes(make([]byte, 32))
	}
	shared := sha256.Sum256(buf)
	return shared[:Aes256KeySize], nil
}

func pkcs7Pad(data []byte) []byte {
	// With PKCS#7, we’re always going to pad,
	// so if our block length was 16, and our plaintext length was 16,
	// then it would be padded with 16 bytes of 16 at the end.
	n := DefaultBlockSize - (len(data) % DefaultBlockSize)
	pb := make([]byte, len(data)+n)
	copy(pb, data)
	copy(pb[len(data):], bytes.Repeat([]byte{byte(n)}, n))
	return pb
}

func pkcs7Unpad(data []byte) []byte {
	if len(data) > 0 && len(data)%DefaultBlockSize == 0 {
		c := data[len(data)-1]
		if n := int(c); n > 0 && n <= DefaultBlockSize {
			ok := true
			for i := 0; i < n; i++ {
				if data[len(data)-n+i] != c {
					ok = false
					break
				}
			}
			if ok {
				return data[:len(data)-n]
			}
		}
	}
	return data
}
