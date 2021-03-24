/*
ellipses.go
-John Taylor
Dec-29-2020

Insert an ellipses into the middle of a long string to shorten it

*/

package ellipsis

// Shorten will insert "..." into the middle of a long string
// to shorten it to a length of w
func Shorten(s string, w int) string {
	if len(s) <= w {
		return s
	}
	if len(s) <= 5 {
		return s
	}
	if w <= 3 {
		return s[len(s)-w:]
	}

	maxsz := (w - 3) / 2
	extra := 0
	if w%2 == 0 {
		extra = 1
	}

	a := s[0:maxsz]
	z := s[len(s)-maxsz-extra:]

	return a + "..." + z
}
