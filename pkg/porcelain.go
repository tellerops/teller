package pkg

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/samber/lo"

	"github.com/fatih/color"
	"github.com/jftuga/ellipsis"
	"github.com/spectralops/teller/pkg/core"
	"github.com/spectralops/teller/pkg/utils"
)

type Porcelain struct {
	Out io.Writer
}

func (p *Porcelain) StartWizard() (*core.WizardAnswers, error) {
	wd, _ := os.Getwd()
	workingfolder := utils.LastSegment(wd)

	providers := BuiltinProviders{}
	providerNames := providers.ProviderHumanToMachine()
	displayProviders := lo.Keys(providerNames)
	sort.Strings(displayProviders)
	// the questions to ask
	var qs = []*survey.Question{
		{
			Name: "project",
			Prompt: &survey.Input{
				Message: "Project name?",
				Default: workingfolder,
			},
			Validate: survey.Required,
		},
		{
			Name: "providers",
			Prompt: &survey.MultiSelect{
				Message:  "Select your secret providers",
				PageSize: 10,
				Options:  displayProviders,
			},
		},
		{
			Name: "confirm",
			Prompt: &survey.Confirm{
				Message: "Would you like extra confirmation before accessing secrets?",
			},
		},
	}

	answers := core.WizardAnswers{}
	// perform the questions
	err := survey.Ask(qs, &answers)
	if err != nil {
		return nil, err
	}

	if len(answers.Providers) == 0 {
		return nil, fmt.Errorf("you need to select at least one provider")
	}

	answers.ProviderKeys = map[string]bool{}
	for _, plabel := range answers.Providers {
		answers.ProviderKeys[providerNames[plabel]] = true
	}

	return &answers, nil
}

func (p *Porcelain) DidCreateNewFile(fname string) {
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Fprintf(p.Out, "Created file: %v\n", green(fname))
}

func (p *Porcelain) AskForConfirmation(msg string) bool {
	yesno := false
	prompt := &survey.Confirm{
		Message: msg,
	}
	_ = survey.AskOne(prompt, &yesno)
	return yesno
}

func (p *Porcelain) VSpace(size int) {
	fmt.Fprint(p.Out, strings.Repeat("\n", size))
}

func (p *Porcelain) PrintContext(projectName, loadedFrom string) {
	green := color.New(color.FgGreen).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()

	fmt.Fprintf(p.Out, "-*- %s: loaded variables for %s using %s -*-\n", white("teller"), green(projectName), green(loadedFrom))
}

func (p *Porcelain) PrintEntries(entries []core.EnvEntry) {
	var buf bytes.Buffer
	yellow := color.New(color.FgYellow).SprintFunc()
	gray := color.New(color.FgHiBlack).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	for i := range entries {
		v := entries[i]
		ep := ellipsis.Shorten(v.ResolvedPath, 30) //nolint: gomnd
		if !v.IsFound {
			fmt.Fprintf(&buf, "[%s %s %s] %s\n", yellow(v.ProviderName), gray(ep), red("missing"), green(v.Key))
		} else {
			fmt.Fprintf(&buf, "[%s %s] %s %s %s\n", yellow(v.ProviderName), gray(ep), green(v.Key), gray("="), maskedValue(v.Value))
		}
	}

	out := buf.String()

	fmt.Fprint(p.Out, out)
}
func maskedValue(v string) string {
	return fmt.Sprintf("%s*****", v[:int(math.Min(float64(len(v)), 2))]) //nolint: gomnd
}

func (p *Porcelain) PrintMatches(matches []core.Match) {
	yellow := color.New(color.FgYellow).SprintFunc()
	gray := color.New(color.FgHiBlack).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Path < matches[j].Path
	})

	for _, m := range matches { //nolint
		sevcolor := white
		if m.Entry.Severity == core.High {
			sevcolor = red
		}
		if m.Entry.Severity == core.Medium {
			sevcolor = yellow
		}
		fmt.Printf("[%s] %s (%s,%s): found match for %s/%s (%s)\n", sevcolor(m.Entry.Severity), green(m.Path), yellow(m.LineNumber), yellow(m.MatchIndex), gray(m.Entry.ProviderName), red(m.Entry.Key), gray(maskedValue(m.Entry.Value)))
	}
}

func (p *Porcelain) PrintMatchSummary(findings []core.Match, entries []core.EnvEntry, elapsed time.Duration) {
	yellow := color.New(color.FgYellow).SprintFunc()
	goodbad := color.New(color.FgGreen).SprintFunc()
	if len(findings) > 0 {
		goodbad = color.New(color.FgRed).SprintFunc()
	}

	fmt.Printf("Scanning for %v entries: found %v matches in %v\n", yellow(len(entries)), goodbad(len(findings)), goodbad(elapsed))
}

func (p *Porcelain) PrintDrift(drifts []core.DriftedEntry) {
	green := color.New(color.FgGreen).SprintFunc()
	gray := color.New(color.FgHiBlack).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	if len(drifts) > 0 {
		fmt.Fprintf(p.Out, "Drifts detected: %v\n\n", len(drifts))

		for i := range drifts {
			d := drifts[i]
			if d.Diff == "changed" {
				fmt.Fprintf(p.Out, "%v [%v] %v %v %v != %v %v %v\n", d.Diff, d.Source.Source, green(d.Source.ProviderName), green(d.Source.Key), gray(maskedValue(d.Source.Value)), red(d.Target.ProviderName), red(d.Target.Key), gray(maskedValue(d.Target.Value)))
			} else {
				fmt.Fprintf(p.Out, "%v [%v] %v %v %v ??\n", d.Diff, d.Source.Source, green(d.Source.ProviderName), green(d.Source.Key), gray(maskedValue(d.Source.Value)))
			}
		}
	}

}

func (p *Porcelain) DidPutKVP(kvp core.KeyPath, pname string, sync bool) {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	gray := color.New(color.FgHiBlack).SprintFunc()
	if sync {
		fmt.Fprintf(p.Out, "Synced %v (%v): OK.\n", green(pname), gray(kvp.Path))
	} else {
		fmt.Fprintf(p.Out, "Put %v (%v) in %v: OK.\n", yellow(kvp.Env), gray(kvp.Path), green(pname))
	}
}

func (p *Porcelain) NoPutKVP(k, pname string) {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	fmt.Fprintf(p.Out, "Put %v in %v: no such key '%v' in mapping\n", yellow(k), green(pname), yellow(k))
}

func (p *Porcelain) DidDeleteKP(kp core.KeyPath, pname string) {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	gray := color.New(color.FgHiBlack).SprintFunc()
	fmt.Fprintf(p.Out, "Delete %v (%v) in %v: OK.\n", yellow(kp.Env), gray(kp.Path), green(pname))
}

func (p *Porcelain) NoDeleteKP(k, pname string) {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	fmt.Fprintf(p.Out, "Delete %v in %v: no such key '%v' in mapping\n", yellow(k), green(pname), yellow(k))
}

func (p *Porcelain) DidDeleteP(path, pname string) {
	green := color.New(color.FgGreen).SprintFunc()
	gray := color.New(color.FgHiBlack).SprintFunc()
	fmt.Fprintf(p.Out, "Delete mapping in path %v in %v: OK.\n", gray(path), green(pname))
}
