package client

import (
	"context"
	"fmt"
	"net/http"
)

// BillingService handles communication with the billing and credits related methods
// of the OpenRouter API.
type BillingService service // Use type alias as established in client.go

// GetCredits retrieves the total credits purchased and used for the authenticated user account.
// Note that these values may be cached and up to 60 seconds stale.
// Response struct CreditsGetResponse defined below. Core data struct CreditsData defined in types_billing.go.
func (s *BillingService) GetCredits(ctx context.Context) (response CreditsGetResponse, err error) {
	path := "credits"
	err = s.client.doRequest(ctx, http.MethodGet, path, nil, &response, false)
	return
}

// CreateCoinbaseCharge initiates a cryptocurrency payment process via Coinbase Commerce.
// It creates a charge based on the requested USD amount and returns the necessary
// transaction details (like calldata) required to fulfill the payment on-chain.
// Request struct CoinbaseChargeRequest and core response data CoinbaseChargeResponseData defined in types_billing.go.
// Response wrapper CoinbaseChargeResponse defined below.
func (s *BillingService) CreateCoinbaseCharge(ctx context.Context, request CoinbaseChargeRequest) (response CoinbaseChargeResponse, err error) {
	if request.Amount <= 0 {
		err = fmt.Errorf("charge amount must be positive")
		return
	}
	if request.Sender == "" {
		err = fmt.Errorf("sender address is required")
		return
	}
	// Add validation for supported chain IDs if needed

	path := "credits/coinbase"
	err = s.client.doRequest(ctx, http.MethodPost, path, request, &response, false) // false = use standard API key
	return
}

// CreditsGetResponse wraps the CreditsData in a 'data' field.
type CreditsGetResponse struct {
	Data CreditsData `json:"data"` // CreditsData is defined in types_billing.go
}

// CoinbaseChargeResponse wraps the CoinbaseChargeResponseData in a 'data' field.
type CoinbaseChargeResponse struct {
	Data CoinbaseChargeResponseData `json:"data"` // CoinbaseChargeResponseData is defined in types_billing.go
}
