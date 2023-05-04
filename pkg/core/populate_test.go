package core

import (
	"os"
	"testing"

	"github.com/alecthomas/assert"
)

func TestPopulatePath(t *testing.T) {
	os.Setenv("TELLER_TEST_FOO", "test-foo")

	p := NewPopulate(map[string]string{
		"foo":                "bar",
		"teller-env":         "env:TELLER_TEST_FOO",
		"teller-env-default": "env:TELLER_TEST_DEFAULT,default-value",
	})

	assert.Equal(t, p.FindAndReplace("foo/{{foo}}/qux/{{foo}}"), "foo/bar/qux/bar")
	assert.Equal(t, p.FindAndReplace("foo/qux"), "foo/qux")
	assert.Equal(t, p.FindAndReplace("foo/{{none}}"), "foo/{{none}}")
	assert.Equal(t, p.FindAndReplace("foo/{{teller-env}}"), "foo/test-foo")
	assert.Equal(t, p.FindAndReplace("foo/{{teller-env-default}}"), "foo/default-value")
	assert.Equal(t, p.FindAndReplace(""), "")

	kp := KeyPath{
		Path:     "{{foo}}/hey",
		Env:      "env",
		Decrypt:  true,
		Optional: true,
	}
	assert.Equal(t, p.KeyPath(kp), KeyPath{
		Path:     "bar/hey",
		Env:      "env",
		Decrypt:  true,
		Optional: true,
	})
	kp = KeyPath{
		Path:     "{{teller-env}}/hey",
		Env:      "env",
		Decrypt:  true,
		Optional: true,
	}
	assert.Equal(t, p.KeyPath(kp), KeyPath{
		Path:     "test-foo/hey",
		Env:      "env",
		Decrypt:  true,
		Optional: true,
	})
	kp = KeyPath{
		Path:     "{{teller-env-default}}/hey",
		Env:      "env",
		Decrypt:  true,
		Optional: true,
	}
	assert.Equal(t, p.KeyPath(kp), KeyPath{
		Path:     "default-value/hey",
		Env:      "env",
		Decrypt:  true,
		Optional: true,
	})
}

func TestPopulateDefaultValue(t *testing.T) {
	key, value := parseDefaultValue("TELLER_TEST_FOO")
	assert.Equal(t, key, "TELLER_TEST_FOO")
	assert.Equal(t, value, "")

	key, value = parseDefaultValue("TELLER_TEST_FOO, default,,value")
	assert.Equal(t, key, "TELLER_TEST_FOO")
	assert.Equal(t, value, "default,,value")
}
