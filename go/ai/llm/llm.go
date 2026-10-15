package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mooss/jen/go/ai/models"
)

const chatCompletionsURL = "https://openrouter.ai/api/v1/chat/completions"

// maxInvalidChunkPreviewLen caps the payload preview included in parse errors to avoid noisy logs
// and leaking large response content.
const maxInvalidChunkPreviewLen = 200

// Chunk is a single SSE `ChatStreamChunk`.
type Chunk struct {
	Choices []Choice  `json:"choices"`
	Usage   *Usage    `json:"usage,omitempty"`
	Error   *apiError `json:"error,omitempty"`
}

// Choice holds one streamed delta.
type Choice struct {
	Delta        Delta   `json:"delta"`
	FinishReason *string `json:"finish_reason"`
}

// Delta carries either answer text or reasoning text.
// They must stay separate and never be concatenated.
type Delta struct {
	Role      string `json:"role,omitempty"`
	Content   string `json:"content,omitempty"`
	Reasoning string `json:"reasoning,omitempty"`
}

// Usage only signals the terminal usage chunk.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}

type apiError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func (e *apiError) Error() string {
	return fmt.Sprintf("openrouter: code %d: %s", e.Code, e.Message)
}

type chatRequest struct {
	Model     string        `json:"model"`
	Messages  []chatMessage `json:"messages"`
	Stream    bool          `json:"stream"`
	Reasoning chatReasoning `json:"reasoning"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatReasoning struct {
	Effort  string `json:"effort"`
	Summary string `json:"summary"`
}

// Stream starts a streaming completion and returns the raw chunk and error channels.
func Stream(ctx context.Context, spec models.Spec, apiKey, prompt string) (<-chan Chunk, <-chan error, error) {
	req, err := newRequest(ctx, spec, apiKey, prompt)
	if err != nil {
		return nil, nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}

	if err := checkStatus(resp); err != nil {
		resp.Body.Close()
		return nil, nil, err
	}

	chunks, errs := scan(ctx, resp)
	return chunks, errs, nil
}

func newRequest(ctx context.Context, spec models.Spec, apiKey, prompt string) (*http.Request, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("openrouter: missing API key")
	}
	if spec.Provider != "openrouter" {
		return nil, fmt.Errorf("openrouter: unsupported provider %q: expected %q", spec.Provider, "openrouter")
	}

	body, err := json.Marshal(chatRequest{
		Model:     spec.ID(),
		Messages:  []chatMessage{{Role: "user", Content: prompt}},
		Stream:    true,
		Reasoning: chatReasoning{Effort: "medium", Summary: "auto"},
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", chatCompletionsURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	return req, nil
}

func checkStatus(resp *http.Response) error {
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	data, _ := io.ReadAll(io.LimitReader(resp.Body, 4*1024))
	return fmt.Errorf("openrouter: unexpected status %s: %s", resp.Status, strings.TrimSpace(string(data)))
}

func scan(ctx context.Context, resp *http.Response) (<-chan Chunk, <-chan error) {
	chunks := make(chan Chunk)
	errs := make(chan error, 1)

	go func() {
		defer close(chunks)
		defer close(errs)
		defer resp.Body.Close()

		reader := bufio.NewReader(resp.Body)

		for {
			line, readErr := reader.ReadString('\n')
			if payload, ok := ssePayload(line); ok {
				if payload == "[DONE]" {
					return
				}
				chunk, err := parseChunk(payload)
				if err != nil {
					errs <- err
					return
				}
				if chunk.Error != nil {
					errs <- chunk.Error
					return
				}
				select {
				case chunks <- chunk:
				case <-ctx.Done():
					errs <- ctx.Err()
					return
				}
			}

			if readErr != nil {
				if ctx.Err() != nil {
					errs <- ctx.Err()
				} else if readErr != io.EOF {
					errs <- fmt.Errorf("openrouter: stream read: %w", readErr)
				}
				return
			}
		}
	}()

	return chunks, errs
}

func ssePayload(line string) (string, bool) {
	if !strings.HasPrefix(line, "data:") {
		return "", false
	}
	payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
	if payload == "" {
		return "", false
	}
	return payload, true
}

func truncatePreview(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(s) <= max {
		return s
	}
	// Truncate on rune boundary to avoid splitting UTF-8.
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "… [truncated]"
}

func parseChunk(payload string) (Chunk, error) {
	var chunk Chunk
	if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
		preview := truncatePreview(payload, maxInvalidChunkPreviewLen)
		return Chunk{}, fmt.Errorf("openrouter: invalid chunk %q: %w", preview, err)
	}
	return chunk, nil
}
