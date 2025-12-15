package mailer

import "context"

type Mailer interface {
	SendResetPasswordEmail(ctx context.Context, toEmail, resetURL string) error
	SendVoucherEmail(ctx context.Context, toEmail, voucherUrl string) error
}
