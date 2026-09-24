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
	openAIDefaultBaseURL = "https://api.openai.com"
	// openAIMaxTokens bounds a single reply. Reasoning-capable models
	// (e.g. DeepSeek's deepseek-flash, routed through this same client —
	// see deepseek.go) spend an invisible, variable chunk of this same
	// budget on reasoning_content before any visible content — observed
	// up to ~70% of total tokens for an ordinary question — so this
	// needs real headroom, not just enough for the visible reply text.
	openAIMaxTokens = 4096
)

// OpenAIClient implements Client against OpenAI's Chat Completions API
// (https://platform.openai.com/docs/api-reference/chat/create), using
// server-sent events for streaming. BaseURL is overridable so any
// OpenAI-compatible chat endpoint works too.
type OpenAIClient struct {
	name       string
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
	stream     bool
}

func NewOpenAIClient(apiKey, model, baseURL string, httpClient *http.Client) *OpenAIClient {
	if baseURL == "" {
		baseURL = openAIDefaultBaseURL
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &OpenAIClient{name: "openai", apiKey: apiKey, model: model, baseURL: baseURL, httpClient: httpClient, stream: true}
}

type openAIRequest struct {
	Model     string          `json:"model"`
	Messages  []openAIMessage `json:"messages"`
	Stream    bool            `json:"stream"`
	MaxTokens int             `json:"max_completion_tokens"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (c *OpenAIClient) StreamChat(ctx context.Context, messages []Message, onDelta func(string) error) error {
	body, err := json.Marshal(openAIRequest{
		Model:     c.model,
		Messages:  toOpenAIMessages(messages),
		Stream:    c.stream,
		MaxTokens: openAIMaxTokens,
	})
	if err != nil {
		return fmt.Errorf("encode %s request: %w", c.name, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build %s request: %w", c.name, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call %s: %w", c.name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("%s returned %d: %s", c.name, resp.StatusCode, string(data))
	}

	if !c.stream {
		return parseOpenAICompletion(resp.Body, onDelta, c.name)
	}
	return parseOpenAISSE(resp.Body, onDelta, c.name)
}

func toOpenAIMessages(messages []Message) []openAIMessage {
	out := make([]openAIMessage, 0, len(messages))
	for _, m := range messages {
		out = append(out, openAIMessage{Role: m.Role, Content: m.Content})
	}
	return out
}

type openAICompletion struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func parseOpenAICompletion(body io.Reader, onDelta func(string) error, name string) error {
	var parsed openAICompletion
	if err := json.NewDecoder(body).Decode(&parsed); err != nil {
		return fmt.Errorf("decode %s response: %w", name, err)
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return fmt.Errorf("%s error: %s", name, parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return fmt.Errorf("%s returned no choices", name)
	}
	text := parsed.Choices[0].Message.Content
	if text == "" {
		return nil
	}
	return onDelta(text)
}

type openAIStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// parseOpenAISSE reads OpenAI's Chat Completions event stream and calls
// onDelta for every non-empty content delta. The stream ends on the
// sentinel `data: [DONE]` line.
func parseOpenAISSE(body io.Reader, onDelta func(string) error, name string) error {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}

		var chunk openAIStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			return fmt.Errorf("decode %s event: %w", name, err)
		}
		if chunk.Error != nil && chunk.Error.Message != "" {
			return fmt.Errorf("%s stream error: %s", name, chunk.Error.Message)
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		if text := chunk.Choices[0].Delta.Content; text != "" {
			if err := onDelta(text); err != nil {
				return err
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read %s stream: %w", name, err)
	}
	return nil
}
