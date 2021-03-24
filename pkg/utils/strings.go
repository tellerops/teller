package utils

import "strings"

func LastSegment(s string) string {
	segs := strings.Split(s, "/")
	return segs[len(segs)-1]
}
