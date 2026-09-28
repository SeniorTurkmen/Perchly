package email

import (
	"context"
	"log"

	"perchly-backend/internal/apierror"
)

// ConsoleSender just logs the code instead of sending real email. Used
// automatically when SMTP isn't configured (see New in config wiring),
// so local dev/tests can exercise the full request-code -> verify-code
// flow without real Gmail credentials.
type ConsoleSender struct{}

func NewConsoleSender() *ConsoleSender {
	return &ConsoleSender{}
}

func (s *ConsoleSender) SendVerificationCode(_ context.Context, toEmail, code string, locale apierror.Locale) error {
	subject, htmlBody, err := buildVerificationEmail(code, locale)
	if err != nil {
		return err
	}
	log.Printf("email: [ConsoleSender] verification code for %s (locale=%s): %s\nsubject: %s\nbody:\n%s", toEmail, locale, code, subject, htmlBody)
	return nil
}
