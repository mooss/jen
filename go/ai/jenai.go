package main

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/mooss/bagend/go/flag"
	"github.com/mooss/jen/go/ai/config"
	"github.com/mooss/jen/go/ai/llm"
	"github.com/mooss/jen/go/ai/models"
	"github.com/mooss/jen/go/ai/prompts"
	"gopkg.in/yaml.v3"
)

func main() {
	if err := ensureConfig(); err != nil {
		fatal(err)
	}
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
		fmt.Println(pretty(map[string]any{
			"model":  spec,
			"config": cfg,
		}))
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
	if err != nil {
		return err
	}
	if alreadyExists {
		return ensureConfigYaml()
	}

	dir, err := config.ConfigDir()
	if err != nil {
		return err
	}
	fmt.Println("Initialized config file in", dir)

	err = writeConfigFile("models.yaml", models.EmbeddedBytes, 0644)
	if err != nil {
		return err
	}

	err = writeConfigFile("prompts.yaml", prompts.EmbeddedBytes, 0644)
	if err != nil {
		return err
	}

	return ensureConfigYaml()
}

func ensureConfigYaml() error {
	dir, err := config.ConfigDir()
	if err != nil {
		return err
	}
	cfgPath := filepath.Join(dir, "config.yaml")

	ferr, exists := fileExists(cfgPath)
	if ferr != nil {
		return ferr
	}
	if !exists {
		return writeConfigFile("config.yaml", []byte{}, 0600)
	}

	info, err := os.Stat(cfgPath)
	if err != nil {
		return err
	}

	if info.Mode().Perm() != 0600 {
		if err := os.Chmod(cfgPath, 0600); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "warning: permissions of %q changed to 0600 to make it non world-readable to protect secrets within\n", cfgPath)
	}

	return nil
}

// ensureConfigDir tests if the directory already exists and creates it if it doesn't.
// Returns true when the directory already exists.
func ensureConfigDir() (error, bool) {
	dir, err := config.ConfigDir()
	if err != nil {
		return err, false
	}

	err, exists := fileExists(dir)
	if err != nil {
		return err, false
	}

	if exists {
		info, statErr := os.Stat(dir)
		if statErr != nil {
			return statErr, true
		}

		if info.Mode().Perm() != 0700 {
			if chmodErr := os.Chmod(dir, 0700); chmodErr != nil {
				return chmodErr, true
			}
			fmt.Fprintf(
				os.Stderr,
				"warning: permissions of config directory %q changed to 0700 to make it non world-readable to protect secrets within\n",
				dir,
			)
		}

		return err, true
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
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

func writeConfigFile(path string, data []byte, perm os.FileMode) error {
	dir, err := config.ConfigDir()
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, path), data, perm); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}
	return nil
}

func conf() (*config.Jenai, *flag.Parser) {
	cfg := config.Jenai{}
	dir, err := config.ConfigDir()
	if err != nil {
		fatal(err)
	}
	noerr0(cfg.LoadFile(filepath.Join(dir, "config.yaml")))
	parser := cfg.RegisterCLI()
	flag.WithHelp(os.Args[0], "PROMPT ...ARGS")(parser)

	return &cfg, parser
}

func run(cfg *config.Jenai, lib prompts.Library) {
	prompt := noerr(cfg.BuildPrompt(lib))

	if cfg.DryRun {
		fmt.Println(prompt)
		os.Exit(0)
	}

	if prompt.Empty() {
		fatal(errors.New("the prompt is empty"))
	}

	spec := noerr(modelSpec(cfg))
	apiKey := noerr(cfg.ApiKeyFor(spec.Provider))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	content, full := streamAndDisplay(ctx, cfg, spec, apiKey, prompt.String())
	writeTees(cfg, prompt, full, content, spec)
}

func streamAndDisplay(ctx context.Context, cfg *config.Jenai, spec models.Spec, apiKey, promptStr string) (content, full string) {
	var contentBuilder, reasoningBuilder strings.Builder
	var reasoningStarted, reasoningClosed bool

	onReasoning := func(s string) {
		reasoningBuilder.WriteString(s)
		if cfg.Thoughtless {
			return
		}
		if !reasoningStarted {
			fmt.Fprint(os.Stdout, "<think>\n")
			reasoningStarted = true
		}
		fmt.Fprint(os.Stdout, s)
	}

	onContent := func(s string) {
		contentBuilder.WriteString(s)
		if reasoningStarted && !reasoningClosed && !cfg.Thoughtless {
			fmt.Fprint(os.Stdout, "\n</think>\n")
			reasoningClosed = true
		}
		fmt.Fprint(os.Stdout, s)
	}

	chunks, errs, err := llm.Stream(ctx, spec, apiKey, promptStr)
	noerr0(err)
	for chunk := range chunks {
		if len(chunk.Choices) == 0 {
			if chunk.Usage != nil {
				continue // Expected final usage-only chunk from `IncludeUsage: true`.
			}
			fmt.Fprintln(os.Stderr, "warning: chunk with empty choices and no usage")
			continue
		}

		c := chunk.Choices[0]
		if c.Delta.Reasoning != nil && c.Delta.Reasoning.Content != "" {
			onReasoning(c.Delta.Reasoning.Content)
		}
		if c.Delta.Content != "" {
			onContent(c.Delta.Content)
		}

	}
	noerr0(<-errs)

	if reasoningStarted && !reasoningClosed {
		noerr0(errors.New("reasoning never closed"))
	}

	content = contentBuilder.String()
	reasoning := reasoningBuilder.String()

	full = content
	if reasoning != "" {
		full = "<think>\n" + reasoning + "\n</think>\n" + content
	}

	if content != "" && !strings.HasSuffix(content, "\n") {
		fmt.Fprintln(os.Stdout)
	}

	content = withTrailingNewline(content)
	full = withTrailingNewline(full)

	return content, full
}

func withTrailingNewline(s string) string {
	if s != "" && !strings.HasSuffix(s, "\n") {
		return s + "\n"
	}
	return s
}

func writeTees(cfg *config.Jenai, prompt config.Prompt, full, content string, spec models.Spec) {
	if cfg.TeeFile != "" {
		if err := tee(cfg, prompt, full, spec); err != nil {
			fmt.Fprintf(os.Stderr,
				"can't tee to %s (will proceed nonetheless): %s", cfg.TeeFile, err)
		}
	}

	if cfg.ThoughtlessTeeFile != "" {
		if err := thoughtlessTee(cfg, content); err != nil {
			fmt.Fprintf(os.Stderr,
				"can't tee to %s (will proceed nonetheless): %s", cfg.ThoughtlessTeeFile, err)
		}
	}
}

func tee(cfg *config.Jenai, prompt config.Prompt, answer string, spec models.Spec) error {
	metadata, err := structToMap(cfg)
	if err != nil {
		return err
	}

	metadata["date"] = time.Now().Format("2006-01-02")
	metadata["model"] = spec.QualifiedName()
	metadata["prompt"] = prompt.Static()
	delete(metadata, "positional") // Noise that should be present in `prompt` anyway.

	yamlmeta, err := yaml.Marshal(metadata)
	if err != nil {
		return err
	}

	content := strings.Join([]string{"---", string(yamlmeta) + "---\n", answer}, "\n")
	return os.WriteFile(cfg.TeeFile, []byte(content), 0644)
}

func thoughtlessTee(cfg *config.Jenai, answer string) error {
	return os.WriteFile(cfg.ThoughtlessTeeFile, []byte(answer), 0644)
}

func modelSpec(cfg *config.Jenai) (models.Spec, error) {
	fromZoo := func() (models.Spec, error) { return models.Get(cfg.Model) }

	// Qualified name in provider:author/model formet.
	provider, rest, found := strings.Cut(cfg.Model, ":")
	if !found {
		return fromZoo()
	}

	author, model, found := strings.Cut(rest, "/")
	if !found {
		return fromZoo()
	}

	return models.Spec{Provider: provider, Author: author, Model: model}, nil
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
	return string(noerr(yaml.Marshal(data)))
}

func structToMap(data any) (map[string]any, error) {
	yamlbytes, err := yaml.Marshal(data)
	if err != nil {
		return nil, err
	}

	result := make(map[string]any)
	err = yaml.Unmarshal(yamlbytes, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
