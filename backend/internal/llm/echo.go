package llm

import (
	"context"
	"strings"
)

// EchoClient is a zero-dependency Client for local development and tests
// without a real provider or API key: it streams the last user message
// back, word by word. Select it with LLM_PROVIDER=echo. It also serves as
// the reference example for adding a new provider: implement Client, add
// a case in New, done.
type EchoClient struct{}

func NewEchoClient() *EchoClient {
	return &EchoClient{}
}

func (c *EchoClient) StreamChat(ctx context.Context, messages []Message, onDelta func(string) error) error {
	var lastUser string
	for _, m := range messages {
		if m.Role == "user" {
			lastUser = m.Content
		}
	}

	words := strings.Fields(lastUser)
	if len(words) == 0 {
		return onDelta("...")
	}

	for i, word := range words {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		chunk := word
		if i < len(words)-1 {
			chunk += " "
		}
		if err := onDelta(chunk); err != nil {
			return err
		}
	}
	return nil
}
