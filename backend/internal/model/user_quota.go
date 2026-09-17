package model

import "time"

// UserQuota is one user's daily message allowance for one persona.
// DailyLimit is per-row (not a global constant) specifically so a future
// admin dashboard can grant an individual user/persona a custom limit by
// updating this column directly.
type UserQuota struct {
	UserID            string    `json:"user_id"`
	PersonaID         string    `json:"persona_id"`
	MessageCountToday int       `json:"message_count_today"`
	DailyLimit        int       `json:"daily_limit"`
	LastResetAt       time.Time `json:"last_reset_at"`
}
