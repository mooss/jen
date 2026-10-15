package llm

import (
	"context"
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

func streamParams(spec models.Spec, prompt string) anyllm.CompletionParams {
	return anyllm.CompletionParams{
		Model:           spec.ID(),
		Messages:        []anyllm.Message{{Role: anyllm.RoleUser, Content: prompt}},
		ReasoningEffort: anyllm.ReasoningEffortMedium,
		StreamOptions:   &anyllm.StreamOptions{IncludeUsage: true},
	}
}

// Stream starts a streaming completion and returns the raw chunk and error channels.
func Stream(ctx context.Context, spec models.Spec, apiKey, prompt string) (<-chan anyllm.ChatCompletionChunk, <-chan error, error) {
	p, err := NewProvider(spec, apiKey)
	if err != nil {
		return nil, nil, err
	}

	if cp, ok := p.(anyllm.CapabilityProvider); ok {
		if !cp.Capabilities().CompletionStreaming {
			return nil, nil, fmt.Errorf("provider %q does not support streaming completion", spec.Provider)
		}
	}

	chunks, errs := p.CompletionStream(ctx, streamParams(spec, prompt))
	return chunks, errs, nil
}
