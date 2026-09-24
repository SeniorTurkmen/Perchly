package config

import (
	"crypto/rand"
	"encoding/hex"
)

// randomSecret returns a hex-encoded random secret of n bytes, used as a
// fallback JWT signing key when JWT_SECRET isn't set.
func randomSecret(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// randomKeyBytes returns n raw random bytes, used as a fallback
// LLM_TOKEN_ENCRYPTION_KEY when that env var isn't set.
func randomKeyBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}
