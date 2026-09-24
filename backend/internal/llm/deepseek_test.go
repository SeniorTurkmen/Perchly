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

func TestDeepSeekClient_StreamChat_UsesOpenAICompatibleAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("path = %q, want /v1/chat/completions", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer ds_test" {
			t.Errorf("Authorization = %q, want Bearer ds_test", got)
		}

		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		var req openAIRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if req.Model != "deepseek-chat" {
			t.Errorf("model = %q, want deepseek-chat", req.Model)
		}
		if !req.Stream {
			t.Error("expected stream=true")
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"Merhaba!\"}}]}\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()

	client := NewDeepSeekClient("ds_test", "deepseek-chat", server.URL, server.Client())

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

func TestNew_DeepSeek_Defaults(t *testing.T) {
	client, err := New(Config{Provider: "deepseek", APIKey: "ds_test"})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	oa, ok := client.(*OpenAIClient)
	if !ok {
		t.Fatalf("client type = %T, want *OpenAIClient", client)
	}
	if oa.model != deepseekDefaultModel {
		t.Errorf("model = %q, want %q", oa.model, deepseekDefaultModel)
	}
	if oa.baseURL != deepseekDefaultBaseURL {
		t.Errorf("baseURL = %q, want %q", oa.baseURL, deepseekDefaultBaseURL)
	}
	if oa.name != "deepseek" {
		t.Errorf("name = %q, want deepseek", oa.name)
	}
	if !oa.stream {
		t.Error("expected stream=true")
	}
}

func TestNew_DeepSeek_RequiresAPIKey(t *testing.T) {
	_, err := New(Config{Provider: "deepseek"})
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.Contains(err.Error(), "LLM_API_KEY") {
		t.Errorf("error = %v, want it to mention LLM_API_KEY", err)
	}
}
