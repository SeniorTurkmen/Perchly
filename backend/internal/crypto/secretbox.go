// Package crypto provides the one primitive the codebase needs for
// storing third-party secrets (LLM provider API keys) at rest: AES-256-GCM
// authenticated encryption under a single server-held key. It is not a
// general-purpose crypto package — bcrypt (password hashing) and
// crypto/rand+sha256 (opaque session tokens) already live where they're
// used (internal/auth) and stay there; this package exists only for
// values that must be decrypted back to their original form later.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

// ErrInvalidKeySize is returned by NewSecretBox when key isn't exactly
// 32 bytes — AES-256 requires it, and silently accepting a shorter key
// would quietly downgrade to AES-128 or panic deep inside crypto/aes.
var ErrInvalidKeySize = errors.New("crypto: key must be exactly 32 bytes")

// SecretBox encrypts and decrypts values with a single fixed AES-256-GCM
// key. One instance is created at startup from LLM_TOKEN_ENCRYPTION_KEY
// (see internal/config) and shared by every repository that stores an
// encrypted secret.
type SecretBox struct {
	aead cipher.AEAD
}

// NewSecretBox builds a SecretBox from a raw 32-byte key.
func NewSecretBox(key []byte) (*SecretBox, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKeySize
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: build AES cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: build GCM: %w", err)
	}
	return &SecretBox{aead: aead}, nil
}

// Encrypt returns the ciphertext and a freshly generated nonce for
// plaintext. Both must be stored — Decrypt needs the exact nonce Encrypt
// generated, so callers persist it alongside the ciphertext (see
// llm_credentials.api_key_nonce) rather than deriving or reusing one.
func (b *SecretBox) Encrypt(plaintext []byte) (ciphertext, nonce []byte, err error) {
	nonce = make([]byte, b.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, fmt.Errorf("crypto: generate nonce: %w", err)
	}
	ciphertext = b.aead.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// Decrypt reverses Encrypt. An error here (rather than a wrong-but-valid
// result) means the ciphertext, nonce, or key don't match — GCM
// authenticates the data, so tampering or using the wrong key/nonce is
// always detected, never silently decrypted into garbage.
func (b *SecretBox) Decrypt(ciphertext, nonce []byte) ([]byte, error) {
	plaintext, err := b.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("crypto: decrypt: %w", err)
	}
	return plaintext, nil
}

// DecodeKey parses the base64 form LLM_TOKEN_ENCRYPTION_KEY is set as
// (generate one with `openssl rand -base64 32`) into the raw 32-byte key
// NewSecretBox needs.
func DecodeKey(encoded string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("crypto: decode key: %w", err)
	}
	if len(key) != 32 {
		return nil, ErrInvalidKeySize
	}
	return key, nil
}
