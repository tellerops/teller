package gokeepasslib

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
)

// Inner header bytes
const (
	InnerHeaderTerminator byte = 0x00 // Inner header terminator byte
	InnerHeaderIRSID      byte = 0x01 // Inner header InnerRandomStreamID byte
	InnerHeaderIRSKey     byte = 0x02 // Inner header InnerRandomStreamKey byte
	InnerHeaderBinary     byte = 0x03 // Inner header binary byte
)

// InnerHeader is the container of crypt options and binaries, only for Kdbx v4
type InnerHeader struct {
	InnerRandomStreamID  uint32
	InnerRandomStreamKey []byte
	Binaries             Binaries
}

// DBContent is a container for all elements of a keepass database
type DBContent struct {
	RawData     []byte       `xml:"-"` // XML encoded original data
	InnerHeader *InnerHeader `xml:"-"`
	XMLName     xml.Name     `xml:"KeePassFile"`
	Meta        *MetaData    `xml:"Meta"`
	Root        *RootData    `xml:"Root"`
}

func (c *DBContent) setKdbxFormatVersion(version formatVersion) {
	c.Meta.setKdbxFormatVersion(version)
	c.Root.setKdbxFormatVersion(version)
}

type DBContentOption func(*DBContent)

func WithDBContentFormattedTime(formatted bool) DBContentOption {
	return func(content *DBContent) {
		WithMetaDataFormattedTime(formatted)(content.Meta)
		WithRootDataFormattedTime(formatted)(content.Root)
	}
}

func withDBContentKDBX4InnerHeader(content *DBContent) {
	innerRandomStreamKey := make([]byte, 64)
	rand.Read(innerRandomStreamKey)

	content.InnerHeader = &InnerHeader{
		InnerRandomStreamID:  ChaChaStreamID,
		InnerRandomStreamKey: innerRandomStreamKey,
	}
}

// NewContent creates a new database content with some good defaults
func NewContent(options ...DBContentOption) *DBContent {
	// Not necessary create InnerHeader because this will be a KDBX v3.1
	content := &DBContent{
		Meta: NewMetaData(),
		Root: NewRootData(),
	}

	for _, option := range options {
		option(content)
	}

	return content
}

// readFrom reads the InnerHeader from an io.Reader
func (ih *InnerHeader) readFrom(r io.Reader) error {
	binaryCount := 0 // Var used to count and index every binary
ForLoop:
	for {
		var headerType byte
		var length int32
		var data []byte

		if err := binary.Read(r, binary.LittleEndian, &headerType); err != nil {
			return err
		}
		if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
			return err
		}
		data = make([]byte, length)
		if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
			return err
		}

		switch headerType {
		case InnerHeaderTerminator:
			// End of inner header
			break ForLoop
		case InnerHeaderIRSID:
			// Found InnerRandomStream ID
			ih.InnerRandomStreamID = binary.LittleEndian.Uint32(data)
		case InnerHeaderIRSKey:
			// Found InnerRandomStream Key
			ih.InnerRandomStreamKey = data
		case InnerHeaderBinary:
			// Found a binary
			var protection byte
			reader := bytes.NewReader(data)

			binary.Read(reader, binary.LittleEndian, &protection) // Read memory protection flag
			content, _ := ioutil.ReadAll(reader)                  // Read content

			ih.Binaries = append(ih.Binaries, Binary{
				ID:               binaryCount,
				MemoryProtection: protection,
				Content:          content,
			})

			binaryCount = binaryCount + 1
		default:
			return ErrUnknownInnerHeaderID(headerType)
		}
	}
	return nil
}

// writeTo the InnerHeader to the given io.Writer
func (ih *InnerHeader) writeTo(w io.Writer) error {
	irsID := make([]byte, 4)
	binary.LittleEndian.PutUint32(irsID, ih.InnerRandomStreamID)

	if err := writeToInnerHeader(w, InnerHeaderIRSID, irsID); err != nil {
		return err
	}
	if err := writeToInnerHeader(w, InnerHeaderIRSKey, ih.InnerRandomStreamKey); err != nil {
		return err
	}

	for _, item := range ih.Binaries {
		buf := []byte{item.MemoryProtection}
		buf = append(buf, item.Content...)
		if err := writeToInnerHeader(w, InnerHeaderBinary, buf); err != nil {
			return err
		}
	}
	// End inner header
	if err := binary.Write(w, binary.LittleEndian, InnerHeaderTerminator); err != nil {
		return err
	}
	return binary.Write(w, binary.LittleEndian, uint32(0))
}

// writeToInnerHeader is an helper to write an inner header item to the given io.Writer
func writeToInnerHeader(w io.Writer, id uint8, data []byte) error {
	if len(data) > 0 {
		if err := binary.Write(w, binary.LittleEndian, id); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, uint32(len(data))); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, data); err != nil {
			return err
		}
	}
	return nil
}

func (ih InnerHeader) String() string {
	return fmt.Sprintf(
		"1) InnerRandomStreamID: %d\n"+
			"2) InnerRandomStreamKey: %x\n"+
			"3) Binaries: %s\n",
		ih.InnerRandomStreamID,
		ih.InnerRandomStreamKey,
		ih.Binaries,
	)
}

// ErrEndOfInnerHeaders is the error returned when the end of inner header is read
var ErrEndOfInnerHeaders = errors.New("gokeepasslib: inner header id was 0, end of inner headers")

// ErrUnknownInnerHeaderID is the error returned if an unknown inner header is read
type ErrUnknownInnerHeaderID byte

func (e ErrUnknownInnerHeaderID) Error() string {
	return fmt.Sprintf("gokeepasslib: unknown inner header ID of %d", int(e))
}
