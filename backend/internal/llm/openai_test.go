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

// openAISSEFixture is a byte-accurate example of what OpenAI's Chat
// Completions API actually sends for a short streamed reply.
const openAISSEFixture = `data: {"id":"chatcmpl-1","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}

data: {"id":"chatcmpl-1","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"Merhaba"},"finish_reason":null}]}

data: {"id":"chatcmpl-1","object":"chat.completion.chunk","choices":[{"index":0,"delta":{"content":"!"},"finish_reason":null}]}

data: {"id":"chatcmpl-1","object":"chat.completion.chunk","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: [DONE]

`

func TestOpenAIClient_StreamChat_ParsesContentDeltas(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("path = %q, want /v1/chat/completions", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("Authorization = %q, want Bearer test-key", got)
		}

		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		var req openAIRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if !req.Stream {
			t.Error("expected stream=true")
		}
		if len(req.Messages) != 2 {
			t.Fatalf("messages = %d, want 2 (system stays in the array)", len(req.Messages))
		}
		if req.Messages[0].Role != "system" {
			t.Errorf("first message role = %q, want system", req.Messages[0].Role)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(openAISSEFixture))
	}))
	defer server.Close()

	client := NewOpenAIClient("test-key", "gpt-4o-mini", server.URL, server.Client())

	var got strings.Builder
	err := client.StreamChat(context.Background(), []Message{
		{Role: "system", Content: "Sen yardımsever bir asistansın."},
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

func TestOpenAIClient_StreamChat_PropagatesStreamError(t *testing.T) {
	const errorFixture = `data: {"error":{"message":"Overloaded","type":"server_error"}}

`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(errorFixture))
	}))
	defer server.Close()

	client := NewOpenAIClient("test-key", "gpt-4o-mini", server.URL, server.Client())

	err := client.StreamChat(context.Background(), []Message{{Role: "user", Content: "hi"}}, func(string) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.Contains(err.Error(), "Overloaded") {
		t.Errorf("error = %v, want it to mention the provider message", err)
	}
}

func TestOpenAIClient_StreamChat_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"Incorrect API key provided"}}`))
	}))
	defer server.Close()

	client := NewOpenAIClient("bad-key", "gpt-4o-mini", server.URL, server.Client())

	err := client.StreamChat(context.Background(), []Message{{Role: "user", Content: "hi"}}, func(string) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error = %v, want it to mention the status code", err)
	}
}
