package email

import (
	"context"
	"fmt"
	"log"

	"gopkg.in/gomail.v2"
)

// SMTPSender implements Sender over plain SMTP (e.g. Gmail with an App
// Password) via gomail.v2.
type SMTPSender struct {
	host     string
	port     int
	username string
	fromName string
	dialer   *gomail.Dialer
}

func NewSMTPSender(host string, port int, username, password, fromName string) *SMTPSender {
	return &SMTPSender{
		host:     host,
		port:     port,
		username: username,
		fromName: fromName,
		dialer:   gomail.NewDialer(host, port, username, password),
	}
}

// SendVerificationCode sends the 6-digit code email. On any SMTP failure
// (connection, auth, etc.) it logs the real error server-side and
// returns a generic error — callers must never surface SMTP details to
// the end user.
func (s *SMTPSender) SendVerificationCode(ctx context.Context, toEmail, code string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	subject, htmlBody, err := buildVerificationEmail(code)
	if err != nil {
		return fmt.Errorf("build verification email: %w", err)
	}

	m := gomail.NewMessage()
	m.SetAddressHeader("From", s.username, s.fromName)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)

	if err := s.dialer.DialAndSend(m); err != nil {
		log.Printf("email: SMTP send failed (host=%s:%d): %v", s.host, s.port, err)
		return fmt.Errorf("send verification email: delivery failed")
	}
	return nil
}
