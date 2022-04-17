package gokeepasslib

import (
	"encoding/xml"

	w "github.com/tobischo/gokeepasslib/v3/wrappers"
)

// DeletedObjectData is the structure for a deleted object
type DeletedObjectData struct {
	XMLName      xml.Name       `xml:"DeletedObject"`
	UUID         UUID           `xml:"UUID"`
	DeletionTime *w.TimeWrapper `xml:"DeletionTime"`
}

func (d *DeletedObjectData) setKdbxFormatVersion(version formatVersion) {
	d.DeletionTime.Formatted = !isKdbx4(version)
}
