package mercadopago

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"
)

// parsea fecha/hora ingresada como hora local AR
func parseUserDateTime(dateStr, timeStr string) (time.Time, error) {
	// "12.11" -> "12:11"
	timeStr = strings.ReplaceAll(timeStr, ".", ":")

	loc, err := time.LoadLocation("America/Argentina/Buenos_Aires")
	if err != nil {
		return time.Time{}, err
	}

	layout := "02/01/2006 15:04" // "18/11/2025 12:11"
	return time.ParseInLocation(layout, dateStr+" "+timeStr, loc)
}

func buildWindow(t time.Time, mins int) (time.Time, time.Time) {
	d := time.Duration(mins) * time.Minute
	return t.Add(-d), t.Add(d)
}

func extractDNI(p *MercadoPagoPayment) *string {
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

// ReconcileOthers implementa la opción 2: otros bancos/billeteras
func (c *Client) ReconcileOthers(
	ctx context.Context,
	req ReconcileOthersRequest,
) (*ReconcileOthersResult, error) {

	if strings.TrimSpace(req.Date) == "" ||
		strings.TrimSpace(req.Time) == "" ||
		req.Amount <= 0 {
		return nil, fmt.Errorf("fecha, hora y monto son obligatorios")
	}

	tLocal, err := parseUserDateTime(req.Date, req.Time)
	if err != nil {
		return nil, fmt.Errorf("formato fecha/hora inválido: %w", err)
	}

	begin, end := buildWindow(tLocal, 10)

	payments, err := c.searchPaymentsInWindow(ctx, begin, end, req.Amount)
	if err != nil {
		return nil, err
	}

	loc, _ := time.LoadLocation("America/Argentina/Buenos_Aires")

	var candidates []MercadoPagoPayment

	for _, p := range payments {
		// monto (por las dudas chequeamos los dos)
		if math.Abs(p.TransactionDetails.TotalPaidAmount-req.Amount) > 0.01 &&
			math.Abs(p.TransactionAmount-req.Amount) > 0.01 {
			continue
		}

		// ultimos 4 de tarjeta
		if req.Last4 != nil && *req.Last4 != "" {
			last4 := extractCardLast4(&p)
			if last4 == nil || *last4 != *req.Last4 {
				continue
			}
		}

		// DNI
		if req.DNI != nil && *req.DNI != "" {
			dni := extractDNI(&p)
			if dni == nil || *dni != *req.DNI {
				continue
			}
		}

		// tiempo (usamos date_approved si está, sino date_created)
		var tMP time.Time
		if !p.DateApproved.IsZero() {
			tMP = p.DateApproved
		} else {
			tMP = p.DateCreated
		}

		tMPLocal := tMP.In(loc)
		diff := tMPLocal.Sub(tLocal)
		if diff < -15*time.Minute || diff > 15*time.Minute {
			continue
		}

		candidates = append(candidates, p)
	}

	if len(candidates) == 0 {
		return nil, nil // no encontrado
	}
	if len(candidates) > 1 {
		return nil, fmt.Errorf("se encontraron múltiples pagos que podrían matchear el comprobante")
	}

	p := candidates[0]

	return &ReconcileOthersResult{
		PaymentID:       p.ID,
		Status:          p.Status,
		TotalPaidAmount: p.TransactionDetails.TotalPaidAmount,
		DateApproved:    p.DateApproved,
		PayerEmail:      p.Payer.Email,
		PayerDNI:        extractDNI(&p),
		CardLast4:       extractCardLast4(&p),
	}, nil
}
