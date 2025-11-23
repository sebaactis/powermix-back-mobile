package mercadopago

import "time"

type AdditionalInfo struct {
	IPAddress  string `json:"ip_address"`
	TrackingID string `json:"tracking_id"`
}

type ChargesExecutionInfo struct {
	InternalExecution InternalExecution `json:"internal_execution"`
}

type InternalExecution struct {
	Date        time.Time `json:"date"`
	ExecutionID string    `json:"execution_id"`
}

type Identification struct {
	Number *string `json:"number"`
	Type   *string `json:"type"`
}

type Payer struct {
	ID             int64          `json:"id"`
	Email          *string        `json:"email"`
	FirstName      *string        `json:"first_name"`
	LastName       *string        `json:"last_name"`
	Identification Identification `json:"identification"`
}

type Collector struct {
	Email          *string        `json:"email"`
	FirstName      *string        `json:"first_name"`
	ID             int64          `json:"id"`
	Identification Identification `json:"identification"`
	LastName       *string        `json:"last_name"`
	OperatorID     *int64         `json:"operator_id"`
	Phone          *string        `json:"phone"`
}

type PaymentMethod struct {
	ID       string `json:"id"`
	IssuerID string `json:"issuer_id"`
	Type     string `json:"type"`
}

type ApplicationData struct {
	Name            *string `json:"name"`
	OperatingSystem *string `json:"operating_system"`
	Version         *string `json:"version"`
}

type BusinessInfo struct {
	Branch  string `json:"branch"`
	SubUnit string `json:"sub_unit"`
	Unit    string `json:"unit"`
}

type Location struct {
	Source  string `json:"source"`
	StateID string `json:"state_id"`
}

type BankInfoCollector struct {
	AccountHolderName *string `json:"account_holder_name"`
	AccountID         *string `json:"account_id"`
	LongName          *string `json:"long_name"`
	TransferAccountID *string `json:"transfer_account_id"`
}

type BankInfoPayer struct {
	AccountID         *string        `json:"account_id"`
	Branch            *string        `json:"branch"`
	ExternalAccountID *string        `json:"external_account_id"`
	ID                *string        `json:"id"`
	Identification    Identification `json:"identification"`
	IsEndConsumer     *bool          `json:"is_end_consumer"`
	LongName          *string        `json:"long_name"`
}

type BankInfo struct {
	Collector              BankInfoCollector `json:"collector"`
	IsSameBankAccountOwner *bool             `json:"is_same_bank_account_owner"`
	OriginBankID           *string           `json:"origin_bank_id"`
	OriginWalletID         *string           `json:"origin_wallet_id"`
	Payer                  BankInfoPayer     `json:"payer"`
}

type InfringementNotification struct {
	Status *string `json:"status"`
	Type   *string `json:"type"`
}

type TransactionData struct {
	BankInfo                 BankInfo                 `json:"bank_info"`
	BankTransferID           *string                  `json:"bank_transfer_id"`
	E2EID                    *string                  `json:"e2e_id"`
	FinancialInstitution     *string                  `json:"financial_institution"`
	InfringementNotification InfringementNotification `json:"infringement_notification"`
	MerchantCategoryCode     *string                  `json:"merchant_category_code"`
	QRCode                   *string                  `json:"qr_code"`
	TicketURL                *string                  `json:"ticket_url"`
	TransactionID            *string                  `json:"transaction_id"`
}

type PointOfInteraction struct {
	ApplicationData ApplicationData `json:"application_data"`
	BusinessInfo    BusinessInfo    `json:"business_info"`
	Location        Location        `json:"location"`
	SubType         string          `json:"sub_type"`
	TransactionData TransactionData `json:"transaction_data"`
	Type            string          `json:"type"`
}

type TransactionDetails struct {
	AcquirerReference        *string `json:"acquirer_reference"`
	BankTransferID           *string `json:"bank_transfer_id"`
	ExternalResourceURL      *string `json:"external_resource_url"`
	FinancialInstitution     *string `json:"financial_institution"`
	InstallmentAmount        float64 `json:"installment_amount"`
	NetReceivedAmount        float64 `json:"net_received_amount"`
	OverpaidAmount           float64 `json:"overpaid_amount"`
	PayableDeferralPeriod    *string `json:"payable_deferral_period"`
	PaymentMethodReferenceID *string `json:"payment_method_reference_id"`
	TotalPaidAmount          float64 `json:"total_paid_amount"`
	TransactionID            *string `json:"transaction_id"`
}

// Card info tipado
type CardInfo struct {
	FirstSixDigits *string `json:"first_six_digits"`
	LastFourDigits *string `json:"last_four_digits"`
}

type MercadoPagoPayment struct {
	AccountsInfo           interface{}            `json:"accounts_info"`
	AcquirerReconciliation []interface{}          `json:"acquirer_reconciliation"`
	AdditionalInfo         AdditionalInfo         `json:"additional_info"`
	AuthorizationCode      *string                `json:"authorization_code"`
	BinaryMode             bool                   `json:"binary_mode"`
	BrandID                *string                `json:"brand_id"`
	BuildVersion           string                 `json:"build_version"`
	CallForAuthorizeID     *string                `json:"call_for_authorize_id"`
	Captured               bool                   `json:"captured"`
	Card                   CardInfo               `json:"card"`
	ChargesDetails         []interface{}          `json:"charges_details"`
	ChargesExecutionInfo   ChargesExecutionInfo   `json:"charges_execution_info"`
	Collector              Collector              `json:"collector"`
	CorporationID          *string                `json:"corporation_id"`
	CounterCurrency        *string                `json:"counter_currency"`
	CouponAmount           float64                `json:"coupon_amount"`
	CurrencyID             string                 `json:"currency_id"`
	DateApproved           time.Time              `json:"date_approved"`
	DateCreated            time.Time              `json:"date_created"`
	DateLastUpdated        time.Time              `json:"date_last_updated"`
	DateOfExpiration       *time.Time             `json:"date_of_expiration"`
	DeductionSchema        *string                `json:"deduction_schema"`
	Description            string                 `json:"description"`
	DifferentialPricingID  *string                `json:"differential_pricing_id"`
	ExternalReference      *string                `json:"external_reference"`
	FeeDetails             []interface{}          `json:"fee_details"`
	FinancingGroup         *string                `json:"financing_group"`
	ID                     int64                  `json:"id"`
	Installments           int                    `json:"installments"`
	IntegratorID           *string                `json:"integrator_id"`
	IssuerID               string                 `json:"issuer_id"`
	LiveMode               bool                   `json:"live_mode"`
	MarketplaceOwner       *int64                 `json:"marketplace_owner"`
	MerchantAccountID      *string                `json:"merchant_account_id"`
	MerchantNumber         *string                `json:"merchant_number"`
	Metadata               map[string]interface{} `json:"metadata"`
	MoneyReleaseDate       time.Time              `json:"money_release_date"`
	MoneyReleaseSchema     *string                `json:"money_release_schema"`
	MoneyReleaseStatus     string                 `json:"money_release_status"`
	NotificationURL        *string                `json:"notification_url"`
	OperationType          string                 `json:"operation_type"`
	Order                  map[string]interface{} `json:"order"`
	PayerID                int64                  `json:"payer_id"`
	Payer                  Payer                  `json:"payer"`
	PaymentMethod          PaymentMethod          `json:"payment_method"`
	PaymentMethodID        string                 `json:"payment_method_id"`
	PaymentTypeID          string                 `json:"payment_type_id"`
	PlatformID             *string                `json:"platform_id"`
	PointOfInteraction     PointOfInteraction     `json:"point_of_interaction"`
	PosID                  *string                `json:"pos_id"`
	ProcessingMode         string                 `json:"processing_mode"`
	Refunds                []interface{}          `json:"refunds"`
	ReleaseInfo            interface{}            `json:"release_info"`
	ShippingAmount         float64                `json:"shipping_amount"`
	SponsorID              *string                `json:"sponsor_id"`
	StatementDescriptor    *string                `json:"statement_descriptor"`
	Status                 string                 `json:"status"`
	StatusDetail           string                 `json:"status_detail"`
	StoreID                *string                `json:"store_id"`
	Tags                   *string                `json:"tags"`
	TaxesAmount            float64                `json:"taxes_amount"`
	TransactionAmount      float64                `json:"transaction_amount"`
	TransactionAmountRefunded float64             `json:"transaction_amount_refunded"`
	TransactionDetails     TransactionDetails     `json:"transaction_details"`
}

// Para opción 1
type PaymentDTO struct {
	DateApproved    time.Time `json:"date_approved"`
	OperationType   string    `json:"operation_type"`
	Status          string    `json:"status"`
	TotalPaidAmount float64   `json:"total_paid_amount"`
}

func (mp *MercadoPagoPayment) ToDTO() *PaymentDTO {
	return &PaymentDTO{
		DateApproved:    mp.DateApproved,
		OperationType:   mp.OperationType,
		Status:          mp.Status,
		TotalPaidAmount: mp.TransactionDetails.TotalPaidAmount,
	}
}

// Para opción 2 (otros bancos / billeteras)
type ReconcileOthersRequest struct {
	Date   string  `json:"date"`   // "18/11/2025"
	Time   string  `json:"time"`   // "12:11" o "12.11"
	Amount float64 `json:"amount"` // 2890

	Last4 *string `json:"last4"` // ultimos 4 - opcional
	DNI   *string `json:"dni"`   // dni - opcional
}

type ReconcileOthersResult struct {
	PaymentID       int64     `json:"payment_id"`
	Status          string    `json:"status"`
	TotalPaidAmount float64   `json:"total_paid_amount"`
	DateApproved    time.Time `json:"date_approved"`
	PayerEmail      *string   `json:"payer_email"`
	PayerDNI        *string   `json:"payer_dni"`
	CardLast4       *string   `json:"card_last4"`
}
