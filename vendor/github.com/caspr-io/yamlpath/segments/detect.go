package segments

import (
	"fmt"
	"regexp"
)

const (
	KeySegment = "^[:a-zA-Z0-9_\\.-]+$"
	// KeySearchSegment     = "^\\[\\.=[a-zA-Z][a-zA-Z0-9_-]*\\]$"
	ExplicitIndexSegment = "^\\[[0-9]+\\]$"
	ImplicitIndexSegment = "^[0-9]+$"
	SliceSegment         = "^\\[[0-9]+:[0-9]+\\]$"
	ValueSearchSegment   = "^\\[\\.[=\\^\\$\\%].+\\]$"
)

var regexps map[string]*regexp.Regexp = map[string]*regexp.Regexp{ //nolint:gochecknoglobals
	KeySegment: regexp.MustCompile(KeySegment),
	// KeySearchSegment:     regexp.MustCompile(KeySearchSegment),
	ExplicitIndexSegment: regexp.MustCompile(ExplicitIndexSegment),
	ImplicitIndexSegment: regexp.MustCompile(ImplicitIndexSegment),
	SliceSegment:         regexp.MustCompile(SliceSegment),
	ValueSearchSegment:   regexp.MustCompile(ValueSearchSegment),
}

func DetectSegment(s string) (YamlPathSegment, error) {
	switch {
	case regexps[ImplicitIndexSegment].MatchString(s), regexps[ExplicitIndexSegment].MatchString(s):
		return ParseIndexSegment(s)
	case regexps[SliceSegment].MatchString(s):
		return ParseSliceSegment(s)
	case regexps[ValueSearchSegment].MatchString(s):
		return ParseStringValueSearchSegment(s)
	case regexps[KeySegment].MatchString(s):
		return ParseKeySegment(s)
	}

	return nil, fmt.Errorf("segment '%s' not supported yet", s)
}
