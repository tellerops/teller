package segments

import (
	"fmt"
)

type Key struct {
	key string
}

func ParseKeySegment(s string) (YamlPathSegment, error) {
	return &Key{
		key: s,
	}, nil
}

func (s *Key) NavigateMap(m map[string]interface{}) (interface{}, error) {
	if v, ok := m[s.key]; ok {
		return v, nil
	}

	return nil, fmt.Errorf("could not find key '%s' in yaml", s.key)
}

func (s *Key) NavigateArray(l []interface{}) (interface{}, error) {
	result := []interface{}{}

	for _, v := range l {
		r, err := NavigateYaml(v, s)
		if err != nil {
			return nil, err
		}

		result = append(result, r)
	}

	return result, nil
}
