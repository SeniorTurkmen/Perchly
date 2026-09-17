package model

import "time"

type EmailVerificationCode struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	Code         string     `json:"-"`
	ExpiresAt    time.Time  `json:"expires_at"`
	AttemptCount int        `json:"-"`
	UsedAt       *time.Time `json:"-"`
	CreatedAt    time.Time  `json:"created_at"`
}
