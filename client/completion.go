package client

import (
	"context"
	"fmt"
	"net/http"
)

// CompletionService handles communication with the legacy completion related methods
// of the OpenRouter API. This endpoint is deprecated.
type CompletionService service // Use type alias as established in client.go

// Create sends a request to generate a text completion based on the provided prompt.
// This method interacts with the DEPRECATED /api/v1/completions endpoint.
// Prefer using the ChatService for modern models and features.
func (s *CompletionService) Create(ctx context.Context, request CompletionRequest) (response CompletionResponse, err error) {
	// Check if the required fields are present
	if request.Model == "" {
		err = fmt.Errorf("model is required for legacy completion request")
		return
	}
	if request.Prompt == "" {
		err = fmt.Errorf("prompt is required for legacy completion request")
		return
	}
	// Warn user about deprecation? Optionally add logging here.
	// log.Println("Warning: Using deprecated /v1/completions endpoint.")

	// Explicitly set stream to false if the user somehow set it,
	// as this implementation doesn't support streaming for the legacy endpoint.
	request.Stream = false

	path := "completions"
	err = s.client.doRequest(ctx, http.MethodPost, path, request, &response, false) // false = use standard API key
	return
}

// Note: Streaming for the legacy /completions endpoint is not explicitly
// documented in the provided text and is generally less common than chat streaming.
// Therefore, a CreateStream equivalent is omitted here. If streaming is needed
// for this endpoint, similar logic to ChatService.CreateStream would be required,
// assuming the API supports SSE for it.
