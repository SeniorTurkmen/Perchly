// Package llm abstracts away the specific LLM provider behind a single
// Client interface, so swapping providers (or adding a new one) never
// touches the chat service or HTTP handlers — only this package.
package llm

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Message is one turn in a chat, independent of any provider's wire
// format. Role is "system", "user", or "assistant".
type Message struct {
	Role    string
	Content string
}

// Client streams a chat completion from an LLM provider. onDelta is
// called synchronously for every incremental chunk of text as it arrives,
// so the caller can forward each chunk immediately (e.g. over SSE).
// Returning an error from onDelta aborts the stream early (e.g. the
// client disconnected); StreamChat returns that same error.
type Client interface {
	StreamChat(ctx context.Context, messages []Message, onDelta func(delta string) error) error
}

// Config selects and configures a provider. Provider is read from
// LLM_PROVIDER (see internal/config): "anthropic" (default) calls the
// real Anthropic API and requires APIKey; "openai" calls OpenAI's Chat
// Completions API (or any OpenAI-compatible base URL) and also requires
// APIKey; "huggingface" calls Hugging Face Inference Providers (OpenAI-
// compatible router at https://router.huggingface.co) with an HF token;
// "echo" needs no API key and just streams the user's own message back,
// for local dev/tests. Adding a real new provider is just another case
// in New plus a type implementing Client.
type Config struct {
	Provider string
	APIKey   string
	Model    string
	BaseURL  string
}

// New builds the Client for cfg.Provider.
func New(cfg Config) (Client, error) {
	switch cfg.Provider {
	case "", "anthropic":
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("LLM_API_KEY is required for the %q provider", "anthropic")
		}
		model := cfg.Model
		if model == "" {
			model = "claude-sonnet-5"
		}
		httpClient := &http.Client{Timeout: 60 * time.Second}
		return NewAnthropicClient(cfg.APIKey, model, cfg.BaseURL, httpClient), nil
	case "openai":
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("LLM_API_KEY is required for the %q provider", "openai")
		}
		model := cfg.Model
		if model == "" {
			model = "gpt-4o-mini"
		}
		httpClient := &http.Client{Timeout: 60 * time.Second}
		return NewOpenAIClient(cfg.APIKey, model, cfg.BaseURL, httpClient), nil
	case "huggingface":
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("LLM_API_KEY is required for the %q provider", "huggingface")
		}
		model := cfg.Model
		if model == "" {
			model = huggingfaceDefaultModel
		}
		httpClient := &http.Client{Timeout: 60 * time.Second}
		return NewHuggingFaceClient(cfg.APIKey, model, cfg.BaseURL, httpClient), nil
	case "echo":
		return NewEchoClient(), nil
	default:
		return nil, fmt.Errorf("unknown LLM provider %q", cfg.Provider)
	}
}

// unconfiguredClient makes a missing/invalid LLM configuration fail loudly
// and only when actually used, instead of preventing the whole server
// (personas, health, etc.) from starting.
type unconfiguredClient struct {
	err error
}

func NewUnconfiguredClient(err error) Client {
	return &unconfiguredClient{err: err}
}

func (c *unconfiguredClient) StreamChat(context.Context, []Message, func(string) error) error {
	return c.err
}
