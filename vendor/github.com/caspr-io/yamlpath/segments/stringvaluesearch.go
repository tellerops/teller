package segments

import (
	"fmt"
	"strings"
)

// [.=foo], [.^foo], [.$foo], [.%foo]
type StringValueSearch struct {
	operator byte
	pattern  string
}

func ParseStringValueSearchSegment(s string) (YamlPathSegment, error) {
	operator := s[2]
	pattern := s[3 : len(s)-1]

	return &StringValueSearch{
		operator: operator,
		pattern:  pattern,
	}, nil
}

func (s *StringValueSearch) NavigateMap(m map[string]interface{}) (interface{}, error) {
	for k, v := range m {
		if s.valueMatches(k) {
			return v, nil
		}
	}

	return nil, fmt.Errorf("could not find matching key in hash for pattern '[.%s%s]", string(s.operator), s.pattern)
}

func (s *StringValueSearch) NavigateArray(l []interface{}) (interface{}, error) {
	for _, i := range l {
		switch v := i.(type) {
		case string:
			if s.valueMatches(v) {
				return v, nil
			}

			continue
		default:
			return nil, fmt.Errorf("could not search for value '%s' as list does not contain strings", s.pattern)
		}
	}

	return nil, fmt.Errorf("could not find match for search part '[.%s%s]'", string(s.operator), s.pattern)
}

func (s *StringValueSearch) valueMatches(v string) bool {
	switch s.operator {
	case '^':
		return strings.HasPrefix(v, s.pattern)
	case '$':
		return strings.HasSuffix(v, s.pattern)
	case '%':
		return strings.Contains(v, s.pattern)
	case '=':
		return v == s.pattern
	default:
		return false
	}
}
