// internal/platform/mailer/resend_mailer.go
package mailer

import (
	"context"
	"fmt"
	"strings"
	"time"

	resend "github.com/resend/resend-go/v2"
)

type ResendMailer struct {
	client  *resend.Client
	from    string
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

func (m *ResendMailer) SendEmailContact(ctx context.Context, contactRequest *ContactRequest) error {
	esc := func(s string) string {
		replacer := strings.NewReplacer(
			"&", "&amp;",
			"<", "&lt;",
			">", "&gt;",
			`"`, "&quot;",
			"'", "&#39;",
		)
		return replacer.Replace(strings.TrimSpace(s))
	}

	name := esc(contactRequest.Name)
	email := esc(contactRequest.Email)
	category := esc(contactRequest.Category)
	message := esc(contactRequest.Message)

	// Colorcito según categoría (opcional)
	categoryColor := "#8B003A"
	switch strings.ToLower(contactRequest.Category) {
	case "pagos", "pago", "comprobante":
		categoryColor = "#0E7490"
	case "cuenta", "login", "acceso":
		categoryColor = "#6D28D9"
	case "voucher", "premio":
		categoryColor = "#16A34A"
	case "bug", "error", "problema":
		categoryColor = "#DC2626"
	}

	html := fmt.Sprintf(`
<!doctype html>
<html lang="es">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <meta name="color-scheme" content="light only" />
    <title>Nuevo contacto</title>
  </head>
  <body style="margin:0; padding:0; background:#F6F7FB;">
    <div style="display:none; max-height:0; overflow:hidden; opacity:0; color:transparent;">
      Nuevo mensaje de soporte (%s) - %s
    </div>

    <table role="presentation" cellpadding="0" cellspacing="0" border="0" width="100%%" style="background:#F6F7FB; padding:24px 12px;">
      <tr>
        <td align="center">
          <table role="presentation" cellpadding="0" cellspacing="0" border="0" width="600" style="width:600px; max-width:600px;">
            
            <!-- Header -->
            <tr>
              <td style="padding: 10px 6px 16px 6px;">
                <div style="font-family: system-ui,-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;">
                  <div style="font-weight:800; font-size:18px; color:#111827;">
                    %s
                  </div>
                  <div style="margin-top:6px; font-size:13px; color:#6B7280;">
                    Nuevo contacto desde el formulario de ayuda
                  </div>
                </div>
              </td>
            </tr>

            <!-- Card -->
            <tr>
              <td style="background:#FFFFFF; border:1px solid #E5E7EB; border-radius:16px; overflow:hidden;">
                
                <!-- Top bar -->
                <div style="padding:18px 18px 0 18px; font-family: system-ui,-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;">
                  <div style="display:inline-block; padding:6px 10px; border-radius:999px; background:%s; color:#FFFFFF; font-size:12px; font-weight:700;">
                    Categoría: %s
                  </div>
                </div>

                <!-- Content -->
                <div style="padding:18px; font-family: system-ui,-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif; color:#111827;">
                  <h2 style="margin:8px 0 10px 0; font-size:18px; line-height:1.25;">
                    Datos del contacto
                  </h2>

                  <table role="presentation" cellpadding="0" cellspacing="0" border="0" width="100%%" style="border-collapse:collapse; font-size:14px;">
                    <tr>
                      <td style="padding:10px 0; color:#6B7280; width:120px;">Nombre</td>
                      <td style="padding:10px 0; font-weight:700;">%s</td>
                    </tr>
                    <tr>
                      <td style="padding:10px 0; color:#6B7280; width:120px;">Email</td>
                      <td style="padding:10px 0;">
                        <a href="mailto:%s" style="color:%s; text-decoration:none; font-weight:700;">%s</a>
                      </td>
                    </tr>
                  </table>

                  <div style="margin-top:14px; padding:14px; border-radius:12px; background:#F9FAFB; border:1px solid #E5E7EB;">
                    <div style="font-size:12px; font-weight:800; color:#6B7280; letter-spacing:.02em;">
                      MENSAJE
                    </div>
                    <div style="margin-top:10px; font-size:14px; line-height:1.6; white-space:pre-wrap;">
                      %s
                    </div>
                  </div>

                  <div style="margin-top:16px; font-size:12px; color:#6B7280; line-height:1.5;">
                    Tip: respondé a este mail o escribile al usuario tocando su email.
                  </div>
                </div>

                <!-- Footer inside card -->
                <div style="padding:14px 18px; background:#111827;">
                  <div style="font-family: system-ui,-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif; font-size:12px; color:#D1D5DB;">
                    Enviado automáticamente por %s · %s
                  </div>
                </div>

              </td>
            </tr>

            <!-- Outer footer -->
            <tr>
              <td style="padding:14px 6px 0 6px; text-align:center;">
                <div style="font-family: system-ui,-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif; font-size:11px; color:#9CA3AF;">
                  Si no esperabas este correo, podés ignorarlo.
                </div>
              </td>
            </tr>

          </table>
        </td>
      </tr>
    </table>
  </body>
</html>
`, category, name, m.appName, categoryColor, category, name, email, categoryColor, email, message, m.appName, time.Now().Format("02/01/2006 15:04"))

	params := &resend.SendEmailRequest{
		From:    "no-reply@powermixstation.com.ar",
		To:      []string{"sebaactis@gmail.com"},
		Subject: "¡Ganaste un voucher!",
		Html:    html,
	}

	_, err := m.client.Emails.Send(params)
	return err
}
