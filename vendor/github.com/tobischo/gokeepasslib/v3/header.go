package gokeepasslib

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
)

// BaseSignature is the valid base signature for kdbx files
var BaseSignature = [...]byte{0x03, 0xd9, 0xa2, 0x9a}

// SecondarySignature is the valid version signature for kdbx files
var SecondarySignature = [...]byte{0x67, 0xfb, 0x4b, 0xb5}

// DefaultKDBX3Sig is the full valid default signature struct for new databases (Kdbx v3.1)
var DefaultKDBX3Sig = Signature{BaseSignature, SecondarySignature, 1, 3}

// DefaultKDBX4Sig is the full valid default signature struct for new databases (Kdbx v4.0)
var DefaultKDBX4Sig = Signature{BaseSignature, SecondarySignature, 0, 4}

// DefaultSig is the full valid default signature struct for new databases (Kdbx v3.1)
var DefaultSig = DefaultKDBX3Sig

// Compression flags
const (
	NoCompressionFlag   uint32 = 0 // No compression flag
	GzipCompressionFlag uint32 = 1 // Gzip compression flag
)

// CipherAES is the AES cipher ID
var CipherAES = []byte{0x31, 0xC1, 0xF2, 0xE6, 0xBF, 0x71, 0x43, 0x50, 0xBE, 0x58, 0x05, 0x21, 0x6A, 0xFC, 0x5A, 0xFF}

// CipherTwoFish is the TwoFish cipher ID
var CipherTwoFish = []byte{0xAD, 0x68, 0xF2, 0x9F, 0x57, 0x6F, 0x4B, 0xB9, 0xA3, 0x6A, 0xD4, 0x7A, 0xF9, 0x65, 0x34, 0x6C}

// CipherChaCha20 is the ChaCha20 cipher ID
var CipherChaCha20 = []byte{0xD6, 0x03, 0x8A, 0x2B, 0x8B, 0x6F, 0x4C, 0xB5, 0xA5, 0x24, 0x33, 0x9A, 0x31, 0xDB, 0xB5, 0x9A}

// KdfAES3 is the AES key derivation function ID for Kdbx v3.1
var KdfAES3 = []byte{0xC9, 0xD9, 0xF3, 0x9A, 0x62, 0x8A, 0x44, 0x60, 0xBF, 0x74, 0x0D, 0x08, 0xC1, 0x8A, 0x4F, 0xEA}

// KdfAES4 is the AES key derivation function ID for Kdbx v4
var KdfAES4 = []byte{0x7C, 0x02, 0xBB, 0x82, 0x79, 0xA7, 0x4A, 0xC0, 0x92, 0x7D, 0x11, 0x4A, 0x00, 0x64, 0x82, 0x38}

// KdfArgon2 is the Argon2 key derivation function ID
var KdfArgon2 = []byte{0xEF, 0x63, 0x6D, 0xDF, 0x8C, 0x29, 0x44, 0x4B, 0x91, 0xF7, 0xA9, 0xA4, 0x03, 0xE3, 0x0A, 0x0C}

// DBHeader is the header of a database
type DBHeader struct {
	RawData     []byte
	Signature   *Signature
	FileHeaders *FileHeaders
}

// Signature holds the Keepass File Signature.
// The first 4 Bytes are the Base Signature,
// followed by 4 Bytes for the Version of the Format
// which is followed by 4 Bytes for the File Version
type Signature struct {
	BaseSignature      [4]byte
	SecondarySignature [4]byte
	MinorVersion       uint16
	MajorVersion       uint16
}

// FileHeaders contains every field of the header
type FileHeaders struct {
	Comment             []byte             // FieldID: 1
	CipherID            []byte             // FieldID: 2
	CompressionFlags    uint32             // FieldID: 3
	MasterSeed          []byte             // FieldID: 4
	TransformSeed       []byte             // FieldID: 5 (KDBX 3.1)
	TransformRounds     uint64             // FieldID: 6 (KDBX 3.1)
	EncryptionIV        []byte             // FieldID: 7
	ProtectedStreamKey  []byte             // FieldID: 8 (KDBX 3.1)
	StreamStartBytes    []byte             // FieldID: 9 (KDBX 3.1)
	InnerRandomStreamID uint32             // FieldID: 10 (KDBX 3.1)
	KdfParameters       *KdfParameters     // FieldID: 11 (KDBX 4)
	PublicCustomData    *VariantDictionary // FieldID: 12 (KDBX 4)
}

// KdfParameters contains every field of the KdfParameters header field
type KdfParameters struct {
	RawData     *VariantDictionary // Raw data of KdfParameters
	UUID        []byte             // $UUID - KDF ID
	Rounds      uint64             // R - Rounds
	Salt        [32]byte           // S - Salt (Argon 2) / Seed (AES)
	Parallelism uint32             // P - Parallelism
	Memory      uint64             // M - Memory
	Iterations  uint64             // I - Iterations
	Version     uint32             // V - Version
	SecretKey   []byte             // K - Secret key
	AssocData   []byte             // A - AssocData
}

// VariantDictionary is a structure used into KdfParameters and PublicCustomData
type VariantDictionary struct {
	Version uint16
	Items   []*VariantDictionaryItem
}

// VariantDictionaryItem is an item of a VariantDictionary
type VariantDictionaryItem struct {
	Type        byte
	NameLength  int32
	Name        []byte
	ValueLength int32
	Value       []byte
}

// NewHeader creates a new Header with good defaults
func NewHeader() *DBHeader {
	return NewKDBX3Header()
}

// NewKDBX3Header creates a new Header with good defaults for KDBX3
func NewKDBX3Header() *DBHeader {
	return &DBHeader{
		Signature:   &DefaultKDBX3Sig,
		FileHeaders: NewKDBX3FileHeaders(),
	}
}

// NewKDBX4Header creates a new Header with good defaults for KDBX4
func NewKDBX4Header() *DBHeader {
	return &DBHeader{
		Signature:   &DefaultKDBX4Sig,
		FileHeaders: NewKDBX4FileHeaders(),
	}
}

// NewFileHeaders creates a new FileHeaders with good defaults
func NewFileHeaders() *FileHeaders {
	return NewKDBX3FileHeaders()
}

// NewKDBX3FileHeaders creates a new FileHeaders with good defaults for KDBX3
func NewKDBX3FileHeaders() *FileHeaders {
	masterSeed := make([]byte, 32)
	rand.Read(masterSeed)

	transformSeed := make([]byte, 32)
	rand.Read(transformSeed)

	encryptionIV := make([]byte, 16)
	rand.Read(encryptionIV)

	protectedStreamKey := make([]byte, 32)
	rand.Read(protectedStreamKey)

	streamStartBytes := make([]byte, 32)
	rand.Read(streamStartBytes)

	return &FileHeaders{
		CipherID:            CipherAES,
		CompressionFlags:    GzipCompressionFlag,
		MasterSeed:          masterSeed,
		TransformSeed:       transformSeed,
		TransformRounds:     6000,
		EncryptionIV:        encryptionIV,
		ProtectedStreamKey:  protectedStreamKey,
		StreamStartBytes:    streamStartBytes,
		InnerRandomStreamID: SalsaStreamID,
	}
}

// NewKDBX4FileHeaders creates a new FileHeaders with good defaults for KDBX4
func NewKDBX4FileHeaders() *FileHeaders {
	masterSeed := make([]byte, 32)
	rand.Read(masterSeed)

	encryptionIV := make([]byte, 12)
	rand.Read(encryptionIV)

	var salt [32]byte
	rand.Read(salt[:])

	return &FileHeaders{
		CipherID:         CipherChaCha20,
		CompressionFlags: GzipCompressionFlag,
		MasterSeed:       masterSeed,
		EncryptionIV:     encryptionIV,
		KdfParameters: &KdfParameters{
			UUID:        KdfArgon2,
			Rounds:      0,
			Salt:        salt,
			Parallelism: 2,
			Memory:      1048576,
			Iterations:  2,
			Version:     19,
		},
	}
}

// readFrom reads the header from an io.Reader
func (h *DBHeader) readFrom(r io.Reader) error {
	// Save read data into a buffer that will be the RawData
	buffer := bytes.NewBuffer([]byte{})
	tR := io.TeeReader(r, buffer)

	// Read signature
	h.Signature = new(Signature)
	if err := binary.Read(tR, binary.LittleEndian, h.Signature); err != nil {
		return err
	}

	// Read file headers
	h.FileHeaders = new(FileHeaders)
	for {
		var err error
		if h.IsKdbx4() {
			err = h.FileHeaders.readHeader4(tR)
		} else {
			err = h.FileHeaders.readHeader31(tR)
		}

		// Update RawData buffer
		h.RawData = buffer.Bytes()

		if err != nil {
			if err == ErrEndOfHeaders {
				break
			}
			return err
		}
	}
	return nil
}

// readHeader4 reads a header of a KDBX v4 database
func (fh *FileHeaders) readHeader4(r io.Reader) error {
	var id uint8
	var length uint32
	var data []byte

	if err := binary.Read(r, binary.LittleEndian, &id); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return err
	}
	data = make([]byte, length)
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return err
	}

	return fh.readFileHeader(id, data)
}

// readHeader4 reads a header of a KDBX v3.1 database
func (fh *FileHeaders) readHeader31(r io.Reader) error {
	var id uint8
	var length uint16
	var data []byte

	if err := binary.Read(r, binary.LittleEndian, &id); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return err
	}
	data = make([]byte, length)
	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		return err
	}

	return fh.readFileHeader(id, data)
}

// readFileHeader reads a header value and puts it into the right variable
func (fh *FileHeaders) readFileHeader(id uint8, data []byte) error {
	switch id {
	case 0:
		return ErrEndOfHeaders
	case 1:
		fh.Comment = data
	case 2:
		fh.CipherID = data
	case 3:
		fh.CompressionFlags = binary.LittleEndian.Uint32(data)
	case 4:
		fh.MasterSeed = data
	case 5:
		fh.TransformSeed = data
	case 6:
		fh.TransformRounds = binary.LittleEndian.Uint64(data)
	case 7:
		fh.EncryptionIV = data
	case 8:
		fh.ProtectedStreamKey = data
	case 9:
		fh.StreamStartBytes = data
	case 10:
		fh.InnerRandomStreamID = binary.LittleEndian.Uint32(data)
	case 11:
		fh.KdfParameters = new(KdfParameters)
		return fh.KdfParameters.readKdfParameters(data)
	case 12:
		fh.PublicCustomData = new(VariantDictionary)
		return fh.PublicCustomData.readVariantDictionary(data)
	default:
		return ErrUnknownHeaderID(id)
	}
	return nil
}

// readKdfParameters reads a variant dictionary and puts values into KdfParameters
func (k *KdfParameters) readKdfParameters(data []byte) error {
	dict := new(VariantDictionary)
	if err := dict.readVariantDictionary(data); err != nil {
		return err
	}

	k.RawData = dict
	for _, item := range dict.Items {
		switch string(item.Name) {
		case "$UUID":
			k.UUID = item.Value
		case "R":
			k.Rounds = binary.LittleEndian.Uint64(item.Value)
		case "S":
			copy(k.Salt[:], item.Value[:32])
		case "P":
			k.Parallelism = binary.LittleEndian.Uint32(item.Value)
		case "M":
			k.Memory = binary.LittleEndian.Uint64(item.Value)
		case "I":
			k.Iterations = binary.LittleEndian.Uint64(item.Value)
		case "V":
			k.Version = binary.LittleEndian.Uint32(item.Value)
		case "K":
			k.SecretKey = item.Value
		case "A":
			k.AssocData = item.Value
		default:
			return ErrUnknownParameterID(string(item.Name))
		}
	}
	return nil
}

const (
	variantDictionaryTypeUInt32 = 0x4
	variantDictionaryTypeUInt64 = 0x5
	variantDictionaryTypeFlag   = 0x08
	variantDictionaryTypeInt32  = 0x0C
	variantDictionaryTypeInt64  = 0x0D
	variantDictionaryTypeString = 0x18
	variantDictionaryTypeBinary = 0x42
)

// updateRawData converts the kdf parameters into rawdata again
func (k *KdfParameters) updateRawData() error {
	dict := new(VariantDictionary)
	dict.Version = 256
	dict.Items = make([]*VariantDictionaryItem, 0, 9)

	if len(k.UUID) > 0 {
		uuidItem := &VariantDictionaryItem{
			Type:  variantDictionaryTypeBinary,
			Name:  []byte("$UUID"),
			Value: k.UUID,
		}
		dict.Items = append(dict.Items, uuidItem)
	}

	if k.Rounds > 0 {
		roundsItem := &VariantDictionaryItem{
			Type:  variantDictionaryTypeUInt64,
			Name:  []byte("R"),
			Value: make([]byte, 8),
		}
		binary.LittleEndian.PutUint64(roundsItem.Value, k.Rounds)
		dict.Items = append(dict.Items, roundsItem)
	}

	if k.Version > 0 {
		versionItem := &VariantDictionaryItem{
			Type:  variantDictionaryTypeUInt32,
			Name:  []byte("V"),
			Value: make([]byte, 4),
		}
		binary.LittleEndian.PutUint32(versionItem.Value, k.Version)
		dict.Items = append(dict.Items, versionItem)
	}

	if k.Iterations > 0 {
		iterationsItem := &VariantDictionaryItem{
			Type:  variantDictionaryTypeUInt64,
			Name:  []byte("I"),
			Value: make([]byte, 8),
		}
		binary.LittleEndian.PutUint64(iterationsItem.Value, k.Iterations)
		dict.Items = append(dict.Items, iterationsItem)
	}

	if k.Memory > 0 {
		memoryItem := &VariantDictionaryItem{
			Type:  variantDictionaryTypeUInt64,
			Name:  []byte("M"),
			Value: make([]byte, 8),
		}
		binary.LittleEndian.PutUint64(memoryItem.Value, k.Memory)
		dict.Items = append(dict.Items, memoryItem)
	}

	if k.Parallelism > 0 {
		parallelismItem := &VariantDictionaryItem{
			Type:  variantDictionaryTypeUInt32,
			Name:  []byte("P"),
			Value: make([]byte, 4),
		}
		binary.LittleEndian.PutUint32(parallelismItem.Value, k.Parallelism)
		dict.Items = append(dict.Items, parallelismItem)
	}

	if len(k.Salt) > 0 {
		saltItem := &VariantDictionaryItem{
			Type:  variantDictionaryTypeBinary,
			Name:  []byte("S"),
			Value: make([]byte, 32),
		}
		copy(saltItem.Value[:32], k.Salt[:])
		dict.Items = append(dict.Items, saltItem)
	}

	if len(k.SecretKey) > 0 {
		secretKeyItem := &VariantDictionaryItem{
			Type:  variantDictionaryTypeBinary,
			Name:  []byte("K"),
			Value: k.SecretKey,
		}
		dict.Items = append(dict.Items, secretKeyItem)
	}

	if len(k.AssocData) > 0 {
		assocDataItem := &VariantDictionaryItem{
			Type:  variantDictionaryTypeBinary,
			Name:  []byte("K"),
			Value: k.AssocData,
		}
		dict.Items = append(dict.Items, assocDataItem)
	}

	// Set NameLength, ValueLength and writes data to the result
	i := 0
	for _, item := range dict.Items {
		item.NameLength = int32(len(item.Name))
		item.ValueLength = int32(len(item.Value))

		if item.ValueLength > 0 {
			dict.Items[i] = item
			i++
		}
	}

	k.RawData = dict

	return nil
}

// readVariantDictionary reads a variant dictionary
func (vd *VariantDictionary) readVariantDictionary(data []byte) error {
	r := bytes.NewReader(data)

	if err := binary.Read(r, binary.LittleEndian, &vd.Version); err != nil {
		return err
	}

	for {
		vdi := new(VariantDictionaryItem)
		if err := binary.Read(r, binary.LittleEndian, &vdi.Type); err != nil {
			return err
		}

		if vdi.Type != 0x00 {
			if err := binary.Read(r, binary.LittleEndian, &vdi.NameLength); err != nil {
				return err
			}
			vdi.Name = make([]byte, vdi.NameLength)
			if err := binary.Read(r, binary.LittleEndian, &vdi.Name); err != nil {
				return err
			}

			if err := binary.Read(r, binary.LittleEndian, &vdi.ValueLength); err != nil {
				return err
			}
			vdi.Value = make([]byte, vdi.ValueLength)
			if err := binary.Read(r, binary.LittleEndian, &vdi.Value); err != nil {
				return err
			}

			vd.Items = append(vd.Items, vdi)
		} else {
			break
		}
	}
	return nil
}

// writeTo writes the header to the given io.Writer
func (h *DBHeader) writeTo(w io.Writer) error {
	var buffer bytes.Buffer
	mw := io.MultiWriter(w, &buffer)

	binary.Write(mw, binary.LittleEndian, h.Signature)

	if h.IsKdbx4() {
		h.FileHeaders.writeTo4(mw)
	} else {
		h.FileHeaders.writeTo31(mw)
	}

	h.RawData = buffer.Bytes()

	return nil
}

// writeTo4 writes a Kdbx v4 structured file header to the given io.Writer
func (fh FileHeaders) writeTo4(w io.Writer) error {
	compressionFlags := make([]byte, 4)
	binary.LittleEndian.PutUint32(compressionFlags, fh.CompressionFlags)

	if err := writeTo4Header(w, 1, fh.Comment); err != nil {
		return err
	}
	if err := writeTo4Header(w, 2, fh.CipherID); err != nil {
		return err
	}
	if err := writeTo4Header(w, 3, compressionFlags); err != nil {
		return err
	}
	if err := writeTo4Header(w, 4, fh.MasterSeed); err != nil {
		return err
	}
	if err := writeTo4Header(w, 7, fh.EncryptionIV); err != nil {
		return err
	}
	fh.KdfParameters.updateRawData()
	if err := writeTo4VariantDictionary(w, 11, fh.KdfParameters.RawData); err != nil {
		return err
	}
	if err := writeTo4VariantDictionary(w, 12, fh.PublicCustomData); err != nil {
		return err
	}
	// End of header
	return writeTo4Header(w, 0, []byte{0x0D, 0x0A, 0x0D, 0x0A})
}

// writeTo4Header is an helper to write a file header with the correct Kdbx v4 structure to the given io.Writer
func writeTo4Header(w io.Writer, id uint8, data []byte) error {
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

// writeTo4VariantDictionary is an helper to write a variant dictionary to the given io.Writer
func writeTo4VariantDictionary(w io.Writer, id uint8, data *VariantDictionary) error {
	if data != nil {
		var buffer bytes.Buffer
		if err := binary.Write(&buffer, binary.LittleEndian, data.Version); err != nil {
			return err
		}

		for _, item := range data.Items {
			if err := binary.Write(&buffer, binary.LittleEndian, item.Type); err != nil {
				return err
			}
			if err := binary.Write(&buffer, binary.LittleEndian, item.NameLength); err != nil {
				return err
			}
			if err := binary.Write(&buffer, binary.LittleEndian, item.Name); err != nil {
				return err
			}
			if err := binary.Write(&buffer, binary.LittleEndian, item.ValueLength); err != nil {
				return err
			}
			if err := binary.Write(&buffer, binary.LittleEndian, item.Value); err != nil {
				return err
			}
		}
		if err := binary.Write(&buffer, binary.LittleEndian, []byte{0x00}); err != nil {
			return err
		}

		// Write to original writer
		if err := binary.Write(w, binary.LittleEndian, id); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, uint32(buffer.Len())); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, buffer.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

// writeTo31 writes a Kdbx v3.1 structured file header to the given io.Writer
func (fh FileHeaders) writeTo31(w io.Writer) error {
	compressionFlags := make([]byte, 4)
	binary.LittleEndian.PutUint32(compressionFlags, fh.CompressionFlags)

	transformRounds := make([]byte, 8)
	binary.LittleEndian.PutUint64(transformRounds, fh.TransformRounds)

	innerRandomStreamID := make([]byte, 4)
	binary.LittleEndian.PutUint32(innerRandomStreamID, fh.InnerRandomStreamID)

	if err := writeTo31Header(w, 1, fh.Comment); err != nil {
		return err
	}
	if err := writeTo31Header(w, 2, fh.CipherID); err != nil {
		return err
	}
	if err := writeTo31Header(w, 3, compressionFlags); err != nil {
		return err
	}
	if err := writeTo31Header(w, 4, fh.MasterSeed); err != nil {
		return err
	}
	if err := writeTo31Header(w, 5, fh.TransformSeed); err != nil {
		return err
	}
	if err := writeTo31Header(w, 6, transformRounds); err != nil {
		return err
	}
	if err := writeTo31Header(w, 7, fh.EncryptionIV); err != nil {
		return err
	}
	if err := writeTo31Header(w, 8, fh.ProtectedStreamKey); err != nil {
		return err
	}
	if err := writeTo31Header(w, 9, fh.StreamStartBytes); err != nil {
		return err
	}
	if err := writeTo31Header(w, 10, innerRandomStreamID); err != nil {
		return err
	}
	// End of header
	return writeTo31Header(w, 0, []byte{0x0D, 0x0A, 0x0D, 0x0A})
}

// writeTo31Header is an helper to write a file header with the correct Kdbx v3.1 structure to the given io.Writer
func writeTo31Header(w io.Writer, id uint8, data []byte) error {
	if len(data) > 0 {
		if err := binary.Write(w, binary.LittleEndian, id); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, uint16(len(data))); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, data); err != nil {
			return err
		}
	}
	return nil
}

// Get a VariantDictionaryItem via its key
func (vd *VariantDictionary) Get(key string) *VariantDictionaryItem {
	for _, item := range vd.Items {
		if string(item.Name) == key {
			return item
		}
	}
	return nil
}

type formatVersion int

func isKdbx4(v formatVersion) bool {
	return v == 4
}

// IsKdbx4 returns true if the header version equals to 4
func (h *DBHeader) IsKdbx4() bool {
	return isKdbx4(formatVersion(h.Signature.MajorVersion))
}

func (h *DBHeader) formatVersion() formatVersion {
	return formatVersion(h.Signature.MajorVersion)
}

// GetSha256 returns the Sha256 hash of the header
func (h *DBHeader) GetSha256() [32]byte {
	return sha256.Sum256(h.RawData)
}

// ValidateSha256 validates the given hash with the Sha256 of the header
func (h *DBHeader) ValidateSha256(hash [32]byte) error {
	sha := h.GetSha256()
	if !reflect.DeepEqual(sha, hash) {
		return errors.New("Sha256 of header mismatching")
	}
	return nil
}

// GetHmacSha256 returns the HMAC-Sha256 hash of the header
func (h *DBHeader) GetHmacSha256(hmacKey []byte) (ret [32]byte) {
	hash := hmac.New(sha256.New, hmacKey)
	hash.Write(h.RawData)
	copy(ret[:32], hash.Sum(nil)[:32])
	return
}

// ValidateHmacSha256 validates the given hash with the HMAC-Sha256 of the header
func (h *DBHeader) ValidateHmacSha256(hmacKey []byte, hash [32]byte) error {
	hmacSha := h.GetHmacSha256(hmacKey)
	if !reflect.DeepEqual(hmacSha, hash) {
		return errors.New("HMAC-Sha256 of header mismatching")
	}
	return nil
}

func (h DBHeader) String() string {
	return fmt.Sprintf("Signature: %s\nFileHeaders: %s",
		h.Signature,
		h.FileHeaders,
	)
}

func (s Signature) String() string {
	return fmt.Sprintf("Base: %x, Secondary: %x, Format Version: %d.%d",
		s.BaseSignature,
		s.SecondarySignature,
		s.MajorVersion,
		s.MinorVersion,
	)
}

func (fh FileHeaders) String() string {
	return fmt.Sprintf(
		"(1) Comment: %x\n"+
			"(2) CipherID: %x\n"+
			"(3) CompressionFlags: %d\n"+
			"(4) MasterSeed: %x\n"+
			"(5) TransformSeed: %x\n"+
			"(6) TransformRounds: %d\n"+
			"(7) EncryptionIV: %x\n"+
			"(8) ProtectedStreamKey: %x\n"+
			"(9) StreamStartBytes: %x\n"+
			"(10) InnerRandomStreamID: %x\n"+
			"(11) KdfParameters: \n%s\n"+
			"(12) PublicCustomData: \n%s\n",
		fh.Comment,
		fh.CipherID,
		fh.CompressionFlags,
		fh.MasterSeed,
		fh.TransformSeed,
		fh.TransformRounds,
		fh.EncryptionIV,
		fh.ProtectedStreamKey,
		fh.StreamStartBytes,
		fh.InnerRandomStreamID,
		fh.KdfParameters,
		fh.PublicCustomData,
	)
}

func (k *KdfParameters) String() string {
	return fmt.Sprintf(
		"  (1) UUID: %x\n"+
			"  (2) Rounds: %d\n"+
			"  (3) Salt: %x\n"+
			"  (4) Parallelism: %d\n"+
			"  (5) Memory: %d\n"+
			"  (6) Iterations: %d\n"+
			"  (7) Version: %d\n"+
			"  (8) SecretKey: %x\n"+
			"  (9) AssocData: %x",
		k.UUID,
		k.Rounds,
		k.Salt,
		k.Parallelism,
		k.Memory,
		k.Iterations,
		k.Version,
		k.SecretKey,
		k.AssocData,
	)
}

func (vd VariantDictionary) String() string {
	var buffer bytes.Buffer
	for _, item := range vd.Items {
		buffer.WriteString(item.String())
	}
	return buffer.String()
}

func (vdi VariantDictionaryItem) String() string {
	return fmt.Sprintf(
		"Type: %x, NameLength: %d, Name: %s, ValueLength: %d, Value: %x\n",
		vdi.Type,
		vdi.NameLength,
		string(vdi.Name),
		vdi.ValueLength,
		vdi.Value,
	)
}

// ErrInvalidSignature is the error returned if the file signature is invalid
type ErrInvalidSignature struct {
	Name     string
	Is       interface{}
	Shouldbe interface{}
}

func (e ErrInvalidSignature) Error() string {
	return fmt.Sprintf(
		"gokeepasslib: invalid signature. %s is %x. Should be %x",
		e.Name,
		e.Is,
		e.Shouldbe,
	)
}

// ErrEndOfHeaders is the error returned when end of headers is read
var ErrEndOfHeaders = errors.New("gokeepasslib: header id was 0, end of headers")

// ErrUnknownHeaderID is the error returned if an unknown header is read
type ErrUnknownHeaderID int

func (i ErrUnknownHeaderID) Error() string {
	return fmt.Sprintf("gokeepasslib: unknown header ID of %d", i)
}

// ErrUnknownParameterID is the error returned if an unknown kdf parameter is read
type ErrUnknownParameterID string

func (s ErrUnknownParameterID) Error() string {
	return fmt.Sprintf("gokeepasslib: unknown kdf parameter '%s'", string(s))
}
