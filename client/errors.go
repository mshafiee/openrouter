package client

import (
	"encoding/json"
	"errors" // Import standard errors package
	"fmt"
)

// APIError represents an error response from the OpenRouter API.
// It includes the HTTP status code, a message, and potentially more detailed
// information parsed from the error response body.
type APIError struct {
	StatusCode int
	Message    string
	// Details contains the structured error information if the API returned
	// the standard JSON error format (defined in types_common.go).
	// It might be nil if the error body couldn't be parsed or didn't follow the expected structure.
	Details *ErrorDetails // ErrorDetails definition is in types_common.go
	RawBody []byte
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Details != nil && e.Details.Message != "" {
		return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Details.Message)
	}
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}

// --- Utility Functions (Keep these as they operate on APIError) ---

// IsAPIError checks if an error is specifically an *APIError.
func IsAPIError(err error) bool {
	_, ok := err.(*APIError)
	return ok
}

// GetAPIError attempts to retrieve the underlying *APIError from a given error.
// Returns the *APIError and true if successful, otherwise nil and false.
func GetAPIError(err error) (*APIError, bool) {
	var apiErr *APIError
	// Use standard library errors.As for robust checking
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}

// GetErrorMetadata attempts to extract the Metadata field from an APIError.
// Returns the metadata map and true if the error is an APIError with non-nil Details and Metadata.
func GetErrorMetadata(err error) (map[string]interface{}, bool) {
	if apiErr, ok := GetAPIError(err); ok {
		if apiErr.Details != nil && apiErr.Details.Metadata != nil {
			return apiErr.Details.Metadata, true
		}
	}
	return nil, false
}

// --- Specific Error Metadata Structs (Keep these definitions or move to types_errors.go/types_common.go) ---
// Decide where these best fit. Keeping them here is okay if they are only used with APIError helpers.
// Moving to types_common.go is also reasonable if they might be reused elsewhere.

// ModerationErrorMetadata represents the structure within `error.metadata`
// when an input is flagged by moderation.
type ModerationErrorMetadata struct {
	Reasons      []string `json:"reasons"`
	FlaggedInput string   `json:"flagged_input"`
	ProviderName string   `json:"provider_name"`
	ModelSlug    string   `json:"model_slug"`
}

// ProviderErrorMetadata represents the structure within `error.metadata`
// when a downstream provider encounters an error.
type ProviderErrorMetadata struct {
	ProviderName string      `json:"provider_name"`
	Raw          interface{} `json:"raw"` // Raw error from the provider
}

// TryGetModerationMetadata attempts to parse the metadata of an APIError as ModerationErrorMetadata.
func TryGetModerationMetadata(err error) (*ModerationErrorMetadata, bool) {
	metadata, ok := GetErrorMetadata(err)
	if !ok {
		return nil, false
	}
	bytes, marshalErr := json.Marshal(metadata)
	if marshalErr != nil {
		return nil, false
	}
	var modMeta ModerationErrorMetadata
	if unmarshalErr := json.Unmarshal(bytes, &modMeta); unmarshalErr == nil {
		if len(modMeta.Reasons) > 0 || modMeta.FlaggedInput != "" {
			return &modMeta, true
		}
	}
	return nil, false
}

// TryGetProviderErrorMetadata attempts to parse the metadata of an APIError as ProviderErrorMetadata.
func TryGetProviderErrorMetadata(err error) (*ProviderErrorMetadata, bool) {
	metadata, ok := GetErrorMetadata(err)
	if !ok {
		return nil, false
	}
	bytes, marshalErr := json.Marshal(metadata)
	if marshalErr != nil {
		return nil, false
	}
	var provMeta ProviderErrorMetadata
	if unmarshalErr := json.Unmarshal(bytes, &provMeta); unmarshalErr == nil {
		if provMeta.ProviderName != "" {
			return &provMeta, true
		}
	}
	return nil, false
}
