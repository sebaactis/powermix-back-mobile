package mercadopago

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	Token  string
	Client *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		Token:  token,
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

// ---------- Opción 1: validar pago por ID ----------

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

// ---------- Search por ventana de fecha/hora + monto (para opción 2) ----------

type mpSearchResponse struct {
	Results []MercadoPagoPayment `json:"results"`
	Paging  struct {
		Total  int `json:"total"`
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	} `json:"paging"`
}

func (c *Client) searchPaymentsInWindow(
	ctx context.Context,
	begin, end time.Time,
	amount float64,
) ([]MercadoPagoPayment, error) {

	baseURL := "https://api.mercadopago.com/v1/payments/search"

	beginStr := begin.Format(time.RFC3339)
	endStr := end.Format(time.RFC3339)

	limit := 50
	offset := 0
	maxPages := 5

	var all []MercadoPagoPayment

	for page := 0; page < maxPages; page++ {
		params := url.Values{}
		params.Add("range", "date_created")
		params.Add("begin_date", beginStr)
		params.Add("end_date", endStr)
		params.Add("status", "approved")
		params.Add("sort", "date_created")
		params.Add("criteria", "desc")
		params.Add("limit", fmt.Sprintf("%d", limit))
		params.Add("offset", fmt.Sprintf("%d", offset))
		params.Add("transaction_amount", fmt.Sprintf("%.2f", amount))

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"?"+params.Encode(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.Token)

		resp, err := c.Client.Do(req)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("mercado pago search error %d: %s", resp.StatusCode, string(body))
		}

		var mpResp mpSearchResponse
		if err := json.Unmarshal(body, &mpResp); err != nil {
			return nil, err
		}

		all = append(all, mpResp.Results...)

		if len(mpResp.Results) < limit || offset+limit >= mpResp.Paging.Total {
			break
		}
		offset += limit
	}

	return all, nil
}
