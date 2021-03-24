package pkg

import (
	"bytes"
	"text/template"

	"github.com/spectralops/teller/pkg/core"
)

type Templating struct {
}
type viewmodel struct {
	Teller *core.EnvEntryLookup
}

func (t *Templating) New() *Templating {
	return &Templating{}
}

func (t *Templating) ForTemplate(tmpl string, entries []core.EnvEntry) (string, error) {
	lookup := core.EnvEntryLookup{
		Entries: entries,
	}

	tt, err := template.New("t").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var output bytes.Buffer

	err = tt.Execute(&output, viewmodel{Teller: &lookup})
	if err != nil {
		return "", err
	}

	return output.String(), nil
}

func (t *Templating) ForGlob() *Templating {
	return &Templating{}
}
