package client

// This file defines the Go structs corresponding to the JSON response
// schemas for the Model Listing APIs (/v1/models and /v1/models/{author}/{slug}/endpoints).

// --- Structs for GET /models ---

// ModelData represents information about a specific model available on OpenRouter,
// typically returned as part of a list from GET /models.
type ModelData struct {
	ID               string                 `json:"id"`                           // Unique model identifier (e.g., "openai/gpt-4o").
	Name             string                 `json:"name"`                         // User-friendly name.
	Created          *int64                 `json:"created,omitempty"`            // Unix timestamp of creation/addition (pointer for potential null).
	Description      *string                `json:"description,omitempty"`        // Optional model description (pointer for nullability).
	Architecture     *ModelArchitecture     `json:"architecture,omitempty"`       // Details about model architecture.
	TopProvider      *TopProviderInfo       `json:"top_provider,omitempty"`       // Information about the top provider.
	Pricing          *ModelPricing          `json:"pricing"`                      // Pricing details (lowest across providers). Pointer used as Pricing itself might be complex or rarely null.
	ContextLength    int                    `json:"context_length"`               // Maximum context length in tokens.
	PerRequestLimits map[string]interface{} `json:"per_request_limits,omitempty"` // Provider-specific limits (use interface{} for flexibility).
}

// ModelArchitecture describes the technical architecture of a model.
type ModelArchitecture struct {
	InputModalities  []string `json:"input_modalities,omitempty"`  // E.g., ["text", "image"].
	OutputModalities []string `json:"output_modalities,omitempty"` // E.g., ["text"].
	Tokenizer        string   `json:"tokenizer,omitempty"`         // E.g., "GPT", "Claude", "Llama".
	InstructType     *string  `json:"instruct_type,omitempty"`     // Instruction format type, if applicable (pointer for nullability).
}

// TopProviderInfo provides summary information about the leading provider for a model.
type TopProviderInfo struct {
	IsModerated *bool `json:"is_moderated,omitempty"` // Indicates if the top provider applies moderation (pointer for potential null).
	// Add other potential fields if observed: MaxCompletionTokens? ProviderName?
}

// ModelPricing represents the pricing details for a model (typically the best available).
// Prices are strings to avoid float precision issues, representing cost in USD.
type ModelPricing struct {
	Prompt            string `json:"prompt"`                       // Cost per prompt token.
	Completion        string `json:"completion"`                   // Cost per completion token.
	Image             string `json:"image,omitempty"`              // Cost per image (if applicable).
	Request           string `json:"request,omitempty"`            // Cost per request (if applicable).
	InputCacheRead    string `json:"input_cache_read,omitempty"`   // Cost per cached input token read (if applicable).
	InputCacheWrite   string `json:"input_cache_write,omitempty"`  // Cost per cached input token written (if applicable).
	WebSearch         string `json:"web_search,omitempty"`         // Cost per web search result used (if applicable).
	InternalReasoning string `json:"internal_reasoning,omitempty"` // Cost per reasoning/thinking token (if applicable).
}

// --- Structs for GET /models/{author}/{slug}/endpoints ---

// ModelEndpointData represents detailed information about a model and its specific provider endpoints.
// Returned by GET /models/{author}/{slug}/endpoints.
type ModelEndpointData struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Created      *float64           `json:"created,omitempty"` // Example showed 1.1, using float64 pointer. Could be int64 timestamp too.
	Description  *string            `json:"description,omitempty"`
	Architecture *ModelArchitecture `json:"architecture,omitempty"` // Reusing the architecture struct.
	Endpoints    []EndpointDetail   `json:"endpoints,omitempty"`    // List of providers serving this model.
}

// EndpointDetail describes a specific provider's offering for a model.
type EndpointDetail struct {
	Name                string           `json:"name"`                           // User-friendly name of the provider/endpoint.
	ContextLength       *float64         `json:"context_length,omitempty"`       // Context length for this specific endpoint (example showed 1.1).
	Pricing             *EndpointPricing `json:"pricing,omitempty"`              // Pricing specific to this provider endpoint.
	ProviderName        string           `json:"provider_name"`                  // Canonical provider name (e.g., "OpenAI", "Anthropic").
	SupportedParameters []string         `json:"supported_parameters,omitempty"` // List of features/params supported (e.g., "tools", "response_format:json_schema").
	// Add other fields if observed: Quantization? IsModerated?
}

// EndpointPricing represents pricing specific to a provider endpoint.
// Prices are strings to avoid float precision issues, representing cost in USD.
type EndpointPricing struct {
	Request    string `json:"request,omitempty"`    // Cost per request.
	Image      string `json:"image,omitempty"`      // Cost per image.
	Prompt     string `json:"prompt,omitempty"`     // Cost per prompt token.
	Completion string `json:"completion,omitempty"` // Cost per completion token.
	// Add other pricing fields like cache, web_search if they can vary per-endpoint
}

// Note: The response wrapper structures are defined in model.go:
// type ModelListResponse struct {
//	 Data []ModelData `json:"data"`
// }
// type ModelEndpointsResponse struct {
//	 Data ModelEndpointData `json:"data"`
// }
