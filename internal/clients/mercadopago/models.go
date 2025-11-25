package mercadopago

import "time"

type Identification struct {
	Number *string `json:"number"`
	Type   *string `json:"type"`
}

type Payer struct {
	ID             interface{}    `json:"id"`
	Email          *string        `json:"email"`
	FirstName      *string        `json:"first_name"`
	LastName       *string        `json:"last_name"`
	Identification Identification `json:"identification"`
}

type Cardholder struct {
	Identification Identification `json:"identification"`
	Name           *string        `json:"name"`
}

type CardInfo struct {
	FirstSixDigits *string     `json:"first_six_digits"`
	LastFourDigits *string     `json:"last_four_digits"`
	Cardholder     *Cardholder `json:"cardholder"`
}

type BankInfoPayer struct {
	Identification Identification `json:"identification"`
}

type BankInfo struct {
	Payer BankInfoPayer `json:"payer"`
}

type TransactionData struct {
	BankInfo BankInfo `json:"bank_info"`
}

type PointOfInteraction struct {
	TransactionData TransactionData `json:"transaction_data"`
}

type TransactionDetails struct {
	TotalPaidAmount float64 `json:"total_paid_amount"`
}

type MercadoPagoPayment struct {
	ID int64 `json:"id"`

	Status        string `json:"status"`
	OperationType string `json:"operation_type"`

	DateApproved *time.Time `json:"date_approved"`
	DateCreated  time.Time  `json:"date_created"`

	TransactionAmount  float64            `json:"transaction_amount"`
	TransactionDetails TransactionDetails `json:"transaction_details"`

	Payer Payer    `json:"payer"`
	Card  CardInfo `json:"card"`

	PointOfInteraction PointOfInteraction `json:"point_of_interaction"`

	PaymentMethodId string `json:"payment_method_id"`
	PaymentTypeId   string `json:"payment_type_id"`
}

type PaymentDTO struct {
	DateApproved    time.Time `json:"date_approved"`
	OperationType   string    `json:"operation_type"`
	Status          string    `json:"status"`
	TotalPaidAmount float64   `json:"total_paid_amount"`
}

func (mp *MercadoPagoPayment) ToDTO() *PaymentDTO {
	t := mp.DateCreated
	if mp.DateApproved != nil {
		t = *mp.DateApproved
	}

	return &PaymentDTO{
		DateApproved:    t,
		OperationType:   mp.OperationType,
		Status:          mp.Status,
		TotalPaidAmount: mp.TransactionDetails.TotalPaidAmount,
	}
}

type ReconcileOthersRequest struct {
	Date   string  `json:"date"` // "18/11/2025"
	Time   string  `json:"time"` // "12:11" o "12.11"
	Amount float64 `json:"amount"`

	Last4 *string `json:"last4"`
	DNI   *string `json:"dni"`
}

type ReconcileOthersResult struct {
	PaymentID       int64     `json:"payment_id"`
	Status          string    `json:"status"`
	TotalPaidAmount float64   `json:"total_paid_amount"`
	OperationType   string    `json:"operation_type"`
	DateApproved    time.Time `json:"date_approved"`
	PayerEmail      *string   `json:"payer_email"`
	PayerDNI        *string   `json:"payer_dni"`
	CardLast4       *string   `json:"card_last4"`
	CardId          *string   `json:"card_id"`
	CardType        *string   `json:"card_type"`
}
