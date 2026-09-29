package email

import (
	"strings"
	"testing"

	"perchly-backend/internal/apierror"
)

func TestBuildVerificationEmail(t *testing.T) {
	subject, htmlBody, err := buildVerificationEmail("482913", apierror.LocaleTR)
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

func TestBuildVerificationEmail_PerLocale(t *testing.T) {
	tests := []struct {
		locale      apierror.Locale
		wantSubject string
		wantInBody  string
	}{
		{apierror.LocaleTR, "Perchly giriş kodun", "10 dakika"},
		{apierror.LocaleEN, "Your Perchly login code", "10 minutes"},
		{apierror.LocaleDE, "Dein Perchly-Anmeldecode", "10 Minuten"},
		{apierror.LocaleES, "Tu código de acceso a Perchly", "10 minutos"},
		{apierror.LocaleFR, "Ton code de connexion Perchly", "10 minutes"},
		{apierror.LocaleRU, "Твой код входа в Perchly", "10 минут"},
		{apierror.LocaleZH, "你的 Perchly 登录验证码", "10 分钟"},
		{apierror.LocaleAR, "رمز الدخول الخاص بك في Perchly", "10 دقائق"},
	}

	for _, tt := range tests {
		t.Run(string(tt.locale), func(t *testing.T) {
			subject, htmlBody, err := buildVerificationEmail("111222", tt.locale)
			if err != nil {
				t.Fatalf("buildVerificationEmail: %v", err)
			}
			if subject != tt.wantSubject {
				t.Errorf("subject = %q, want %q", subject, tt.wantSubject)
			}
			if !strings.Contains(htmlBody, tt.wantInBody) {
				t.Errorf("expected body to contain %q, got: %s", tt.wantInBody, htmlBody)
			}
			if !strings.Contains(htmlBody, "111222") {
				t.Errorf("expected the code to appear in the email body")
			}
		})
	}
}

func TestBuildVerificationEmail_UnknownLocaleFallsBackToTurkish(t *testing.T) {
	subject, htmlBody, err := buildVerificationEmail("999000", apierror.Locale("xx"))
	if err != nil {
		t.Fatalf("buildVerificationEmail: %v", err)
	}
	if subject != "Perchly giriş kodun" {
		t.Errorf("subject = %q, want Turkish fallback", subject)
	}
	if !strings.Contains(htmlBody, "10 dakika") {
		t.Errorf("expected Turkish fallback footnote in body, got: %s", htmlBody)
	}
}
