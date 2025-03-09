package main

import (
	"encoding/json"
	"fmt"
	"os"

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

	if err := cfg.ParseCLI(parser, os.Args[1:]); err != nil {
		fatal(err)
	}

	/////////////////////////////
	// Highjack execution flow //
	// That is to handle the flags that trigger an action and exit immediately.

	if cfg.List {
		for name := range prompts.Map {
			fmt.Println(name)
		}
		os.Exit(0)
	}

	exec(cfg)
}

func conf() (*config.Jenai, *flag.Parser) {
	cfg := config.Jenai{}
	parser := cfg.RegisterCLI()
	flag.WithHelp(os.Args[0], "PROMPT ...ARGS")(parser)

	return &cfg, parser
}

func fatal(err error) {
	fmt.Println("Error:", err)
	os.Exit(1)
}

func exec(cfg *config.Jenai) {
	prompt, err := cfg.BuildPrompt()
	if err != nil {
		fatal(err)
	}

	if cfg.DryRun {
		fmt.Println(prompt)
		os.Exit(0)
	}

	spec, err := modelSpec(cfg)
	if err != nil {
		fatal(err)
	}
	fmt.Println("Model:", pretty(spec))
	fmt.Println("Config:", pretty(cfg))
}

func modelSpec(cfg *config.Jenai) (models.Spec, error) {
	spec, exists := models.ModelSpecs[cfg.Model]
	if !exists {
		return spec, fmt.Errorf("unknown model: %s", cfg.Model)
	}

	return spec, nil
}

func pretty(data any) string {
	pretty, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fatal(err)
	}

	return string(pretty)
}
