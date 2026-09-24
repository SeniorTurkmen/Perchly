package model

import "time"

// AdminUser is an account for the internal admin dashboard — distinct
// from User (app end users). There's no self-serve signup; rows are
// created via the createadmin CLI (see cmd/createadmin).
type AdminUser struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}
