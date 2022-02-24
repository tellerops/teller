package utils

import "strings"

func LastSegment(s string) string {
	segs := strings.Split(s, "/")
	return segs[len(segs)-1]
}

func Merge(first, second map[string]string) map[string]string {
	target := make(map[string]string)
	for k, v := range first {
		target[k] = v
	}
	for k, v := range second {
		target[k] = v
	}

	return target
}
