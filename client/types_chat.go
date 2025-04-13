package client

// This file defines the Go structs corresponding to the JSON request and response
// schemas for the OpenRouter Chat Completions API (/v1/chat/completions).

// ChatCompletionRequest represents the request body for the chat completions endpoint.
type ChatCompletionRequest struct {
	Messages          []Message            `json:"messages,omitempty"` // Use omitempty if Prompt can be used instead
	Prompt            string               `json:"prompt,omitempty"`   // Legacy prompt, use Messages instead. Either Messages or Prompt required.
	Model             string               `json:"model,omitempty"`    // Optional: Defaults to user's default if not specified.
	ResponseFormat    *ResponseFormat      `json:"response_format,omitempty"`
	Stop              *StringOrArray       `json:"stop,omitempty"`
	Stream            bool                 `json:"stream,omitempty"` // Default: false
	MaxTokens         *int                 `json:"max_tokens,omitempty"`
	Temperature       *float64             `json:"temperature,omitempty"` // Range [0, 2], Default: 1.0
	Tools             []Tool               `json:"tools,omitempty"`
	ToolChoice        *ToolChoice          `json:"tool_choice,omitempty"`
	Seed              *int                 `json:"seed,omitempty"`
	TopP              *float64             `json:"top_p,omitempty"`              // Range (0, 1], Default: 1.0
	TopK              *int                 `json:"top_k,omitempty"`              // Range [1, inf), Default: 0 (disabled)
	FrequencyPenalty  *float64             `json:"frequency_penalty,omitempty"`  // Range [-2, 2], Default: 0.0
	PresencePenalty   *float64             `json:"presence_penalty,omitempty"`   // Range [-2, 2], Default: 0.0
	RepetitionPenalty *float64             `json:"repetition_penalty,omitempty"` // Range (0, 2], Default: 1.0
	LogitBias         map[string]float64   `json:"logit_bias,omitempty"`         // Map token ID (string) to bias [-100, 100]
	Logprobs          *bool                `json:"logprobs,omitempty"`           // Request log probabilities
	TopLogprobs       *int                 `json:"top_logprobs,omitempty"`       // Range [0, 20]. Requires logprobs=true.
	MinP              *float64             `json:"min_p,omitempty"`              // Range [0, 1], Default: 0.0
	TopA              *float64             `json:"top_a,omitempty"`              // Range [0, 1], Default: 0.0
	Prediction        *PredictionContent   `json:"prediction,omitempty"`         // OpenAI latency optimization hint
	Transforms        []string             `json:"transforms,omitempty"`         // OpenRouter-only: ["middle-out"]
	Models            []string             `json:"models,omitempty"`             // OpenRouter-only: Fallback models
	Route             string               `json:"route,omitempty"`              // OpenRouter-only: Routing strategy (e.g., "fallback")
	Provider          *ProviderPreferences `json:"provider,omitempty"`           // OpenRouter-only: Provider routing preferences
	Reasoning         *ReasoningConfig     `json:"reasoning,omitempty"`          // OpenRouter-only: Control reasoning tokens
	MaxPrice          *MaxPrice            `json:"max_price,omitempty"`          // OpenRouter-only: Max acceptable price
	Modalities        []string             `json:"modalities,omitempty"`         // For image generation, e.g., ["image", "text"]
}

// ResponseFormat specifies the desired output format (e.g., JSON).
type ResponseFormat struct {
	Type       string            `json:"type"`                  // "json_object" or "json_schema"
	JSONSchema *JSONSchemaFormat `json:"json_schema,omitempty"` // Used when type is "json_schema"
}

// JSONSchemaFormat defines the schema for structured JSON output.
type JSONSchemaFormat struct {
	Name        string      `json:"name"`             // Required: A name for the schema.
	Strict      bool        `json:"strict,omitempty"` // Recommended: Ensure strict adherence.
	Description string      `json:"description,omitempty"`
	Schema      interface{} `json:"schema"` // Required: The actual JSON Schema object (can be map[string]interface{} or a specific struct).
}

// PredictionContent is used for OpenAI's predicted outputs feature.
type PredictionContent struct {
	Type    string `json:"type"` // Should be "content"
	Content string `json:"content"`
}

// Message represents a single message in the chat conversation.
type Message struct {
	Role       string     `json:"role"`                   // "system", "user", "assistant", or "tool"
	Content    Content    `json:"content,omitempty"`      // Can be string or array of ContentPart. Nullable for assistant tool calls.
	Name       string     `json:"name,omitempty"`         // Optional name for user/assistant, function name for tool.
	ToolCallID string     `json:"tool_call_id,omitempty"` // Required for role "tool".
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`   // Present for role "assistant" when making tool calls.
	Reasoning  *string    `json:"reasoning,omitempty"`    // Only in ResponseMessage: Reasoning steps if enabled.
}

// ResponseMessage is an alias for Message used in responses to include Reasoning field easily.
type ResponseMessage = Message

// DeltaMessage represents the delta content in a streaming response chunk.
type DeltaMessage struct {
	Role      *string             `json:"role,omitempty"` // Role usually appears only in the first chunk for assistant
	Content   *string             `json:"content,omitempty"`
	ToolCalls []StreamingToolCall `json:"tool_calls,omitempty"`
	Reasoning *string             `json:"reasoning,omitempty"` // Reasoning steps chunk.
}

// Content is a flexible type that can hold either a simple string
// or an array of ContentPart for multimodal input.
type Content interface{} // Can be string or []ContentPart

// ContentPart represents a part of a multimodal message content (text or image).
type ContentPart interface {
	contentType() string // Marker method
}

// TextContentPart represents a text part of a message.
type TextContentPart struct {
	Type string `json:"type"` // Always "text"
	Text string `json:"text"`
}

func (t TextContentPart) contentType() string { return t.Type }

// ImageContentPart represents an image part of a message.
type ImageContentPart struct {
	Type     string       `json:"type"` // Always "image_url"
	ImageURL ImageContent `json:"image_url"`
}

func (t ImageContentPart) contentType() string { return t.Type }

// ImageContent holds the URL or base64 data for an image.
type ImageContent struct {
	URL    string `json:"url"`              // URL or data URI (e.g., data:image/jpeg;base64,...)
	Detail string `json:"detail,omitempty"` // "auto", "low", "high", Default: "auto"
}

// CacheControlContentPart represents a text part with an Anthropic cache control directive.
// Note: This embeds TextContentPart logic for JSON marshaling but adds CacheControl.
type CacheControlContentPart struct {
	Type         string        `json:"type"` // Always "text"
	Text         string        `json:"text"`
	CacheControl *CacheControl `json:"cache_control"`
}

func (t CacheControlContentPart) contentType() string { return t.Type }

// CacheControl specifies caching behavior (Anthropic-specific).
type CacheControl struct {
	Type string `json:"type"` // e.g., "ephemeral"
}

// --- Chat Completion Response ---

// ChatCompletionResponse represents the full response for a non-streaming chat completion request.
type ChatCompletionResponse struct {
	ID                string         `json:"id"`
	Choices           []ChatChoice   `json:"choices"`
	Created           int64          `json:"created"` // Unix timestamp (seconds)
	Model             string         `json:"model"`   // Final model used
	Object            string         `json:"object"`  // "chat.completion"
	SystemFingerprint string         `json:"system_fingerprint,omitempty"`
	Usage             *ResponseUsage `json:"usage,omitempty"`       // Usage is required in non-streaming OpenAI spec
	Annotations       []Annotation   `json:"annotations,omitempty"` // For web search citations
}

// ChatCompletionChunk represents a single chunk in a streaming response.
type ChatCompletionChunk struct {
	ID                string                `json:"id"`
	Choices           []StreamingChatChoice `json:"choices"`
	Created           int64                 `json:"created"`
	Model             string                `json:"model"`
	Object            string                `json:"object"` // "chat.completion.chunk"
	SystemFingerprint string                `json:"system_fingerprint,omitempty"`
	Usage             *ResponseUsage        `json:"usage,omitempty"` // Included in the *last* chunk before [DONE]
}

// ChatChoice represents a single completion choice in a non-streaming response.
type ChatChoice struct {
	FinishReason       *string              `json:"finish_reason"`                  // "stop", "length", "tool_calls", "content_filter", "error", or null
	NativeFinishReason *string              `json:"native_finish_reason,omitempty"` // Raw reason from provider
	Index              int                  `json:"index"`                          // Choice index, typically 0
	Message            ResponseMessage      `json:"message"`                        // The generated message
	Logprobs           *LogprobResult       `json:"logprobs,omitempty"`
	Error              *InlineErrorResponse `json:"error,omitempty"` // Error specific to this choice
}

// StreamingChatChoice represents a single delta choice in a streaming response chunk.
type StreamingChatChoice struct {
	FinishReason       *string              `json:"finish_reason,omitempty"` // Only in the final delta chunk for a choice
	NativeFinishReason *string              `json:"native_finish_reason,omitempty"`
	Index              int                  `json:"index"`
	Delta              DeltaMessage         `json:"delta"`              // The message delta
	Logprobs           *LogprobResult       `json:"logprobs,omitempty"` // Logprobs can also be streamed
	Error              *InlineErrorResponse `json:"error,omitempty"`    // Error specific to this choice stream
}
