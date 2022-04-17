package gokeepasslib

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
)

// ErrInvalidUUIDLength is an error which is returned during unmarshaling if the UUID does not have 16 bytes length
var ErrInvalidUUIDLength = errors.New("gokeepasslib: length of decoded UUID was not 16")

// UUID stores a universal identifier for each group+entry
type UUID [16]byte

// NewUUID returns a new randomly generated UUID
func NewUUID() UUID {
	var id UUID
	rand.Read(id[:])
	return id
}

// Compare allowes to check whether two instance of UUID are equal in value.
// This is used for searching a uuid
func (u UUID) Compare(c UUID) bool {
	for i, v := range c {
		if u[i] != v {
			return false
		}
	}
	return true
}

// MarshalText is a marshaler method to encode uuid content as base 64 and return it
func (u UUID) MarshalText() ([]byte, error) {
	text := make([]byte, 24)
	base64.StdEncoding.Encode(text, u[:])
	return text, nil
}

// UnmarshalText unmarshals a byte slice into a UUID by decoding the given data from base64
func (u *UUID) UnmarshalText(text []byte) error {
	id := make([]byte, base64.StdEncoding.DecodedLen(len(text)))
	length, err := base64.StdEncoding.Decode(id, text)
	if err != nil {
		return err
	}
	if length == 0 {
		*u = NewUUID()
		return nil
	}
	if length != 16 {
		return ErrInvalidUUIDLength
	}
	copy((*u)[:], id[:16])
	return nil
}
