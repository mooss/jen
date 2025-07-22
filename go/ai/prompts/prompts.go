// Package prompts centralizes prompt definition.
package prompts

import (
	_ "embed"
	"fmt"

	"github.com/mooss/jen/go/utils"
)

//go:embed prompts.yaml
var EmbeddedBytes []byte

type Library struct {
	Prompts      map[string]string `yaml:"prompts"`
	Personas     map[string]string `yaml:"personas"`
	Instructions map[string]string `yaml:"instructions"`
	Section1     map[string]string `yaml:"section1"`
}

var Embedded = utils.OnceErr(func() (Library, error) { return FromYAML(EmbeddedBytes) })

func FromYAML(data []byte) (Library, error) {
	res, err := utils.FromYAML[Library](data)
	return utils.Wrapf(res, err, "failed to load prompt library from YAML")
}

// RawPrompt returns the content of requested prompt, if it exists.
func (lib Library) RawPrompt(name string) (string, error) {
	source, exists := lib.Prompts[name]
	if !exists {
		return "", fmt.Errorf("unknown prompt: %s", name)
	}

	return source, nil
}
