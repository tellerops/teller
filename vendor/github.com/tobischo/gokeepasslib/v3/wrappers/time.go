package wrappers

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

// The lower value for KDBX4 times is 0001-01-01.
// Since the time values are stored as seconds since that time,
// we need an offset to calculate that value.
// zeroUnixOffset represents time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
// which can be used as offset here when using `.Unix()` as the conversion into an integer
// from `time.Time`.
// Since nix time counts since 1970-01-01, any value before that would be negative,
// which also makes this offset negative. Substracting this value from any other
// time value will increase it by as much. So even a value before 1970 would be correct
// converted back and forth.
// This value is set directly as int64 const to avoid having to initialize time.Time values
// all the time
const zeroUnixOffset int64 = -62135596800

// TimeWrapper is a time.Time wrapper that provides xml marshaling and unmarshaling
type TimeWrapper struct {
	Formatted bool      // True for Kdbx v3.1 (formatted as RFC3339)
	Time      time.Time // Time value
}

type TimeOption func(*TimeWrapper)

func WithKDBX4Formatting(t *TimeWrapper) {
	WithFormatted(false)(t)
}

func WithFormatted(formatted bool) TimeOption {
	return func(t *TimeWrapper) {
		t.Formatted = formatted
	}
}

// Now returns a TimeWrapper instance with the current time in UTC
func Now(options ...TimeOption) TimeWrapper {
	t := TimeWrapper{
		Formatted: true,
		Time:      time.Now().In(time.UTC),
	}

	for _, option := range options {
		option(&t)
	}

	return t
}

// MarshalText marshals time into an RFC3339 compliant value in UTC (Kdbx v3.1)
// On Kdbx v4 it calculates the timestamp subtracting seconds from the time date and encode it with base64
func (tw TimeWrapper) MarshalText() ([]byte, error) {
	t := time.Time(tw.Time).In(time.UTC)
	if y := t.Year(); y < 0 || y >= 10000 {
		return nil, ErrYearOutsideOfRange
	}

	var ret []byte
	if tw.Formatted {
		// Kdbx v3.1
		b := make([]byte, 0, len(time.RFC3339))
		ret = t.AppendFormat(b, time.RFC3339)
	} else {
		// Kdbx v4 - Count since year 1
		total := t.Unix() - zeroUnixOffset

		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, uint64(total))
		ret = make([]byte, base64.StdEncoding.EncodedLen(len(buf)))
		base64.StdEncoding.Encode(ret, buf)
	}
	return ret, nil
}

// UnmarshalText take a string of format time.RFC3339 and marshals it into the TimeWrapper value (Kdbx v3.1)
// On Kdbx v4 it calculates the time with given seconds via data byte array (base64 encoded)
func (tw *TimeWrapper) UnmarshalText(data []byte) error {
	var formatted bool
	// Check for RFC string (KDBX 3.1), if it fail try with KDBX 4
	t, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		// KDBX v4
		// In version 4 the time is a base64 timestamp of seconds passed since 1/1/0001
		var buf int64

		decoded := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
		_, err = base64.StdEncoding.Decode(decoded, data)
		if err != nil {
			return err
		}
		if err := binary.Read(bytes.NewReader(decoded), binary.LittleEndian, &buf); err != nil {
			return err
		}

		// Count since year 1
		t = time.Unix(zeroUnixOffset+buf, 0)
		formatted = false
	} else {
		formatted = true
	}
	*tw = TimeWrapper{
		Formatted: formatted,
		Time:      t,
	}
	return nil
}

func (tw TimeWrapper) String() string {
	return fmt.Sprintf(
		"Formatted: %v, Time: %v",
		tw.Formatted,
		tw.Time,
	)
}

// ErrYearOutsideOfRange is the error returned when the year is outside 0 and 9999
var ErrYearOutsideOfRange = errors.New("Wrappers.Time.MarshalText: year outside of range [0,9999]")
