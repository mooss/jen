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

// Wrapf wraps a non-nil err and prefixes it with a formatted string.
// It is meant to be a shortcut for wrapping a 2-value error result.
func Wrapf[T any](res T, err error, format string, args ...any) (T, error) {
	if err != nil {
		return res, fmt.Errorf(format+": %w", append(args, err)...)
	}

	return res, nil
}
