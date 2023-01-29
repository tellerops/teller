package core

import (
	"fmt"
	"os"
	"strings"
)

const (
	populateFromEnvironment = "env:"
	populateWithDefault     = ","
)

type Opts map[string]string
type Populate struct {
	rules map[string]string
}

func NewPopulate(opts Opts) *Populate {
	rules := make(map[string]string)
	for k, v := range opts {
		val := v
		if strings.HasPrefix(v, populateFromEnvironment) {
			evar := strings.TrimPrefix(v, populateFromEnvironment)
			evar, defaultValue := parseDefaultValue(evar)
			val = os.Getenv(evar)
			if val == "" {
				val = defaultValue
			}
		}
		rules[fmt.Sprintf("{{%s}}", k)] = val
	}

	return &Populate{
		rules: rules,
	}
}

func (p *Populate) FindAndReplace(path string) string {
	populated := path
	for k, v := range p.rules {
		populated = strings.ReplaceAll(populated, k, v)
	}
	return populated
}

func (p *Populate) KeyPath(kp KeyPath) KeyPath {
	return kp.SwitchPath(p.FindAndReplace(kp.Path))
}

// parseDefaultValue returns that field name and the default value if `populateWithDefault` was found
// Example 1: FOO,BAR -> the function return FOO, BAR
// Example 2: FOO -> the function return FOO, "" (empty value)
func parseDefaultValue(evar string) (key, defaultValue string) {
	if strings.Contains(evar, populateWithDefault) {
		data := strings.SplitN(evar, populateWithDefault, 2) //nolint
		if len(data) == 2 {                                  //nolint
			return data[0], strings.TrimSpace(data[1])
		}
	}
	return evar, ""
}
