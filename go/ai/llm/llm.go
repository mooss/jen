package llm

import (
	"context"
	"errors"
	"fmt"

	anyllm "github.com/mozilla-ai/any-llm-go"
	"github.com/mozilla-ai/any-llm-go/providers"
	"github.com/mozilla-ai/any-llm-go/providers/openai"

	"github.com/mooss/jen/go/ai/models"
)

// NewProvider returns a provider for the given model spec.
func NewProvider(spec models.Spec, apiKey string) (anyllm.Provider, error) {
	switch spec.Provider {
	case "openrouter":
		return newOpenRouter(apiKey)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", spec.Provider)
	}
}

func newOpenRouter(apiKey string) (anyllm.Provider, error) {
	return openai.NewCompatible(openai.CompatibleConfig{
		Name:           "openrouter",
		DefaultBaseURL: "https://openrouter.ai/api/v1",
		RequireAPIKey:  true,
		Capabilities:   providers.Capabilities{Completion: true, CompletionStreaming: true},
	}, anyllm.WithAPIKey(apiKey))
}

// Complete sends the prompt to the model and returns content and reasoning.
func Complete(
	ctx context.Context, spec models.Spec, apiKey string, prompt string,
) (content, reasoning string, err error) {
	p, err := NewProvider(spec, apiKey)
	if err != nil {
		return "", "", err
	}

	resp, err := p.Completion(ctx, anyllm.CompletionParams{
		Model:           spec.ID(),
		Messages:        []anyllm.Message{{Role: anyllm.RoleUser, Content: prompt}},
		ReasoningEffort: anyllm.ReasoningEffortMedium,
	})
	if err != nil {
		return "", "", err
	}

	if len(resp.Choices) == 0 {
		return "", "", errors.New("empty completion")
	}

	answer := resp.Choices[0].Message
	content = answer.ContentString()

	if answer.Reasoning != nil {
		reasoning = answer.Reasoning.Content
	}

	return content, reasoning, nil
}
