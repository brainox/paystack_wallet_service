package paystack

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/brainox/paystack_wallet_service/external/external_models"
)

const (
	PaystackBaseURL = "https://api.paystack.co"
)

type PaystackService struct {
	secretKey string
	client    *http.Client
}

func NewPaystackService(secretKey string) *PaystackService {
	return &PaystackService{
		secretKey: secretKey,
		client:    &http.Client{},
	}
}

// InitializeTransaction initializes a Paystack transaction
func (s *PaystackService) InitializeTransaction(email string, amount int, reference string) (*external_models.InitializeTransactionResponse, error) {
	url := fmt.Sprintf("%s/transaction/initialize", PaystackBaseURL)

	payload := external_models.InitializeTransactionRequest{
		Amount:    amount, // Amount in kobo (smallest currency unit)
		Email:     email,
		Reference: reference,
		Currency:  "NGN",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.secretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("paystack error: %s", string(body))
	}

	var result external_models.InitializeTransactionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !result.Status {
		return nil, fmt.Errorf("paystack returned error: %s", result.Message)
	}

	return &result, nil
}

// VerifyTransaction verifies a Paystack transaction
func (s *PaystackService) VerifyTransaction(reference string) (*external_models.VerifyTransactionResponse, error) {
	url := fmt.Sprintf("%s/transaction/verify/%s", PaystackBaseURL, reference)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.secretKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("paystack error: %s", string(body))
	}

	var result external_models.VerifyTransactionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !result.Status {
		return nil, fmt.Errorf("paystack returned error: %s", result.Message)
	}

	return &result, nil
}

// ValidateWebhookSignature validates the Paystack webhook signature
func (s *PaystackService) ValidateWebhookSignature(body []byte, signature string) bool {
	hash := hmac.New(sha512.New, []byte(s.secretKey))
	hash.Write(body)
	expectedSignature := hex.EncodeToString(hash.Sum(nil))
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}
