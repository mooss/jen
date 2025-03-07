package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mooss/bagend/go/flag"
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
	Prompt       string
	Positional   []string
}

func fatal(err error) {
	fmt.Println("Error:", err)
	os.Exit(1)
}

func main() {
	var (
		cfg  Config
		list bool
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

	// Shift prompt name if not in oneshot or paste mode.
	if !cfg.OneShot && !cfg.Paste && len(cfg.Positional) > 0 {
		cfg.Prompt = cfg.Positional[0]
		cfg.Positional = cfg.Positional[1:]
	}

	// Validate mutual exclusivity.
	if cfg.Paste && cfg.OneShot {
		fmt.Println("Error: --paste and --oneshot are mutually exclusive")
		os.Exit(1)
	}

	if cfg.Session == "/last" {
		fmt.Println("Warning: /last session handling not yet implemented")
	}

	pretty, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fatal(err)
	}

	fmt.Printf("Config: %s\n", string(pretty))
}
