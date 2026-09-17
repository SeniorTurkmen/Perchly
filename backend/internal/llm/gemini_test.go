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

// geminiSSEFixture is a byte-accurate example of Gemini's
// streamGenerateContent?alt=sse payload for a short reply.
const geminiSSEFixture = `data: {"candidates":[{"content":{"parts":[{"text":"Merhaba"}],"role":"model"},"index":0}]}

data: {"candidates":[{"content":{"parts":[{"text":"!"}],"role":"model"},"finishReason":"STOP","index":0}]}

`

func TestGeminiClient_StreamChat_ParsesTextDeltas(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/v1beta/models/gemini-2.5-flash:streamGenerateContent") {
			t.Errorf("path = %q, want streamGenerateContent on gemini-2.5-flash", r.URL.Path)
		}
		if got := r.URL.Query().Get("alt"); got != "sse" {
			t.Errorf("alt = %q, want sse", got)
		}
		if got := r.Header.Get("x-goog-api-key"); got != "test-key" {
			t.Errorf("x-goog-api-key = %q, want test-key", got)
		}

		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		var req geminiRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if req.SystemInstruction == nil || len(req.SystemInstruction.Parts) == 0 {
			t.Fatal("expected systemInstruction from the system message")
		}
		if req.SystemInstruction.Parts[0].Text != "Sen yardımsever bir asistansın." {
			t.Errorf("systemInstruction = %q", req.SystemInstruction.Parts[0].Text)
		}
		if len(req.Contents) != 2 {
			t.Fatalf("contents = %d, want 2", len(req.Contents))
		}
		if req.Contents[0].Role != "user" {
			t.Errorf("first role = %q, want user", req.Contents[0].Role)
		}
		if req.Contents[1].Role != "model" {
			t.Errorf("second role = %q, want model", req.Contents[1].Role)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(geminiSSEFixture))
	}))
	defer server.Close()

	client := NewGeminiClient("test-key", "gemini-2.5-flash", server.URL, server.Client())

	var got strings.Builder
	err := client.StreamChat(context.Background(), []Message{
		{Role: "system", Content: "Sen yardımsever bir asistansın."},
		{Role: "user", Content: "Selam"},
		{Role: "assistant", Content: "Önceki yanıt"},
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

func TestGeminiClient_StreamChat_MergesConsecutiveSameRole(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		var req geminiRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if len(req.Contents) != 1 {
			t.Fatalf("contents = %d, want 1 merged user turn", len(req.Contents))
		}
		if len(req.Contents[0].Parts) != 2 {
			t.Fatalf("parts = %d, want 2", len(req.Contents[0].Parts))
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(geminiSSEFixture))
	}))
	defer server.Close()

	client := NewGeminiClient("test-key", "gemini-2.5-flash", server.URL, server.Client())
	err := client.StreamChat(context.Background(), []Message{
		{Role: "user", Content: "bir"},
		{Role: "user", Content: "iki"},
	}, func(string) error { return nil })
	if err != nil {
		t.Fatalf("StreamChat returned error: %v", err)
	}
}

func TestGeminiClient_StreamChat_PropagatesStreamError(t *testing.T) {
	const errorFixture = `data: {"error":{"message":"Resource exhausted"}}

`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(errorFixture))
	}))
	defer server.Close()

	client := NewGeminiClient("test-key", "gemini-2.5-flash", server.URL, server.Client())
	err := client.StreamChat(context.Background(), []Message{{Role: "user", Content: "hi"}}, func(string) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.Contains(err.Error(), "Resource exhausted") {
		t.Errorf("error = %v, want it to mention the provider message", err)
	}
}

func TestGeminiClient_StreamChat_BlockedPrompt(t *testing.T) {
	const blocked = `data: {"promptFeedback":{"blockReason":"SAFETY","blockReasonMessage":"Blocked"}}

`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(blocked))
	}))
	defer server.Close()

	client := NewGeminiClient("test-key", "gemini-2.5-flash", server.URL, server.Client())
	err := client.StreamChat(context.Background(), []Message{{Role: "user", Content: "hi"}}, func(string) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.Contains(err.Error(), "Blocked") {
		t.Errorf("error = %v, want it to mention the block message", err)
	}
}

func TestGeminiClient_StreamChat_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"API key not valid"}}`))
	}))
	defer server.Close()

	client := NewGeminiClient("bad-key", "gemini-2.5-flash", server.URL, server.Client())
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

func TestNew_Gemini_Defaults(t *testing.T) {
	client, err := New(Config{Provider: "gemini", APIKey: "test-key"})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	g, ok := client.(*GeminiClient)
	if !ok {
		t.Fatalf("client type = %T, want *GeminiClient", client)
	}
	if g.model != geminiDefaultModel {
		t.Errorf("model = %q, want %q", g.model, geminiDefaultModel)
	}
	if g.baseURL != geminiDefaultBaseURL {
		t.Errorf("baseURL = %q, want %q", g.baseURL, geminiDefaultBaseURL)
	}
}

func TestNew_Gemini_RequiresAPIKey(t *testing.T) {
	_, err := New(Config{Provider: "gemini"})
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.Contains(err.Error(), "LLM_API_KEY") {
		t.Errorf("error = %v, want it to mention LLM_API_KEY", err)
	}
}
