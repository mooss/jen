// Package config defines jenai's configuration.
package config

import (
	"errors"

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

func (jen *Jenai) Parse(args []string) error {
	parser := flag.NewParser()
	parser.Bool("context-above", &jen.ContextAbove, "Put context files and dir above instructions")
	parser.StringSlice("dir", &jen.ContextDirs, "Include all files in directory as context")
	parser.Bool("dry-run", &jen.DryRun, "Print interpolated prompt without sending to LLM").
		Alias("n")
	parser.StringSlice("file", &jen.ContextFiles, "Include specific file(s) as context")
	parser.Bool("interactive", &jen.Interactive, "Start an interactive aichat session").
		Alias("i")
	parser.Bool("list", &jen.List, "list all available prompts").
		Alias("l")
	parser.String("model", &jen.Model, "Model name").
		Alias("m").Default("gemini")
	parser.Bool("oneshot", &jen.OneShot, "Use positional arguments as the prompt").
		Alias("o")
	parser.Bool("paste", &jen.Paste, "Use clipboard content as prompt")
	parser.String("session", &jen.Session, "Reuse or create specific session name (/last for most recent session)")
	parser.String("tee", &jen.TeeFile, "Output to both stdout and FILE")

	if err := parser.Parse(args); err != nil {
		return err
	}

	jen.Positional = parser.Positional
	jen.rawPositional = parser.Positional

	// Load prompt name.
	if jen.PromptMode() && len(jen.Positional) > 0 {
		jen.PromptName = jen.Positional[0]
		jen.Positional = jen.Positional[1:]
	}

	return nil
}

func (jen Jenai) Validate() error {
	// Validate mutual exclusivity.
	if jen.Paste && jen.OneShot {
		return errors.New("Error: --paste and --oneshot are mutually exclusive")
	}

	// Validate that positional arguments are provided when needed.
	if !jen.Paste && len(jen.rawPositional) == 0 && jen.Session == "" {
		return errors.New("Error: No positional arguments provided")
	}

	return nil
}

// PromptMode returns true when the configuration is in prompt mode (i.e. not in the special paste
// or oneshot mode).
func (jen Jenai) PromptMode() bool {
	return !jen.OneShot && !jen.Paste
}

func (jen Jenai) RawPrompt() (string, error) {
	return prompts.Raw(jen.PromptName)
}
