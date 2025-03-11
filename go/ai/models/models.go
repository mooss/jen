// Package models defines models that jenai can use and utilities around them.
package models

import (
	"fmt"
)

type Spec struct {
	// ShortName of the model, e.g., "codestral"
	ShortName string
	// Provider of the model, e.g., "openrouter"
	Provider string
	// Author of the model, e.g., "mistralai"
	Author string
	// Model identifier, e.g., "codestral-2501"
	Model string
}

//nolint:revive
var ModelSpecs = map[string]Spec{
	"codestral":  {ShortName: "codestral", Provider: "openrouter", Author: "mistralai", Model: "codestral-2501"},
	"gemini":     {ShortName: "gemini", Provider: "openrouter", Author: "google", Model: "gemini-2.0-flash-001"},
	"gemini-pro": {ShortName: "gemini-pro", Provider: "openrouter", Author: "google", Model: "gemini-2.0-pro-exp-02-05:free"},
	"haiku":      {ShortName: "haiku", Provider: "openrouter", Author: "anthropic", Model: "claude-3-haiku"},
	"r1":         {ShortName: "r1", Provider: "openrouter", Author: "deepseek", Model: "deepseek-r1:nitro"},
	"sonnet":     {ShortName: "sonnet", Provider: "openrouter", Author: "anthropic", Model: "claude-3.7-sonnet"},
	"r1-70":      {ShortName: "r1-70", Provider: "openrouter", Author: "deepseek", Model: "deepseek-r1-distill-llama-70b:free"},
	"qwco":       {ShortName: "qwco", Provider: "openrouter", Author: "qwen", Model: "qwen-2.5-coder-32b-instruct"},
	"ds-v3":      {ShortName: "ds-v3", Provider: "openrouter", Author: "deepseek", Model: "deepseek-chat"},
}

// Aichat returns the model name as aichat's --model flag expects it.
func (sp Spec) Aichat() string {
	return fmt.Sprintf("%s:%s/%s", sp.Provider, sp.Author, sp.Model)
}
