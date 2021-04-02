package pkg

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"text/template"

	"github.com/spectralops/teller/pkg/core"
	"github.com/thoas/go-funk"
)

type Teller struct {
	Cmd        []string
	Config     *TellerFile
	Porcelain  *Porcelain
	Populate   *core.Populate
	Providers  Providers
	Entries    []core.EnvEntry
	Templating *Templating
}

func NewTeller(tlrfile *TellerFile, cmd []string) *Teller {
	return &Teller{
		Config:     tlrfile,
		Cmd:        cmd,
		Providers:  &BuiltinProviders{},
		Populate:   core.NewPopulate(tlrfile.Opts),
		Porcelain:  &Porcelain{Out: os.Stdout},
		Templating: &Templating{},
	}
}
func bail(e error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", e)
	os.Exit(1)
}

func (tl *Teller) execCmd(cmd string, cmdArgs []string) error {
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

func (tl *Teller) ExportEnv() string {
	var b bytes.Buffer

	fmt.Fprintf(&b, "#!/bin/sh\n")
	for _, v := range tl.Entries {
		fmt.Fprintf(&b, "export %s=%s\n", v.Key, v.Value)
	}
	return b.String()
}

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

	err := tl.execCmd(tl.Cmd[0], tl.Cmd[1:])
	if err != nil {
		bail(err)
	}
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
					entries = append(entries, *ent)
				}

			}
		}
	}

	sort.Sort(core.EntriesByKey(entries))
	tl.Entries = entries
	return nil
}
