package llm

import (
	"net/http"
	"strings"
)

const (
	// Hugging Face's OpenAI-compatible router. The OpenAI SDK docs show
	// this with a trailing /v1; we store the origin only and append
	// /v1/chat/completions the same way OpenAIClient does.
	huggingfaceDefaultBaseURL = "https://router.huggingface.co"
	// Docs default: fastest available partner for this model. Swap the
	// suffix to :cheapest, :preferred, or an explicit partner (:groq).
	huggingfaceDefaultModel = "openai/gpt-oss-120b:fastest"
)

// NewHuggingFaceClient implements Client against Hugging Face Inference
// Providers (https://huggingface.co/docs/inference-providers): an
// OpenAI-compatible Chat Completions API authenticated with an HF token.
// Requests use stream=false (the router returns one JSON completion);
// StreamChat still delivers the text through onDelta so callers don't
// change. BaseURL may be either https://router.huggingface.co or the
// docs form https://router.huggingface.co/v1 — both resolve to the same
// endpoint.
func NewHuggingFaceClient(apiKey, model, baseURL string, httpClient *http.Client) *OpenAIClient {
	if baseURL == "" {
		baseURL = huggingfaceDefaultBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")
	baseURL = strings.TrimSuffix(baseURL, "/v1")
	client := NewOpenAIClient(apiKey, model, baseURL, httpClient)
	client.name = "huggingface"
	client.stream = false
	return client
}
