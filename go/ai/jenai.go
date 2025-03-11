package main

import (
	"encoding/json"
	"fmt"
	"io"
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
	session := noerr(cfg.Session())
	aichat := func(stdin io.Reader) {
		cmd := exec.Command("aichat",
			"--model", spec.Aichat(), "--session", session.Name, "--save-session")
		cmd.Env = append(cmd.Env, "AICHAT_COMPRESS_THRESHOLD=10000",
			"AICHAT_SESSIONS_DIR="+session.Dir)
		cmd.Stdin = stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		noerr0(cmd.Run())
	}

	if prompt != "" {
		aichat(strings.NewReader(prompt))
		if cfg.TeeFile != "" {
			if err := tee(cfg.TeeFile, session); err != nil {
				fmt.Fprintf(os.Stderr,
					"can't tee to %s (will proceed nonetheless): %s", cfg.TeeFile, err)
			}
		}
	}

	// If interactive mode or session specified without prompt, start interactive session.
	if cfg.Interactive || (session.Requested && prompt == "") {
		aichat(os.Stdin)
	}
}

func tee(teefile string, session config.SessionMetadata) error {
	conv, err := session.Load()
	if err != nil {
		return err
	}

	if len(conv.Messages) == 0 {
		return fmt.Errorf("no message in session")
	}

	last := conv.Messages[len(conv.Messages)-1]
	return os.WriteFile(teefile, []byte(last.Content), 0644)
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
