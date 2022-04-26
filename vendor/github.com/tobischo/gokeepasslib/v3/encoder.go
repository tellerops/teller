package gokeepasslib

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/xml"
	"io"
)

// Header to be put before xml content in kdbx file
var xmlHeader = []byte(`<?xml version="1.0" encoding="utf-8" standalone="yes"?>` + "\n")

// Encoder is used to automaticaly encrypt and write a database to a file, network, etc
type Encoder struct {
	w io.Writer
}

// NewEncoder creates a new encoder with writer w, identical to gokeepasslib.Encoder{w}
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode writes db to e's internal writer
func (e *Encoder) Encode(db *Database) error {
	// Unlock protected entries ensuring that we have them prepared in the order that is matching
	// the xml unmarshalling order
	err := db.UnlockProtectedEntries()
	if err != nil {
		return err
	}
	// Re-Lock the protected values mapping to ensure that they are locked in memory and
	// follow the order in which they would be written again
	err = db.LockProtectedEntries()
	if err != nil {
		return err
	}

	// ensure timestamps will be formatted correctly
	db.ensureKdbxFormatVersion()

	// Calculate transformed key to make HMAC and encrypt
	transformedKey, err := db.getTransformedKey()
	if err != nil {
		return err
	}

	// Write header then hashes before decode content (necessary to update HeaderHash)
	// db.Header writeTo will change its hash
	if err = db.Header.writeTo(e.w); err != nil {
		return err
	}

	// Update header hash into db.Hashes then write the data
	hash := db.Header.GetSha256()
	if db.Header.IsKdbx4() {
		db.Hashes.Sha256 = hash

		hmacKey := buildHmacKey(db, transformedKey)
		hmacHash := db.Header.GetHmacSha256(hmacKey)
		db.Hashes.Hmac = hmacHash

		if err = db.Hashes.writeTo(e.w); err != nil {
			return err
		}
	} else {
		db.Content.Meta.HeaderHash = base64.StdEncoding.EncodeToString(hash[:])
	}

	// Encode xml and append header to the top
	rawContent, err := xml.MarshalIndent(db.Content, "", "\t")
	if err != nil {
		return err
	}
	rawContent = append(xmlHeader, rawContent...)

	// Write InnerHeader (Kdbx v4)
	if db.Header.IsKdbx4() {
		var ih bytes.Buffer
		if err = db.Content.InnerHeader.writeTo(&ih); err != nil {
			return err
		}

		rawContent = append(ih.Bytes(), rawContent...)
	}

	// Encode raw content
	encodedContent, err := encodeRawContent(db, rawContent, transformedKey)
	if err != nil {
		return err
	}

	// Writes the encrypted database content
	_, err = e.w.Write(encodedContent)
	return err
}

func encodeRawContent(db *Database, content []byte, transformedKey []byte) (encoded []byte, err error) {
	// Compress if the header compression flag is 1 (gzip)
	if db.Header.FileHeaders.CompressionFlags == GzipCompressionFlag {
		b := new(bytes.Buffer)
		w := gzip.NewWriter(b)

		if _, err = w.Write(content); err != nil {
			return encoded, err
		}

		// Close() needs to be explicitly called to write Gzip stream footer,
		// Flush() is not enough. some gzip decoders treat missing footer as error
		// while some don't). internally Close() also does flush.
		if err = w.Close(); err != nil {
			return encoded, err
		}

		content = b.Bytes()
	}

	// Compose blocks (Kdbx v3.1)
	if !db.Header.IsKdbx4() {
		var blocks bytes.Buffer
		composeContentBlocks31(&blocks, content)

		// Append blocks to StreamStartBytes
		content = append(db.Header.FileHeaders.StreamStartBytes, blocks.Bytes()...)
	}

	// Adds padding to data as required to encrypt properly
	if len(content)%16 != 0 {
		padding := make([]byte, 16-(len(content)%16))
		for i := 0; i < len(padding); i++ {
			padding[i] = byte(len(padding))
		}
		content = append(content, padding...)
	}

	// Encrypt content
	// Decrypt content
	encrypter, err := db.GetEncrypterManager(transformedKey)
	if err != nil {
		return encoded, err
	}
	encrypted := encrypter.Encrypt(content)

	// Compose blocks (Kdbx v4)
	if db.Header.IsKdbx4() {
		var blocks bytes.Buffer
		composeContentBlocks4(&blocks, encrypted, db.Header.FileHeaders.MasterSeed, transformedKey)

		encrypted = blocks.Bytes()
	}
	return encrypted, nil
}
