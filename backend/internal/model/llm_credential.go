package model

import "time"

// LLMCredential is one stored, encrypted API key for an LLM provider,
// managed from the admin dashboard (see internal/crypto.SecretBox for
// how the key is encrypted at rest). The decrypted key never appears on
// this struct — callers that need to actually call the provider decrypt
// APIKeyCiphertext/APIKeyNonce explicitly, only where the client is
// built (internal/llm), never in a response DTO.
type LLMCredential struct {
	ID               string
	Provider         string
	Label            string
	APIKeyCiphertext []byte
	APIKeyNonce      []byte
	APIKeyLast4      string
	BaseURL          *string
	IsActive         bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
