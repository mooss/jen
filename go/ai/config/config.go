// Package config defines jenai's configuration.
package config

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/mooss/bagend/go/flag"
	"github.com/mooss/jen/go/ai/prompts"
)

type Jenai struct {
	// Actual config.
	ContextAbove bool
	ContextDirs  []string
	ContextFiles []string
	DryRun       bool
	Interactive  bool
	List         bool
	Model        string
	OneShot      bool
	Paste        bool
	Positional   []string
	PromptName   string
	Session      string
	TeeFile      string

	// Implementation details.
	rawPositional []string
}

/////////////////////////////////
// Construction and validation //

func (conf *Jenai) RegisterCLI() *flag.Parser {
	parser := flag.NewParser()
	parser.Bool("context-above", &conf.ContextAbove, "Put context files and dir above instructions")
	parser.StringSlice("dir", &conf.ContextDirs, "Include all files in directory as context")
	parser.Bool("dry-run", &conf.DryRun, "Print interpolated prompt without sending to LLM").
		Alias("n")
	parser.StringSlice("file", &conf.ContextFiles, "Include specific file(s) as context")
	parser.Bool("interactive", &conf.Interactive, "Start an interactive aichat session").
		Alias("i")
	parser.Bool("list", &conf.List, "list all available prompts").
		Alias("l")
	parser.String("model", &conf.Model, "Model name").
		Alias("m").Default("gemini")
	parser.Bool("oneshot", &conf.OneShot, "Use positional arguments as the prompt").
		Alias("o")
	parser.Bool("paste", &conf.Paste, "Use clipboard content as prompt")
	parser.String("session", &conf.Session,
		"Reuse or create specific session name (/last for most recent session)")
	parser.String("tee", &conf.TeeFile, "Output to both stdout and FILE")

	return parser
}

// ParseCLI fills the fields from CLI arguments.
func (conf *Jenai) ParseCLI(parser *flag.Parser, args []string) error {
	if err := parser.Parse(args); err != nil {
		return err
	}

	conf.Positional = parser.Positional
	conf.rawPositional = parser.Positional

	// Load prompt name.
	if conf.PromptMode() && len(conf.Positional) > 0 {
		conf.PromptName = conf.Positional[0]
		conf.Positional = conf.Positional[1:]
	}

	return nil
}

// Validate returns an error when the configuration is incoherent.
func (conf Jenai) Validate() error {
	// Validate mutual exclusivity.
	if conf.Paste && conf.OneShot {
		return errors.New("--paste and --oneshot are mutually exclusive")
	}

	// Validate that positional arguments are provided when needed.
	if !conf.Paste && len(conf.rawPositional) == 0 && conf.Session == "" {
		return errors.New("No positional arguments provided")
	}

	return nil
}

//////////////////////////////
// Other high-level methods //

// BuildPrompt returns the complete prompt, taking all sources into account (prompt, clipboard and
// positional argument).
// Prompt and clipboard are mutually exclusive.
// The prompt is evaluated.
func (conf Jenai) BuildPrompt(lib prompts.Library) (string, error) {
	get := func() (string, error) {
		return prompts.NewEvalContext(lib, &conf.Positional).Evaluate(conf.PromptName)
	}

	switch {
	case conf.OneShot:
		get = func() (string, error) { return "", nil }
	case conf.Paste:
		get = ReadClipboard
	}

	buf := []string{}
	if primary, err := get(); err != nil {
		return "", err
	} else if primary != "" {
		buf = append(buf, primary)
	}

	if len(conf.Positional) > 0 {
		buf = append(buf, strings.Join(conf.Positional, " "))
	}

	return strings.Join(buf, "\n\n"), nil
}

/////////////////////
// Utility methods //

// PromptMode returns true when the configuration is in prompt mode (i.e. not in the special paste
// or oneshot mode).
func (conf Jenai) PromptMode() bool {
	return !conf.OneShot && !conf.Paste
}

///////////////////////
// Utility functions //

// ReadClipboard returns the content of the clipboard.
func ReadClipboard() (string, error) {
	cmd := exec.Command("xclip", "-o", "-selection", "clipboard")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get clipboard content: %w", err)
	}

	return string(output), nil
}
