package pkg

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/karrick/godirwalk"
	"github.com/spectralops/teller/pkg/core"
	"github.com/thoas/go-funk"
	"gopkg.in/yaml.v3"
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
		for i := range tl.Entries {
			b := tl.Entries[i]
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
	for i := range tl.Entries {
		v := tl.Entries[i]
		fmt.Fprintf(&b, "export %s=%s\n", v.Key, v.Value)
	}
	return b.String()
}

// Export variables into a .env format (basically a KEY=VAL format, that's also compatible with Docker)
func (tl *Teller) ExportDotenv() string {
	var b bytes.Buffer

	for i := range tl.Entries {
		v := tl.Entries[i]
		fmt.Fprintf(&b, "%s=%s\n", v.Key, v.Value)
	}
	return b.String()
}

func (tl *Teller) ExportYAML() (out string, err error) {
	valmap := map[string]string{}

	for i := range tl.Entries {
		v := tl.Entries[i]
		valmap[v.Key] = v.Value
	}
	content, err := yaml.Marshal(valmap)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (tl *Teller) ExportJSON() (out string, err error) {
	valmap := map[string]string{}

	for i := range tl.Entries {
		v := tl.Entries[i]
		valmap[v.Key] = v.Value
	}
	content, err := json.MarshalIndent(valmap, "", "  ")
	if err != nil {
		return "", err
	}
	return string(content), nil
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
		for i := range entries {
			ent := entries[i]
			if !ent.IsFound || ent.Value == "" || ent.Severity == core.None {
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

// Template Teller vars from a given path (can be file or folder)
func (tl *Teller) Template(from, to string) error {

	fileInfo, err := os.Stat(from)
	if err != nil {
		return fmt.Errorf("invald path. err: %v", err)
	}

	if fileInfo.IsDir() {
		return tl.templateFolder(from, to)
	}

	return tl.templateFile(from, to)
}

// templateFolder scan given folder and inject Teller vars for each search file
func (tl *Teller) templateFolder(from, to string) error {

	err := godirwalk.Walk(from, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			if de.IsDir() {
				return nil
			}
			copyTo := filepath.Join(to, strings.Replace(osPathname, from, "", 1))
			return tl.templateFile(osPathname, copyTo)
		},
		Unsorted: true,
	})

	return err
}

// templateFile inject Teller vars into a single file
func (tl *Teller) templateFile(from, to string) error {
	tfile, err := os.ReadFile(from)
	if err != nil {
		return fmt.Errorf("cannot read template '%v': %v", from, err)
	}

	res, err := tl.Templating.ForTemplate(string(tfile), tl.Entries)
	if err != nil {
		return fmt.Errorf("cannot render template '%v': %v", from, err)
	}

	info, _ := os.Stat(from)

	// crate destination path if not exists
	toFolder := filepath.Dir(to)
	if _, err = os.Stat(toFolder); os.IsNotExist(err) {
		err = os.MkdirAll(toFolder, os.ModePerm)
		if err != nil {
			fmt.Println("1")
			return fmt.Errorf("cannot create folder '%v': %v", to, err)
		}
	}

	err = os.WriteFile(to, []byte(res), info.Mode())
	if err != nil {
		return fmt.Errorf("cannot save to '%v': %v", to, err)
	}
	return nil
}

func updateParams(ent *core.EnvEntry, from *core.KeyPath, pname string) {
	ent.ProviderName = pname
	ent.Source = from.Source
	ent.Sink = from.Sink

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

func (tl *Teller) CollectFromProvider(pname string) ([]core.EnvEntry, error) {
	entries := []core.EnvEntry{}
	conf, ok := tl.Config.Providers[pname]
	p, err := tl.Providers.GetProvider(pname)
	if err != nil && ok && conf.Kind != "" {
		// ok, maybe same provider, with 'kind'?
		p, err = tl.Providers.GetProvider(conf.Kind)
	}

	// still no provider? bail.
	if err != nil {
		return nil, err
	}

	if conf.EnvMapping != nil {
		es, err := p.GetMapping(tl.Populate.KeyPath(*conf.EnvMapping))
		if err != nil {
			return nil, err
		}

		//nolint
		for k, v := range es {
			// optionally remap environment variables synced from the provider
			if val, ok := conf.EnvMapping.Remap[v.Key]; ok {
				es[k].Key = val
			}
			updateParams(&es[k], conf.EnvMapping, pname)
		}

		entries = append(entries, es...)
	}

	if conf.Env != nil {
		//nolint
		for k, v := range *conf.Env {
			ent, err := p.Get(tl.Populate.KeyPath(v.WithEnv(k)))
			if err != nil {
				if v.Optional {
					continue
				} else {
					return nil, err
				}
			} else {
				//nolint
				updateParams(ent, &v, pname)
				entries = append(entries, *ent)
			}
		}
	}
	return entries, nil
}

func (tl *Teller) CollectFromProviderMap(ps *ProvidersMap) ([]core.EnvEntry, error) {
	entries := []core.EnvEntry{}
	for pname := range *ps {
		pents, err := tl.CollectFromProvider(pname)
		if err != nil {
			return nil, err
		}
		entries = append(entries, pents...)
	}

	sort.Sort(core.EntriesByKey(entries))
	return entries, nil
}

// The main "load all variables from all providers" logic. Walks over all definitions in the tellerfile
// and then: fetches, converts, creates a new EnvEntry. We're also mapping the sensitivity aspects of it.
// Note that for a similarly named entry - last one wins.
func (tl *Teller) Collect() error {
	t := tl.Config
	entries, err := tl.CollectFromProviderMap(&t.Providers)
	if err != nil {
		return err
	}

	tl.Entries = entries
	tl.Redactor = NewRedactor(entries)
	return nil
}

func (tl *Teller) Drift(providerNames []string) []core.DriftedEntry {
	sources := map[string]core.EnvEntry{}
	targets := map[string][]core.EnvEntry{}
	filtering := len(providerNames) > 0
	for i := range tl.Entries {
		ent := tl.Entries[i]
		if filtering && !funk.ContainsString(providerNames, ent.ProviderName) {
			continue
		}
		if ent.Source != "" {
			sources[ent.Source+":"+ent.Key] = ent
		} else if ent.Sink != "" {
			k := ent.Sink + ":" + ent.Key
			ents := targets[k]
			if ents == nil {
				targets[k] = []core.EnvEntry{ent}
			} else {
				targets[k] = append(ents, ent)
			}
		}
	}

	drifts := []core.DriftedEntry{}

	//nolint
	for sk, source := range sources {
		ents := targets[sk]
		if ents == nil {
			drifts = append(drifts, core.DriftedEntry{Diff: "missing", Source: source})
		}

		for _, e := range ents {
			if e.Value != source.Value {
				drifts = append(drifts, core.DriftedEntry{Diff: "changed", Source: source, Target: e})
			}
		}
	}

	sort.Sort(core.DriftedEntriesBySource(drifts))
	return drifts
}

func (tl *Teller) GetProviderByName(pname string) (*MappingConfig, core.Provider, error) {
	pcfg, ok := tl.Config.Providers[pname]
	if !ok {
		return nil, nil, fmt.Errorf("provider %v not found", pname)
	}
	p := pname
	if pcfg.Kind != "" {
		p = pcfg.Kind
	}
	provider, err := tl.Providers.GetProvider(p)
	return &pcfg, provider, err
}

func (tl *Teller) Put(kvmap map[string]string, providerNames []string, sync bool, directPath string) error {
	for _, pname := range providerNames {
		pcfg, provider, err := tl.GetProviderByName(pname)
		if err != nil {
			return fmt.Errorf("cannot create provider %v: %v", pname, err)
		}

		useDirectPath := directPath != ""

		// XXXWIP design - decide porcelain or not, errors?
		if sync {
			var kvp core.KeyPath
			if useDirectPath {
				kvp = core.KeyPath{Path: directPath}
			} else {
				if pcfg.EnvMapping == nil {
					return fmt.Errorf("there is no env sync mapping for provider '%v'", pname)
				}
				kvp = *pcfg.EnvMapping
			}
			kvpResolved := tl.Populate.KeyPath(kvp)
			err := provider.PutMapping(kvpResolved, kvmap)
			if err != nil {
				return fmt.Errorf("cannot put (sync) %v in provider %v: %v", kvpResolved.Path, pname, err)
			}
			tl.Porcelain.DidPutKVP(kvpResolved, pname, true)
		} else {
			if pcfg.Env == nil {
				return fmt.Errorf("there is no specific key mapping to map to for provider '%v'", pname)
			}

			for k, v := range kvmap {
				// get the kvp for specific mapping
				ok := false
				var kvp core.KeyPath

				if useDirectPath {
					kvp = core.KeyPath{Path: directPath}
					ok = true
				} else {
					kvp, ok = (*pcfg.Env)[k]
				}

				if ok {
					kvpResolved := tl.Populate.KeyPath(kvp.WithEnv(k))
					err := provider.Put(kvpResolved, v)
					if err != nil {
						return fmt.Errorf("cannot put %v in provider %v: %v", k, pname, err)
					}
					tl.Porcelain.DidPutKVP(kvpResolved, pname, false)
				} else {
					tl.Porcelain.NoPutKVP(k, pname)
				}
			}
		}
	}

	return nil
}

func (tl *Teller) Sync(from string, to []string, sync bool) error {
	entries, err := tl.CollectFromProvider(from)
	if err != nil {
		return err
	}
	kvmap := map[string]string{}
	for i := range entries {
		ent := entries[i]
		kvmap[ent.Key] = ent.Value
	}

	err = tl.Put(kvmap, to, sync, "")
	return err
}

func (tl *Teller) MirrorDrift(source, target string) ([]core.DriftedEntry, error) {
	drifts := []core.DriftedEntry{}
	sourceEntries, err := tl.CollectFromProvider(source)
	if err != nil {
		return nil, err
	}

	targetEntries, err := tl.CollectFromProvider(target)
	if err != nil {
		return nil, err
	}

	for i := range sourceEntries {
		sent := sourceEntries[i]
		tent := funk.Find(targetEntries, func(ent core.EnvEntry) bool {
			return sent.Key == ent.Key
		})
		if tent == nil {
			drifts = append(drifts, core.DriftedEntry{Diff: "missing", Source: sent})
			continue
		}
		tentry := tent.(core.EnvEntry)
		if sent.Value != tentry.Value {
			drifts = append(drifts, core.DriftedEntry{Diff: "changed", Source: sent, Target: tentry})
		}
	}

	return drifts, nil
}

func (tl *Teller) Delete(keys []string, providerNames []string, directPath string) error {
	for _, pname := range providerNames {
		pcfg, provider, err := tl.GetProviderByName(pname)
		if err != nil {
			return fmt.Errorf("cannot create provider %v: %v", pname, err)
		}

		useDirectPath := directPath != ""

		if pcfg.Env == nil {
			return fmt.Errorf("there is no specific key mapping to map to for provider '%v'", pname)
		}

		for _, key := range keys {
			// get the kp for specific mapping
			var (
				kp core.KeyPath
				ok bool
			)

			if useDirectPath {
				kp = core.KeyPath{Path: directPath}
				ok = true
			} else {
				kp, ok = (*pcfg.Env)[key]
			}

			if !ok {
				tl.Porcelain.NoDeleteKP(key, pname)
				continue
			}

			kpResolved := tl.Populate.KeyPath(kp.WithEnv(key))
			err := provider.Delete(kpResolved)
			if err != nil {
				return fmt.Errorf("cannot delete %v in provider %q: %v", key, pname, err)
			}

			tl.Porcelain.DidDeleteKP(kp, pname)
		}
	}

	return nil
}
