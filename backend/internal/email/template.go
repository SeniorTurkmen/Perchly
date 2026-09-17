package email

import (
	"bytes"
	"fmt"
	"html/template"
)

const verificationSubject = "Perchly giriş kodun"

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
                Merhaba, Perchly'e giriş yapmak için aşağıdaki kodu kullan.
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
                Bu kod 10 dakika geçerli. Bu isteği sen yapmadıysan bu e-postayı yok sayabilirsin.
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
	Code string
}

// buildVerificationEmail renders the subject and HTML body for a
// verification-code email. Kept separate from any SMTP/dialing code so
// the template itself is testable without a network connection.
func buildVerificationEmail(code string) (subject, htmlBody string, err error) {
	var buf bytes.Buffer
	if err := verificationHTMLTemplate.Execute(&buf, verificationTemplateData{Code: code}); err != nil {
		return "", "", fmt.Errorf("render verification email template: %w", err)
	}
	return verificationSubject, buf.String(), nil
}
