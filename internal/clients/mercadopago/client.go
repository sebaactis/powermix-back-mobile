package mercadopago

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
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

type mpSearchResponse struct {
	Results []MercadoPagoPayment `json:"results"`
	Paging  struct {
		Total  int `json:"total"`
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	} `json:"paging"`
}

// Metodo para buscar un pago solo con ID MP
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

// Metodo para buscar un pago con datos varios si no tenemos ID MP (pago por otras billeteras por ej)
func (c *Client) ReconcileOthers(
	ctx context.Context,
	req ReconcileOthersRequest,
) (*ReconcileOthersResult, error) {

	if strings.TrimSpace(req.Date) == "" ||
		strings.TrimSpace(req.Time) == "" ||
		req.Amount <= 0 {
		return nil, fmt.Errorf("la fecha, hora y monto del comprobante son datos obligatorios")
	}

	tLocal, err := parseUserDateTime(req.Date, req.Time)
	if err != nil {
		return nil, fmt.Errorf("formato fecha/hora inválido: %w", err)
	}

	begin, end := buildWindow(tLocal, 2)

	payments, err := c.searchPaymentsInWindow(ctx, begin, end, req.Amount)
	if err != nil {
		return nil, err
	}

	loc, _ := time.LoadLocation("America/Argentina/Buenos_Aires")

	var candidates []MercadoPagoPayment

	for _, p := range payments {

		if math.Abs(p.TransactionDetails.TotalPaidAmount-req.Amount) > 0.01 &&
			math.Abs(p.TransactionAmount-req.Amount) > 0.01 {
			continue
		}

		if req.Last4 != nil && *req.Last4 != "" {
			last4 := extractCardLast4(&p)
			if last4 == nil || *last4 != *req.Last4 {
				continue
			}
		}

		if req.DNI != nil && *req.DNI != "" {
			dni := extractDNI(&p)
			if dni == nil || *dni != *req.DNI {
				continue
			}
		}

		var tMP time.Time
		if !p.DateApproved.IsZero() {
			tMP = *p.DateApproved
		} else {
			tMP = p.DateCreated
		}

		tMPLocal := tMP.In(loc)
		diff := tMPLocal.Sub(tLocal)
		if diff < -3*time.Minute || diff > 3*time.Minute {
			continue
		}

		candidates = append(candidates, p)
	}

	if len(candidates) == 0 {
		return nil, nil
	}
	if len(candidates) > 1 {
		return nil, fmt.Errorf("se encontraron múltiples pagos que podrían matchear el comprobante, por favor, de ser posible, agregue mas información")
	}

	p := candidates[0]

	return &ReconcileOthersResult{
		PaymentID:       p.ID,
		Status:          p.Status,
		TotalPaidAmount: p.TransactionDetails.TotalPaidAmount,
		OperationType:   p.OperationType,
		DateApproved:    *p.DateApproved,
		PayerEmail:      p.Payer.Email,
		PayerDNI:        extractDNI(&p),
		CardLast4:       extractCardLast4(&p),
		CardId:          &p.PaymentMethodId,
		CardType:        &p.PaymentTypeId,
	}, nil
}

//************* FUNCIONES PRIVADAS DEL CLIENTE *************//

func (c *Client) searchPaymentsInWindow(
	ctx context.Context,
	begin, end time.Time,
	amount float64,
) ([]MercadoPagoPayment, error) {

	baseURL := "https://api.mercadopago.com/v1/payments/search"

	beginStr := formatMPDate(begin)
	endStr := formatMPDate(end)

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

func parseUserDateTime(dateStr, timeStr string) (time.Time, error) {

	timeStr = strings.ReplaceAll(timeStr, ".", ":")

	loc, err := time.LoadLocation("America/Argentina/Buenos_Aires")
	if err != nil {
		return time.Time{}, err
	}

	layout := "02/01/2006 15:04"
	return time.ParseInLocation(layout, dateStr+" "+timeStr, loc)
}

func buildWindow(t time.Time, mins int) (time.Time, time.Time) {
	d := time.Duration(mins) * time.Minute
	return t.Add(-d), t.Add(d)
}

func extractDNI(p *MercadoPagoPayment) *string {

	if p.Card.Cardholder != nil &&
		p.Card.Cardholder.Identification.Number != nil &&
		*p.Card.Cardholder.Identification.Number != "" {
		return p.Card.Cardholder.Identification.Number
	}

	if p.Payer.Identification.Number != nil && *p.Payer.Identification.Number != "" {
		return p.Payer.Identification.Number
	}

	if p.PointOfInteraction.TransactionData.BankInfo.Payer.Identification.Number != nil &&
		*p.PointOfInteraction.TransactionData.BankInfo.Payer.Identification.Number != "" {
		return p.PointOfInteraction.TransactionData.BankInfo.Payer.Identification.Number
	}

	return nil
}

func extractCardLast4(p *MercadoPagoPayment) *string {
	if p.Card.LastFourDigits != nil && *p.Card.LastFourDigits != "" {
		return p.Card.LastFourDigits
	}
	return nil
}

func formatMPDate(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}
