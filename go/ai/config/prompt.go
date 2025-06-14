package config

import "strings"

type Prompt struct {
	// Context is the content of the included paths.
	Context string

	// ContextAbove is true when the context files should be included at the top of the prompt.
	ContextAbove bool

	// Paths is the names of the included paths.
	Paths []string

	// Positional is the joined positional arguments.
	Positional string

	// Primary is the compiled named prompt or the content of the clipboard.
	Primary string
}

func (p Prompt) Empty() bool {
	return p.Context == "" && p.Positional == "" && p.Primary == ""
}

// String returns the full content of the prompt.
func (p Prompt) String() string {
	return strings.Join(p.addContext(p.static()), "\n\n")
}

func (p Prompt) static() []string {
	buf := []string{}
	if len(p.Primary) > 0 {
		buf = append(buf, p.Primary)
	}

	if len(p.Positional) > 0 {
		buf = append(buf, p.Positional)
	}

	return buf
}

// Static returns the static part of the prompt (without the context).
func (p Prompt) Static() string {
	return strings.Join(p.static(), "\n\n")
}

func (p Prompt) addContext(buf []string) []string {
	if len(p.Context) == 0 {
		return buf
	}

	if p.ContextAbove {
		buf = append([]string{p.Context}, buf...)
	} else {
		buf = append(buf, p.Context)
	}

	return buf
}
