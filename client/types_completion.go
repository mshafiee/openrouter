package client

// This file defines the Go structs corresponding to the JSON request and response
// schemas for the OpenRouter Legacy Text Completions API (/v1/completions).
// This endpoint is marked as deprecated.

// CompletionRequest represents the request body for the legacy text completions endpoint.
type CompletionRequest struct {
	Model             string               `json:"model"`            // Required
	Prompt            string               `json:"prompt"`           // Required
	Stream            bool                 `json:"stream,omitempty"` // Default: false (Streaming not implemented in client for this legacy endpoint)
	MaxTokens         *int                 `json:"max_tokens,omitempty"`
	Temperature       *float64             `json:"temperature,omitempty"` // Range [0, 2], Default: 1.0
	Seed              *int                 `json:"seed,omitempty"`
	TopP              *float64             `json:"top_p,omitempty"`              // Range (0, 1], Default: 1.0
	TopK              *int                 `json:"top_k,omitempty"`              // Range [1, inf), Default: 0 (disabled)
	FrequencyPenalty  *float64             `json:"frequency_penalty,omitempty"`  // Range [-2, 2], Default: 0.0
	PresencePenalty   *float64             `json:"presence_penalty,omitempty"`   // Range [-2, 2], Default: 0.0
	RepetitionPenalty *float64             `json:"repetition_penalty,omitempty"` // Range (0, 2], Default: 1.0
	LogitBias         map[string]float64   `json:"logit_bias,omitempty"`         // Map token ID (string) to bias [-100, 100]
	Logprobs          *bool                `json:"logprobs,omitempty"`           // Request log probabilities (May not be fully supported/documented for legacy)
	TopLogprobs       *int                 `json:"top_logprobs,omitempty"`       // Range [0, 20] (May not be fully supported/documented for legacy)
	MinP              *float64             `json:"min_p,omitempty"`              // Range [0, 1], Default: 0.0
	TopA              *float64             `json:"top_a,omitempty"`              // Range [0, 1], Default: 0.0
	Transforms        []string             `json:"transforms,omitempty"`         // OpenRouter-only: ["middle-out"]
	Models            []string             `json:"models,omitempty"`             // OpenRouter-only: Fallback models
	Provider          *ProviderPreferences `json:"provider,omitempty"`           // OpenRouter-only: Provider routing preferences
	Reasoning         *ReasoningConfig     `json:"reasoning,omitempty"`          // OpenRouter-only: Control reasoning tokens (unlikely applicable here)
	// Note: Stop sequences were not explicitly listed for this endpoint in the
	// POST /completions section, but might be supported. Add if needed.
	// Stop              *StringOrArray     `json:"stop,omitempty"`
}

// CompletionResponse represents the response body for the legacy text completions endpoint.
type CompletionResponse struct {
	ID      string             `json:"id,omitempty"`      // Optional in example response
	Choices []CompletionChoice `json:"choices,omitempty"` // Optional in example response
	// Based on standard OpenAI patterns, Usage and Model might be included
	Usage *ResponseUsage `json:"usage,omitempty"`
	Model string         `json:"model,omitempty"`
	// Object field ("text_completion") might also be present
	Object string `json:"object,omitempty"`
}

// CompletionChoice represents a single completion choice in the legacy response.
type CompletionChoice struct {
	Text         string         `json:"text,omitempty"`
	Index        int            `json:"index,omitempty"` // Example shows 1, typically 0-based
	FinishReason string         `json:"finish_reason,omitempty"`
	Logprobs     *LogprobResult `json:"logprobs,omitempty"` // Include if logprobs requested
}

// Note: Streaming response types for the legacy endpoint are omitted as
// streaming was not explicitly documented for it and the client implementation
// in completion.go does not support it.
