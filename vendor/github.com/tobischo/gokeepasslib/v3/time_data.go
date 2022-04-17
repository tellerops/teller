package gokeepasslib

import (
	w "github.com/tobischo/gokeepasslib/v3/wrappers"
)

type TimeDataOption func(*TimeData)

func WithTimeDataFormattedTime(formatted bool) TimeDataOption {
	return func(td *TimeData) {
		td.CreationTime.Formatted = formatted
		td.LastModificationTime.Formatted = formatted
		td.LastAccessTime.Formatted = formatted
		td.LocationChanged.Formatted = formatted
		td.Expires = w.NewBoolWrapper(false)
	}
}

// TimeData contains all metadata related to times for groups and entries
// e.g. the last modification time or the creation time
type TimeData struct {
	CreationTime         *w.TimeWrapper `xml:"CreationTime"`
	LastModificationTime *w.TimeWrapper `xml:"LastModificationTime"`
	LastAccessTime       *w.TimeWrapper `xml:"LastAccessTime"`
	ExpiryTime           *w.TimeWrapper `xml:"ExpiryTime"`
	Expires              w.BoolWrapper  `xml:"Expires"`
	UsageCount           int64          `xml:"UsageCount"`
	LocationChanged      *w.TimeWrapper `xml:"LocationChanged"`
}

func (td *TimeData) setKdbxFormatVersion(version formatVersion) {
	if td.CreationTime != nil {
		td.CreationTime.Formatted = !isKdbx4(version)
	}
	if td.LastModificationTime != nil {
		td.LastModificationTime.Formatted = !isKdbx4(version)
	}
	if td.LastAccessTime != nil {
		td.LastAccessTime.Formatted = !isKdbx4(version)
	}
	if td.ExpiryTime != nil {
		td.ExpiryTime.Formatted = !isKdbx4(version)
	}
	if td.LocationChanged != nil {
		td.LocationChanged.Formatted = !isKdbx4(version)
	}
}

// NewTimeData returns a TimeData struct with good defaults (no expire time, all times set to now)
func NewTimeData(options ...TimeDataOption) TimeData {
	now := w.Now()
	td := TimeData{
		CreationTime:         &now,
		LastModificationTime: &now,
		LastAccessTime:       &now,
		LocationChanged:      &now,
		Expires:              w.NewBoolWrapper(false),
		UsageCount:           0,
	}

	for _, option := range options {
		option(&td)
	}

	return td
}
