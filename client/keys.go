package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	// "time" // time might be needed by structs if they were defined here
)

// KeysService handles communication with the API key management related methods
// of the OpenRouter API. Some methods require a Provisioning API key to be configured.
type KeysService service // Use type alias as established in client.go

// === Operations on the Current API Key (Standard Auth) ===

// GetCurrent retrieves information about the API key used for the current authentication session.
// Response wrapper KeyInfoResponse defined below. Core data struct KeyInfoData defined in types_keys.go.
func (s *KeysService) GetCurrent(ctx context.Context) (response KeyInfoResponse, err error) {
	path := "key"
	err = s.client.doRequest(ctx, http.MethodGet, path, nil, &response, false) // false = use standard API key
	return
}

// === Operations Requiring a Provisioning Key ===

// List retrieves a list of API keys associated with the account.
// Requires authentication using a Provisioning API key.
// Request params ListKeysParams defined in types_keys.go. Response wrapper ListKeysResponse defined below.
func (s *KeysService) List(ctx context.Context, params ListKeysParams) (response ListKeysResponse, err error) {
	if s.client.provisioningKey == "" {
		err = fmt.Errorf("provisioning key is required to list API keys")
		return
	}

	path := "keys"
	query := url.Values{}
	if params.Offset != nil {
		query.Set("offset", strconv.Itoa(*params.Offset))
	}
	if params.IncludeDisabled != nil {
		query.Set("include_disabled", strconv.FormatBool(*params.IncludeDisabled))
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	err = s.client.doRequest(ctx, http.MethodGet, path, nil, &response, true) // true = use provisioning API key
	return
}

// Create creates a new API key for the account.
// Requires authentication using a Provisioning API key.
// Request struct CreateApiKeyRequest defined in types_keys.go. Response wrapper CreateApiKeyResponse defined below.
func (s *KeysService) Create(ctx context.Context, request CreateApiKeyRequest) (response CreateApiKeyResponse, err error) {
	if s.client.provisioningKey == "" {
		err = fmt.Errorf("provisioning key is required to create an API key")
		return
	}
	if request.Name == "" {
		err = fmt.Errorf("key name is required")
		return
	}

	path := "keys"
	err = s.client.doRequest(ctx, http.MethodPost, path, request, &response, true) // true = use provisioning API key
	return
}

// Get retrieves details about a specific API key identified by its hash.
// Requires authentication using a Provisioning API key.
// Response wrapper GetApiKeyResponse defined below. Core data struct ApiKeyData defined in types_keys.go.
func (s *KeysService) Get(ctx context.Context, keyHash string) (response GetApiKeyResponse, err error) {
	if s.client.provisioningKey == "" {
		err = fmt.Errorf("provisioning key is required to get API key details")
		return
	}
	if keyHash == "" {
		err = fmt.Errorf("key hash cannot be empty")
		return
	}

	path := fmt.Sprintf("/keys/%s", url.PathEscape(keyHash))
	err = s.client.doRequest(ctx, http.MethodGet, path, nil, &response, true) // true = use provisioning API key
	return
}

// Update modifies properties of an existing API key identified by its hash.
// Requires authentication using a Provisioning API key.
// Request struct UpdateApiKeyRequest defined in types_keys.go. Response wrapper UpdateApiKeyResponse defined below.
func (s *KeysService) Update(ctx context.Context, keyHash string, request UpdateApiKeyRequest) (response UpdateApiKeyResponse, err error) {
	if s.client.provisioningKey == "" {
		err = fmt.Errorf("provisioning key is required to update an API key")
		return
	}
	if keyHash == "" {
		err = fmt.Errorf("key hash cannot be empty")
		return
	}
	if request.Name == nil && request.Label == nil && request.Limit == nil && request.Disabled == nil {
		err = fmt.Errorf("at least one field (name, label, limit, disabled) must be provided for update")
		return
	}

	path := fmt.Sprintf("/keys/%s", url.PathEscape(keyHash))
	err = s.client.doRequest(ctx, http.MethodPatch, path, request, &response, true) // true = use provisioning API key
	return
}

// Delete removes an API key identified by its hash. This action is irreversible.
// Requires authentication using a Provisioning API key.
// Response wrapper DeleteApiKeyResponse defined below.
func (s *KeysService) Delete(ctx context.Context, keyHash string) (response DeleteApiKeyResponse, err error) {
	if s.client.provisioningKey == "" {
		err = fmt.Errorf("provisioning key is required to delete an API key")
		return
	}
	if keyHash == "" {
		err = fmt.Errorf("key hash cannot be empty")
		return
	}

	path := fmt.Sprintf("/keys/%s", url.PathEscape(keyHash))
	err = s.client.doRequest(ctx, http.MethodDelete, path, nil, &response, true) // true = use provisioning API key
	return
}

// --- Response Wrapper Struct Definitions ---

// KeyInfoResponse wraps the KeyInfoData in a 'data' field.
type KeyInfoResponse struct {
	Data KeyInfoData `json:"data"` // KeyInfoData is defined in types_keys.go
}

// ListKeysResponse wraps the list of ApiKeyData in a 'data' field.
type ListKeysResponse struct {
	Data []ApiKeyData `json:"data"` // ApiKeyData is defined in types_keys.go
}

// CreateApiKeyResponse wraps the ApiKeyData (including the new key string) in a 'data' field.
type CreateApiKeyResponse struct {
	Data ApiKeyData `json:"data"` // ApiKeyData is defined in types_keys.go
}

// GetApiKeyResponse wraps the ApiKeyData in a 'data' field.
type GetApiKeyResponse struct {
	Data ApiKeyData `json:"data"` // ApiKeyData is defined in types_keys.go
}

// UpdateApiKeyResponse wraps the updated ApiKeyData in a 'data' field.
type UpdateApiKeyResponse struct {
	Data ApiKeyData `json:"data"` // ApiKeyData is defined in types_keys.go
}

// DeleteApiKeyResponse wraps the success indicator in a 'data' field.
type DeleteApiKeyResponse struct {
	Data struct {
		Success bool `json:"success"`
	} `json:"data"`
}
