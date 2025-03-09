package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mooss/bagend/go/flag"
	"github.com/mooss/jen/go/ai/config"
	"github.com/mooss/jen/go/ai/models"
	"github.com/mooss/jen/go/ai/prompts"
)

func main() {
	cfg, parser := conf()

	if len(os.Args) == 1 { // No arguments, print help.
		fmt.Print(parser.Help())
		os.Exit(0)
	}

	dumpConfig := false
	parser.Bool("dump-config", &dumpConfig, "print the config and exit")

	noerr0(cfg.ParseCLI(parser, os.Args[1:]))

	/////////////////////////////
	// Highjack execution flow //
	// That is to handle the flags that trigger an action and exit immediately.

	library := noerr(prompts.Embedded())

	if cfg.List {
		for name := range library.Prompts {
			fmt.Println(name)
		}
		os.Exit(0)
	}

	if dumpConfig {
		spec := noerr(modelSpec(cfg))
		fmt.Println("Model:", pretty(spec))
		fmt.Println("Config:", pretty(cfg))
		os.Exit(0)
	}

	///////////////
	// Execution //

	run(cfg, library)
}

//////////////////////////
// High-level utilities //

func conf() (*config.Jenai, *flag.Parser) {
	cfg := config.Jenai{}
	parser := cfg.RegisterCLI()
	flag.WithHelp(os.Args[0], "PROMPT ...ARGS")(parser)

	return &cfg, parser
}

func run(cfg *config.Jenai, lib prompts.Library) {
	noerr0(cfg.Validate())
	prompt := noerr(cfg.BuildPrompt(lib))

	if cfg.DryRun {
		fmt.Println(prompt)
		os.Exit(0)
	}

	spec := noerr(modelSpec(cfg))

	// Send the prompt through aichat.
	cmd := exec.Command("aichat", "--model", spec.Aichat())
	cmd.Stdin = strings.NewReader(prompt)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	noerr0(cmd.Run())
}

func modelSpec(cfg *config.Jenai) (models.Spec, error) {
	spec, exists := models.ModelSpecs[cfg.Model]
	if !exists {
		return spec, fmt.Errorf("unknown model: %s", cfg.Model)
	}

	return spec, nil
}

////////////////
// Primitives //

func fatal(err error) {
	fmt.Println("Error:", err)
	os.Exit(1)
}

func noerr[T any](res T, err error) T {
	if err != nil {
		fatal(err)
	}
	return res
}

func noerr0(err error) { noerr(0, err) }

func pretty(data any) string {
	return string(noerr(json.MarshalIndent(data, "", "  ")))
}
