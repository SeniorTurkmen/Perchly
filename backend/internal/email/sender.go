// Package email sends transactional email. It's deliberately a thin
// interface (see Sender) over one concrete SMTP implementation today, so
// switching to a provider like Resend or SES later only means writing a
// new type that implements Sender and wiring it in cmd/api/main.go — no
// change to AuthService or anything that calls SendVerificationCode.
package email

import "context"

// Sender is anything that can deliver a verification-code email.
type Sender interface {
	SendVerificationCode(ctx context.Context, toEmail, code string) error
}
