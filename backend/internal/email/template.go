package email

import (
	"bytes"
	"fmt"
	"html/template"

	"perchly-backend/internal/apierror"
)

// verificationCopy is one locale's subject/intro/footnote for the
// verification-code email — {{.Code}} itself needs no translation, so
// only these three strings vary. Mirrors the map[apierror.Locale]struct{}
// + Turkish-fallback shape used across the rest of the codebase (see
// apierror.Message): Turkish is never a map entry, it's the literal
// fallback value below (verificationCopyTR).
type verificationCopy struct {
	Subject  string
	Intro    string
	Footnote string
}

var verificationCopyTR = verificationCopy{
	Subject:  "Perchly giriş kodun",
	Intro:    "Merhaba, Perchly'e giriş yapmak için aşağıdaki kodu kullan.",
	Footnote: "Bu kod 10 dakika geçerli. Bu isteği sen yapmadıysan bu e-postayı yok sayabilirsin.",
}

var verificationCopyByLocale = map[apierror.Locale]verificationCopy{
	apierror.LocaleEN: {
		Subject:  "Your Perchly login code",
		Intro:    "Hi there — use the code below to log in to Perchly.",
		Footnote: "This code is valid for 10 minutes. If you didn't request this, you can safely ignore this email.",
	},
	apierror.LocaleDE: {
		Subject:  "Dein Perchly-Anmeldecode",
		Intro:    "Hallo, verwende den folgenden Code, um dich bei Perchly anzumelden.",
		Footnote: "Dieser Code ist 10 Minuten lang gültig. Wenn du das nicht angefordert hast, kannst du diese E-Mail einfach ignorieren.",
	},
	apierror.LocaleES: {
		Subject:  "Tu código de acceso a Perchly",
		Intro:    "Hola, usa el siguiente código para iniciar sesión en Perchly.",
		Footnote: "Este código es válido durante 10 minutos. Si no solicitaste esto, puedes ignorar este correo.",
	},
	apierror.LocaleFR: {
		Subject:  "Ton code de connexion Perchly",
		Intro:    "Bonjour, utilise le code ci-dessous pour te connecter à Perchly.",
		Footnote: "Ce code est valable 10 minutes. Si tu n'es pas à l'origine de cette demande, tu peux ignorer cet e-mail.",
	},
	apierror.LocaleRU: {
		Subject:  "Твой код входа в Perchly",
		Intro:    "Привет! Используй код ниже, чтобы войти в Perchly.",
		Footnote: "Этот код действителен 10 минут. Если ты не запрашивал его, просто проигнорируй это письмо.",
	},
	apierror.LocaleZH: {
		Subject:  "你的 Perchly 登录验证码",
		Intro:    "你好,使用以下验证码登录 Perchly。",
		Footnote: "此验证码 10 分钟内有效。如果不是你本人发起的请求,可以忽略这封邮件。",
	},
	apierror.LocaleAR: {
		Subject:  "رمز الدخول الخاص بك في Perchly",
		Intro:    "مرحبًا، استخدم الرمز أدناه لتسجيل الدخول إلى Perchly.",
		Footnote: "هذا الرمز صالح لمدة 10 دقائق. إذا لم تطلب هذا الرمز، يمكنك تجاهل هذه الرسالة بأمان.",
	},
}

// verificationCopyFor resolves locale's subject/intro/footnote, falling
// back to Turkish for LocaleTR or any locale with no entry above — same
// fallback shape as apierror.Message.
func verificationCopyFor(locale apierror.Locale) verificationCopy {
	if locale == apierror.LocaleTR {
		return verificationCopyTR
	}
	if c, ok := verificationCopyByLocale[locale]; ok {
		return c
	}
	return verificationCopyTR
}

var verificationHTMLTemplate = template.Must(template.New("verification").Parse(`
<!DOCTYPE html>
<html>
<body style="margin:0;padding:0;background-color:#f5f4f2;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;">
  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="padding:40px 16px;">
    <tr>
      <td align="center">
        <table role="presentation" width="100%" style="max-width:420px;background-color:#ffffff;border-radius:20px;padding:40px 32px;">
          <tr>
            <td align="center" style="padding-bottom:8px;">
              <span style="font-size:20px;font-weight:600;color:#1a1a1a;">Perchly</span>
            </td>
          </tr>
          <tr>
            <td align="center" style="padding-bottom:24px;">
              <p style="margin:0;font-size:15px;line-height:1.5;color:#5a5a5a;">
                {{.Intro}}
              </p>
            </td>
          </tr>
          <tr>
            <td align="center" style="padding:20px 0;">
              <span style="display:inline-block;font-size:36px;font-weight:700;letter-spacing:8px;color:#1a1a1a;background-color:#f5f4f2;border-radius:14px;padding:16px 28px;">
                {{.Code}}
              </span>
            </td>
          </tr>
          <tr>
            <td align="center" style="padding-top:24px;">
              <p style="margin:0;font-size:13px;line-height:1.5;color:#8a8a8a;">
                {{.Footnote}}
              </p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>
`))

type verificationTemplateData struct {
	Code     string
	Intro    string
	Footnote string
}

// buildVerificationEmail renders the subject and HTML body for a
// verification-code email in locale. Kept separate from any SMTP/dialing
// code so the template itself is testable without a network connection.
func buildVerificationEmail(code string, locale apierror.Locale) (subject, htmlBody string, err error) {
	vc := verificationCopyFor(locale)

	var buf bytes.Buffer
	if err := verificationHTMLTemplate.Execute(&buf, verificationTemplateData{
		Code: code, Intro: vc.Intro, Footnote: vc.Footnote,
	}); err != nil {
		return "", "", fmt.Errorf("render verification email template: %w", err)
	}
	return vc.Subject, buf.String(), nil
}
