package core

import (
	"fmt"
	"os"
	"strings"
)

const (
	populateFromEnvironment = "env:"
	populateWithDefault     = ":default:"
)

type Opts map[string]string
type Populate struct {
	opts map[string]string
}

func NewPopulate(opts Opts) *Populate {
	return &Populate{
		opts: opts,
	}
}

func (p *Populate) FindAndReplace(path string) string {
	populated := path
	for k, v := range p.opts {
		val := v
		if strings.HasPrefix(v, populateFromEnvironment) {
			evar := strings.TrimPrefix(v, populateFromEnvironment)
			evar, defaultValue := p.parseDefaultValue(evar)
			val = os.Getenv(evar)
			if val == "" {
				val = defaultValue
			}
		}
		populated = strings.ReplaceAll(populated, fmt.Sprintf("{{%s}}", k), val)
	}
	return populated
}

func (p *Populate) KeyPath(kp KeyPath) KeyPath {
	path := p.FindAndReplace(kp.Path)
	populated := path
	for k, v := range p.opts {
		val := v
		if strings.HasPrefix(v, populateFromEnvironment) {
			evar := strings.TrimPrefix(v, populateFromEnvironment)
			evar, defaultValue := p.parseDefaultValue(evar)
			val = os.Getenv(evar)
			if val == "" {
				val = defaultValue
			}
		}
		populated = strings.ReplaceAll(populated, fmt.Sprintf("{{%s}}", k), val)
	}
	return kp.SwitchPath(path)
}

// parseDefaultValue returns that field name and the default value if `populateWithDefault` was found
// Example 1: FOO:default:BAR -> the function return FOO, BAR
// Example 2: FOO -> the function return FOO, "" (empty value)
func (p *Populate) parseDefaultValue(evar string) (string, string) {
	if strings.Contains(evar, populateWithDefault) {
		data := strings.Split(evar, populateWithDefault)
		if len(data) == 2 {
			return data[0], data[1]
		}
	}
	return evar, ""
}
