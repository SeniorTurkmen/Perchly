package service

import (
	"context"
	"errors"
	"testing"

	"perchly-backend/internal/crypto"
	"perchly-backend/internal/llm"
	"perchly-backend/internal/model"
	"perchly-backend/internal/repository"
)

type fakeLLMCredentialLookup struct {
	credentials map[string]model.LLMCredential
	calls       int
}

func (f *fakeLLMCredentialLookup) GetByID(_ context.Context, id string) (model.LLMCredential, error) {
	f.calls++
	c, ok := f.credentials[id]
	if !ok {
		return model.LLMCredential{}, repository.ErrLLMCredentialNotFound
	}
	return c, nil
}

type fakeLLMModelLookup struct {
	models map[string]model.LLMModel
	calls  int
}

func (f *fakeLLMModelLookup) GetByID(_ context.Context, id string) (model.LLMModel, error) {
	f.calls++
	m, ok := f.models[id]
	if !ok {
		return model.LLMModel{}, repository.ErrLLMModelNotFound
	}
	return m, nil
}

func testSecretBox(t *testing.T) *crypto.SecretBox {
	t.Helper()
	key, err := crypto.DecodeKey("MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=") // 32 raw bytes
	if err != nil {
		t.Fatalf("decode test key: %v", err)
	}
	box, err := crypto.NewSecretBox(key)
	if err != nil {
		t.Fatalf("NewSecretBox: %v", err)
	}
	return box
}

// encryptedFixture returns ciphertext/nonce for plaintext under box, for
// building a model.LLMCredential fixture that decrypts back correctly.
func encryptedFixture(t *testing.T, box *crypto.SecretBox, plaintext string) (ciphertext, nonce []byte) {
	t.Helper()
	ciphertext, nonce, err := box.Encrypt([]byte(plaintext))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	return ciphertext, nonce
}

func TestLLMClientFactory_ForPersona_NilModelIDUsesFallback(t *testing.T) {
	box := testSecretBox(t)
	fallback := &fakeLLMClient{}
	factory := NewLLMClientFactory(&fakeLLMCredentialLookup{}, &fakeLLMModelLookup{}, box, fallback)

	client, err := factory.ForPersona(context.Background(), model.Persona{LLMModelID: nil})
	if err != nil {
		t.Fatalf("ForPersona() error = %v", err)
	}
	if client != fallback {
		t.Errorf("ForPersona() returned %v, want the fallback client", client)
	}
}

func TestLLMClientFactory_ForModel_BuildsAndCaches(t *testing.T) {
	box := testSecretBox(t)
	ciphertext, nonce := encryptedFixture(t, box, "test-key")

	credentials := &fakeLLMCredentialLookup{credentials: map[string]model.LLMCredential{
		"cred-1": {ID: "cred-1", Provider: "echo", IsActive: true, APIKeyCiphertext: ciphertext, APIKeyNonce: nonce},
	}}
	models := &fakeLLMModelLookup{models: map[string]model.LLMModel{
		"model-1": {ID: "model-1", CredentialID: "cred-1", ModelName: "echo-model", IsActive: true},
	}}
	factory := NewLLMClientFactory(credentials, models, box, &fakeLLMClient{})

	ctx := context.Background()
	client1, err := factory.ForModel(ctx, "model-1")
	if err != nil {
		t.Fatalf("ForModel() error = %v", err)
	}
	if client1 == nil {
		t.Fatal("ForModel() returned a nil client")
	}

	client2, err := factory.ForModel(ctx, "model-1")
	if err != nil {
		t.Fatalf("ForModel() second call error = %v", err)
	}
	if client1 != client2 {
		t.Errorf("ForModel() second call returned a different client instance, want the cached one")
	}
	if credentials.calls != 1 || models.calls != 1 {
		t.Errorf("repo calls = (credentials:%d, models:%d), want exactly 1 each (second ForModel should hit the cache)", credentials.calls, models.calls)
	}
}

func TestLLMClientFactory_ForModel_InactiveModelFails(t *testing.T) {
	box := testSecretBox(t)
	ciphertext, nonce := encryptedFixture(t, box, "test-key")

	credentials := &fakeLLMCredentialLookup{credentials: map[string]model.LLMCredential{
		"cred-1": {ID: "cred-1", Provider: "echo", IsActive: true, APIKeyCiphertext: ciphertext, APIKeyNonce: nonce},
	}}
	models := &fakeLLMModelLookup{models: map[string]model.LLMModel{
		"model-1": {ID: "model-1", CredentialID: "cred-1", ModelName: "echo-model", IsActive: false},
	}}
	factory := NewLLMClientFactory(credentials, models, box, &fakeLLMClient{})

	if _, err := factory.ForModel(context.Background(), "model-1"); err == nil {
		t.Fatal("ForModel() with an inactive model: expected an error, got nil")
	}
}

func TestLLMClientFactory_ForModel_InactiveCredentialFails(t *testing.T) {
	box := testSecretBox(t)
	ciphertext, nonce := encryptedFixture(t, box, "test-key")

	credentials := &fakeLLMCredentialLookup{credentials: map[string]model.LLMCredential{
		"cred-1": {ID: "cred-1", Provider: "echo", IsActive: false, APIKeyCiphertext: ciphertext, APIKeyNonce: nonce},
	}}
	models := &fakeLLMModelLookup{models: map[string]model.LLMModel{
		"model-1": {ID: "model-1", CredentialID: "cred-1", ModelName: "echo-model", IsActive: true},
	}}
	factory := NewLLMClientFactory(credentials, models, box, &fakeLLMClient{})

	if _, err := factory.ForModel(context.Background(), "model-1"); err == nil {
		t.Fatal("ForModel() with an inactive credential: expected an error, got nil")
	}
}

func TestLLMClientFactory_ForModel_NotFoundPropagatesError(t *testing.T) {
	box := testSecretBox(t)
	factory := NewLLMClientFactory(&fakeLLMCredentialLookup{}, &fakeLLMModelLookup{}, box, &fakeLLMClient{})

	_, err := factory.ForModel(context.Background(), "does-not-exist")
	if err == nil {
		t.Fatal("ForModel() for an unknown model id: expected an error, got nil")
	}
	if !errors.Is(err, repository.ErrLLMModelNotFound) {
		t.Errorf("ForModel() error = %v, want it to wrap repository.ErrLLMModelNotFound", err)
	}
}

func TestLLMClientFactory_InvalidateAll_ForcesRebuild(t *testing.T) {
	box := testSecretBox(t)
	ciphertext, nonce := encryptedFixture(t, box, "test-key")

	credentials := &fakeLLMCredentialLookup{credentials: map[string]model.LLMCredential{
		"cred-1": {ID: "cred-1", Provider: "echo", IsActive: true, APIKeyCiphertext: ciphertext, APIKeyNonce: nonce},
	}}
	models := &fakeLLMModelLookup{models: map[string]model.LLMModel{
		"model-1": {ID: "model-1", CredentialID: "cred-1", ModelName: "echo-model", IsActive: true},
	}}
	factory := NewLLMClientFactory(credentials, models, box, &fakeLLMClient{})

	ctx := context.Background()
	if _, err := factory.ForModel(ctx, "model-1"); err != nil {
		t.Fatalf("first ForModel() error = %v", err)
	}
	factory.InvalidateAll()
	if _, err := factory.ForModel(ctx, "model-1"); err != nil {
		t.Fatalf("second ForModel() error = %v", err)
	}

	if credentials.calls != 2 || models.calls != 2 {
		t.Errorf("repo calls after invalidation = (credentials:%d, models:%d), want exactly 2 each (a fresh rebuild)", credentials.calls, models.calls)
	}
}

// fakeLLMClient is a minimal llm.Client for identity comparisons in
// tests (== on the fallback pointer) — its StreamChat is never actually
// invoked by these tests.
type fakeLLMClient struct{}

func (f *fakeLLMClient) StreamChat(context.Context, []llm.Message, func(string) error) error {
	return nil
}
