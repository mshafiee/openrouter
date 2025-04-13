// Package client provides a Go client for the OpenRouter API.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"       // Keep for debug logging added earlier, remove if not needed
	"net/http" // Keep for debug logging added earlier, remove if not needed
	"net/url"
	"runtime"
	"strings"
)

const (
	defaultBaseURL = "https://openrouter.ai/api/v1/"
	libraryVersion = "0.2.0"
	mediaTypeJSON  = "application/json"
)

// defaultUserAgent is the default User-Agent header for requests.
var defaultUserAgent = fmt.Sprintf("openrouter-go/%s (%s; %s)", libraryVersion, runtime.GOOS, runtime.GOARCH)

// Client manages communication with the OpenRouter API.
type Client struct {
	httpClient      *http.Client       // HTTP client used to communicate with the API.
	apiKey          string             // Primary API Key for standard operations.
	provisioningKey string             // Optional API Key specifically for key management.
	baseURL         *url.URL           // Base URL for API requests.
	userAgent       string             // User-Agent header for requests.
	defaultHeaders  http.Header        // Headers added to every request.
	common          service            // Reuse a single struct instead of allocating one for each service on the heap.
	Chat            *ChatService       // Access Chat Completion methods.
	Completion      *CompletionService // Access Legacy Completion methods (Deprecated).
	Generation      *GenerationService // Access Generation metadata methods.
	Model           *ModelService      // Access Model listing methods.
	Billing         *BillingService    // Access Billing and Credits methods.
	Auth            *AuthService       // Access Auth methods (PKCE flow).
	Keys            *KeysService       // Access API Key management methods.
}

// service is a base type for API services.
type service struct {
	client *Client
}

// Option is a functional option type for configuring the Client.
// Implementations are in options.go
type Option func(*Client) error

// NewClient creates a new OpenRouter API client.
// An apiKey is required for most operations.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key cannot be empty")
	}

	// Initialize with default values
	baseURL, err := url.Parse(defaultBaseURL)
	if err != nil {
		// This should not happen with a hardcoded valid URL
		return nil, fmt.Errorf("failed to parse default base URL: %w", err)
	}

	c := &Client{
		httpClient:     http.DefaultClient,
		apiKey:         apiKey,
		baseURL:        baseURL,
		userAgent:      defaultUserAgent,
		defaultHeaders: make(http.Header),
	}

	// Apply functional options (defined in options.go)
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	// Initialize services
	c.common.client = c
	c.Chat = (*ChatService)(&c.common)
	c.Completion = (*CompletionService)(&c.common)
	c.Generation = (*GenerationService)(&c.common)
	c.Model = (*ModelService)(&c.common)
	c.Billing = (*BillingService)(&c.common)
	c.Auth = (*AuthService)(&c.common)
	c.Keys = (*KeysService)(&c.common)

	return c, nil
}

// --- Core Request Logic ---

// doRequest performs an HTTP request to the OpenRouter API using standard or provisioning key.
func (c *Client) doRequest(ctx context.Context, method, path string, reqData, resData interface{}, useProvisioningKey bool) error {
	// 1. Determine API Key
	var effectiveAPIKey string
	if useProvisioningKey {
		if c.provisioningKey == "" {
			return fmt.Errorf("provisioning key required but not set for path %s", path)
		}
		effectiveAPIKey = c.provisioningKey
	} else {
		effectiveAPIKey = c.apiKey
	}

	// 2. Construct URL
	// Use ResolveReference which correctly handles joining base URL and relative path
	relURL, err := url.Parse(path)
	if err != nil {
		return fmt.Errorf("failed to parse relative path %q: %w", path, err)
	}
	fullURL := c.baseURL.ResolveReference(relURL)

	// 3. Prepare Request Body (if applicable)
	var reqBody io.Reader
	var reqBytes []byte // Store bytes for potential logging/debugging
	if reqData != nil {
		reqBytes, err = json.Marshal(reqData)
		if err != nil {
			return fmt.Errorf("failed to marshal request data for %s: %w", path, err)
		}
		reqBody = bytes.NewReader(reqBytes)
	}

	// 4. Create HTTP Request
	req, err := http.NewRequestWithContext(ctx, method, fullURL.String(), reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request for %s: %w", path, err)
	}

	// 5. Set Headers
	if reqBody != nil {
		req.Header.Set("Content-Type", mediaTypeJSON)
	}
	req.Header.Set("Accept", mediaTypeJSON)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Authorization", "Bearer "+effectiveAPIKey)

	for key, values := range c.defaultHeaders {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 6. Execute Request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Wrap network errors for potential context checking later
			return fmt.Errorf("failed to execute request to %s: %w", path, err)
		}
	}
	defer resp.Body.Close()

	// 7. Read Body *before* further status code checks
	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		// Return an APIError even if status was < 400, indicating body read failure
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("API request returned status %d but failed to read response body: %v", resp.StatusCode, readErr),
			RawBody:    nil,
		}
	}

	// 8. Handle Status Code
	if resp.StatusCode >= 400 {
		// Error status code, try to parse standard error JSON from bodyBytes
		var errResp ErrorResponseWrapper
		if jsonErr := json.Unmarshal(bodyBytes, &errResp); jsonErr == nil && errResp.Error != nil {
			// Parsed standard error format
			return &APIError{
				StatusCode: resp.StatusCode,
				Message:    fmt.Sprintf("API error: %s", errResp.Error.Message),
				Details:    errResp.Error,
				RawBody:    bodyBytes,
			}
		}
		// If parsing error JSON failed, return a generic error with the raw body
		rawBodyStr := string(bodyBytes)
		if len(rawBodyStr) > 512 {
			rawBodyStr = rawBodyStr[:512] + "..."
		}
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("API error (status %d): %s", resp.StatusCode, rawBodyStr),
			RawBody:    bodyBytes,
		}
	}

	// 9. Decode Success Response Body (if applicable and status is not 204)
	if resData != nil && resp.StatusCode != http.StatusNoContent {
		// Check Content-Type before attempting to decode JSON
		contentType := resp.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, mediaTypeJSON) {
			rawBodyStr := string(bodyBytes)
			if len(rawBodyStr) > 512 {
				rawBodyStr = rawBodyStr[:512] + "..."
			}
			// Return an error if the content type is not JSON, even with 2xx status
			return &APIError{
				StatusCode: resp.StatusCode,
				Message:    fmt.Sprintf("API request successful (status %d) but received unexpected Content-Type '%s' instead of '%s'. Body: %s", resp.StatusCode, contentType, mediaTypeJSON, rawBodyStr),
				RawBody:    bodyBytes,
			}
		}

		// Attempt to decode the bodyBytes we already read
		if err := json.Unmarshal(bodyBytes, resData); err != nil {
			rawBodyStr := string(bodyBytes)
			if len(rawBodyStr) > 512 {
				rawBodyStr = rawBodyStr[:512] + "..."
			}
			// Return a more specific error if decoding fails
			return fmt.Errorf("failed to decode successful response body (status %d, content-type %s) from %s: %w. Body: %s", resp.StatusCode, contentType, path, err, rawBodyStr)
		}
	} else if resData != nil && resp.StatusCode == http.StatusNoContent {
		// Handle case where we expect data but get 204. Correct behavior is to not decode.
		// log.Printf("Warning: Received HTTP 204 No Content for %s when expecting response data.", path)
	}

	return nil // Success
}

// doStreamingRequest performs an HTTP request expecting a Server-Sent Events stream.
func (c *Client) doStreamingRequest(ctx context.Context, method, path string, reqData interface{}) (*http.Response, error) {
	effectiveAPIKey := c.apiKey

	relURL, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse relative path %q: %w", path, err)
	}
	fullURL := c.baseURL.ResolveReference(relURL)

	var reqBody io.Reader
	if reqData != nil {
		reqBytes, err := json.Marshal(reqData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request data for streaming %s: %w", path, err)
		}
		reqBody = bytes.NewReader(reqBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL.String(), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create streaming request for %s: %w", path, err)
	}

	// Set Headers specific to streaming
	if reqBody != nil {
		req.Header.Set("Content-Type", mediaTypeJSON)
	}
	req.Header.Set("Accept", "text/event-stream") // Expect SSE
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Authorization", "Bearer "+effectiveAPIKey)
	for key, values := range c.defaultHeaders {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Execute Request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			return nil, fmt.Errorf("failed to execute streaming request to %s: %w", path, err)
		}
	}

	// Check status code *before* assuming stream is valid
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		bodyBytes, readErr := io.ReadAll(resp.Body)
		// Use the existing APIError creation logic from doRequest
		if readErr != nil {
			return nil, &APIError{
				StatusCode: resp.StatusCode,
				Message:    fmt.Sprintf("API error (status %d) but failed to read error response body: %v", resp.StatusCode, readErr),
			}
		}
		var errResp ErrorResponseWrapper
		if jsonErr := json.Unmarshal(bodyBytes, &errResp); jsonErr == nil && errResp.Error != nil {
			return nil, &APIError{
				StatusCode: resp.StatusCode,
				Message:    fmt.Sprintf("API error: %s", errResp.Error.Message),
				Details:    errResp.Error,
				RawBody:    bodyBytes,
			}
		}
		// Generic error if body wasn't standard JSON error
		rawBodyStr := string(bodyBytes)
		if len(rawBodyStr) > 512 {
			rawBodyStr = rawBodyStr[:512] + "..."
		}
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("API error (status %d): %s", resp.StatusCode, rawBodyStr),
			RawBody:    bodyBytes,
		}
	}

	// Check Content-Type *after* confirming StatusOK
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/event-stream") {
		resp.Body.Close() // Close the body as it's not the expected stream
		// Attempt to read body as potential error message or unexpected content
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("expected Content-Type 'text/event-stream' (status %d), got '%s', and failed to read body: %w", resp.StatusCode, contentType, readErr)
		}
		rawBodyStr := string(bodyBytes)
		if len(rawBodyStr) > 512 {
			rawBodyStr = rawBodyStr[:512] + "..."
		}
		// Return a standard error, not APIError, as status was OK but content type wrong
		return nil, fmt.Errorf("expected Content-Type 'text/event-stream' (status %d), got '%s'. Body: %s", resp.StatusCode, contentType, rawBodyStr)
	}

	// Return the raw response for the stream handler to process
	return resp, nil
}

// doUnauthenticatedRequest performs a request without adding the Authorization header.
// Used for OAuth PKCE code exchange.
func (c *Client) doUnauthenticatedRequest(ctx context.Context, method, path string, reqData, resData interface{}) error {
	// 1. Construct URL
	relURL, err := url.Parse(path)
	if err != nil {
		return fmt.Errorf("failed to parse relative path %q: %w", path, err)
	}
	fullURL := c.baseURL.ResolveReference(relURL)

	// 2. Prepare Request Body
	var reqBody io.Reader
	var reqBytes []byte
	if reqData != nil {
		reqBytes, err = json.Marshal(reqData)
		if err != nil {
			return fmt.Errorf("failed to marshal request data for %s: %w", path, err)
		}
		reqBody = bytes.NewReader(reqBytes)
	}

	// 3. Create HTTP Request
	req, err := http.NewRequestWithContext(ctx, method, fullURL.String(), reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request for %s: %w", path, err)
	}

	// 4. Set Headers (No Authorization)
	if reqBody != nil {
		req.Header.Set("Content-Type", mediaTypeJSON)
	}
	req.Header.Set("Accept", mediaTypeJSON)
	req.Header.Set("User-Agent", c.userAgent)
	// NO Authorization header added

	for key, values := range c.defaultHeaders {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 5. Execute Request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return fmt.Errorf("failed to execute request to %s: %w", path, err)
		}
	}
	defer resp.Body.Close()

	// 6. Read Body
	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("API request returned status %d but failed to read response body: %v", resp.StatusCode, readErr),
		}
	}

	// 7. Handle Status Code
	if resp.StatusCode >= 400 {
		var errResp ErrorResponseWrapper
		if jsonErr := json.Unmarshal(bodyBytes, &errResp); jsonErr == nil && errResp.Error != nil {
			return &APIError{
				StatusCode: resp.StatusCode,
				Message:    fmt.Sprintf("API error: %s", errResp.Error.Message),
				Details:    errResp.Error,
				RawBody:    bodyBytes,
			}
		}
		rawBodyStr := string(bodyBytes)
		if len(rawBodyStr) > 512 {
			rawBodyStr = rawBodyStr[:512] + "..."
		}
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("API error (status %d): %s", resp.StatusCode, rawBodyStr),
			RawBody:    bodyBytes,
		}
	}

	// 8. Decode Success Response Body
	if resData != nil && resp.StatusCode != http.StatusNoContent {
		contentType := resp.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, mediaTypeJSON) {
			rawBodyStr := string(bodyBytes)
			if len(rawBodyStr) > 512 {
				rawBodyStr = rawBodyStr[:512] + "..."
			}
			return &APIError{
				StatusCode: resp.StatusCode,
				Message:    fmt.Sprintf("API request successful (status %d) but received unexpected Content-Type '%s' instead of '%s'. Body: %s", resp.StatusCode, contentType, mediaTypeJSON, rawBodyStr),
				RawBody:    bodyBytes,
			}
		}

		if err := json.Unmarshal(bodyBytes, resData); err != nil {
			rawBodyStr := string(bodyBytes)
			if len(rawBodyStr) > 512 {
				rawBodyStr = rawBodyStr[:512] + "..."
			}
			return fmt.Errorf("failed to decode successful response body (status %d, content-type %s) from %s: %w. Body: %s", resp.StatusCode, contentType, path, err, rawBodyStr)
		}
	}

	return nil // Success
}

// IsProvisioning returns true if a provisioning key was set on the client.
func (c *Client) IsProvisioning() bool {
	return c.provisioningKey != ""
}
