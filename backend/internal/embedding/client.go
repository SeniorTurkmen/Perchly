// Package embedding abstracts away the specific embedding provider behind
// a single Client interface, mirroring internal/llm: swapping providers
// only ever means a new file here plus a case in New.
package embedding

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Client turns text into an embedding vector. Implementations must return
// vectors matching the fixed dimensionality of the message_embeddings
// table (1536) — a mismatch fails at insert time with a clear Postgres
// error rather than silently truncating or padding.
type Client interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

// Config selects and configures a provider. Provider is read from
// EMBEDDING_PROVIDER (see internal/config): "openai" (default) calls the
// real OpenAI embeddings API and requires APIKey; "echo" needs no API key
// and produces a deterministic bag-of-words embedding, for local dev/tests.
type Config struct {
	Provider string
	APIKey   string
	Model    string
	BaseURL  string
}

func New(cfg Config) (Client, error) {
	switch cfg.Provider {
	case "", "openai":
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("EMBEDDING_API_KEY is required for the %q provider", "openai")
		}
		model := cfg.Model
		if model == "" {
			model = "text-embedding-3-small"
		}
		httpClient := &http.Client{Timeout: 30 * time.Second}
		return NewOpenAIClient(cfg.APIKey, model, cfg.BaseURL, httpClient), nil
	case "echo":
		return NewEchoClient(), nil
	default:
		return nil, fmt.Errorf("unknown embedding provider %q", cfg.Provider)
	}
}

// unconfiguredClient makes a missing/invalid embedding configuration fail
// loudly and only when actually used, instead of preventing the whole
// server from starting — embeddings are an enhancement (better recall in
// ContextBuilder), not a hard requirement for chatting.
type unconfiguredClient struct {
	err error
}

func NewUnconfiguredClient(err error) Client {
	return &unconfiguredClient{err: err}
}

func (c *unconfiguredClient) Embed(context.Context, string) ([]float32, error) {
	return nil, c.err
}
