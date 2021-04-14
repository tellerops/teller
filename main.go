package main

import (
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/spectralops/teller/pkg"
)

var CLI struct {
	Config string `short:"c" help:"Path to teller.yml"`

	Run struct {
		Redact bool     `optional name:"redact" help:"Redact output of the child process"`
		Cmd    []string `arg name:"cmd" help:"Command to execute"`
	} `cmd help:"Run a command"`

	Version struct {
	} `cmd short:"v" help:"Teller version"`

	New struct {
	} `cmd help:"Create a new teller configuration file"`

	Show struct {
	} `cmd help:"Print in a human friendly, secure format"`

	Sh struct {
	} `cmd help:"Print ready to be eval'd exports for your shell"`

	Env struct {
	} `cmd help:"Print in a .env format for Docker and others"`

	Template struct {
		TemplateFile string `arg name:"template_file" help:"Input template file (Go template format)"`
		OutFile      string `arg name:"out_file" help:"Output file"`
	} `cmd help:"Inject vars into a template file"`

	Redact struct {
		In  string `optional name:"in" help:"Input file"`
		Out string `optional name:"out" help:"Output file"`
	} `cmd help:"Scans your codebase for sensitive keys"`

	Scan struct {
		Path   string `arg optional name:"path" help:"Scan root, default: '.'"`
		Silent bool   `optional name:"silent" help:"No text, just exit code"`
	} `cmd help:"Scans your codebase for sensitive keys"`
}

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

//nolint
func main() {
	ctx := kong.Parse(&CLI)

	// below commands don't require a tellerfile
	//nolint
	switch ctx.Command() {
	case "version":
		fmt.Printf("Teller %v\n", version)
		fmt.Printf("Revision %v, date: %v\n", commit, date)
		os.Exit(0)
	}

	//
	// load or create new file
	//
	telleryml := ".teller.yml"
	if CLI.Config != "" {
		telleryml = CLI.Config
	}

	if _, err := os.Stat(telleryml); os.IsNotExist(err) {
		teller := pkg.Teller{
			Porcelain: &pkg.Porcelain{Out: os.Stderr},
		}
		err = teller.SetupNewProject(telleryml)
		if err != nil {
			fmt.Printf("Error: %v", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	tlrfile, err := pkg.NewTellerFile(telleryml)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	teller := pkg.NewTeller(tlrfile, CLI.Run.Cmd, CLI.Run.Redact)
	err = teller.Collect()
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	// all of the below require a tellerfile
	switch ctx.Command() {
	case "run <cmd>":
		if len(CLI.Run.Cmd) < 1 {
			fmt.Println("Error: No command given")
			os.Exit(1)
		}
		teller.Exec()

	case "redact":
		// redact (stdin)
		// redact --in FILE --out FOUT
		// redact --in FILE (stdout)
		var fin io.Reader = os.Stdin
		var fout io.Writer = os.Stdout

		if CLI.Redact.In != "" {
			f, err := os.Open(CLI.Redact.In)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			fin = f
		}

		if CLI.Redact.Out != "" {
			f, err := os.Create(CLI.Redact.Out)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			fout = f
		}

		if err := teller.RedactLines(fin, fout); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	case "sh":
		fmt.Print(teller.ExportEnv())

	case "env":
		fmt.Print(teller.ExportDotenv())

	case "show":
		teller.PrintEnvKeys()

	case "scan":
		findings, err := teller.Scan(CLI.Scan.Path, CLI.Scan.Silent)

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		num := len(findings)
		if num > 0 {
			os.Exit(1)
		}

	case "template <template_file> <out_file>":
		err := teller.TemplateFile(CLI.Template.TemplateFile, CLI.Template.OutFile)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	default:
		teller.PrintEnvKeys()

	}
}
