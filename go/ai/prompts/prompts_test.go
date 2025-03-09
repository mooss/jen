//nolint:revive
package prompts

import (
	"fmt"
	"maps"
	"slices"
	"testing"

	"github.com/mooss/bagend/go/fun/eager/lie"
)

type basic struct {
	title       string
	name        string
	expectError bool
}

type runfunc func(string) (string, error)
type evalfunc func(string, ...any) (string, error)

func getterTest(t *testing.T, run runfunc, what string, cases []basic) {
	t.Helper()

	cases = append(cases, basic{"Non-existent prompt", "%%!unknown_prompt", true})

	for _, tt := range cases {
		t.Run(fmt.Sprintf("%s (%s)", tt.title, what), func(t *testing.T) {
			_, err := run(tt.name)
			if tt.expectError && err == nil {
				t.Errorf("Expected error for %s %q, got nil", what, tt.name)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error for %s %q: %v", what, tt.name, err)
			}
		})
	}
}

func TestGettersAndEvaluators(t *testing.T) {
	lib, err := Embedded()
	if err != nil {
		t.Fatalf("Failed to load embedded prompt library: %v", err)
	}

	// All existing prompts are safe to test.
	getterTest(t, lib.RawPrompt, "prompts", lie.Map(func(prompt string) basic {
		return basic{"Existing prompt: " + prompt, prompt, false}
	}, slices.Collect(maps.Keys(lib.Prompts))))

	/////////////////////
	// Evaluator tests //

	mkvalid := func(what, name string) basic {
		return basic{fmt.Sprintf("Existing %s: %s", what, name), name, false}
	}
	evalTest := func(fun evalfunc, what string, name string) {
		getterTest(
			t,
			func(name string) (string, error) { return fun(name, "arg1", "arg2") },
			what, []basic{mkvalid(what, name)},
		)
	}

	eval := NewEvalContext(lib, nil)
	eval.tmpl.Funcs(map[string]any{"git": func(...any) string { return "wow such diff" }})

	evalTest(eval.persona, "persona", "jaded_dev")
	evalTest(eval.instruction, "instruction", "code_review")
	evalTest(eval.section1, "section1", "git_diff")
}

func TestTemplateFunctions(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		data        any
		expectError bool
	}{
		{"Valid template", "{{ join \", \" (consume_args) }}", []string{"arg1", "arg2"}, false},
		{"Invalid template", "{{ unknown_func }}", nil, true},
	}

	lib, err := Embedded()
	if err != nil {
		t.Fatalf("Failed to load embedded prompts: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewEvalContext(lib, &[]string{"arg1", "arg2"})
			_, err := ctx.execute(tt.template, tt.data)
			if tt.expectError && err == nil {
				t.Errorf("Expected error for template %q, got nil", tt.template)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error for template %q: %v", tt.template, err)
			}
		})
	}
}
