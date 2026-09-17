package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const openAICompletionFixture = `{"id":"chatcmpl-1","object":"chat.completion","choices":[{"index":0,"message":{"role":"assistant","content":"Merhaba!"},"finish_reason":"stop"}]}
`

func TestHuggingFaceClient_StreamChat_UsesOpenAICompatibleRouter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("path = %q, want /v1/chat/completions", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer hf_test" {
			t.Errorf("Authorization = %q, want Bearer hf_test", got)
		}

		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		var req openAIRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if req.Model != "openai/gpt-oss-120b:fastest" {
			t.Errorf("model = %q, want openai/gpt-oss-120b:fastest", req.Model)
		}
		if req.Stream {
			t.Error("expected stream=false")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(openAICompletionFixture))
	}))
	defer server.Close()

	client := NewHuggingFaceClient("hf_test", "openai/gpt-oss-120b:fastest", server.URL, server.Client())

	var got strings.Builder
	err := client.StreamChat(context.Background(), []Message{
		{Role: "user", Content: "Selam"},
	}, func(delta string) error {
		got.WriteString(delta)
		return nil
	})
	if err != nil {
		t.Fatalf("StreamChat returned error: %v", err)
	}
	if want := "Merhaba!"; got.String() != want {
		t.Errorf("streamed text = %q, want %q", got.String(), want)
	}
}

func TestHuggingFaceClient_StripsTrailingV1FromBaseURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("path = %q, want /v1/chat/completions (docs base URL already includes /v1)", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(openAICompletionFixture))
	}))
	defer server.Close()

	client := NewHuggingFaceClient("hf_test", "openai/gpt-oss-120b:fastest", server.URL+"/v1", server.Client())
	err := client.StreamChat(context.Background(), []Message{{Role: "user", Content: "hi"}}, func(string) error {
		return nil
	})
	if err != nil {
		t.Fatalf("StreamChat returned error: %v", err)
	}
}

func TestNew_HuggingFace_Defaults(t *testing.T) {
	client, err := New(Config{Provider: "huggingface", APIKey: "hf_test"})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	oa, ok := client.(*OpenAIClient)
	if !ok {
		t.Fatalf("client type = %T, want *OpenAIClient", client)
	}
	if oa.model != huggingfaceDefaultModel {
		t.Errorf("model = %q, want %q", oa.model, huggingfaceDefaultModel)
	}
	if oa.baseURL != huggingfaceDefaultBaseURL {
		t.Errorf("baseURL = %q, want %q", oa.baseURL, huggingfaceDefaultBaseURL)
	}
	if oa.name != "huggingface" {
		t.Errorf("name = %q, want huggingface", oa.name)
	}
	if oa.stream {
		t.Error("expected stream=false")
	}
}

func TestNew_HuggingFace_RequiresAPIKey(t *testing.T) {
	_, err := New(Config{Provider: "huggingface"})
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.Contains(err.Error(), "LLM_API_KEY") {
		t.Errorf("error = %v, want it to mention LLM_API_KEY", err)
	}
}
