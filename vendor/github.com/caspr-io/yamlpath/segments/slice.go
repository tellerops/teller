package segments

import (
	"fmt"
	"strconv"
	"strings"
)

type Slice struct {
	startIdx int
	endIdx   int
}

func ParseSliceSegment(s string) (YamlPathSegment, error) {
	idxs := strings.Split(s[1:len(s)-1], ":")

	start, err := strconv.Atoi(idxs[0])
	if err != nil {
		return nil, fmt.Errorf("part '%s' is not an index into an array. %w", s, err)
	}

	end, err := strconv.Atoi(idxs[1])
	if err != nil {
		return nil, fmt.Errorf("part '%s' is not an index into an array. %w", s, err)
	}

	if start > end {
		return nil, fmt.Errorf("cannot take slice with reversed indexes '%s'", s)
	}

	return &Slice{
		startIdx: start,
		endIdx:   end,
	}, nil
}

func (s *Slice) NavigateMap(m map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("cannot slice map")
}

func (s *Slice) NavigateArray(l []interface{}) (interface{}, error) {
	if s.startIdx >= len(l) {
		return nil, fmt.Errorf("start slice index out of bounds '%d' for array length '%d'", s.startIdx, len(l))
	}

	if s.endIdx > len(l) {
		return nil, fmt.Errorf("end slice index out of bounds '%d' for array length '%d'", s.endIdx, len(l))
	}

	slice := []interface{}{}
	for i := s.startIdx; i < s.endIdx; i++ {
		slice = append(slice, l[i])
	}

	return slice, nil
}
