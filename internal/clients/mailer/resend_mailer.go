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

func (m *ResendMailer) SendVoucherEmail(ctx context.Context, toEmail, voucherUrl string) error {
    html := fmt.Sprintf(`
        <div style="font-family: system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;">
            <h2>Ganaste un voucher - %s</h2>
            <p>Por cargar tus comprobantes ganaste un voucher para canjear por un pedido gratis.</p>

            <!-- Imagen clickeable -->
            <div style="margin: 16px 0; text-align: center;">
                <a href="%s" target="_blank" rel="noopener noreferrer">
                    <img src="%s" alt="Tu voucher"
                        style="max-width: 260px; height: auto; display: block; margin: 0 auto;" />
                </a>
            </div>

            <!-- Link separado para abrir en grande -->
            <p style="text-align: center; margin-top: 8px;">
                <a href="%s" target="_blank" rel="noopener noreferrer"
                    style="display: inline-block; padding: 8px 14px;
                           background: #8B003A; color: #ffffff; text-decoration: none;
                           border-radius: 6px; font-weight: 600; font-size: 14px;">
                    Ver voucher en pantalla completa
                </a>
            </p>
        </div>
    `, m.appName, voucherUrl, voucherUrl, voucherUrl)

    params := &resend.SendEmailRequest{
        From:    "no-reply@powermixstation.com.ar",
        To:      []string{toEmail},
        Subject: "¡Ganaste un voucher!",
        Html:    html,
    }

    _, err := m.client.Emails.Send(params)
    return err
}


