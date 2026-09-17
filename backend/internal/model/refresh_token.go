package model

import "time"

// RefreshToken is stored with only TokenHash — the raw token is handed
// to the client once, at issuance, and never persisted or logged.
type RefreshToken struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"-"`
	CreatedAt time.Time  `json:"created_at"`
}
