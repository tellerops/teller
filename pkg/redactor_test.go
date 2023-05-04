package pkg

import (
	"bytes"
	"io"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/spectralops/teller/pkg/core"
)

func TestRedactorOverlap(t *testing.T) {
	cases := []struct {
		name    string
		entries []core.EnvEntry
		s       string
		sr      string
	}{
		{
			name: "overlap",
			// in this case we don't want '123' to appear in the clear after all redactions are made.
			// it can happen if the smaller secret get replaced first because both
			// secrets overlap. we need to ensure the wider secrets always get
			// replaced first.
			entries: []core.EnvEntry{
				{
					ProviderName: "test",
					ResolvedPath: "/some/path",
					Key:          "OTHER_KEY",
					Value:        "hello",
					RedactWith:   "**OTHER_KEY**",
				},
				{
					ProviderName: "test",
					ResolvedPath: "/some/path",
					Key:          "SOME_KEY",
					Value:        "hello123",
					RedactWith:   "**SOME_KEY**",
				},
			},
			s: `
func Foobar(){
	secret := "hello"
	callService(secret, "hello123")
	// hello, hello123
}
`,
			sr: `
func Foobar(){
	secret := "**OTHER_KEY**"
	callService(secret, "**SOME_KEY**")
	// **OTHER_KEY**, **SOME_KEY**
}
`,
		},
		{
			name: "multiple",
			entries: []core.EnvEntry{
				{
					ProviderName: "test",
					ResolvedPath: "/some/path",
					Key:          "SOME_KEY",
					Value:        "shazam",
					RedactWith:   "**SOME_KEY**",
				},
				{
					ProviderName: "test",
					ResolvedPath: "/some/path",
					Key:          "OTHER_KEY",
					Value:        "loot",
					RedactWith:   "**OTHER_KEY**",
				},
			},
			s: `
func Foobar(){
	secret := "loot"
	callService(secret, "shazam")
}
`,
			sr: `
func Foobar(){
	secret := "**OTHER_KEY**"
	callService(secret, "**SOME_KEY**")
}
`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			redactor := NewRedactor(buf, c.entries)

			_, err := io.WriteString(redactor, c.s)
			assert.NoError(t, err)

			err = redactor.Close()
			assert.NoError(t, err)

			assert.Equal(t, buf.String(), c.sr)
		})
	}
}
func TestRedactorMultiple(t *testing.T) {
	entries := []core.EnvEntry{
		{
			ProviderName: "test",
			ResolvedPath: "/some/path",
			Key:          "SOME_KEY",
			Value:        "shazam",
			RedactWith:   "**SOME_KEY**",
		},
		{
			ProviderName: "test",
			ResolvedPath: "/some/path",
			Key:          "OTHER_KEY",
			Value:        "loot",
			RedactWith:   "**OTHER_KEY**",
		},
	}
	s := `
func Foobar(){
	secret := "loot"
	callService(secret, "shazam")
}
`
	sr := `
func Foobar(){
	secret := "**OTHER_KEY**"
	callService(secret, "**SOME_KEY**")
}
`

	buf := bytes.NewBuffer(nil)
	redactor := NewRedactor(buf, entries)

	_, err := io.WriteString(redactor, s)
	assert.NoError(t, err)

	err = redactor.Close()
	assert.NoError(t, err)

	assert.Equal(t, buf.String(), sr)
}

func TestRedactor(t *testing.T) {
	entries := []core.EnvEntry{
		{
			ProviderName: "test",
			ResolvedPath: "/some/path",
			Key:          "SOME_KEY",
			Value:        "shazam",
			RedactWith:   "**NOPE**",
		},
	}
	s := `
func Foobar(){
	secret := "shazam"
	callService(secret, "shazam")
}
`
	sr := `
func Foobar(){
	secret := "**NOPE**"
	callService(secret, "**NOPE**")
}
`

	buf := bytes.NewBuffer(nil)
	redactor := NewRedactor(buf, entries)

	_, err := io.WriteString(redactor, s)
	assert.NoError(t, err)

	err = redactor.Close()
	assert.NoError(t, err)

	assert.Equal(t, buf.String(), sr)
}

func TestRedactor_Close(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	redactor := NewRedactor(buf, nil)

	assert.NoError(t, redactor.Close())
	// can be close more than once.
	assert.NoError(t, redactor.Close())
}
