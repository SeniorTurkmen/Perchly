// Package email sends transactional email. It's deliberately a thin
// interface (see Sender) over one concrete SMTP implementation today, so
// switching to a provider like Resend or SES later only means writing a
// new type that implements Sender and wiring it in cmd/api/main.go — no
// change to AuthService or anything that calls SendVerificationCode.
package email

import (
	"context"

	"perchly-backend/internal/apierror"
)

// Sender is anything that can deliver a verification-code email, in the
// caller's resolved locale (see apierror.LocaleFromContext).
type Sender interface {
	SendVerificationCode(ctx context.Context, toEmail, code string, locale apierror.Locale) error
}
