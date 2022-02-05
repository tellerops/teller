package segments

import (
	"fmt"
	"strconv"
)

type Index struct {
	idx int
}

func ParseIndexSegment(s string) (YamlPathSegment, error) {
	if s[0] == '[' {
		s = s[1 : len(s)-1]
	}

	idx, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}

	return &Index{
		idx: idx,
	}, nil
}

func (p *Index) NavigateMap(m map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("cannot index map")
}

func (p *Index) NavigateArray(l []interface{}) (interface{}, error) {
	if len(l) <= p.idx {
		return nil, fmt.Errorf("out of bounds %d (len %d)", p.idx, len(l))
	}

	return l[p.idx], nil
}
