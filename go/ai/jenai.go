package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/mooss/bagend/go/flag"
	"github.com/mooss/jen/go/prompts"
)

type Config struct {
	ContextAbove bool
	ContextDirs  []string
	DryRun       bool
	ContextFiles []string
	Interactive  bool
	Model        string
	OneShot      bool
	Paste        bool
	Session      string
	TeeFile      string
	PromptName   string
	RawPrompt    string
	Positional   []string
}

func fatal(err error) {
	fmt.Println("Error:", err)
	os.Exit(1)
}

func main() {
	/////////////////////////////////
	// Declare and parse arguments //

	var (
		cfg  Config
		list bool
		err  error
	)

	parser := flag.NewParser()
	parser.Bool("context-above", &cfg.ContextAbove, "Put context files and dir above instructions")
	parser.StringSlice("dir", &cfg.ContextDirs, "Include all files in directory as context")
	parser.Bool("dry-run", &cfg.DryRun, "Print interpolated prompt without sending to LLM").
		Alias("n")
	parser.StringSlice("file", &cfg.ContextFiles, "Include specific file(s) as context")
	parser.Bool("interactive", &cfg.Interactive, "Start an interactive aichat session").
		Alias("i")
	parser.Bool("list", &list, "list all available prompts").
		Alias("l")
	parser.String("model", &cfg.Model, "Model name").
		Alias("m").Default("gemini")
	parser.Bool("oneshot", &cfg.OneShot, "Use positional arguments as the prompt").
		Alias("o")
	parser.Bool("paste", &cfg.Paste, "Use clipboard content as prompt")
	parser.String("session", &cfg.Session, "Reuse or create specific session name (/last for most recent session)")
	parser.String("tee", &cfg.TeeFile, "Output to both stdout and FILE")

	if err := parser.Parse(os.Args[1:]); err != nil {
		fatal(err)
	}

	cfg.Positional = parser.Positional

	/////////////////////////////
	// Highjack execution flow //
	// That is to handle the flags that trigger an action and exit immediately.

	if list {
		for name := range prompts.Map {
			fmt.Println(name)
		}
		os.Exit(0)
	}

	////////////////////////////////////
	// Validate and process arguments //

	// Validate mutual exclusivity.
	if cfg.Paste && cfg.OneShot {
		fatal(errors.New("Error: --paste and --oneshot are mutually exclusive"))
	}

	// Validate that positional arguments are provided when needed.
	if !cfg.Paste && len(cfg.Positional) == 0 && cfg.Session == "" {
		fatal(errors.New("Error: No positional arguments provided"))
	}

	if !cfg.OneShot && !cfg.Paste && len(cfg.Positional) > 0 {
		// Shift prompt name.
		cfg.PromptName = cfg.Positional[0]
		cfg.Positional = cfg.Positional[1:]

		cfg.RawPrompt, err = prompts.Raw(cfg.PromptName)
		if err != nil {
			fatal(err)
		}
	}

	/////////////////////
	// Execution logic //

	if cfg.DryRun {
		fmt.Println(cfg.RawPrompt)
		os.Exit(0)
	}

	pretty, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fatal(err)
	}

	fmt.Printf("Config: %s\n", string(pretty))
}
