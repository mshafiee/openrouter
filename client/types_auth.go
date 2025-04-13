package client

// This file defines the Go structs corresponding to the JSON request and response
// schemas for the OpenRouter Authentication API, specifically the OAuth PKCE
// code exchange endpoint (/v1/auth/keys).

// ExchangeCodeRequest represents the request body for exchanging an OAuth authorization code
// for an OpenRouter API key during the PKCE flow.
type ExchangeCodeRequest struct {
	// Code is the authorization code received from the OpenRouter redirect after user authentication.
	// This field is required.
	Code string `json:"code"`

	// CodeVerifier is the original plain text string used to generate the code_challenge
	// sent in the initial authorization request.
	// Required if code_challenge_method was 'S256' or 'plain'. Omit if no challenge was used.
	CodeVerifier string `json:"code_verifier,omitempty"`

	// CodeChallengeMethod indicates the method used to generate the code_challenge ('S256' or 'plain').
	// Required if code_challenge was used in the initial authorization request. Omit otherwise.
	CodeChallengeMethod string `json:"code_challenge_method,omitempty"` // Valid values: "S256", "plain"
}

// ExchangeCodeResponse represents the successful response body after exchanging the code.
// It contains the newly generated user-controlled API key.
type ExchangeCodeResponse struct {
	// Key is the OpenRouter API key obtained through the code exchange.
	// This key should be stored securely and used for subsequent API requests.
	Key string `json:"key"`

	// UserID is the identifier for the user associated with the API key.
	// This field is optional in the response.
	UserID string `json:"user_id,omitempty"`
}

// Note: The actual API call logic for this endpoint is implemented in auth.go,
// specifically in the AuthService.ExchangeCode method.
// That method uses these request and response structs.
