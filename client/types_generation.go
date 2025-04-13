package client

import "time"

// This file defines the Go struct corresponding to the JSON response
// schema for the Get Generation Metadata API (/v1/generation).

// GenerationData represents the detailed metadata for a single generation request.
// It's the core object returned within the 'data' field of the GET /generation response.
type GenerationData struct {
	ID                     string    `json:"id"`                       // Unique generation ID.
	TotalCost              float64   `json:"total_cost"`               // Final cost in USD after discounts/fees.
	CreatedAt              time.Time `json:"created_at"`               // Timestamp when the generation record was created.
	Model                  string    `json:"model"`                    // Final model used for the generation.
	Origin                 string    `json:"origin,omitempty"`         // Source of the request (e.g., 'api', 'chat').
	Usage                  float64   `json:"usage"`                    // Cost before discounts/fees in USD.
	IsBYOK                 bool      `json:"is_byok"`                  // True if the request used a Bring-Your-Own-Key provider.
	UpstreamID             *string   `json:"upstream_id"`              // ID assigned by the upstream provider, if available (pointer for nullability).
	CacheDiscount          *float64  `json:"cache_discount"`           // Amount discounted/charged due to prompt caching (pointer for nullability).
	AppID                  *int      `json:"app_id"`                   // Internal application identifier (pointer for nullability).
	Streamed               bool      `json:"streamed"`                 // True if the request was streamed.
	Cancelled              bool      `json:"cancelled"`                // True if the streaming request was cancelled.
	ProviderName           *string   `json:"provider_name"`            // Name of the provider that fulfilled the request (pointer for nullability).
	Latency                *float64  `json:"latency"`                  // Time to first token in seconds (pointer for nullability).
	ModerationLatency      *float64  `json:"moderation_latency"`       // Time spent on moderation checks in seconds (pointer for nullability).
	GenerationTime         *float64  `json:"generation_time"`          // Total time taken for the generation in seconds (pointer for nullability).
	FinishReason           *string   `json:"finish_reason"`            // Normalized finish reason (pointer for nullability).
	NativeFinishReason     *string   `json:"native_finish_reason"`     // Raw finish reason from the provider (pointer for nullability).
	TokensPrompt           *int      `json:"tokens_prompt"`            // Normalized prompt token count (GPT-4o tokenizer) (pointer for nullability).
	TokensCompletion       *int      `json:"tokens_completion"`        // Normalized completion token count (GPT-4o tokenizer) (pointer for nullability).
	NativeTokensPrompt     *int      `json:"native_tokens_prompt"`     // Prompt token count using native tokenizer (pointer for nullability).
	NativeTokensCompletion *int      `json:"native_tokens_completion"` // Completion token count using native tokenizer (pointer for nullability).
	NativeTokensReasoning  *int      `json:"native_tokens_reasoning"`  // Reasoning token count using native tokenizer (pointer for nullability).
	NumMediaPrompt         *int      `json:"num_media_prompt"`         // Number of media items (e.g., images) in the prompt (pointer for nullability).
	NumMediaCompletion     *int      `json:"num_media_completion"`     // Number of media items (e.g., images) in the completion (pointer for nullability).
	NumSearchResults       *int      `json:"num_search_results"`       // Number of web search results used, if applicable (pointer for nullability).
}

// Note: The response structure for GET /generation is defined in generation.go
// as:
// type GenerationGetResponse struct {
//	 Data GenerationData `json:"data"`
// }
// This GenerationData struct definition corresponds to the inner 'data' object.
