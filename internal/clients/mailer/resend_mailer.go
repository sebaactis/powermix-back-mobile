// internal/platform/mailer/resend_mailer.go
package mailer

import (
	"context"
	"fmt"

	resend "github.com/resend/resend-go/v2"
)

type ResendMailer struct {
	client *resend.Client
	from   string
	appName string
}

func NewResendMailer(apiKey, from, appName string) *ResendMailer {
	client := resend.NewClient(apiKey)
	return &ResendMailer{
		client:  client,
		from:    from,    
		appName: appName,
	}
}

func (m *ResendMailer) SendResetPasswordEmail(ctx context.Context, toEmail, resetURL string) error {
	html := fmt.Sprintf(`
		<div style="font-family: system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;">
			<h2>Recuperar contraseña - %s</h2>
			<p>Recibimos un pedido para restablecer tu contraseña.</p>
			<p>Si fuiste vos, hacé clic en el siguiente botón:</p>
			<p>
				<a href="%s" style="display:inline-block;padding:10px 18px;background:#8B003A;color:#ffffff;text-decoration:none;border-radius:6px;font-weight:600;">
					Restablecer contraseña
				</a>
			</p>
			<p>Si no pediste esto, podés ignorar este correo.</p>
		</div>
	`, m.appName, resetURL)

	params := &resend.SendEmailRequest{
		From:    "no-reply@powermixstation.com.ar",
		To:      []string{toEmail},
		Subject: "Recuperar contraseña",
		Html:    html,
	}

	_, err := m.client.Emails.Send(params)
	return err
}
