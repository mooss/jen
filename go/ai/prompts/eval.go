// Package prompts implements prompt evaluation using Go templates.
package prompts

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"text/template"
)

// EvalContext provides the context for evaluating templates.
type EvalContext struct {
	Library
	PositionalArguments *[]string
	tmpl                *template.Template
}

func NewEvalContext(lib Library, args *[]string) *EvalContext {
	return &EvalContext{
		Library:             lib,
		PositionalArguments: args,
	}
}

// Evaluate executes the requested prompt within the current context.
func (ctx *EvalContext) Evaluate(prompt string) (string, error) {
	content, err := ctx.RawPrompt(prompt)
	if err != nil {
		return "", err
	}

	ctx.tmpl = template.New("prompt").Funcs(ctx.functions())
	return ctx.execute(content, nil)
}

func (ctx *EvalContext) functions() template.FuncMap {
	return template.FuncMap{
		"git":     gitCommand,
		"join":    joinStrings,
		"strings": stringsFlat,

		"ins":  ctx.instruction,
		"per":  ctx.persona,
		"sec1": ctx.section1,

		"consume_args": ctx.consumeArgs,
	}
}

func (ctx *EvalContext) execute(content string, dot any) (string, error) {
	var err error
	ctx.tmpl, err = ctx.tmpl.Parse(content)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := ctx.tmpl.Execute(&buf, dot); err != nil {
		return "", err
	}

	return buf.String(), nil
}

////////////////////////
// Template functions //

// gitCommand executes a git command and returns its output.
func gitCommand(args ...any) (string, error) {
	strargs, err := stringsFlat(args...)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git", strargs...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %v failed: %w", strargs, err)
	}

	return string(output), nil
}

// joinStrings joins strings with a separator.
func joinStrings(sep string, elems []string) string {
	return strings.Join(elems, sep)
}

// stringsFlat recursively flattens nested slices of strings.
func stringsFlat(args ...any) ([]string, error) {
	var res []string

	for _, arg := range args {
		switch concrete := arg.(type) {
		case string:
			res = append(res, concrete)
		case []string:
			res = append(res, concrete...)
		case []any:
			flattened, err := stringsFlat(concrete...)
			if err != nil {
				return nil, err
			}

			res = append(res, flattened...)
		default:
			return nil, fmt.Errorf("cannot flatten %T to []string", arg)
		}
	}

	return res, nil
}

//////////////////////
// Template methods //

// consumeArgs consumes and returns the positional arguments.
// The goal is to avoid duplication so that when they are used inside a prompt, they are not
// additionally appended at the end.
func (ctx *EvalContext) consumeArgs() ([]string, error) {
	if len(*ctx.PositionalArguments) == 0 {
		return nil, fmt.Errorf("no positional arguments")
	}

	res := *ctx.PositionalArguments
	*ctx.PositionalArguments = nil
	return res, nil
}

// persona returns the persona template with the given name.
func (ctx *EvalContext) persona(name string, args ...any) (string, error) {
	content, exists := ctx.Personas[name]
	if !exists {
		return "", fmt.Errorf("unknown persona: %s", name)
	}

	return ctx.execute("# Persona\n\n"+content, args)
}

// instruction returns the instruction template with the given name.
func (ctx *EvalContext) instruction(name string, args ...any) (string, error) {
	content, exists := ctx.Instructions[name]
	if !exists {
		return "", fmt.Errorf("unknown instruction: %s", name)
	}

	return ctx.execute("# Instructions\n\n"+content, args)
}

// section1 returns the section1 template with the given name and arguments.
func (ctx *EvalContext) section1(name string, args ...string) (string, error) {
	content, exists := ctx.Section1[name]
	if !exists {
		return "", fmt.Errorf("unknown section1: %s", name)
	}

	return ctx.execute(content, args)
}
