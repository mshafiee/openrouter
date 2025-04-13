package client

import (
	"fmt"
	"net/http"
	"net/url"
	// "runtime" // Not needed here anymore
)

// Option is a functional option type for configuring the Client.
// Defined canonically in client.go, re-declared here for context or assume imported.
// type Option func(*Client) error

// WithHTTPClient sets the HTTP client to use for requests.
// If not specified, http.DefaultClient is used.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) error {
		if httpClient == nil {
			return fmt.Errorf("HTTP client cannot be nil")
		}
		c.httpClient = httpClient
		return nil
	}
}

// WithBaseURL sets the base URL for API requests.
// If not specified, the default OpenRouter API URL (https://openrouter.ai/api/v1) is used.
// The provided URL string must be an absolute URL.
func WithBaseURL(baseURLStr string) Option {
	return func(c *Client) error {
		parsedURL, err := url.Parse(baseURLStr)
		if err != nil {
			return fmt.Errorf("failed to parse base URL %q: %w", baseURLStr, err)
		}
		if !parsedURL.IsAbs() {
			return fmt.Errorf("base URL must be absolute: %q", baseURLStr)
		}
		c.baseURL = parsedURL
		return nil
	}
}

// WithProvisioningKey sets the Provisioning API Key, used specifically for
// the key management endpoints (/keys, /keys/{hash}).
// Standard API operations will still use the main API key provided to NewClient.
func WithProvisioningKey(key string) Option {
	return func(c *Client) error {
		// No validation needed here, empty string is acceptable if not using provisioning endpoints
		c.provisioningKey = key
		return nil
	}
}

// WithUserAgent sets a custom User-Agent header string for all requests.
// If not specified, a default User-Agent like "openrouter-go/version (OS; Arch)" is used.
func WithUserAgent(userAgent string) Option {
	return func(c *Client) error {
		if userAgent == "" {
			return fmt.Errorf("user agent cannot be empty")
		}
		c.userAgent = userAgent
		return nil
	}
}

// WithDefaultHeader adds a default header key-value pair to be sent with every request.
// This can be called multiple times to add multiple headers. If the same key is provided
// multiple times, the last value set will take precedence (due to using http.Header.Set internally).
// Useful for headers like 'HTTP-Referer' or 'X-Title' for site ranking.
func WithDefaultHeader(key, value string) Option {
	return func(c *Client) error {
		if key == "" {
			return fmt.Errorf("header key cannot be empty")
		}
		if c.defaultHeaders == nil {
			c.defaultHeaders = make(http.Header)
		}
		c.defaultHeaders.Set(key, value)
		return nil
	}
}
