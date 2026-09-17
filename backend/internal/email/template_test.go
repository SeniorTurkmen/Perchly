package email

import (
	"strings"
	"testing"
)

func TestBuildVerificationEmail(t *testing.T) {
	subject, htmlBody, err := buildVerificationEmail("482913")
	if err != nil {
		t.Fatalf("buildVerificationEmail: %v", err)
	}

	if subject == "" {
		t.Error("expected a non-empty subject")
	}
	if !strings.Contains(htmlBody, "482913") {
		t.Errorf("expected the code to appear in the email body, got: %s", htmlBody)
	}
	if !strings.Contains(htmlBody, "Perchly") {
		t.Errorf("expected the Perchly brand name in the email body")
	}
	if !strings.Contains(htmlBody, "10 dakika") {
		t.Errorf("expected the 10-minute validity notice in the email body")
	}
}
