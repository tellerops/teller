package gokeepasslib

import (
	"encoding/binary"
	"fmt"
	"io"
)

// DBHashes stores the hashes of a Kdbx v4 database
type DBHashes struct {
	Sha256 [32]byte
	Hmac   [32]byte
}

// NewHashes creates a new DBHashes based on the given header
func NewHashes(header *DBHeader) *DBHashes {
	return &DBHashes{
		Sha256: header.GetSha256(),
	}
}

// readFrom reads the hashes from an io.Reader
func (h *DBHashes) readFrom(r io.Reader) error {
	return binary.Read(r, binary.LittleEndian, h)
}

// writeTo writes the hashes to the given io.Writer
func (h DBHashes) writeTo(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, h)
}

func (h DBHashes) String() string {
	return fmt.Sprintf(
		"(1) Sha256: %x\n"+
			"(2) Hmac: %x\n",
		h.Sha256,
		h.Hmac,
	)
}
