// Package utils provides utility that don't warrant their dedicated package for now.
// This is mostly a purgatory for utilities meant to eventually go to bagend.
package utils

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func FromYAML[T any](data []byte) (T, error) {
	var res T
	if err := yaml.Unmarshal(data, &res); err != nil {
		var zero T
		return zero, err
	}

	return res, nil
}

// Wrap returns a function that will wrap an error using prefix.
// It is meant to be a shortcut for wrapping a 2-value error result.
func Wrap[T any](res T, err error, prefix string) (T, error) {
	if err != nil {
		return res, fmt.Errorf("%s: %w", prefix, err)
	}

	return res, nil
}
