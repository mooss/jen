package config_test

import (
	"testing"

	"github.com/mooss/jen/go/ai/config"
)

func TestPromptMode(t *testing.T) {
	tests := []struct {
		name     string
		jen      config.Jenai
		expected bool
	}{
		{"Prompt mode", config.Jenai{}, true},
		{"OneShot mode", config.Jenai{OneShot: true}, false},
		{"Paste mode", config.Jenai{Paste: true}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.jen.PromptMode()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
