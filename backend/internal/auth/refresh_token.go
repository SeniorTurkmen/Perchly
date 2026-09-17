package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// GenerateRefreshToken returns a new random opaque refresh token, plus
// its SHA-256 hash for storage. Only the hash is ever persisted — the
// raw value is returned to the client once, at issuance, and cannot be
// recovered from the stored hash.
func GenerateRefreshToken() (raw string, hash string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}
	raw = base64.RawURLEncoding.EncodeToString(buf)
	return raw, HashRefreshToken(raw), nil
}

// HashRefreshToken hashes a raw refresh token the same way
// GenerateRefreshToken does, so a presented token can be looked up by
// hash without ever storing (or needing to decrypt) the raw value.
func HashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
