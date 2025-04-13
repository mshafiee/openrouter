package client

import (
	"context"
	"fmt"
	"net/http"
)

// AuthService handles communication with the authentication related methods
// of the OpenRouter API, specifically the PKCE flow.
type AuthService service // Use type alias as established in client.go

// ExchangeCode performs the final step of the OAuth PKCE flow.
// It exchanges the authorization code (obtained from the redirect URL after user approval)
// for a user-controlled OpenRouter API key.
// This request does NOT require prior authentication (no Bearer token is sent).
// Request and response structs are defined in types_auth.go.
func (s *AuthService) ExchangeCode(ctx context.Context, request ExchangeCodeRequest) (response ExchangeCodeResponse, err error) {
	if request.Code == "" {
		err = fmt.Errorf("authorization code cannot be empty")
		return
	}
	if request.CodeVerifier != "" && request.CodeChallengeMethod != "S256" && request.CodeChallengeMethod != "plain" && request.CodeChallengeMethod != "" {
		err = fmt.Errorf("invalid code_challenge_method: %q, must be 'S256', 'plain', or empty", request.CodeChallengeMethod)
		return
	}
	if request.CodeVerifier == "" && request.CodeChallengeMethod != "" {
		err = fmt.Errorf("code_challenge_method (%q) provided without code_verifier", request.CodeChallengeMethod)
		return
	}

	path := "auth/keys"
	// Use the special unauthenticated request helper from client.go
	err = s.client.doUnauthenticatedRequest(ctx, http.MethodPost, path, request, &response)
	return
}
