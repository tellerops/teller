package gokeepasslib

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"

	w "github.com/tobischo/gokeepasslib/v3/wrappers"
)

// Binaries Stores a slice of binaries in the metadata header of a database
// This will be used only on KDBX 3.1
// Since KDBX 4, binaries are stored into the InnerHeader
type Binaries []Binary

// Binary stores a binary found in the metadata header of a database
type Binary struct {
	ID               int           `xml:"ID,attr"`         // Index of binary (Manually counted on KDBX v4)
	MemoryProtection byte          `xml:"-"`               // Memory protection flag (Only KDBX v4)
	Content          []byte        `xml:",innerxml"`       // Binary content
	Compressed       w.BoolWrapper `xml:"Compressed,attr"` // Compressed flag (Only KDBX v3.1)
}

// BinaryReference stores a reference to a binary which appears in the xml of an entry
type BinaryReference struct {
	Name  string `xml:"Key"`
	Value struct {
		ID int `xml:"Ref,attr"`
	} `xml:"Value"`
}

// Find returns a reference to a binary with the same ID as id, or nil if none if found
func (bs Binaries) Find(id int) *Binary {
	for i := range bs {
		if bs[i].ID == id {
			return &bs[i]
		}
	}
	return nil
}

// Find returns a reference to a binary in the database db with the same id as br, or nil if none is found
func (br *BinaryReference) Find(db *Database) *Binary {
	if db.Header.IsKdbx4() {
		return db.Content.InnerHeader.Binaries.Find(br.Value.ID)
	}
	return db.Content.Meta.Binaries.Find(br.Value.ID)
}

// Add appends binary data to the slice
func (bs *Binaries) Add(c []byte) *Binary {
	binary := Binary{
		Compressed: w.NewBoolWrapper(true),
	}
	if len(*bs) == 0 {
		binary.ID = 0
	} else {
		binary.ID = (*bs)[len(*bs)-1].ID + 1
	}
	binary.SetContent(c)
	*bs = append(*bs, binary)
	return &(*bs)[len(*bs)-1]
}

// GetContentBytes returns a bytes slice containing content of a binary
func (b Binary) GetContentBytes() ([]byte, error) {
	// Check for base64 content (KDBX 3.1), if it fail try with KDBX 4
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(b.Content)))
	_, err := base64.StdEncoding.Decode(decoded, b.Content)
	if err != nil {
		// KDBX 4 doesn't encode it
		decoded = b.Content[:]
	}

	if b.Compressed.Bool {
		reader, err := gzip.NewReader(bytes.NewReader(decoded))
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		bts, err := ioutil.ReadAll(reader)
		if err != nil && err != io.ErrUnexpectedEOF {
			return nil, err
		}
		return bts, nil
	}
	return decoded, nil
}

// GetContentString returns the content of a binary as a string
func (b Binary) GetContentString() (string, error) {
	data, err := b.GetContentBytes()

	if err != nil {
		return "", err
	}

	return string(data), nil
}

// GetContent returns a string which is the plaintext content of a binary
//
// Deprecated: use GetContentString() instead
func (b Binary) GetContent() (string, error) {
	return b.GetContentString()
}

// SetContent encodes and (if Compressed=true) compresses c and sets b's content
func (b *Binary) SetContent(c []byte) error {
	buff := &bytes.Buffer{}
	writer := base64.NewEncoder(base64.StdEncoding, buff)
	if b.Compressed.Bool {
		writer = gzip.NewWriter(writer)
	}
	_, err := writer.Write(c)
	if err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	b.Content = buff.Bytes()
	return nil
}

// CreateReference creates a reference with the same id as b with filename f
func (b Binary) CreateReference(f string) BinaryReference {
	return NewBinaryReference(f, b.ID)
}

// NewBinaryReference creates a new BinaryReference with the given name and id
func NewBinaryReference(name string, id int) BinaryReference {
	ref := BinaryReference{}
	ref.Name = name
	ref.Value.ID = id
	return ref
}

func (b Binary) String() string {
	return fmt.Sprintf(
		"ID: %d, MemoryProtection: %x, Compressed:%#v, Content:%x",
		b.ID,
		b.MemoryProtection,
		b.Compressed,
		b.Content,
	)
}
func (br BinaryReference) String() string {
	return fmt.Sprintf("ID: %d, File Name: %s", br.Value.ID, br.Name)
}
