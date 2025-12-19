// Package config defines jenai's configuration.
package config

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/mooss/bagend/go/flag"
	"github.com/mooss/jen/go/ai/prompts"
)

type Jenai struct {
	// Actual config.
	Context     Context
	DryRun      bool
	Interactive bool
	List        bool
	ListModels  bool
	Model       string
	OneShot     bool
	Paste       bool
	Positional  []string
	session     SessionMetadata
	TeeFile     string
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
	parser.Bool("interactive", &conf.Interactive, "Start an interactive aichat session").
		Alias("i")
	parser.Bool("list", &conf.List, "list all available prompts").
		Alias("l")
	parser.Bool("list-models", &conf.ListModels, "list all available models").
		Alias("lm")
	parser.Bool("linum", &conf.Context.LineNumbers, "Print files with line numbers")
	parser.String("model", &conf.Model, "Model name (short name from --lm or provider:author/model)").
		Alias("m").Default("ds3.2")
	parser.Bool("oneshot", &conf.OneShot, "Use positional arguments as the prompt").
		Alias("o")
	parser.Bool("paste", &conf.Paste, "Use clipboard content as prompt")
	parser.String("session", &conf.session.Name,
		"Reuse or create specific session name (/last for most recent session)")
	parser.String("tee", &conf.TeeFile, "Output first answer to both stdout and FILE (overwritten)")

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

/////////////
// Session //

// Session returns a session object to manipulate a chat session.
func (conf *Jenai) Session() (SessionMetadata, error) {
	if conf.session.Dir == "" {
		if err := conf.session.prepare(); err != nil {
			return SessionMetadata{}, err
		}
	}

	return conf.session, nil
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
