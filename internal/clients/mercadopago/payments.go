package mercadopago

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) ValidatePaymentExists(ctx context.Context, paymentID string) (*PaymentDTO, error) {
	url := fmt.Sprintf("https://api.mercadopago.com/v1/payments/%s", paymentID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("mercado pago error: %s", string(body))
	}

	var payment MercadoPagoPayment
	if err := json.NewDecoder(resp.Body).Decode(&payment); err != nil {
		return nil, fmt.Errorf("error decoding payment: %w", err)
	}

	return payment.ToDTO(), nil
}
