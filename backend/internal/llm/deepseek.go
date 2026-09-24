package llm

import "net/http"

const (
	deepseekDefaultBaseURL = "https://api.deepseek.com"
	deepseekDefaultModel   = "deepseek-chat"
)

// NewDeepSeekClient implements Client against DeepSeek's Chat Completions
// API (https://api-docs.deepseek.com/api/create-chat-completion), which is
// wire-compatible with OpenAI's — same request/response shape, same SSE
// streaming — so this just points OpenAIClient at DeepSeek's base URL,
// the same approach NewHuggingFaceClient uses for its router.
func NewDeepSeekClient(apiKey, model, baseURL string, httpClient *http.Client) *OpenAIClient {
	if baseURL == "" {
		baseURL = deepseekDefaultBaseURL
	}
	client := NewOpenAIClient(apiKey, model, baseURL, httpClient)
	client.name = "deepseek"
	return client
}
