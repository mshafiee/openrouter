package client

import "time" // Import time package for CreatedAt/UpdatedAt

// This file defines the Go structs corresponding to the JSON request and response
// schemas for the OpenRouter API Key Management APIs (/v1/key and /v1/keys/*).

// --- Structs for GET /key (Current Key Info) ---

// KeyInfoData represents information about the currently authenticated API key.
// Returned within the 'data' field of GET /key response.
type KeyInfoData struct {
	Label             string        `json:"label"`
	Usage             float64       `json:"usage"`
	Limit             *float64      `json:"limit"`
	LimitRemaining    *float64      `json:"limit_remaining"`
	IsFreeTier        bool          `json:"is_free_tier"`
	IsProvisioningKey bool          `json:"is_provisioning_key"`
	RateLimit         RateLimitInfo `json:"rate_limit"`
}

// RateLimitInfo represents the rate limit structure for an API key.
// Could also be in types_common.go.
type RateLimitInfo struct {
	Requests int    `json:"requests"`
	Interval string `json:"interval"`
}

// --- Structs for /keys/* (Provisioning Key Management) ---

// ListKeysParams defines optional query parameters for the List operation.
type ListKeysParams struct {
	Offset          *int  `url:"offset,omitempty"`           // Offset for pagination.
	IncludeDisabled *bool `url:"include_disabled,omitempty"` // Whether to include disabled keys.
}

// ApiKeyData represents detailed information about an API key, typically used
// in responses from the provisioning endpoints (GET /keys, POST /keys, GET/PATCH /keys/{hash}).
type ApiKeyData struct {
	Name      string    `json:"name"`
	Label     *string   `json:"label"` // Pointer for nullability
	Limit     *float64  `json:"limit"` // Pointer for nullability
	Disabled  bool      `json:"disabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Hash      string    `json:"hash"`
	Key       string    `json:"key,omitempty"`   // ONLY present in the POST /keys Create response.
	Usage     *float64  `json:"usage,omitempty"` // Pointer as it might not always be present
}

// CreateApiKeyRequest represents the request body for creating a new API key
// via the POST /keys endpoint (requires provisioning key auth).
type CreateApiKeyRequest struct {
	Name  string   `json:"name"` // Required
	Label *string  `json:"label,omitempty"`
	Limit *float64 `json:"limit,omitempty"` // Pointer for nullability
}

// UpdateApiKeyRequest represents the request body for updating an existing API key
// via the PATCH /keys/{hash} endpoint (requires provisioning key auth).
type UpdateApiKeyRequest struct {
	Name     *string  `json:"name,omitempty"`
	Label    *string  `json:"label,omitempty"`
	Limit    *float64 `json:"limit,omitempty"` // Pointer allows client to represent unset
	Disabled *bool    `json:"disabled,omitempty"`
}

// --- Placeholder for Response Wrappers ---
// The actual response wrapper structs (e.g., KeyInfoResponse, ListKeysResponse)
// are defined in keys.go as they are simple wrappers specific to those methods.
