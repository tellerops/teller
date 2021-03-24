package core

import (
	"fmt"
	"os"
	"strings"
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
		if strings.HasPrefix(v, "env:") {
			evar := strings.TrimPrefix(v, "env:")
			val = os.Getenv(evar)
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
		if strings.HasPrefix(v, "env:") {
			evar := strings.TrimPrefix(v, "env:")
			val = os.Getenv(evar)
		}
		populated = strings.ReplaceAll(populated, fmt.Sprintf("{{%s}}", k), val)
	}
	return kp.SwitchPath(path)
}
