package gokeepasslib

import (
	"encoding/xml"

	w "github.com/tobischo/gokeepasslib/v3/wrappers"
)

type EntryOption func(*Entry)

func WithEntryFormattedTime(formatted bool) EntryOption {
	return func(e *Entry) {
		WithTimeDataFormattedTime(formatted)(&e.Times)
	}
}

// Entry is the structure which holds information about a parsed entry in a keepass database
type Entry struct {
	UUID            UUID              `xml:"UUID"`
	IconID          int64             `xml:"IconID"`
	CustomIconUUID  UUID              `xml:"CustomIconUUID"`
	ForegroundColor string            `xml:"ForegroundColor"`
	BackgroundColor string            `xml:"BackgroundColor"`
	OverrideURL     string            `xml:"OverrideURL"`
	Tags            string            `xml:"Tags"`
	Times           TimeData          `xml:"Times"`
	Values          []ValueData       `xml:"String,omitempty"`
	AutoType        AutoTypeData      `xml:"AutoType"`
	Histories       []History         `xml:"History"`
	Binaries        []BinaryReference `xml:"Binary,omitempty"`
	CustomData      []CustomData      `xml:"CustomData>Item"`
}

// NewEntry return a new entry with time data and uuid set
func NewEntry(options ...EntryOption) Entry {
	entry := Entry{}
	entry.Times = NewTimeData()
	entry.UUID = NewUUID()

	for _, option := range options {
		option(&entry)
	}

	return entry
}

func (e *Entry) setKdbxFormatVersion(version formatVersion) {
	(&e.Times).setKdbxFormatVersion(version)

	for i := range e.Histories {
		(&e.Histories[i]).setKdbxFormatVersion(version)
	}
}

// Get returns the value in e corresponding with key k, or an empty string otherwise
func (e *Entry) Get(key string) *ValueData {
	for i := range e.Values {
		if e.Values[i].Key == key {
			return &e.Values[i]
		}
	}
	return nil
}

// GetContent returns the content of the value belonging to the given key in string form
func (e *Entry) GetContent(key string) string {
	val := e.Get(key)
	if val == nil {
		return ""
	}
	return val.Value.Content
}

// GetIndex returns the index of the Value belonging to the given key, or -1 if none is found
func (e *Entry) GetIndex(key string) int {
	for i := range e.Values {
		if e.Values[i].Key == key {
			return i
		}
	}
	return -1
}

// GetPassword returns the password of an entry
func (e *Entry) GetPassword() string {
	return e.GetContent("Password")
}

// GetPasswordIndex returns the index in the values slice belonging to the password
func (e *Entry) GetPasswordIndex() int {
	return e.GetIndex("Password")
}

// GetTitle returns the title of an entry
func (e *Entry) GetTitle() string {
	return e.GetContent("Title")
}

// History stores information about changes made to an entry,
// in the form of a list of previous versions of that entry
type History struct {
	Entries []Entry `xml:"Entry"`
}

func (h *History) setKdbxFormatVersion(version formatVersion) {
	for i := range h.Entries {
		(&h.Entries[i]).setKdbxFormatVersion(version)
	}
}

// ValueData is a structure containing key value pairs of information stored in an entry
type ValueData struct {
	Key   string `xml:"Key"`
	Value V      `xml:"Value"`
}

// V is a wrapper for the content of a value, so that it can store whether it is protected
type V struct {
	Content   string        `xml:",chardata"`
	Protected w.BoolWrapper `xml:"Protected,attr,omitempty"`
}

// AutoTypeData is a structure containing auto type settings of an entry
type AutoTypeData struct {
	Enabled                 w.BoolWrapper         `xml:"Enabled"`
	DataTransferObfuscation int64                 `xml:"DataTransferObfuscation"`
	DefaultSequence         string                `xml:"DefaultSequence"`
	Associations            []AutoTypeAssociation `xml:"Association,omitempty"`
}

// AutoTypeAssociation is a structure that store the keystroke sequence of a window for AutoTypeData
type AutoTypeAssociation struct {
	Window            string `xml:"Window"`
	KeystrokeSequence string `xml:"KeystrokeSequence"`
}

// CustomData is the structure for plugins custom data
type CustomData struct {
	XMLName xml.Name `xml:"Item"`
	Key     string   `xml:"Key"`
	Value   string   `xml:"Value"`
}
