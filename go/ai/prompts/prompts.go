// Package prompts centralizes prompt definition.
package prompts

import (
	_ "embed"
	"fmt"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed prompts.yaml
var embeddedBytes []byte

type Library struct {
	Prompts      map[string]string `yaml:"prompts"`
	Personas     map[string]string `yaml:"personas"`
	Instructions map[string]string `yaml:"instructions"`
	Section1     map[string]string `yaml:"section1"`
}

var Embedded = onceErr(func() (Library, error) { return FromYAML(embeddedBytes) })

func FromYAML(data []byte) (Library, error) {
	var res Library
	if err := yaml.Unmarshal(data, &res); err != nil {
		return Library{}, fmt.Errorf("failed to load prompt library from YAML bytes: %w", err)
	}

	return res, nil
}

// Interpolate loads the requested prompt and executes it.
// Prompt execution is not implemented yet.
func (lib Library) Interpolate(prompt string) (string, error) {
	source, exists := lib.Prompts[prompt]
	if !exists {
		return "", fmt.Errorf("unknown prompt: %s", prompt)
	}

	return source, nil
}

///////////////
// Utilities //

func onceErr[T any](mk func() (T, error)) func() (T, error) {
	type cache struct {
		value T
		err   error
	}
	get := sync.OnceValue(func() cache {
		value, err := mk()
		return cache{value, err}
	})
	return func() (T, error) {
		res := get()
		return res.value, res.err
	}
}
