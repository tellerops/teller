package gokeepasslib

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
)

// Block size of 1MB - https://keepass.info/help/kb/kdbx_4.html#dataauth
const blockSplitRate = 1048576

type BlockHMACBuilder struct {
	baseKey []byte
}

func NewBlockHMACBuilder(masterSeed []byte, transformedKey []byte) *BlockHMACBuilder {
	keyBuilder := sha512.New()
	keyBuilder.Write(masterSeed)
	keyBuilder.Write(transformedKey)
	keyBuilder.Write([]byte{0x01})
	baseKey := keyBuilder.Sum(nil)

	return &BlockHMACBuilder{
		baseKey: baseKey,
	}
}

func (b *BlockHMACBuilder) BuildHMAC(index uint64, length uint32, data []byte) []byte {
	blockKeyBuilder := sha512.New()
	binary.Write(blockKeyBuilder, binary.LittleEndian, index)
	blockKeyBuilder.Write(b.baseKey)
	blockKey := blockKeyBuilder.Sum(nil)

	mac := hmac.New(sha256.New, blockKey)
	binary.Write(mac, binary.LittleEndian, index)
	binary.Write(mac, binary.LittleEndian, length)
	mac.Write(data)
	return mac.Sum(nil)
}

// decomposeContentBlocks4 decodes the content data block by block (Kdbx v4)
// Used to extract data blocks from the entire content
func decomposeContentBlocks4(r io.Reader, masterSeed []byte, transformedKey []byte) ([]byte, error) {
	var contentData []byte
	// Get all the content
	content, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	hmacBuilder := NewBlockHMACBuilder(masterSeed, transformedKey)

	index := uint64(0)
	offset := uint32(0)
	for {
		var blockHMAC [32]byte
		var length uint32

		copy(blockHMAC[:], content[offset:offset+32])
		offset = offset + 32

		buf := bytes.NewBuffer(content[offset : offset+4])
		binary.Read(buf, binary.LittleEndian, &length)

		offset = offset + 4

		data := make([]byte, length)
		endOfData := offset + length
		copy(data, content[offset:endOfData])
		offset = endOfData

		calculatedHMAC := hmacBuilder.BuildHMAC(index, length, data)

		if subtle.ConstantTimeCompare(calculatedHMAC, blockHMAC[:]) == 0 {
			return nil, fmt.Errorf("Failed to verify HMAC for block %d", index)
		}

		// Add to blocks
		contentData = append(contentData, data...)

		if length == 0 {
			break
		}

		index += 1
	}
	return contentData, nil
}

// decomposeContentBlocks31 decodes the content data block by block (Kdbx v3.1)
// Used to extract data blocks from the entire content
func decomposeContentBlocks31(r io.Reader) ([]byte, error) {
	var contentData []byte
	// Get all the content
	content, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	offset := uint32(0)
	for {
		var hash [32]byte
		var length uint32
		var data []byte

		// Skipping Index, uint32
		offset = offset + 4

		copy(hash[:], content[offset:offset+32])
		offset = offset + 32

		length = binary.LittleEndian.Uint32(content[offset : offset+4])
		offset = offset + 4

		if length > 0 {
			data = make([]byte, length)
			copy(data, content[offset:offset+length])
			offset = offset + length

			// Add to decoded blocks
			contentData = append(contentData, data...)
		} else {
			break
		}
	}
	return contentData, nil
}

// composeContentBlocks4 composes every content block into a HMAC-LENGTH-DATA block scheme (Kdbx v4)
func composeContentBlocks4(w io.Writer, contentData []byte, masterSeed []byte, transformedKey []byte) {
	hmacBuilder := NewBlockHMACBuilder(masterSeed, transformedKey)

	offset := 0
	endOffset := 0
	var index = uint64(0)
	for {
		remainingLength := len(contentData[offset:])

		if remainingLength >= blockSplitRate {
			endOffset = offset + blockSplitRate
		} else {
			endOffset = offset + remainingLength
		}

		length := endOffset - offset
		data := make([]byte, length)
		copy(data, contentData[offset:endOffset])
		uLength := uint32(length)

		blockHMAC := hmacBuilder.BuildHMAC(index, uLength, data)

		w.Write(blockHMAC)
		binary.Write(w, binary.LittleEndian, uLength)
		w.Write(data)

		offset = endOffset

		if length == 0 {
			break
		}

		index += 1
	}
	binary.Write(w, binary.LittleEndian, [32]byte{})
	binary.Write(w, binary.LittleEndian, uint32(0))
}

// composeBlocks31 composes every content block into a INDEX-SHA-LENGTH-DATA block scheme (Kdbx v3.1)
func composeContentBlocks31(w io.Writer, contentData []byte) {
	index := uint32(0)
	offset := 0
	for offset < len(contentData) {
		var hash [32]byte
		var length uint32
		var data []byte

		if len(contentData[offset:]) >= blockSplitRate {
			data = append(data, contentData[offset:]...)
		} else {
			data = append(data, contentData...)
		}

		length = uint32(len(data))
		hash = sha256.Sum256(data)

		binary.Write(w, binary.LittleEndian, index)
		binary.Write(w, binary.LittleEndian, hash)
		binary.Write(w, binary.LittleEndian, length)
		binary.Write(w, binary.LittleEndian, data)
		index++
		offset = offset + blockSplitRate
	}
	binary.Write(w, binary.LittleEndian, index)
	binary.Write(w, binary.LittleEndian, [32]byte{})
	binary.Write(w, binary.LittleEndian, uint32(0))
}
