package model

import "time"

// AdminSession is stored with only TokenHash — the raw token is handed
// to the admin dashboard once, at login, and never persisted or logged.
// Same shape as RefreshToken, scoped to AdminUser instead of User.
type AdminSession struct {
	ID          string     `json:"id"`
	AdminUserID string     `json:"admin_user_id"`
	TokenHash   string     `json:"-"`
	ExpiresAt   time.Time  `json:"expires_at"`
	RevokedAt   *time.Time `json:"-"`
	CreatedAt   time.Time  `json:"created_at"`
}
