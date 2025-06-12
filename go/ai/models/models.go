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
	"gpt-4.1":    {ShortName: "gpt-4.1", Provider: "openrouter", Author: "openai", Model: "gpt-4.1"},
	"gpt-4.1-mini":    {ShortName: "gpt-4.1-mini", Provider: "openrouter", Author: "openai", Model: "gpt-4.1-mini"},
	"gpt-4.1-nano":    {ShortName: "gpt-4.1-nano", Provider: "openrouter", Author: "openai", Model: "gpt-4.1-nano"},

	"gemini-flash":    {ShortName: "gemini-flash", Provider: "openrouter", Author: "google", Model: "gemini-2.5-flash-preview-05-20:nitro"},
	"gemini-pro": {ShortName: "gemini-pro", Provider: "openrouter", Author: "google", Model: "gemini-2.5-pro-preview:nitro"},

	"r1":         {ShortName: "r1", Provider: "openrouter", Author: "deepseek", Model: "deepseek-r1-0528:nitro"},
	"ds-v3":      {ShortName: "ds-v3", Provider: "openrouter", Author: "deepseek", Model: "deepseek-chat:nitro"},

	"haiku":      {ShortName: "haiku", Provider: "openrouter", Author: "anthropic", Model: "claude-3.5-haiku"},
	"sonnet":     {ShortName: "sonnet", Provider: "openrouter", Author: "anthropic", Model: "claude-sonnet-4"},
	"opus":     {ShortName: "opus", Provider: "openrouter", Author: "anthropic", Model: "claude-opus-4"},

	"qwco":       {ShortName: "qwco", Provider: "openrouter", Author: "qwen", Model: "qwen-2.5-coder-32b-instruct"},

	"codestral": {ShortName: "codestral", Provider: "openrouter", Author: "mistralai", Model: "codestral-2501"},
}

// Aichat returns the model name as aichat's --model flag expects it.
func (sp Spec) Aichat() string {
	return fmt.Sprintf("%s:%s/%s", sp.Provider, sp.Author, sp.Model)
}
