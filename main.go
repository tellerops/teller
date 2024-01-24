package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/alecthomas/kong"
	"github.com/spectralops/teller/pkg"
	"github.com/spectralops/teller/pkg/logging"
	"github.com/spectralops/teller/pkg/providers"
	"github.com/spectralops/teller/pkg/utils"
)

var CLI struct {
	Config   string `short:"c" help:"Path to teller YAML file"`
	LogLevel string `short:"l"  help:"Application log level"`

	Run struct {
		Redact bool     `optional name:"redact" help:"Redact output of the child process"`
		Cmd    []string `arg name:"cmd" help:"Command to execute"`
	} `cmd help:"Run a command"`

	Version struct {
	} `cmd aliases:"v" help:"Teller version"`

	New struct {
	} `cmd help:"Create a new teller configuration file"`

	Show struct {
	} `cmd help:"Print in a human friendly, secure format"`

	Providers struct {
		Path string `optional name:"path" help:"Path for saving providers JSON file"`
	} `cmd help:"Export providers metadata to a local JSON file" hidden: ""`

	Yaml struct {
	} `cmd help:"Print values in a YAML format (suitable for GCloud)"`

	JSON struct {
	} `cmd help:"Print values in a JSON format"`

	Sh struct {
	} `cmd help:"Print ready to be eval'd exports for your shell"`

	Env struct {
	} `cmd help:"Print in a .env format for Docker and others"`

	Template struct {
		TemplatePath string `arg name:"template_path" help:"Path to the template source (Go template format)"`
		Out          string `arg name:"out" help:"Output file"`
	} `cmd help:"Inject vars from a template by given source path (single file or folder)"`

	Redact struct {
		In  string `optional name:"in" help:"Input file"`
		Out string `optional name:"out" help:"Output file"`
	} `cmd help:"Redacts secrets from a process output"`

	Scan struct {
		Path   string `arg optional name:"path" help:"Scan root, default: '.'"`
		Silent bool   `optional name:"silent" help:"No text, just exit code"`
	} `cmd help:"Scans your codebase for sensitive keys"`

	GraphDrift struct {
		Providers []string `arg optional name:"providers" help:"A list of providers to check for drift"`
	} `cmd help:"Detect secret and value drift between providers"`

	Put struct {
		Kvs       map[string]string `arg name:"kvs" help:"A list of key/value pairs, where key is from your tellerfile mapping"`
		Providers []string          `name:"providers" help:"A list of providers to put the new value into"`
		Sync      bool              `optional name:"sync" help:"Sync all given k/vs to the env_sync key"`
		Path      string            `optional name:"path" help:"Take literal path and not from config"`
	} `cmd help:"Put a new value"`

	Copy struct {
		From string   `name:"from" help:"A provider name to sync from"`
		To   []string `name:"to" help:"A list of provider names to copy values from the source provider to"`
		Sync bool     `optional name:"sync" help:"Sync all given k/vs to the env_sync key"`
	} `cmd help:"Sync data from a source provider directly to multiple target providers"`

	MirrorDrift struct {
		Source string `name:"source" help:"A source to check drift against"`
		Target string `name:"target" help:"A target to check against source"`
	} `cmd help:"Check same-key (mirror) value drift between source and target"`

	Delete struct {
		Keys      []string `arg optional name:"keys" help:"A list of keys, where key is from your tellerfile mapping"`
		Providers []string `name:"providers" help:"A list of providers to delete the key from"`
		Path      string   `optional name:"path" help:"Take literal path and not from config"`
		AllKeys   bool     `optional name:"all-keys" help:"Deletes all keys for a given path. Applicable only when used together with the 'path' flag"`
	} `cmd help:"Delete a secret"`
}

var (
	version         = "dev"
	commit          = "none"
	date            = "unknown"
	defaultLogLevel = "error"
)

//nolint
func main() {
	ctx := kong.Parse(&CLI)

	logger := logging.GetRoot()
	if CLI.LogLevel != "" {
		defaultLogLevel = CLI.LogLevel
	}
	logger.SetLevel(defaultLogLevel)

	// below commands don't require a tellerfile
	//nolint
	switch ctx.Command() {
	case "version":
		fmt.Printf("Teller %v\n", version)
		fmt.Printf("Revision %v, date: %v\n", commit, date)
		os.Exit(0)
	case "providers":
		providersMetaList := providers.GetAllProvidersMeta()
		providersMetaJSON, err := providers.GenerateProvidersMetaJSON(version, providersMetaList)
		if err != nil {
			logger.WithError(err).Fatal("could not get providers meta, %s", err)
		}

		saveErr := utils.WriteFileInPath("providers-meta.json", CLI.Providers.Path, []byte(providersMetaJSON))
		if saveErr != nil {
			logger.WithError(err).Fatal("could not save providers meta to a local file, %s", saveErr)
		}
		fmt.Printf("Providers meta has been exported successfully\n")

		os.Exit(0)
	}

	//
	// load or create new file
	//
	const (
		defaultTellerFile = ".teller.yml"
		// Alternative default teller file, it uses official YAML extension
		// See https://github.com/tellerops/teller/issues/162
		secondDefaultTellerFile = ".teller.yaml"
	)

	telleryml := defaultTellerFile
	if CLI.Config != "" {
		telleryml = CLI.Config
	}

	if ctx.Command() == "new" {
		teller := pkg.Teller{
			Porcelain: &pkg.Porcelain{Out: os.Stderr},
			Logger:    logger,
		}
		if _, err := os.Stat(telleryml); err == nil && !teller.Porcelain.AskForConfirmation(fmt.Sprintf("The file %s already exists. Do you want to override the configuration with new settings?", telleryml)) {
			os.Exit(0)
		}

		err := teller.SetupNewProject(telleryml)
		if err != nil {
			logger.WithError(err).Fatal("could not create configuration")
		}
		os.Exit(0)
	}

	tlrfile, err := pkg.NewTellerFile(telleryml)
	if isDefaultFilePathErr(CLI.Config, err) {
		tlrfile, err = pkg.NewTellerFile(secondDefaultTellerFile)
	}
	if err != nil {
		logger.WithError(err).WithField("file", telleryml).Fatal("could not read file")
	}

	teller := pkg.NewTeller(tlrfile, CLI.Run.Cmd, CLI.Run.Redact, logger)

	// below commands don't require collecting
	//nolint
	switch ctx.Command() {
	case "put <kvs>":
		err := teller.Put(CLI.Put.Kvs, CLI.Put.Providers, CLI.Put.Sync, CLI.Put.Path)
		if err != nil {
			logger.WithError(err).Fatal("put command field")
		}
		os.Exit(0)
	case "copy":
		err := teller.Sync(CLI.Copy.From, CLI.Copy.To, CLI.Copy.Sync)
		if err != nil {
			logger.WithError(err).WithFields(map[string]interface{}{
				"from":      CLI.Copy.From,
				"to":        CLI.Copy.To,
				"sync_flag": CLI.Copy.Sync,
			}).Fatal("could not copy data between providers")
		}
		os.Exit(0)
	case "mirror-drift":
		drifts, err := teller.MirrorDrift(CLI.MirrorDrift.Source, CLI.MirrorDrift.Target)
		if err != nil {
			logger.WithError(err).Fatal("mirror-drift command field")
		}
		if len(drifts) > 0 {
			teller.Porcelain.PrintDrift(drifts)
			os.Exit(1)
		}
		os.Exit(0)
	case "delete":
		err := teller.Delete(CLI.Delete.Keys, CLI.Delete.Providers, CLI.Delete.Path, CLI.Delete.AllKeys)
		if err != nil {
			logger.WithError(err).Fatal("could not delete key")
		}
		os.Exit(0)
	case "delete <keys>":
		err := teller.Delete(CLI.Delete.Keys, CLI.Delete.Providers, CLI.Delete.Path, CLI.Delete.AllKeys)
		if err != nil {
			logger.WithError(err).Fatal("could not delete keys")
		}
		os.Exit(0)
	}
	// collecting

	err = teller.Collect()
	if err != nil {
		logger.WithError(err).Fatal("could not load all variables from the given existing providers")
	}

	// all of the below require a tellerfile
	switch ctx.Command() {
	case "run <cmd>":
		if len(CLI.Run.Cmd) < 1 {
			logger.Fatal("Error: No command given")
		}
		teller.Exec()

	case "graph-drift <providers>":
		fallthrough
	case "graph-drift":
		drifts := teller.Drift(CLI.GraphDrift.Providers)
		if len(drifts) > 0 {
			teller.Porcelain.PrintDrift(drifts)
			os.Exit(1)
		}

	case "redact":
		// redact (stdin)
		// redact --in FILE --out FOUT
		// redact --in FILE (stdout)
		var fin io.Reader = os.Stdin
		var fout io.Writer = os.Stdout

		if CLI.Redact.In != "" {
			f, err := os.Open(CLI.Redact.In)
			if err != nil {
				logger.WithError(err).Fatal("could not open file")
			}
			fin = f
		}

		if CLI.Redact.Out != "" {
			f, err := os.Create(CLI.Redact.Out)
			if err != nil {
				logger.WithError(err).Fatal("could not create file")
			}

			fout = f
		}

		if err := teller.RedactLines(fin, fout); err != nil {
			logger.WithError(err).Fatal("could not redact lines")
		}

	case "sh":
		fmt.Print(teller.ExportEnv())

	case "env":
		fmt.Print(teller.ExportDotenv())

	case "yaml":
		out, err := teller.ExportYAML()
		if err != nil {
			logger.WithError(err).Fatal("could not export to YAML")
		}
		fmt.Print(out)

	case "json":
		out, err := teller.ExportJSON()
		if err != nil {
			logger.WithError(err).Fatal("could not export to JSON")
		}
		fmt.Print(out)

	case "show":
		teller.PrintEnvKeys()

	case "scan":
		findings, err := teller.Scan(CLI.Scan.Path, CLI.Scan.Silent)

		if err != nil {
			logger.WithError(err).WithField("path", CLI.Scan.Path).Fatal("scan error")
		}
		num := len(findings)
		if num > 0 {
			os.Exit(1)
		}

	case "template <template_path> <out>":
		err := teller.Template(CLI.Template.TemplatePath, CLI.Template.Out)
		if err != nil {
			logger.WithError(err).WithFields(map[string]interface{}{
				"template_path":   CLI.Template.TemplatePath,
				"template_output": CLI.Template.Out,
			}).Fatal("could not populate template")
		}

	default:
		println(ctx.Command())
		teller.PrintEnvKeys()
	}
}

func isDefaultFilePathErr(config string, err error) bool {
	// Ignore if explicitly set to '.teller.yml'.
	if config != "" {
		return false
	}
	return errors.Is(err, fs.ErrNotExist)
}
