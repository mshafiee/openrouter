package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// GenerationService handles communication with the generation metadata related methods
// of the OpenRouter API.
type GenerationService service // Use type alias as established in client.go

// Get retrieves detailed metadata about a specific generation request using its ID.
// This includes cost, token counts (native and normalized), provider details, etc.
func (s *GenerationService) Get(ctx context.Context, generationID string) (response GenerationGetResponse, err error) {
	if generationID == "" {
		err = fmt.Errorf("generation ID cannot be empty")
		return
	}

	// Construct the path with the query parameter
	path := fmt.Sprintf("/generation?id=%s", url.QueryEscape(generationID))

	err = s.client.doRequest(ctx, http.MethodGet, path, nil, &response, false) // nil request body for GET, false = use standard API key
	return
}

// GenerationGetResponse wraps the actual GenerationData in a 'data' field,
// matching the API response structure.
type GenerationGetResponse struct {
	Data GenerationData `json:"data"`
}
