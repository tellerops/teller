package pkg

import (
	"testing"

	"github.com/alecthomas/assert"
	"github.com/spectralops/teller/pkg/core"
)

func TestRedactorOverlap(t *testing.T) {

	// in this case we dont want '123' to appear in the clear after all redactions are made.
	// it can happen if the smaller secret get replaced first because both
	// secrets overlap. we need to ensure the wider secrets always get
	// replaced first.

	entries := []core.EnvEntry{
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
	}
	redactor := Redactor{Entries: entries}
	s := `
	func Foobar(){
		secret := "hello"
		callService(secret, "hello123")
		// hello, hello123
	}
	`
	sr := `
	func Foobar(){
		secret := "**OTHER_KEY**"
		callService(secret, "**SOME_KEY**")
		// **OTHER_KEY**, **SOME_KEY**
	}
	`

	assert.Equal(t, redactor.Redact(s), sr)
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
	redactor := Redactor{Entries: entries}
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

	assert.Equal(t, redactor.Redact(s), sr)
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
	redactor := Redactor{Entries: entries}
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

	assert.Equal(t, redactor.Redact(s), sr)
}
