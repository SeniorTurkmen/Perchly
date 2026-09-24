package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	geminiDefaultBaseURL = "https://generativelanguage.googleapis.com"
	geminiDefaultModel   = "gemini-2.5-flash"
	// geminiMaxTokens bounds a single reply — see openAIMaxTokens' doc
	// comment for why this was raised from the original 1024. Applies
	// here too: Gemini 2.5 Flash has "thinking" on by default, and
	// thinking tokens count against maxOutputTokens the same way.
	geminiMaxTokens = 4096
)

// GeminiClient implements Client against Gemini's generateContent API
// (https://ai.google.dev/api/generate-content), using SSE streaming
// (`streamGenerateContent?alt=sse`). Auth is an AI Studio / Gemini API
// key via `x-goog-api-key`.
type GeminiClient struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

func NewGeminiClient(apiKey, model, baseURL string, httpClient *http.Client) *GeminiClient {
	if baseURL == "" {
		baseURL = geminiDefaultBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &GeminiClient{apiKey: apiKey, model: model, baseURL: baseURL, httpClient: httpClient}
}

type geminiRequest struct {
	SystemInstruction *geminiContent       `json:"systemInstruction,omitempty"`
	Contents          []geminiContent      `json:"contents"`
	GenerationConfig  geminiGenerationCfg  `json:"generationConfig"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationCfg struct {
	MaxOutputTokens int `json:"maxOutputTokens"`
}

func (c *GeminiClient) StreamChat(ctx context.Context, messages []Message, onDelta func(string) error) error {
	system, contents := splitGeminiMessages(messages)

	reqBody := geminiRequest{
		Contents:         contents,
		GenerationConfig: geminiGenerationCfg{MaxOutputTokens: geminiMaxTokens},
	}
	if system != "" {
		reqBody.SystemInstruction = &geminiContent{Parts: []geminiPart{{Text: system}}}
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("encode gemini request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/v1beta/models/%s:streamGenerateContent?%s",
		c.baseURL,
		url.PathEscape(c.model),
		url.Values{"alt": []string{"sse"}}.Encode(),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call gemini: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("gemini returned %d: %s", resp.StatusCode, string(data))
	}

	return parseGeminiSSE(resp.Body, onDelta)
}

// splitGeminiMessages pulls "system" roles into Gemini's top-level
// systemInstruction and maps assistant → model. Consecutive turns with
// the same role are merged (Gemini requires user/model alternation).
func splitGeminiMessages(messages []Message) (string, []geminiContent) {
	var system strings.Builder
	contents := make([]geminiContent, 0, len(messages))

	for _, m := range messages {
		if m.Role == "system" {
			if system.Len() > 0 {
				system.WriteString("\n\n")
			}
			system.WriteString(m.Content)
			continue
		}

		role := "user"
		if m.Role == "assistant" {
			role = "model"
		}

		if n := len(contents); n > 0 && contents[n-1].Role == role {
			contents[n-1].Parts = append(contents[n-1].Parts, geminiPart{Text: m.Content})
			continue
		}
		contents = append(contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: m.Content}},
		})
	}

	return system.String(), contents
}

type geminiStreamEvent struct {
	Candidates []struct {
		Content struct {
			Parts []geminiPart `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	PromptFeedback *struct {
		BlockReason        string `json:"blockReason"`
		BlockReasonMessage string `json:"blockReasonMessage"`
	} `json:"promptFeedback"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func parseGeminiSSE(body io.Reader, onDelta func(string) error) error {
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

		var event geminiStreamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			return fmt.Errorf("decode gemini event: %w", err)
		}
		if event.Error != nil && event.Error.Message != "" {
			return fmt.Errorf("gemini stream error: %s", event.Error.Message)
		}
		if event.PromptFeedback != nil && event.PromptFeedback.BlockReason != "" {
			msg := event.PromptFeedback.BlockReason
			if event.PromptFeedback.BlockReasonMessage != "" {
				msg = event.PromptFeedback.BlockReasonMessage
			}
			return fmt.Errorf("gemini blocked prompt: %s", msg)
		}
		if len(event.Candidates) == 0 {
			continue
		}

		candidate := event.Candidates[0]
		for _, part := range candidate.Content.Parts {
			if part.Text == "" {
				continue
			}
			if err := onDelta(part.Text); err != nil {
				return err
			}
		}
		if candidate.FinishReason == "SAFETY" {
			return fmt.Errorf("gemini blocked response: SAFETY")
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read gemini stream: %w", err)
	}
	return nil
}
