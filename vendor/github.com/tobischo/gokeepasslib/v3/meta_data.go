package gokeepasslib

import (
	w "github.com/tobischo/gokeepasslib/v3/wrappers"
)

// MemProtection is a structure containing settings for MemoryProtection
type MemProtection struct {
	ProtectTitle    w.BoolWrapper `xml:"ProtectTitle"`
	ProtectUserName w.BoolWrapper `xml:"ProtectUserName"`
	ProtectPassword w.BoolWrapper `xml:"ProtectPassword"`
	ProtectURL      w.BoolWrapper `xml:"ProtectURL"`
	ProtectNotes    w.BoolWrapper `xml:"ProtectNotes"`
}

type MetaDataOption func(*MetaData)

// CustomIcon is the structure needed to store custom icons.  Unsure of what version/format requires this
type CustomIcon struct {
	UUID UUID   `xml:"UUID"` //Entry's CustomIcon UUID should match this
	Data string `xml:"Data"` //base64 encoded PNG icon.  Unknown size constraints
}

func WithMetaDataFormattedTime(formatted bool) MetaDataOption {
	return func(md *MetaData) {
		md.MasterKeyChanged.Formatted = formatted
	}
}

// NewMetaData creates a MetaData struct with some defaults set
func NewMetaData(options ...MetaDataOption) *MetaData {
	now := w.Now()

	md := &MetaData{
		SettingsChanged:        &now,
		MasterKeyChanged:       &now,
		MasterKeyChangeRec:     -1,
		MasterKeyChangeForce:   -1,
		HistoryMaxItems:        10,
		HistoryMaxSize:         6291456, // 6 MB
		MaintenanceHistoryDays: 365,
	}

	for _, option := range options {
		option(md)
	}

	return md
}

// MetaData is the structure for the metadata headers at the top of kdbx files,
// it contains things like the name of the database
type MetaData struct {
	Generator                  string         `xml:"Generator"`
	SettingsChanged            *w.TimeWrapper `xml:"SettingsChanged"`
	HeaderHash                 string         `xml:"HeaderHash,omitempty"`
	DatabaseName               string         `xml:"DatabaseName"`
	DatabaseNameChanged        *w.TimeWrapper `xml:"DatabaseNameChanged"`
	DatabaseDescription        string         `xml:"DatabaseDescription"`
	DatabaseDescriptionChanged *w.TimeWrapper `xml:"DatabaseDescriptionChanged"`
	DefaultUserName            string         `xml:"DefaultUserName"`
	DefaultUserNameChanged     *w.TimeWrapper `xml:"DefaultUserNameChanged"`
	MaintenanceHistoryDays     int64          `xml:"MaintenanceHistoryDays"`
	Color                      string         `xml:"Color"`
	MasterKeyChanged           *w.TimeWrapper `xml:"MasterKeyChanged"`
	MasterKeyChangeRec         int64          `xml:"MasterKeyChangeRec"`
	MasterKeyChangeForce       int64          `xml:"MasterKeyChangeForce"`
	MemoryProtection           MemProtection  `xml:"MemoryProtection"`
	CustomIcons                []CustomIcon   `xml:"CustomIcons>Icon"`
	RecycleBinEnabled          w.BoolWrapper  `xml:"RecycleBinEnabled"`
	RecycleBinUUID             UUID           `xml:"RecycleBinUUID"`
	RecycleBinChanged          *w.TimeWrapper `xml:"RecycleBinChanged"`
	EntryTemplatesGroup        string         `xml:"EntryTemplatesGroup"`
	EntryTemplatesGroupChanged *w.TimeWrapper `xml:"EntryTemplatesGroupChanged"`
	HistoryMaxItems            int64          `xml:"HistoryMaxItems"`
	HistoryMaxSize             int64          `xml:"HistoryMaxSize"`
	LastSelectedGroup          string         `xml:"LastSelectedGroup"`
	LastTopVisibleGroup        string         `xml:"LastTopVisibleGroup"`
	Binaries                   Binaries       `xml:"Binaries>Binary,omitempty"`
	CustomData                 []CustomData   `xml:"CustomData>Item"`
}

func (md *MetaData) setKdbxFormatVersion(version formatVersion) {
	if md.SettingsChanged != nil {
		md.SettingsChanged.Formatted = !isKdbx4(version)
	}
	if md.DatabaseNameChanged != nil {
		md.DatabaseNameChanged.Formatted = !isKdbx4(version)
	}
	if md.DatabaseDescriptionChanged != nil {
		md.DatabaseDescriptionChanged.Formatted = !isKdbx4(version)
	}
	if md.DefaultUserNameChanged != nil {
		md.DefaultUserNameChanged.Formatted = !isKdbx4(version)
	}
	if md.MasterKeyChanged != nil {
		md.MasterKeyChanged.Formatted = !isKdbx4(version)
	}
	if md.RecycleBinChanged != nil {
		md.RecycleBinChanged.Formatted = !isKdbx4(version)
	}
	if md.EntryTemplatesGroupChanged != nil {
		md.EntryTemplatesGroupChanged.Formatted = !isKdbx4(version)
	}
}
