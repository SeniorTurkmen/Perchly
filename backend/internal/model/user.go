package model

import "time"

// User is either anonymous (device-identified, no email) or a verified
// account (email + email_verified_at set). An anonymous user upgrades in
// place to a verified one — same id, same conversations/quota history —
// the first time they successfully verify an email code without an
// existing verified account for that email (see AuthService, Durum B).
type User struct {
	ID              string     `json:"id"`
	Email           *string    `json:"email"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	DisplayName     *string    `json:"display_name"`
	IsAnonymous     bool       `json:"is_anonymous"`
	DeviceID        *string    `json:"-"`
	// Timezone is IANA (e.g. "Europe/Istanbul"); used only by the quota
	// system to reset daily limits at the user's local midnight.
	Timezone  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
