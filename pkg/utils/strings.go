package utils

import "strings"

func LastSegment(s string) string {
	segs := strings.Split(s, "/")
	return segs[len(segs)-1]
}

func Merge(from, into map[string]string) {
	if into == nil {
		into = make(map[string]string)
	}

	for k, v := range from {
		into[k] = v
	}
}
