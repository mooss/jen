// Package config defines jenai's configuration.
package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mooss/bagend/go/flag"
	"github.com/mooss/jen/go/ai/prompts"
	"github.com/mooss/jen/go/utils"
)

func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", "jenai"), nil
}

type Jenai struct {
	ApiKeys            map[string]string `yaml:"-" json:"-"`
	Context            Context           `yaml:"context" json:"context"`
	DryRun             bool              `yaml:"dryrun,omitempty" json:"dryrun,omitempty"`
	List               bool              `yaml:"list,omitempty" json:"list,omitempty"`
	ListModels         bool              `yaml:"listmodels,omitempty" json:"listmodels,omitempty"`
	Model              string            `yaml:"model,omitempty" json:"model,omitempty"`
	OneShot            bool              `yaml:"oneshot,omitempty" json:"oneshot,omitempty"`
	Paste              bool              `yaml:"paste,omitempty" json:"paste,omitempty"`
	Positional         []string          `yaml:"positional,omitempty" json:"positional,omitempty"`
	Thoughtless        bool              `yaml:"thoughtless,omitempty" json:"thoughtless,omitempty"`
	ThoughtlessTeeFile string            `yaml:"thoughtlessteefile,omitempty" json:"thoughtlessteefile,omitempty"`
	TeeFile            string            `yaml:"teefile,omitempty" json:"teefile,omitempty"`
}

/////////////////////////////////
// Construction and validation //

func (conf *Jenai) RegisterCLI() *flag.Parser {
	parser := flag.NewParser()
	parser.Bool("context-above", &conf.Context.Above,
		"Put context files and dir above instructions")
	parser.StringSlice("dir", &conf.Context.Dirs, "Include all files in directory as context")
	parser.Bool("dry-run", &conf.DryRun, "Print interpolated prompt without sending to LLM").
		Alias("n")
	parser.StringSlice("file", &conf.Context.Files, "Include specific file(s) as context")
	parser.Bool("list", &conf.List, "list all available prompts").
		Alias("l")
	parser.Bool("list-models", &conf.ListModels, "list all available models").
		Alias("lm")
	parser.Bool("linum", &conf.Context.LineNumbers, "Print files with line numbers")
	parser.String("model", &conf.Model, "Model name (short name from --lm or provider:author/model)").
		Alias("m").Default("ds41-flash")
	parser.Bool("oneshot", &conf.OneShot, "Use positional arguments as the prompt").
		Alias("o")
	parser.Bool("paste", &conf.Paste, "Use clipboard content as prompt")
	parser.String("tee", &conf.TeeFile, "Output first answer to both stdout and FILE (overwritten)")
	parser.Bool("thoughtless", &conf.Thoughtless,
		"Hide model reasoning").
		Alias("t")
	parser.String("thoughtless-tee", &conf.ThoughtlessTeeFile,
		"Output first answer to FILE without the reasoning block (overwritten)").
		Alias("tt")

	return parser
}

// ParseCLI fills the fields from CLI arguments.
func (conf *Jenai) ParseCLI(parser *flag.Parser, args []string) error {
	if err := parser.Parse(args); err != nil {
		return err
	}

	conf.Positional = parser.Positional
	return nil
}

func (conf *Jenai) LoadFile(path string) error {
	loaded, err := utils.FromYAMLFile[struct {
		ApiKeys map[string]string `yaml:"api-keys" json:"api-keys"`
	}](path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to load config file %q: %w", path, err)
	}

	if loaded.ApiKeys != nil {
		conf.ApiKeys = loaded.ApiKeys
	}
	return nil
}

func (conf *Jenai) ApiKeyFor(provider string) (string, error) {
	if key := conf.ApiKeys[provider]; key != "" {
		return key, nil
	}
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return "", fmt.Errorf(
		"missing api key for %q: put it in %s as:\napi-keys:\n  %s: API_KEY_HERE",
		provider, filepath.Join(dir, "config.yaml"), provider,
	)
}

// BuildPrompt returns the complete prompt, taking all sources into account (prompt, clipboard,
// positional argument, and stdin).
// Prompt and clipboard are mutually exclusive.
// The prompt is evaluated.
func (conf *Jenai) BuildPrompt(lib prompts.Library) (Prompt, error) {
	var (
		clipboard  string
		context    string
		paths      []string
		err        error
		positional []string
		primary    string
		stdin      string
	)

	positional = conf.Positional
	if !conf.OneShot && len(conf.Positional) > 0 {
		positional = conf.Positional[1:]
		primary, err = prompts.NewEvalContext(lib, &positional).Evaluate(conf.Positional[0])
		if err != nil {
			return Prompt{}, err
		}
	}

	if conf.Paste {
		clipboard, err = readClipboard()
		if err != nil {
			return Prompt{}, err
		}
	}

	context, paths, err = conf.Context.Build()
	if err != nil {
		return Prompt{}, err
	}

	stdin, err = readStdin()
	if err != nil {
		return Prompt{}, err
	}

	res := Prompt{
		Clipboard:    clipboard,
		Context:      context,
		ContextAbove: conf.Context.Above,
		Paths:        paths,
		Positional:   strings.Join(positional, " "),
		Primary:      primary,
		Stdin:        stdin,
	}

	return res, nil
}

///////////////////////
// Utility functions //

// readClipboard returns the content of the clipboard.
func readClipboard() (string, error) {
	cmd := exec.Command("xclip", "-o", "-selection", "clipboard")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get clipboard content: %w", err)
	}

	return string(output), nil
}

// readStdin returns the content of stdin, if something was piped in.
func readStdin() (string, error) {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return "", nil // No data being piped.
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("failed to read from stdin: %w", err)
	}
	return string(data), nil
}
