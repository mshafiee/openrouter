# openrouter-go

[![Go Reference](https://pkg.go.dev/badge/github.com/mshafiee/openrouter-go/client.svg)](https://pkg.go.dev/github.com/mshafiee/openrouter-go/client)
[![Go Report Card](https://goreportcard.com/badge/github.com/mshafiee/openrouter-go)](https://goreportcard.com/report/github.com/mshafiee/openrouter-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`openrouter-go` provides an idiomatic Go client library for interacting with the [OpenRouter API](https://openrouter.ai/docs).

OpenRouter offers a unified interface to access hundreds of AI models from various providers, optimizing for cost, performance, and reliability through features like automatic fallbacks and intelligent routing.

## Features

*   **Unified API Access:** Interact with Chat Completions, Legacy Completions, Model Listings, Billing, and more via a single client.
*   **Type-Safe:** Go structs provided for API requests and responses, mirroring the OpenAPI specification.
*   **Streaming Support:** Built-in handling for Server-Sent Events (SSE) for Chat Completions.
*   **API Key Management:** Programmatically manage API keys using Provisioning Keys.
*   **Error Handling:** Custom error type (`APIError`) provides structured access to API error details.
*   **Configurable:** Use functional options to configure HTTP client, base URL, default headers, and API keys.
*   **Context Aware:** All API calls accept `context.Context` for cancellation and timeouts.

## Installation

```bash
go get github.com/mshafiee/openrouter-go
```

Then import the client package in your Go code:

```go
import "github.com/mshafiee/openrouter-go/client"
```

## Authentication

Most API calls require an OpenRouter API key. You can obtain one from the [OpenRouter Keys page](https://openrouter.ai/keys).

It's recommended to load your API key from an environment variable rather than hardcoding it:

```bash
export OPENROUTER_API_KEY="sk-or-v1-..."
```

Instantiate the client with your key:

```go
import (
	"log"
	"os"
	"net/http"
	"time"

	"github.com/mshafiee/openrouter-go/client"
)

func main() {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("Error: OPENROUTER_API_KEY environment variable not set.")
	}

	// Optional: Add default headers for attribution, and custom HTTP client
	opts := []client.Option{
		client.WithDefaultHeader("User-Agent", "github.com/mshafiee/openrouter-go"),
		client.WithDefaultHeader("HTTP-Referer", "https://your-site-url.com"),
		client.WithDefaultHeader("X-Title", "Your Awesome App"),
	}
	// Example: Set a custom HTTP client (optional)
	customTransport := &http.Transport{
		TLSNextProto: make(map[string]func(string, *tls.Conn) http.RoundTripper),
	}
	customHttpClient := &http.Client{
		Timeout:   60 * time.Second,
		Transport: customTransport,
	}
	opts = append(opts, client.WithHTTPClient(customHttpClient))

	orClient, err := client.NewClient(apiKey, opts...)
	if err != nil {
		log.Fatalf("Error creating OpenRouter client: %v", err)
	}

	// Use orClient for API calls...
}
```

### Provisioning API Key

For API key management endpoints (like creating, listing, or deleting keys), you need a **Provisioning API Key**, obtainable from the [OpenRouter Provisioning Keys settings](https://openrouter.ai/settings/provisioning-keys).

Provide it using the `WithProvisioningKey` option:

```go
import "os"
import "github.com/mshafiee/openrouter-go/client"

// ...

	provKey := os.Getenv("OPENROUTER_PROVISIONING_KEY")
	if provKey == "" {
		// Handle error: provisioning key needed for key management
	}

	orClient, err := client.NewClient(
		apiKey, // Standard key still needed for potential non-key-management calls or future use
		client.WithProvisioningKey(provKey),
	)
	// ... use client.Keys methods ...

```

## Quick Start

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/mshafiee/openrouter-go/client"
)

func main() {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("Error: OPENROUTER_API_KEY environment variable not set.")
	}

	orClient, err := client.NewClient(
		apiKey,
		client.WithDefaultHeader("User-Agent", "github.com/mshafiee/openrouter-go"),
		client.WithDefaultHeader("HTTP-Referer", "https://your-site-url.com"),
		client.WithDefaultHeader("X-Title", "Your Awesome App"),
	)
	if err != nil {
		log.Fatalf("Error creating OpenRouter client: %v", err)
	}

	ctx := context.Background()
	chatReq := client.ChatCompletionRequest{
		Model: "openai/gpt-4o-mini", // Or your preferred model
		Messages: []client.Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "What is the capital of France?"},
		},
		MaxTokens:   PtrInt(100), // Use helper for pointer
		Temperature: PtrFloat64(0.7),
	}

	resp, err := orClient.Chat.Create(ctx, chatReq)
	if err != nil {
		var apiErr *client.APIError
		if errors.As(err, &apiErr) {
			log.Fatalf("OpenRouter API Error: Status %d, Message: %s", apiErr.StatusCode, apiErr.Error())
		}
		log.Fatalf("Error calling Chat.Create: %v", err)
	}

	if len(resp.Choices) > 0 {
		if contentStr, ok := resp.Choices[0].Message.Content.(string); ok {
			fmt.Println("Response:", contentStr)
		} else {
			fmt.Println("Response: (non-string content)")
		}
		if resp.Usage != nil {
			fmt.Printf("Usage: Prompt Tokens: %d, Completion Tokens: %d\n", resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
		}
	} else {
		fmt.Println("No choices returned.")
	}
}

// Helpers for pointer fields
func PtrInt(i int) *int             { return &i }
func PtrFloat64(f float64) *float64 { return &f }
```

## Usage Examples

### Chat Completion (Streaming)

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/mshafiee/openrouter-go/client"
)

func main() {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("Error: OPENROUTER_API_KEY environment variable not set.")
	}
	orClient, err := client.NewClient(apiKey)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	chatReq := client.ChatCompletionRequest{
		Model: "mistralai/mixtral-8x7b-instruct", // Updated model name
		Messages: []client.Message{
			{Role: "user", Content: "Write a short story about a curious robot exploring a garden."},
		},
		Stream:    true,
		MaxTokens: PtrInt(200),
	}

	stream, err := orClient.Chat.CreateStream(ctx, chatReq)
	if err != nil {
		log.Fatalf("Error creating stream: %v", err)
	}
	defer stream.Close() // IMPORTANT: Ensure stream is closed

	fmt.Println("Streaming Response:")
	var fullResponse strings.Builder
	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("\nStream finished.")
			break
		}
		if err != nil {
			// Check stream.Err() for potentially stored error
			if streamErr := stream.Err(); streamErr != nil {
				log.Printf("Stream error state: %v", streamErr)
			}
			log.Fatalf("Error receiving stream chunk: %v", err)
		}

		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta
			if delta.Content != nil {
				fmt.Print(*delta.Content)
				fullResponse.WriteString(*delta.Content)
			}
			if chunk.Choices[0].FinishReason != nil {
				log.Printf("\nStream Finish Reason: %s", *chunk.Choices[0].FinishReason)
			}
		}
		if chunk.Usage != nil {
			log.Printf("\nStream Usage: %+v", *chunk.Usage)
		}
	}

	if streamErr := stream.Err(); streamErr != nil {
		log.Fatalf("Stream encountered an error: %v", streamErr)
	} else {
		log.Printf("\nFull streamed response collected:\n%s", fullResponse.String())
		if stream.Usage != nil {
			log.Printf("Final Stream Usage: %+v", *stream.Usage)
		}
	}
}

// Helper for pointer fields
func PtrInt(i int) *int { return &i }
```

### List Available Models

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mshafiee/openrouter-go/client"
)

func main() {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	orClient, err := client.NewClient(apiKey)
	if err != nil { log.Fatal(err) }

	ctx := context.Background()
	resp, err := orClient.Model.List(ctx)
	if err != nil {
		log.Fatalf("Error listing models: %v", err)
	}

	fmt.Printf("Found %d models:\n", len(resp.Data))
	for i, model := range resp.Data {
		if i >= 5 { // Print first 5 for brevity
			fmt.Println("...")
			break
		}
		fmt.Printf("- %s (%s) - Context: %d\n", model.Name, model.ID, model.ContextLength)
	}
}
```

### List Endpoints for a Specific Model

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mshafiee/openrouter-go/client"
)

func main() {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	orClient, err := client.NewClient(apiKey)
	if err != nil { log.Fatal(err) }

	ctx := context.Background()
	author := "mistralai"
	slug := "mistral-7b-instruct"
	resp, err := orClient.Model.ListEndpoints(ctx, author, slug)
	if err != nil {
		log.Fatalf("Error listing endpoints for %s/%s: %v", author, slug, err)
	}

	fmt.Printf("Endpoints for %s:\n", resp.Data.Name)
	for _, endpoint := range resp.Data.Endpoints {
		fmt.Printf("- Provider: %s, Context: %.0f, Prompt Price: %s, Completion Price: %s\n",
			endpoint.ProviderName, endpoint.ContextLength, endpoint.Pricing.Prompt, endpoint.Pricing.Completion)
	}
}
```

### Get Credit Balance

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mshafiee/openrouter-go/client"
)

func main() {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	orClient, err := client.NewClient(apiKey)
	if err != nil { log.Fatal(err) }

	ctx := context.Background()
	resp, err := orClient.Billing.GetCredits(ctx)
	if err != nil {
		log.Fatalf("Error getting credits: %v", err)
	}
	fmt.Printf("Credits: Total Purchased: $%.4f, Total Used: $%.4f, Balance: $%.4f\n",
		resp.Data.TotalCredits, resp.Data.TotalUsage, resp.Data.TotalCredits-resp.Data.TotalUsage)
}
```

### API Key Management (Requires Provisioning Key)

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mshafiee/openrouter-go/client"
)

// Helper function for optional float64 pointers
func Ptr[T any](v T) *T {
	return &v
}

func main() {
	// Ensure client is initialized with WithProvisioningKey(...)
	provKey := os.Getenv("OPENROUTER_PROVISIONING_KEY")
	apiKey := os.Getenv("OPENROUTER_API_KEY") // May still be needed if mixing calls
	if provKey == "" {
		log.Fatal("OPENROUTER_PROVISIONING_KEY not set")
	}
	if apiKey == "" {
		log.Fatal("OPENROUTER_API_KEY not set")
	}

	provClient, err := client.NewClient(apiKey, client.WithProvisioningKey(provKey))
	if err != nil { log.Fatal(err) }

	ctx := context.Background()

	// List Keys
	listResp, err := provClient.Keys.List(ctx, nil) // Pass options for pagination/filtering
	if err != nil {
		log.Fatalf("Error listing keys: %v", err)
	}
	fmt.Printf("Found %d API keys:\n", len(listResp.Data))
	var firstKeyHash string // Variable to store the hash of the first key for update/delete demo
	for i, key := range listResp.Data {
		if i == 0 {
			firstKeyHash = key.Hash // Save the hash of the first key found
		}
		fmt.Printf("- Name: %s, Hash: %s, Limit: %v, Disabled: %t\n", key.Name, key.Hash, key.Limit, key.Disabled)
	}

	// Create a Key (Example)
	createReq := client.APIKeyCreateRequest{
		Name:  "My Test Key Go",
		Limit: Ptr(10.0), // Optional limit ($10 USD); Use Ptr helper for optional float64
	}
	createResp, err := provClient.Keys.Create(ctx, createReq)
	if err != nil {
		log.Fatalf("Error creating key: %v", err)
	}
	fmt.Printf("Created Key: Name: %s, Hash: %s, Key: %s\n", createResp.Data.Name, createResp.Data.Hash, createResp.Data.Key)
	// IMPORTANT: Store createResp.Data.Key securely! It's only shown once.
	newKeyHash := createResp.Data.Hash // Save hash for later operations

	// Update the *first* key found during list (if any)
	if firstKeyHash != "" {
		fmt.Printf("\nAttempting to update key with hash: %s\n", firstKeyHash)
		updateReq := client.APIKeyUpdateRequest{
			Name:     Ptr("Updated Test Key Name Go"),
			Disabled: Ptr(true), // Disable the key
			Limit:    Ptr(5.0),  // Change limit to $5
		}
		updateResp, err := provClient.Keys.Update(ctx, firstKeyHash, updateReq)
		if err != nil {
			log.Printf("WARN: Error updating key %s: %v (maybe it was deleted?)", firstKeyHash, err)
		} else {
			fmt.Printf("Updated Key: Name: %s, Hash: %s, Limit: %v, Disabled: %t\n",
				updateResp.Data.Name, updateResp.Data.Hash, updateResp.Data.Limit, updateResp.Data.Disabled)
		}
	} else {
		fmt.Println("\nSkipping update example as no existing keys were listed.")
	}


	// Delete the *newly created* key (Example - use the hash from creation)
	if newKeyHash != "" {
		fmt.Printf("\nAttempting to delete newly created key with hash: %s\n", newKeyHash)
		delResp, err := provClient.Keys.Delete(ctx, newKeyHash)
		if err != nil {
			log.Fatalf("Error deleting key %s: %v", newKeyHash, err)
		}
		fmt.Printf("Deleted key %s: Success: %t\n", newKeyHash, delResp.Data.Success)
	} else {
		fmt.Println("\nSkipping delete example as key creation failed or hash wasn't captured.")
	}
}
```

### Error Handling

Methods return an `error`. If the error originates from the OpenRouter API (non-2xx status code), you can use `errors.As` to check if it's a `*client.APIError`.

```go
import (
	"errors"
	"fmt"
	"log"
	"net/http" // Required for http status constants

	"github.com/mshafiee/openrouter-go/client"
)

func handleApiCall(ctx context.Context, orClient *client.Client) {
	// Example: Replace with actual API call
	chatReq := client.ChatCompletionRequest{
		Model: "invalid/model-name", // Intentionally cause an error
		Messages: []client.Message{
			{Role: "user", Content: "Test"},
		},
	}
	_, err := orClient.Chat.Create(ctx, chatReq)

	if err != nil {
		var apiErr *client.APIError
		if errors.As(err, &apiErr) {
			// Access specific error details
			fmt.Printf("API Error Status: %d\n", apiErr.StatusCode)
			fmt.Printf("API Error Message: %s\n", apiErr.Message)
			if apiErr.Details != nil {
				// Details contains the parsed *ErrorBody struct
				fmt.Printf("API Error Code (from JSON): %d\n", apiErr.Details.Code)
				fmt.Printf("API Error Message (from JSON): %s\n", apiErr.Details.Message)
				if apiErr.Details.Metadata != nil {
					fmt.Printf("API Error Metadata: %v\n", apiErr.Details.Metadata)
				}
			}
			// Optionally print raw body for debugging non-standard errors
			// fmt.Printf("API Raw Response: %s\n", string(apiErr.RawBody))

			// Decide how to handle based on status code or message
			if apiErr.StatusCode == http.StatusTooManyRequests {
				fmt.Println("Rate limit exceeded. Consider backoff and retry.")
			} else if apiErr.StatusCode == http.StatusPaymentRequired {
				fmt.Println("Insufficient credits.")
			}
		} else {
			// Handle other errors (network, context cancellation, etc.)
			log.Printf("Non-API error: %v", err)
		}
		// Stop processing or return error
		return
	}
	// ... process successful response ...
	fmt.Println("API call successful!")
}
```

## Configuration Options

The `client.NewClient` function accepts functional options for customization:

*   `client.WithHTTPClient(*http.Client)`: Use a custom HTTP client.
*   `client.WithBaseURL(string)`: Set a different base URL (e.g., for testing or a proxy).
*   `client.WithProvisioningKey(string)`: Provide the key for key management API calls.
*   `client.WithUserAgent(string)`: Set a custom User-Agent header.
*   `client.WithDefaultHeader(key, value string)`: Add headers sent with every request (e.g., `HTTP-Referer`, `X-Title`).

## API Coverage

This client aims to provide coverage for the OpenRouter API v1, including:

*   Chat Completions (`/chat/completions`) - Create & Stream
*   Legacy Completions (`/completions`) - Create (Deprecated)
*   Generation Metadata (`/generation`) - Get
*   Models (`/models`) - List Models, List Endpoints
*   Billing (`/credits`, `/credits/coinbase`) - Get Credits, Create Coinbase Charge
*   Authentication (`/auth/keys`) - Exchange Auth Code (PKCE)
*   API Keys (`/key`, `/keys`, `/keys/{hash}`) - Get Current, List, Create, Get, Update, Delete (requires Provisioning Key)

Please refer to the [GoDoc Reference](https://pkg.go.dev/github.com/mshafiee/openrouter-go/client) for detailed information on specific methods and types.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request or open an Issue.
<!-- Consider adding a CONTRIBUTING.md file with more details -->

## License

This library is distributed under the MIT License.
