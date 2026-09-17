package email

import (
	"context"
	"log"
)

// ConsoleSender just logs the code instead of sending real email. Used
// automatically when SMTP isn't configured (see New in config wiring),
// so local dev/tests can exercise the full request-code -> verify-code
// flow without real Gmail credentials.
type ConsoleSender struct{}

func NewConsoleSender() *ConsoleSender {
	return &ConsoleSender{}
}

func (s *ConsoleSender) SendVerificationCode(_ context.Context, toEmail, code string) error {
	log.Printf("email: [ConsoleSender] verification code for %s: %s", toEmail, code)
	return nil
}
