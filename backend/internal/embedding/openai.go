package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const openAIDefaultBaseURL = "https://api.openai.com"

// OpenAIClient implements Client against OpenAI's embeddings API
// (https://platform.openai.com/docs/api-reference/embeddings). BaseURL is
// overridable so any OpenAI-compatible embeddings endpoint works too.
type OpenAIClient struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

func NewOpenAIClient(apiKey, model, baseURL string, httpClient *http.Client) *OpenAIClient {
	if baseURL == "" {
		baseURL = openAIDefaultBaseURL
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &OpenAIClient{apiKey: apiKey, model: model, baseURL: baseURL, httpClient: httpClient}
}

type openAIEmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type openAIEmbeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

func (c *OpenAIClient) Embed(ctx context.Context, text string) ([]float32, error) {
	body, err := json.Marshal(openAIEmbeddingRequest{Model: c.model, Input: text})
	if err != nil {
		return nil, fmt.Errorf("encode embedding request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build embedding request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call embedding API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("embedding API returned %d: %s", resp.StatusCode, string(data))
	}

	var parsed openAIEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decode embedding response: %w", err)
	}
	if len(parsed.Data) == 0 {
		return nil, fmt.Errorf("embedding API returned no data")
	}
	return parsed.Data[0].Embedding, nil
}
