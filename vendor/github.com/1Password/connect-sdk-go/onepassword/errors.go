package onepassword

import "fmt"

// Error is an error returned by the Connect API.
type Error struct {
	StatusCode int    `json:"status"`
	Message    string `json:"message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("status %d: %s", e.StatusCode, e.Message)
}

func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return t.Message == e.Message && t.StatusCode == e.StatusCode
}
