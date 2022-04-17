package gokeepasslib

import (
	"encoding/xml"
	"io"

	w "github.com/tobischo/gokeepasslib/v3/wrappers"
)

const (
	groupChildOrderDefault = iota
	groupChildOrderEntryFirst
	groupChildOrderGroupFirst
)

type GroupOption func(*Group)

func WithGroupFormattedTime(formatted bool) GroupOption {
	return func(g *Group) {
		WithTimeDataFormattedTime(formatted)(&g.Times)

		for _, group := range g.Groups {
			WithGroupFormattedTime(formatted)(&group)
		}

		for _, entry := range g.Entries {
			WithEntryFormattedTime(formatted)(&entry)
		}
	}
}

// Group is a structure to store entries in their named groups for organization
type Group struct {
	UUID                    UUID                  `xml:"UUID"`
	Name                    string                `xml:"Name"`
	Notes                   string                `xml:"Notes"`
	IconID                  int64                 `xml:"IconID"`
	CustomIconUUID          UUID                  `xml:"CustomIconUUID"`
	Times                   TimeData              `xml:"Times"`
	IsExpanded              w.BoolWrapper         `xml:"IsExpanded"`
	DefaultAutoTypeSequence string                `xml:"DefaultAutoTypeSequence"`
	EnableAutoType          w.NullableBoolWrapper `xml:"EnableAutoType"`
	EnableSearching         w.NullableBoolWrapper `xml:"EnableSearching"`
	LastTopVisibleEntry     string                `xml:"LastTopVisibleEntry"`
	Entries                 []Entry               `xml:"Entry,omitempty"`
	Groups                  []Group               `xml:"Group,omitempty"`
	groupChildOrder         int                   `xml:"-"`
}

// UnmarshalXML unmarshals the boolean from d
func (g *Group) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for {
		token, err := d.Token()
		if err == io.EOF {
			break
		}
		switch element := token.(type) {
		case xml.StartElement:
			unmarshalGroupToken(g, d, element)
		}
	}

	return nil
}

func unmarshalGroupToken(g *Group, d *xml.Decoder, element xml.StartElement) error {
	switch element.Name.Local {
	case "Entry":
		if g.groupChildOrder == groupChildOrderDefault {
			g.groupChildOrder = groupChildOrderEntryFirst
		}

		var entry Entry
		err := d.DecodeElement(&entry, &element)
		if err != nil {
			return err
		}

		g.Entries = append(g.Entries, entry)
	case "Group":
		if g.groupChildOrder == groupChildOrderDefault {
			g.groupChildOrder = groupChildOrderGroupFirst
		}

		var group Group
		err := d.DecodeElement(&group, &element)
		if err != nil {
			return err
		}

		g.Groups = append(g.Groups, group)
	case "UUID":
		return d.DecodeElement(&g.UUID, &element)
	case "Name":
		return d.DecodeElement(&g.Name, &element)
	case "Notes":
		return d.DecodeElement(&g.Notes, &element)
	case "IconID":
		return d.DecodeElement(&g.IconID, &element)
	case "CustomIconUUID":
		return d.DecodeElement(&g.CustomIconUUID, &element)
	case "Times":
		return d.DecodeElement(&g.Times, &element)
	case "IsExpanded":
		return d.DecodeElement(&g.IsExpanded, &element)
	case "DefaultAutoTypeSequence":
		return d.DecodeElement(&g.DefaultAutoTypeSequence, &element)
	case "EnableAutoType":
		return d.DecodeElement(&g.EnableAutoType, &element)
	case "EnableSearching":
		return d.DecodeElement(&g.EnableSearching, &element)
	case "LastTopVisibleEntry":
		return d.DecodeElement(&g.LastTopVisibleEntry, &element)
	}

	return nil
}

// NewGroup returns a new group with time data and uuid set
func NewGroup(options ...GroupOption) Group {
	group := Group{
		EnableAutoType:  w.NewNullableBoolWrapper(true),
		EnableSearching: w.NewNullableBoolWrapper(true),
		Times:           NewTimeData(),
		UUID:            NewUUID(),
	}

	for _, option := range options {
		option(&group)
	}

	return group
}

func (g *Group) setKdbxFormatVersion(version formatVersion) {
	(&g.Times).setKdbxFormatVersion(version)

	for i := range g.Groups {
		(&g.Groups[i]).setKdbxFormatVersion(version)
	}

	for i := range g.Entries {
		(&g.Entries[i]).setKdbxFormatVersion(version)
	}
}
