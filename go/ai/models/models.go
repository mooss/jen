// Package models defines models that jenai can use and utilities around them.
package models

import (
	_ "embed"
	"fmt"

	"github.com/mooss/jen/go/utils"
)

//go:embed models.yaml
var embeddedBytes []byte

type Spec struct {
	// ShortName of the model, e.g., "r1".
	ShortName string `yaml:"short_name"`
	// Provider of the model, e.g., "openrouter".
	Provider  string `yaml:"provider"`
	// Author of the model, e.g., "deepseek".
	Author    string `yaml:"author"`
	// Model identifier, e.g., "deepseek-r1-0528:nitro".
	Model     string `yaml:"model"`
}

type Zoo struct {
	Models map[string]Spec `yaml:"models"`
}

var loadModels = utils.OnceErr(func() (map[string]Spec, error) { return FromYAML(embeddedBytes) })

func FromYAML(data []byte) (map[string]Spec, error) {
	res, err := utils.FromYAML[Zoo](data)
	return utils.Wrapf(res.Models, err, "failed to load model zoo from YAML")
}

// ModelSpecs returns the map of model short names to their specifications.
func ModelSpecs() (map[string]Spec, error) {
	return loadModels()
}

// Get returns the specification for the given short name.
func Get(shortName string) (Spec, error) {
	specs, err := ModelSpecs()
	if err != nil {
		return Spec{}, err
	}
	spec, ok := specs[shortName]
	if !ok {
		return Spec{}, fmt.Errorf("unknown model: %s", shortName)
	}
	return spec, nil
}

// Aichat returns the model name as aichat's --model flag expects it.
func (sp Spec) Aichat() string {
	return fmt.Sprintf("%s:%s/%s", sp.Provider, sp.Author, sp.Model)
}
