package external_models

import "time"

// Paystack API Models

// InitializeTransactionRequest represents the request to initialize a Paystack transaction
type InitializeTransactionRequest struct {
	Amount      int                    `json:"amount"`                 // Amount in kobo
	Email       string                 `json:"email"`                  // Customer email
	Reference   string                 `json:"reference"`              // Unique transaction reference
	CallbackURL string                 `json:"callback_url,omitempty"` // Optional callback URL
	Currency    string                 `json:"currency"`               // Currency (NGN, USD, etc.)
	Metadata    map[string]interface{} `json:"metadata,omitempty"`     // Optional metadata
}

// InitializeTransactionResponse represents the response from Paystack transaction initialization
type InitializeTransactionResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

// VerifyTransactionResponse represents the response from Paystack transaction verification
type VerifyTransactionResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ID              int64     `json:"id"`
		Domain          string    `json:"domain"`
		Status          string    `json:"status"`
		Reference       string    `json:"reference"`
		Amount          int       `json:"amount"`
		Message         string    `json:"message"`
		GatewayResponse string    `json:"gateway_response"`
		PaidAt          time.Time `json:"paid_at"`
		CreatedAt       time.Time `json:"created_at"`
		Channel         string    `json:"channel"`
		Currency        string    `json:"currency"`
		IPAddress       string    `json:"ip_address"`
		Metadata        string    `json:"metadata"`
		Customer        struct {
			ID           int    `json:"id"`
			Email        string `json:"email"`
			CustomerCode string `json:"customer_code"`
		} `json:"customer"`
		Authorization struct {
			AuthorizationCode string `json:"authorization_code"`
			Bin               string `json:"bin"`
			Last4             string `json:"last4"`
			ExpMonth          string `json:"exp_month"`
			ExpYear           string `json:"exp_year"`
			Channel           string `json:"channel"`
			CardType          string `json:"card_type"`
			Bank              string `json:"bank"`
			CountryCode       string `json:"country_code"`
			Brand             string `json:"brand"`
		} `json:"authorization"`
	} `json:"data"`
}

// WebhookEvent represents a Paystack webhook event
type WebhookEvent struct {
	Event string                 `json:"event"`
	Data  WebhookTransactionData `json:"data"`
}

// WebhookTransactionData represents transaction data in a webhook event
type WebhookTransactionData struct {
	ID              int64     `json:"id"`
	Domain          string    `json:"domain"`
	Status          string    `json:"status"`
	Reference       string    `json:"reference"`
	Amount          int       `json:"amount"`
	Message         string    `json:"message"`
	GatewayResponse string    `json:"gateway_response"`
	PaidAt          time.Time `json:"paid_at"`
	CreatedAt       time.Time `json:"created_at"`
	Channel         string    `json:"channel"`
	Currency        string    `json:"currency"`
	IPAddress       string    `json:"ip_address"`
	Metadata        string    `json:"metadata"`
	Customer        struct {
		ID           int    `json:"id"`
		Email        string `json:"email"`
		CustomerCode string `json:"customer_code"`
	} `json:"customer"`
}
