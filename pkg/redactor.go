package pkg

import (
	"sort"
	"strings"

	"github.com/spectralops/teller/pkg/core"
)

type Redactor struct {
	Entries []core.EnvEntry
}

func NewRedactor(entries []core.EnvEntry) *Redactor {
	return &Redactor{
		Entries: entries,
	}
}

func (r *Redactor) Redact(s string) string {
	redacted := s
	entries := append([]core.EnvEntry(nil), r.Entries...)

	sort.Sort(core.EntriesByValueSize(entries))
	for i := range entries {
		ent := entries[i]
		redacted = strings.ReplaceAll(redacted, ent.Value, ent.RedactWith)
	}

	return redacted
}
