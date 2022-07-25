package logging

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/alecthomas/assert"
)

func TestSetOutput(t *testing.T) {

	buf := new(bytes.Buffer)
	logger := New()

	logger.SetOutput(buf)
	logger.Error("error message")

	assert.NotEmpty(t, buf.String())

}

func TestJSONFormat(t *testing.T) {

	buf := new(bytes.Buffer)
	logger := New()

	logger.SetOutput(buf)
	logger.SetOutputFormat("json")
	logger.Error("error message")

	var js interface{}
	assert.True(t, json.Unmarshal(buf.Bytes(), &js) == nil)

}

func TestTextFormat(t *testing.T) {

	buf := new(bytes.Buffer)
	logger := New()

	logger.SetOutput(buf)
	logger.SetOutputFormat("text")
	logger.Error("error message")

	assert.Contains(t, buf.String(), "level=error")
	assert.Contains(t, buf.String(), "msg=\"error message\"")

}

func TestWithFields(t *testing.T) {

	buf := new(bytes.Buffer)
	logger := New()

	logger.SetOutput(buf)
	logger.SetOutputFormat("text")
	logger.WithFields(map[string]interface{}{"field-a": "test", "field-b": "test"}).Error("error message")

	assert.Contains(t, buf.String(), "field-a=test")
	assert.Contains(t, buf.String(), "field-b=test")

}

func TestWithField(t *testing.T) {

	buf := new(bytes.Buffer)
	logger := New()

	logger.SetOutput(buf)
	logger.SetOutputFormat("text")
	logger.WithField("field-a", "test").Error("error message")

	assert.Contains(t, buf.String(), "field-a=test")

}

func TestWithError(t *testing.T) {

	buf := new(bytes.Buffer)
	logger := New()

	logger.SetOutput(buf)
	logger.SetOutputFormat("text")
	logger.WithError(errors.New("some-error")).Error("error message")

	assert.Contains(t, buf.String(), "error=some-error")

}

func TestLevelSetter(t *testing.T) {

	logger := New()
	logger.SetLevel("debug")
	assert.Equal(t, "debug", logger.GetLevel())

}
