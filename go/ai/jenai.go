package main

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/mooss/bagend/go/flag"
	"github.com/mooss/jen/go/ai/config"
	"github.com/mooss/jen/go/ai/models"
	"github.com/mooss/jen/go/ai/prompts"
	"gopkg.in/yaml.v3"
)

var home = os.Getenv("HOME")
var configDir = filepath.Join(home, ".config", "jenai")

func main() {
	ensureConfig()
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

	if cfg.ListModels { // Align and print in sorted order.
		specs, err := models.ModelSpecs()
		if err != nil {
			fatal(err)
		}
		longest := 0
		for _, spec := range specs {
			if len(spec.ShortName) > longest {
				longest = len(spec.ShortName)
			}
		}

		format := fmt.Sprintf("%%-%ds  (%%s/%%s)\n", longest)
		for _, short := range slices.Sorted(maps.Keys(specs)) {
			spec := specs[short]
			fmt.Printf(format, spec.ShortName, spec.Provider, spec.Author)
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

func ensureConfig() error {
	err, alreadyExists := ensureConfigDir()
	if err != nil || alreadyExists {
		return err
	}

	fmt.Println("Initialized config file in", configDir)

	err = writeConfigFile("models.yaml", models.EmbeddedBytes)
	if err != nil {
		return err
	}

	err = writeConfigFile("prompts.yaml", prompts.EmbeddedBytes)

	return err
}

// ensureConfigDir tests if the directory already exists and creates it if it doesn't.
// Returns true when the directory already exists.
func ensureConfigDir() (error, bool) {
	err, exists := fileExists(home)
	if err != nil || !exists {
		return fmt.Errorf("$HOME is invalid (%s): %w", home, err), false
	}

	err, exists = fileExists(configDir)
	if err != nil || exists {
		return err, true
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err), false
	}

	return nil, false
}

func fileExists(filename string) (error, bool) {
	_, err := os.Stat(filename)
	if err == nil {
		return nil, true
	}

	if os.IsNotExist(err) {
		err = nil
	}

	return err, false
}

func writeConfigFile(path string, data []byte) error {
	if err := os.WriteFile(filepath.Join(configDir, path), data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}
	return nil
}

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

	if !prompt.Empty() {
		aichat(strings.NewReader(prompt.String()))
		if cfg.TeeFile != "" {
			if err := tee(cfg.TeeFile, session, prompt); err != nil {
				fmt.Fprintf(os.Stderr,
					"can't tee to %s (will proceed nonetheless): %s", cfg.TeeFile, err)
			}
		}
	}

	// Handle interactive mode.
	if cfg.Interactive || (session.Requested && prompt.Empty()) {
		aichat(os.Stdin)
	}
}

func tee(teefile string, session config.SessionMetadata, prompt config.Prompt) error {
	conv, err := session.Load()
	if err != nil {
		return err
	}

	if len(conv.Messages) == 0 {
		return fmt.Errorf("no message in session")
	}

	last := conv.Messages[len(conv.Messages)-1]
	metadata, err := yaml.Marshal(map[string]any{
		"date":    time.Now().Format("2006-01-02"),
		"model":   conv.Model,
		"prompt":  prompt.Static(),
		"context": prompt.Paths,
	})
	if err != nil {
		return err
	}

	content := strings.Join([]string{"---", string(metadata) + "...\n", last.Content}, "\n")
	return os.WriteFile(teefile, []byte(content), 0644)
}

func modelSpec(cfg *config.Jenai) (models.Spec, error) {
	// Model specified directly in aichat's provider:author/model formet.
	if parts := strings.Split(cfg.Model, ":"); len(parts) == 2 {
		authorModel := strings.Split(parts[1], "/")
		if len(authorModel) == 2 {
			return models.Spec{
				Provider: parts[0],
				Author:   authorModel[0],
				Model:    authorModel[1],
			}, nil
		}
	}

	// Short name lookup.
	return models.Get(cfg.Model)
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
