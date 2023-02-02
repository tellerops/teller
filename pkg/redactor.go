package pkg

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spectralops/teller/pkg/core"
)

type Redactor struct {
	io.WriteCloser
	err <-chan error
}

func NewRedactor(dist io.Writer, entries []core.EnvEntry) *Redactor {
	entries = append([]core.EnvEntry(nil), entries...)
	sort.Sort(core.EntriesByValueSize(entries))

	r, w := io.Pipe()
	ch := make(chan error)
	go func() {
		defer close(ch)

		s := bufio.NewScanner(r)
		buf := make([]byte, 0, 64*1024)
		s.Buffer(buf, 10*1024*1024) // 10MB lines correlating to 10MB files max (bundles?)

		for s.Scan() {
			line := s.Text()
			for _, ent := range entries {
				line = strings.ReplaceAll(line, ent.Value, ent.RedactWith)
			}
			if _, err := fmt.Fprintln(dist, line); err != nil {
				ch <- r.CloseWithError(err)
				return
			}
		}
		ch <- s.Err()
	}()

	return &Redactor{
		WriteCloser: w,
		err:         ch,
	}
}

func (r *Redactor) Close() error {
	err := r.WriteCloser.Close()
	<-r.err

	return err
}
