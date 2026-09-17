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
)

const (
	anthropicDefaultBaseURL = "https://api.anthropic.com"
	anthropicVersion        = "2023-06-01"
	anthropicMaxTokens      = 1024
)

// AnthropicClient implements Client against Anthropic's Messages API
// (https://docs.anthropic.com/en/api/messages-streaming), using
// server-sent events for streaming.
type AnthropicClient struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

func NewAnthropicClient(apiKey, model, baseURL string, httpClient *http.Client) *AnthropicClient {
	if baseURL == "" {
		baseURL = anthropicDefaultBaseURL
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &AnthropicClient{apiKey: apiKey, model: model, baseURL: baseURL, httpClient: httpClient}
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
	Stream    bool               `json:"stream"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (c *AnthropicClient) StreamChat(ctx context.Context, messages []Message, onDelta func(string) error) error {
	system, chatMessages := splitSystemMessage(messages)

	body, err := json.Marshal(anthropicRequest{
		Model:     c.model,
		MaxTokens: anthropicMaxTokens,
		System:    system,
		Messages:  chatMessages,
		Stream:    true,
	})
	if err != nil {
		return fmt.Errorf("encode anthropic request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build anthropic request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call anthropic: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("anthropic returned %d: %s", resp.StatusCode, string(data))
	}

	return parseAnthropicSSE(resp.Body, onDelta)
}

// splitSystemMessage pulls "system" role messages out into Anthropic's
// separate top-level `system` field, since its `messages` array only
// accepts "user"/"assistant" roles.
func splitSystemMessage(messages []Message) (string, []anthropicMessage) {
	var system strings.Builder
	chatMessages := make([]anthropicMessage, 0, len(messages))

	for _, m := range messages {
		if m.Role == "system" {
			if system.Len() > 0 {
				system.WriteString("\n\n")
			}
			system.WriteString(m.Content)
			continue
		}
		chatMessages = append(chatMessages, anthropicMessage{Role: m.Role, Content: m.Content})
	}

	return system.String(), chatMessages
}

type anthropicStreamEvent struct {
	Type  string `json:"type"`
	Delta struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// parseAnthropicSSE reads Anthropic's event stream and calls onDelta for
// every text_delta chunk. It relies on the `type` field inside each
// event's JSON payload rather than the SSE `event:` line, since Anthropic
// always includes it and it's simpler to parse correctly.
func parseAnthropicSSE(body io.Reader, onDelta func(string) error) error {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var data string
	handleEvent := func() error {
		if data == "" {
			return nil
		}
		defer func() { data = "" }()

		var event anthropicStreamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			return fmt.Errorf("decode anthropic event: %w", err)
		}

		switch event.Type {
		case "content_block_delta":
			if event.Delta.Type == "text_delta" && event.Delta.Text != "" {
				return onDelta(event.Delta.Text)
			}
		case "error":
			return fmt.Errorf("anthropic stream error: %s", event.Error.Message)
		}
		return nil
	}

	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case line == "":
			if err := handleEvent(); err != nil {
				return err
			}
		case strings.HasPrefix(line, "data:"):
			data = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read anthropic stream: %w", err)
	}
	return handleEvent()
}
