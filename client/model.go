package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ModelService handles communication with the model listing related methods
// of the OpenRouter API.
type ModelService service // Use type alias as established in client.go

// List retrieves a list of all models available through the OpenRouter API.
func (s *ModelService) List(ctx context.Context) (response ModelListResponse, err error) {
	path := "models"
	err = s.client.doRequest(ctx, http.MethodGet, path, nil, &response, false) // nil request body for GET, false = use standard API key
	return
}

// ListEndpoints retrieves details about a specific model, including the providers (endpoints)
// that serve it and their specific configurations (pricing, context length, supported parameters).
// Requires the model author (e.g., "openai") and slug (e.g., "gpt-4o").
func (s *ModelService) ListEndpoints(ctx context.Context, author, slug string) (response ModelEndpointsResponse, err error) {
	if author == "" {
		err = fmt.Errorf("model author cannot be empty")
		return
	}
	if slug == "" {
		err = fmt.Errorf("model slug cannot be empty")
		return
	}

	// Construct the path with URL encoding for author and slug
	// Using path parameters as defined: /models/:author/:slug/endpoints
	path := fmt.Sprintf("/models/%s/%s/endpoints", url.PathEscape(author), url.PathEscape(slug))

	err = s.client.doRequest(ctx, http.MethodGet, path, nil, &response, false) // nil request body for GET, false = use standard API key
	return
}

// ModelListResponse wraps the list of ModelData in a 'data' field,
// matching the API response structure for listing models.
type ModelListResponse struct {
	Data []ModelData `json:"data"`
}

// ModelEndpointsResponse wraps the ModelEndpointData in a 'data' field,
// matching the API response structure for listing model endpoints.
type ModelEndpointsResponse struct {
	Data ModelEndpointData `json:"data"`
}
