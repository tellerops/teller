package pkg

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/karrick/godirwalk"
	"github.com/spectralops/teller/pkg/core"
	"github.com/thoas/go-funk"
)

// Teller
// Cmd - command to execute if any given.
// Porcelain - wrapping teller in a nice porcelain; in other words the textual UI for teller.
// Providers - the available providers to use.
// Entries - when loaded, these contains the mapped entries. Load them with Collect()
// Templating - Teller's templating options.
type Teller struct {
	Redact     bool
	Cmd        []string
	Config     *TellerFile
	Porcelain  *Porcelain
	Populate   *core.Populate
	Providers  Providers
	Entries    []core.EnvEntry
	Templating *Templating
	Redactor   *Redactor
}

// Create a new Teller instance, using a tellerfile, and a command to execute (if any)
func NewTeller(tlrfile *TellerFile, cmd []string, redact bool) *Teller {
	return &Teller{
		Redact:     redact,
		Config:     tlrfile,
		Cmd:        cmd,
		Providers:  &BuiltinProviders{},
		Populate:   core.NewPopulate(tlrfile.Opts),
		Porcelain:  &Porcelain{Out: os.Stdout},
		Templating: &Templating{},
		Redactor:   &Redactor{},
	}
}

// shorthand for killing the current process with a bad exist code, but without a Go panic
func bail(e error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", e)
	os.Exit(1)
}

// execute a command, and take care to sanitize the child process environment (conditionally)
func (tl *Teller) execCmd(cmd string, cmdArgs []string, withRedaction bool) error {
	command := exec.Command(cmd, cmdArgs...)
	if !tl.Config.CarryEnv {
		command.Env = funk.Map(tl.Entries, func(ent interface{}) string {
			return fmt.Sprintf("%s=%s", ent.(core.EnvEntry).Key, ent.(core.EnvEntry).Value)
		}).([]string)

		command.Env = append(command.Env, funk.Map([]string{"USER", "HOME", "PATH"}, func(k string) string { return fmt.Sprintf("%s=%s", k, os.Getenv(k)) }).([]string)...)

	} else {
		for _, b := range tl.Entries {
			os.Setenv(b.Key, b.Value)
		}
	}
	if withRedaction {
		out, err := command.CombinedOutput()
		redacted := tl.Redactor.Redact(string(out))
		os.Stdout.Write([]byte(redacted))
		return err
	}

	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}

func (tl *Teller) PrintEnvKeys() {
	tl.Porcelain.PrintContext(tl.Config.Project, tl.Config.LoadedFrom)
	tl.Porcelain.VSpace(1)
	tl.Porcelain.PrintEntries(tl.Entries)
}

// Export variables into a shell sourceable format
func (tl *Teller) ExportEnv() string {
	var b bytes.Buffer

	fmt.Fprintf(&b, "#!/bin/sh\n")
	for _, v := range tl.Entries {
		fmt.Fprintf(&b, "export %s=%s\n", v.Key, v.Value)
	}
	return b.String()
}

// Export variables into a .env format (basically a KEY=VAL format, that's also compatible with Docker)
func (tl *Teller) ExportDotenv() string {
	var b bytes.Buffer

	for _, v := range tl.Entries {
		fmt.Fprintf(&b, "%s=%s\n", v.Key, v.Value)
	}
	return b.String()
}

func renderWizardTemplate(fname string, answers *core.WizardAnswers) error {
	t, err := template.New("t").Parse(TellerFileTemplate)
	if err != nil {
		return err
	}
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	err = t.Execute(f, answers)
	if err != nil {
		return err
	}
	return nil
}

// Start an interactive wizard, that will create a file when completed.
func (tl *Teller) SetupNewProject(fname string) error {
	answers, err := tl.Porcelain.StartWizard()
	if err != nil {
		return err
	}
	err = renderWizardTemplate(fname, answers)
	if err != nil {
		return err
	}

	tl.Porcelain.DidCreateNewFile(fname)
	return nil
}

// Execute a command with teller. This requires all entries to be loaded beforehand with Collect()
func (tl *Teller) RedactLines(r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	//nolint
	buf := make([]byte, 0, 64*1024)
	//nolint
	scanner.Buffer(buf, 10*1024*1024) // 10MB lines correlating to 10MB files max (bundles?)

	for scanner.Scan() {
		if _, err := fmt.Fprintln(w, tl.Redactor.Redact(string(scanner.Bytes()))); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// Execute a command with teller. This requires all entries to be loaded beforehand with Collect()
func (tl *Teller) Exec() {
	tl.Porcelain.PrintContext(tl.Config.Project, tl.Config.LoadedFrom)
	if tl.Config.Confirm != "" {
		tl.Porcelain.VSpace(1)
		tl.Porcelain.PrintEntries(tl.Entries)
		tl.Porcelain.VSpace(1)
		if !tl.Porcelain.AskForConfirmation(tl.Populate.FindAndReplace(tl.Config.Confirm)) {
			return
		}
	}

	err := tl.execCmd(tl.Cmd[0], tl.Cmd[1:], tl.Redact)
	if err != nil {
		bail(err)
	}
}

func hasBindata(line []byte) bool {
	for _, el := range line {
		if el == 0 {
			return true
		}
	}
	return false
}
func checkForMatches(path string, entries []core.EnvEntry) ([]core.Match, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	retval := []core.Match{}

	scanner := bufio.NewScanner(file)
	//nolint
	buf := make([]byte, 0, 64*1024)
	//nolint
	scanner.Buffer(buf, 10*1024*1024) // 10MB lines correlating to 10MB files max (bundles?)

	var lineNumber int = 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Bytes()
		if hasBindata(line) {
			// This is a binary file.  Skip it!
			return retval, nil
		}

		linestr := string(line)
		for _, ent := range entries {
			if ent.Value == "" || ent.Severity == core.None {
				continue
			}
			if matchIndex := strings.Index(linestr, ent.Value); matchIndex != -1 {
				m := core.Match{
					Path: path, Line: linestr, LineNumber: lineNumber, MatchIndex: matchIndex, Entry: ent}
				retval = append(retval, m)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return retval, nil
}

// Scan for entries. Each of the mapped entries is considered highly sensitive unless stated other wise (with sensitive: high|medium|low|none)
// as such, we can offer a security scan to locate those in the current codebase (if the entries are sensitive and are placed inside a vault or
// similar store, what's the purpose of hardcoding these? let's help ourselves and locate throughout all the files in the path given)
func (tl *Teller) Scan(path string, silent bool) ([]core.Match, error) {
	if path == "" {
		path = "."
	}

	start := time.Now()
	findings := []core.Match{}
	err := godirwalk.Walk(path, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			// Following string operation is not most performant way
			// of doing this, but common enough to warrant a simple
			// example here:
			if strings.Contains(osPathname, ".git") {
				return godirwalk.SkipThis
			}
			if de.IsRegular() {
				ms, err := checkForMatches(osPathname, tl.Entries)
				if err == nil {
					findings = append(findings, ms...)
				}
				// else {
				// 	can't open, can't scan
				// 	fmt.Println("error: %v", err)
				// }
			}
			return nil
		},
		Unsorted: true, // (optional) set true for faster yet non-deterministic enumeration (see godoc)
	})

	elapsed := time.Since(start)
	if len(findings) > 0 && !silent {
		tl.Porcelain.PrintMatches(findings)
		tl.Porcelain.VSpace(1)
	}

	if !silent {
		tl.Porcelain.PrintMatchSummary(findings, tl.Entries, elapsed)
	}
	return findings, err
}

func (tl *Teller) TemplateFile(from, to string) error {
	tfile, err := ioutil.ReadFile(from)
	if err != nil {
		return fmt.Errorf("cannot read template '%v': %v", from, err)
	}

	res, err := tl.Templating.ForTemplate(string(tfile), tl.Entries)
	if err != nil {
		return fmt.Errorf("cannot render template '%v': %v", from, err)
	}

	info, _ := os.Stat(from)

	err = ioutil.WriteFile(to, []byte(res), info.Mode())
	if err != nil {
		return fmt.Errorf("cannot save to '%v': %v", to, err)
	}
	return nil
}

func updateParams(ent *core.EnvEntry, from *core.KeyPath) {
	if from.Severity == "" {
		ent.Severity = core.High
	} else {
		ent.Severity = from.Severity
	}

	if from.RedactWith == "" {
		ent.RedactWith = "**REDACTED**"
	} else {
		ent.RedactWith = from.RedactWith
	}
}

// The main "load all variables from all providers" logic. Walks over all definitions in the tellerfile
// and then: fetches, converts, creates a new EnvEntry. We're also mapping the sensitivity aspects of it.
// Note that for a similarly named entry - last one wins.
func (tl *Teller) Collect() error {
	t := tl.Config
	entries := []core.EnvEntry{}
	for pname, conf := range t.Providers {
		p, err := tl.Providers.GetProvider(pname)
		if err != nil {
			return err
		}

		if conf.EnvMapping != nil {
			es, err := p.GetMapping(tl.Populate.KeyPath(*conf.EnvMapping))
			if err != nil {
				return err
			}

			// optionally remap environment variables synced from the provider
			for k, v := range es {
				if val, ok := conf.EnvMapping.Remap[v.Key]; ok {
					es[k].Key = val
					updateParams(&es[k], conf.EnvMapping)
				}
			}

			entries = append(entries, es...)
		}

		if conf.Env != nil {
			for k, v := range *conf.Env {
				ent, err := p.Get(tl.Populate.KeyPath(v.WithEnv(k)))
				if err != nil {
					if v.Optional {
						continue
					} else {
						return err
					}
				} else {
					//nolint
					updateParams(ent, &v)
					entries = append(entries, *ent)
				}
			}
		}
	}

	sort.Sort(core.EntriesByKey(entries))
	tl.Entries = entries
	tl.Redactor = NewRedactor(entries)
	return nil
}
