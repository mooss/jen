package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mooss/jen/go/ai/config"
	"github.com/mooss/jen/go/ai/prompts"
)

func fatal(err error) {
	fmt.Println("Error:", err)
	os.Exit(1)
}

func main() {
	cfg := config.Jenai{}
	if err := cfg.ParseCLI(os.Args[1:]); err != nil {
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

	/////////////////////
	// Execution logic //

	prompt, err := cfg.BuildPrompt()
	if err != nil {
		fatal(err)
	}

	if cfg.DryRun {
		fmt.Println(prompt)
		os.Exit(0)
	}

	pretty, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fatal(err)
	}

	fmt.Printf("Config: %s\n", string(pretty))
}
