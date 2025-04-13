package main

import (
	"context"
	"crypto/tls"
	"encoding/base64" // For image encoding
	"encoding/json"   // Import encoding/json
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings" // For image MIME type check
	"time"

	// Adjust this import path to where your client package resides
	"github.com/mshafiee/openrouter-go/client"
)

func main() {
	// --- Configuration ---
	// Get API Key from environment variable
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("Error: OPENROUTER_API_KEY environment variable not set.")
	}

	log.Printf("DEBUG: Using OPENROUTER_API_KEY starting with: %s...", apiKey[:15]) // Print prefix

	// Optional: Get Provisioning Key if testing key management
	provisioningKey := os.Getenv("OPENROUTER_PROVISIONING_API_KEY")

	// Optional: Site info for ranking
	siteURL := os.Getenv("YOUR_SITE_URL")   // e.g., "https://myapp.com"
	siteName := os.Getenv("YOUR_SITE_NAME") // e.g., "My Awesome App"

	// --- Client Initialization ---
	opts := []client.Option{
		client.WithDefaultHeader("User-Agent", "github.com/mshafiee/openrouter-go"),
	}
	if siteURL != "" {
		opts = append(opts, client.WithDefaultHeader("HTTP-Referer", siteURL))
	}
	if siteName != "" {
		opts = append(opts, client.WithDefaultHeader("X-Title", siteName))
	}
	if provisioningKey != "" {
		opts = append(opts, client.WithProvisioningKey(provisioningKey))
	}
	// Example: Set a custom HTTP client that disables HTTP/2
	customTransport := &http.Transport{
		// Disable HTTP/2 by setting TLSNextProto to empty map
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

	log.Println("OpenRouter Client Initialized.")

	// --- Context for Requests ---
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second) // 90-second timeout
	defer cancel()

	// === Example API Calls ===
	// Choose which examples to run by uncommenting them

	// --- 1. Simple Chat Completion ---
	exampleSimpleChat(ctx, orClient)

	// --- 2. Streaming Chat Completion ---
	exampleStreamingChat(ctx, orClient)

	// --- 3. Chat Completion with Tool Use ---
	exampleChatWithTools(ctx, orClient)

	// --- 4. Get Generation Metadata ---
	// Replace with a real ID from your activity to see cost/token details
	exampleGetGeneration(ctx, orClient, "YOUR_GENERATION_ID_HERE")

	// --- 5. List Models ---
	exampleListModels(ctx, orClient)

	// --- 6. List Model Endpoints ---
	exampleListModelEndpoints(ctx, orClient, "openai", "gpt-4o")

	// --- 7. Get Credit Balance ---
	exampleGetCredits(ctx, orClient)

	// --- 8. Get Current API Key Info ---
	exampleGetCurrentKey(ctx, orClient)

	// --- 9. API Key Management (Requires Provisioning Key) ---
	exampleManageKeys(ctx, orClient) // Uncomment only if Provisioning Key is set

	// --- 10. Multimodal Chat (Image Input) ---
	// Requires a local file named "image.jpg" (or png/webp) in the same directory
	exampleMultimodalChat(ctx, orClient, "./image.jpg")

	// --- 11. Structured Output (JSON Schema) ---
	exampleStructuredOutputChat(ctx, orClient)

	// --- 12. Response Format (JSON Object) ---
	exampleJsonObjectChat(ctx, orClient)

	// --- 13. Provider Routing & Fallback ---
	exampleRoutingChat(ctx, orClient)

	// --- 14. Reasoning Tokens (Anthropic Example) ---
	exampleReasoningChat(ctx, orClient)

	// --- 15. Logprobs ---
	exampleLogprobsChat(ctx, orClient)

	// --- 16. Prompt Caching (Anthropic Example Structure) ---
	// Demonstrates request structure; actual caching benefit requires repeated calls
	// and checking cost via GetGeneration or Activity Page.
	// exampleAnthropicCacheChat(ctx, orClient)

	log.Println("Example execution finished.")
}

// --- Example Functions ---

func exampleSimpleChat(ctx context.Context, orClient *client.Client) {
	log.Println("--- Running Simple Chat Completion Example ---")
	request := client.ChatCompletionRequest{
		Model: "openai/gpt-4o-mini", // Use a fast and capable model
		Messages: []client.Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "What is the weather like in Berlin in May?"},
		},
		MaxTokens:   PtrInt(150), // Use helper PtrInt to get *int
		Temperature: PtrFloat64(0.7),
	}

	response, err := orClient.Chat.Create(ctx, request)
	handleError("Simple Chat Completion", err)
	if err == nil {
		if len(response.Choices) > 0 {
			log.Printf("Chat Response Model: %s", response.Model)
			// Check if FinishReason is nil before dereferencing
			finishReason := "<nil>"
			if response.Choices[0].FinishReason != nil {
				finishReason = *response.Choices[0].FinishReason
			}
			log.Printf("Chat Response Finish Reason: %s", finishReason)
			// Content can be string or []ContentPart, handle type assertion if needed
			// Assuming string content for this simple example
			if contentStr, ok := response.Choices[0].Message.Content.(string); ok {
				log.Printf("Chat Response Content: %s", contentStr)
			} else {
				log.Printf("Chat Response Content: (non-string type received)")
			}

			if response.Usage != nil {
				log.Printf("Chat Usage: %+v", *response.Usage)
			}
		} else {
			log.Println("Chat Response: Received no choices.")
		}
	}
	log.Println("--- Simple Chat Completion Example Finished ---")
}

func exampleStreamingChat(ctx context.Context, orClient *client.Client) {
	log.Println("--- Running Streaming Chat Completion Example ---")
	request := client.ChatCompletionRequest{
		Model: "mistralai/mixtral-8x7b-instruct", // Another good streaming model
		Messages: []client.Message{
			{Role: "user", Content: "Write a short poem about Go programming."},
		},
		Stream:    true, // MUST be true for streaming
		MaxTokens: PtrInt(200),
	}

	stream, err := orClient.Chat.CreateStream(ctx, request)
	handleError("Streaming Chat Start", err)
	if err != nil {
		return
	}
	defer stream.Close() // IMPORTANT: Ensure stream body is closed

	log.Println("Streaming Response:")
	fullResponse := ""
	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			log.Println("\nStream finished.")
			break
		}
		handleError("Streaming Chat Recv", err)
		if err != nil {
			// Check stream.Err() for potentially stored error
			if streamErr := stream.Err(); streamErr != nil {
				log.Printf("Stream error state: %v", streamErr)
			}
			break // Exit loop on receive error
		}

		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta
			if delta.Content != nil {
				fmt.Print(*delta.Content) // Print content chunks as they arrive
				fullResponse += *delta.Content
			}
			// Could also check for delta.Role, delta.ToolCalls, delta.Reasoning here
			if chunk.Choices[0].FinishReason != nil {
				log.Printf("\nStream Finish Reason: %s", *chunk.Choices[0].FinishReason)
			}
		}
		// The last chunk *before* EOF might contain Usage
		if chunk.Usage != nil {
			log.Printf("\nStream Usage: %+v", *chunk.Usage)
		}
	}

	// Check for errors encountered *during* streaming after the loop
	if streamErr := stream.Err(); streamErr != nil {
		log.Printf("Error occurred during stream processing: %v", streamErr)
	} else {
		log.Printf("\nFull streamed response collected:\n%s", fullResponse)
		// Usage might also be available on the stream object *after* EOF
		if stream.Usage != nil {
			log.Printf("Final Stream Usage: %+v", *stream.Usage)
		}
	}

	log.Println("--- Streaming Chat Completion Example Finished ---")
}

// --- Tool Use Example ---

// Simple function to simulate getting current weather
func getCurrentWeather(location string, unit string) (map[string]interface{}, error) {
	log.Printf("[Tool Call] Getting weather for %s in %s\n", location, unit)
	// In a real scenario, call an external weather API here
	temp := 22.0
	desc := "Partly cloudy"
	if unit == "fahrenheit" {
		temp = temp*1.8 + 32
		unit = "F"
	} else {
		unit = "C"
	}
	return map[string]interface{}{
		"location":    location,
		"temperature": fmt.Sprintf("%.1f%s", temp, unit),
		"description": desc,
	}, nil
}

func exampleChatWithTools(ctx context.Context, orClient *client.Client) {
	log.Println("--- Running Chat Completion with Tools Example ---")

	messages := []client.Message{
		{Role: "system", Content: "You are a helpful assistant that can get weather information."},
		{Role: "user", Content: "What's the weather in London like right now in fahrenheit?"},
	}

	tools := []client.Tool{
		{
			Type: "function",
			Function: client.FunctionDescription{
				Name:        "getCurrentWeather",
				Description: "Get the current weather in a given location",
				Parameters: map[string]interface{}{ // Use map[string]interface{} for JSON schema flexibility
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]interface{}{
							"type":        "string",
							"description": "The city and state, e.g. San Francisco, CA",
						},
						"unit": map[string]interface{}{
							"type": "string",
							"enum": []string{"celsius", "fahrenheit"},
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	// --- First API Call: Ask the model, potentially requesting a tool call ---
	log.Println("Round 1: Sending request to model...")

	// Correct way to assign a string value like "auto" to ToolChoice (*interface{})
	var toolChoiceValue client.ToolChoice = "auto" // Assign the string to the interface type
	request1 := client.ChatCompletionRequest{
		Model:      "openai/gpt-4o", // A model known to support tool use well
		Messages:   messages,
		Tools:      tools,
		ToolChoice: &toolChoiceValue, // Pass the address of the interface variable
	}

	response1, err := orClient.Chat.Create(ctx, request1)
	handleError("Tool Use - Round 1", err)
	if err != nil {
		return
	}

	if len(response1.Choices) == 0 {
		log.Println("Tool Use - Round 1: No choices received.")
		return
	}

	// Add assistant's response (which should contain the tool call request)
	assistantMsg1 := response1.Choices[0].Message
	messages = append(messages, assistantMsg1) // Append the *entire* message struct

	// --- Check if a tool call was requested ---
	if assistantMsg1.ToolCalls == nil || len(assistantMsg1.ToolCalls) == 0 {
		log.Println("Tool Use - Round 1: Model did not request a tool call. Response:")
		if contentStr, ok := assistantMsg1.Content.(string); ok {
			log.Println(contentStr)
		} else {
			log.Println("(non-string content received)")
		}
		log.Println("--- Chat Completion with Tools Example Finished ---")
		return
	}

	log.Println("Round 1: Model requested tool call(s).")

	// --- Process Tool Calls ---
	for _, toolCall := range assistantMsg1.ToolCalls {
		if toolCall.Type == "function" {
			log.Printf("Processing function call: %s (ID: %s)\n", toolCall.Function.Name, toolCall.ID)

			// Simple routing based on function name
			if toolCall.Function.Name == "getCurrentWeather" {
				// Decode arguments
				var args struct {
					Location string `json:"location"`
					Unit     string `json:"unit"`
				}
				// Arguments come as a JSON string, need to unmarshal
				err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
				if err != nil {
					log.Printf("Error decoding arguments for %s: %v\n", toolCall.Function.Name, err)
					// Append an error message back to the model? Or handle differently?
					// For simplicity, we'll just log and potentially skip to next call.
					continue // Skip this tool call if args are bad
				}
				// Provide default unit if not specified by the model
				if args.Unit == "" {
					args.Unit = "celsius"
				}

				// Call the actual function
				toolResultData, toolErr := getCurrentWeather(args.Location, args.Unit)
				var toolResultContent string
				if toolErr != nil {
					log.Printf("Error executing tool %s: %v\n", toolCall.Function.Name, toolErr)
					toolResultContent = fmt.Sprintf(`{"error": "Failed to execute tool: %v"}`, toolErr)
				} else {
					// Marshal result back to JSON string for the model
					resultBytes, marshalErr := json.Marshal(toolResultData)
					if marshalErr != nil {
						log.Printf("Error marshaling tool result for %s: %v\n", toolCall.Function.Name, marshalErr)
						toolResultContent = fmt.Sprintf(`{"error": "Failed to encode tool result: %v"}`, marshalErr)
					} else {
						toolResultContent = string(resultBytes)
					}
				}

				// Append the tool result message
				messages = append(messages, client.Message{
					Role:       "tool",
					ToolCallID: toolCall.ID, // Link result to the request
					Name:       toolCall.Function.Name,
					Content:    toolResultContent, // Content must be string representation of the result
				})
				log.Printf("Appended tool result for %s.\n", toolCall.Function.Name)

			} else {
				log.Printf("Warning: Received request for unknown function: %s\n", toolCall.Function.Name)
				// Append a message indicating the tool is unknown?
				messages = append(messages, client.Message{
					Role:       "tool",
					ToolCallID: toolCall.ID,
					Name:       toolCall.Function.Name,
					Content:    `{"error": "Function not found"}`,
				})
			}
		}
	}

	// --- Second API Call: Send tool results back to the model ---
	log.Println("Round 2: Sending tool results back to model...")
	request2 := client.ChatCompletionRequest{
		Model:    "openai/gpt-4o",
		Messages: messages, // Send the whole history including tool req/resp
		Tools:    tools,    // Include tools definition again (required by some models)
		// ToolChoice: "none", // Optionally force model to *not* call tools again
	}

	response2, err := orClient.Chat.Create(ctx, request2)
	handleError("Tool Use - Round 2", err)
	if err == nil {
		if len(response2.Choices) > 0 {
			log.Println("Tool Use - Final Response:")
			if contentStr, ok := response2.Choices[0].Message.Content.(string); ok {
				log.Println(contentStr)
			} else {
				log.Println("(non-string content received)")
			}
		} else {
			log.Println("Tool Use - Round 2: No choices received.")
		}
	}

	log.Println("--- Chat Completion with Tools Example Finished ---")
}

// --- Other Examples ---

func exampleGetGeneration(ctx context.Context, orClient *client.Client, genID string) {
	log.Println("--- Running Get Generation Example ---")
	if genID == "" || genID == "YOUR_GENERATION_ID_HERE" {
		log.Println("Skipping Get Generation: Please provide a valid Generation ID.")
		return
	}
	response, err := orClient.Generation.Get(ctx, genID)
	handleError("Get Generation", err)
	if err == nil {
		log.Printf("Generation Details (%s):\n", response.Data.ID)
		log.Printf("  Model: %s\n", response.Data.Model)
		if response.Data.ProviderName != nil {
			log.Printf("  Provider: %s\n", *response.Data.ProviderName)
		} else {
			log.Printf("  Provider: <nil>\n")
		}
		log.Printf("  Created: %s\n", response.Data.CreatedAt.Format(time.RFC3339))
		log.Printf("  Cost: $%.8f\n", response.Data.TotalCost)
		if response.Data.CacheDiscount != nil {
			log.Printf("  Cache Discount: $%.8f\n", *response.Data.CacheDiscount)
		}
		// Log native token counts if available
		if response.Data.NativeTokensPrompt != nil {
			log.Printf("  Native Prompt Tokens: %d\n", *response.Data.NativeTokensPrompt)
		}
		if response.Data.NativeTokensCompletion != nil {
			log.Printf("  Native Completion Tokens: %d\n", *response.Data.NativeTokensCompletion)
		}
		// Add more fields as needed...
	}
	log.Println("--- Get Generation Example Finished ---")
}

func exampleListModels(ctx context.Context, orClient *client.Client) {
	log.Println("--- Running List Models Example ---")
	response, err := orClient.Model.List(ctx)
	handleError("List Models", err)
	if err == nil {
		log.Printf("Found %d models. First few:\n", len(response.Data))
		limit := 5
		if len(response.Data) < limit {
			limit = len(response.Data)
		}
		for i := 0; i < limit; i++ {
			model := response.Data[i]
			pricePrompt := "<nil>"
			if model.Pricing != nil {
				pricePrompt = model.Pricing.Prompt // Assuming Pricing is always present if model is valid
			}
			log.Printf("  - ID: %s, Name: %s, Context: %d, Price (Prompt): %s\n",
				model.ID, model.Name, model.ContextLength, pricePrompt)
		}
		// To see all models:
		// for _, model := range response.Data { ... }
	}
	log.Println("--- List Models Example Finished ---")
}

func exampleListModelEndpoints(ctx context.Context, orClient *client.Client, author, slug string) {
	log.Println("--- Running List Model Endpoints Example ---")
	response, err := orClient.Model.ListEndpoints(ctx, author, slug)
	handleError(fmt.Sprintf("List Endpoints for %s/%s", author, slug), err)
	if err == nil {
		log.Printf("Model: %s (%s)\n", response.Data.Name, response.Data.ID)
		log.Printf("Found %d endpoints:\n", len(response.Data.Endpoints))
		for _, ep := range response.Data.Endpoints {
			log.Printf("  - Provider: %s (%s)\n", ep.Name, ep.ProviderName)
			if ep.Pricing != nil {
				log.Printf("    Pricing (Prompt/Completion): %s / %s\n", ep.Pricing.Prompt, ep.Pricing.Completion)
			}
			if ep.ContextLength != nil {
				log.Printf("    Context Length: %.0f\n", *ep.ContextLength)
			}
			log.Printf("    Supported Params: %v\n", ep.SupportedParameters)
		}
	}
	log.Println("--- List Model Endpoints Example Finished ---")
}

func exampleGetCredits(ctx context.Context, orClient *client.Client) {
	log.Println("--- Running Get Credits Example ---")
	response, err := orClient.Billing.GetCredits(ctx)
	handleError("Get Credits", err)
	if err == nil {
		balance := response.Data.TotalCredits - response.Data.TotalUsage
		log.Printf("Total Credits Purchased: $%.4f\n", response.Data.TotalCredits)
		log.Printf("Total Credits Used:      $%.4f\n", response.Data.TotalUsage)
		log.Printf("Current Balance:         $%.4f\n", balance)
	}
	log.Println("--- Get Credits Example Finished ---")
}

func exampleGetCurrentKey(ctx context.Context, orClient *client.Client) {
	log.Println("--- Running Get Current Key Info Example ---")
	response, err := orClient.Keys.GetCurrent(ctx)
	handleError("Get Current Key Info", err)
	if err == nil {
		log.Printf("Current Key Info:\n")
		log.Printf("  Label: %s\n", response.Data.Label)
		if response.Data.Limit != nil {
			log.Printf("  Limit: $%.4f\n", *response.Data.Limit)
			if response.Data.LimitRemaining != nil { // Check LimitRemaining separately
				log.Printf("  Remaining: $%.4f\n", *response.Data.LimitRemaining)
			} else {
				log.Printf("  Remaining: <nil>\n")
			}
		} else {
			log.Printf("  Limit: Unlimited\n")
		}
		log.Printf("  Usage: $%.4f\n", response.Data.Usage)
		log.Printf("  Is Provisioning: %t\n", response.Data.IsProvisioningKey)
		log.Printf("  Rate Limit: %d requests / %s\n", response.Data.RateLimit.Requests, response.Data.RateLimit.Interval)
	}
	log.Println("--- Get Current Key Info Example Finished ---")
}

func exampleManageKeys(ctx context.Context, orClient *client.Client) {
	log.Println("--- Running API Key Management Example (Requires Provisioning Key) ---")
	// Use the exported helper method to check if the key is set
	if !orClient.IsProvisioning() {
		log.Println("Skipping Key Management: Provisioning Key not set on client.")
		return
	}

	// 1. List keys
	log.Println("Listing initial keys (limit 5)...")
	listParams := client.ListKeysParams{} // Default: limit 100, include disabled=false
	listResp, err := orClient.Keys.List(ctx, listParams)
	handleError("List Keys", err)
	if err != nil {
		return // Exit if listing fails
	}
	log.Printf("Found %d total keys initially.\n", len(listResp.Data)) // Note: This only shows first page if > 100 keys
	limit := 5
	if len(listResp.Data) < limit {
		limit = len(listResp.Data)
	}
	for i := 0; i < limit; i++ {
		k := listResp.Data[i]
		log.Printf("  - Hash: %s, Name: %s, Label: %s, Disabled: %t\n", k.Hash, k.Name, derefString(k.Label), k.Disabled)
	}

	// 2. Create a new key
	log.Println("Creating a new test key...")
	createReq := client.CreateApiKeyRequest{
		Name:  "Go Client Test Key",
		Label: PtrString("go-test-key"),
		Limit: PtrFloat64(1.50), // $1.50 limit
	}
	createResp, err := orClient.Keys.Create(ctx, createReq)
	handleError("Create Key", err)
	if err != nil {
		return // Exit if create fails
	}
	newKeyHash := createResp.Data.Hash
	newKeyString := createResp.Data.Key // IMPORTANT: Save this key string securely!
	log.Printf("Created key with Hash: %s\n", newKeyHash)
	log.Printf("IMPORTANT: New API Key String: %s (SAVE THIS SECURELY!)\n", newKeyString)

	// 3. Get the created key
	log.Printf("Getting details for key %s...\n", newKeyHash)
	getResp, err := orClient.Keys.Get(ctx, newKeyHash)
	handleError("Get Key", err)
	if err == nil {
		limitVal := "<nil>"
		if getResp.Data.Limit != nil {
			limitVal = fmt.Sprintf("%.2f", *getResp.Data.Limit)
		}
		log.Printf("  Got Key - Name: %s, Limit: %s\n", getResp.Data.Name, limitVal)
	}

	// 4. Update the key (e.g., disable it)
	log.Printf("Updating key %s (disabling it)...\n", newKeyHash)
	updateReq := client.UpdateApiKeyRequest{
		Name:     PtrString("Go Client Test Key [Updated/Disabled]"),
		Disabled: PtrBool(true),
		Limit:    PtrFloat64(2.0), // Update limit
	}
	updateResp, err := orClient.Keys.Update(ctx, newKeyHash, updateReq)
	handleError("Update Key", err)
	if err == nil {
		limitVal := "<nil>"
		if updateResp.Data.Limit != nil {
			limitVal = fmt.Sprintf("%.2f", *updateResp.Data.Limit)
		}
		log.Printf("  Updated Key - Name: %s, Disabled: %t, Limit: %s\n",
			updateResp.Data.Name, updateResp.Data.Disabled, limitVal)
	}

	// 5. Delete the key
	log.Printf("Deleting key %s...\n", newKeyHash)
	deleteResp, err := orClient.Keys.Delete(ctx, newKeyHash)
	handleError("Delete Key", err)
	if err == nil {
		log.Printf("  Delete Key Success: %t\n", deleteResp.Data.Success)
	}

	log.Println("--- API Key Management Example Finished ---")
}

func exampleMultimodalChat(ctx context.Context, orClient *client.Client, imagePath string) {
	log.Println("--- Running Multimodal Chat Completion Example ---")

	// --- Image Preprocessing ---
	// Ensure the image file exists before proceeding
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		// Create a dummy placeholder if it doesn't exist
		log.Printf("Warning: Image file '%s' not found. Creating a dummy placeholder.", imagePath)
		dummyData := []byte{
			0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
			0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
			0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
			0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
			0x42, 0x60, 0x82, // Minimal 1x1 black PNG
		}
		if writeErr := os.WriteFile(imagePath, dummyData, 0644); writeErr != nil {
			log.Printf("Error creating dummy image file: %v. Skipping example.", writeErr)
			handleError("Multimodal Chat - Setup", writeErr) // Log error appropriately
			return
		}
	}

	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		log.Printf("Error reading image file '%s': %v. Skipping example.", imagePath, err)
		handleError("Multimodal Chat - Read Image", err)
		return
	}
	base64Image := base64.StdEncoding.EncodeToString(imageData)
	// Determine MIME type (simplistic check for demo)
	mimeType := "image/png" // Default to PNG for dummy
	if strings.HasSuffix(strings.ToLower(imagePath), ".jpg") || strings.HasSuffix(strings.ToLower(imagePath), ".jpeg") {
		mimeType = "image/jpeg"
	} else if strings.HasSuffix(strings.ToLower(imagePath), ".webp") {
		mimeType = "image/webp"
	}
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Image)
	// --- End Image Preprocessing ---

	request := client.ChatCompletionRequest{
		Model: "openai/gpt-4o", // Or another capable multimodal model
		Messages: []client.Message{
			{
				Role: "user",
				// Content needs to be a slice of ContentPart interfaces
				Content: []client.ContentPart{
					client.TextContentPart{Type: "text", Text: "What is in this image? Describe it."},
					client.ImageContentPart{
						Type: "image_url",
						ImageURL: client.ImageContent{
							URL:    dataURI, // Use the base64 data URI
							Detail: "auto",  // Optional: low, high, auto
						},
					},
				},
			},
		},
		MaxTokens: PtrInt(300),
	}

	response, err := orClient.Chat.Create(ctx, request)
	handleError("Multimodal Chat Completion", err)
	if err == nil {
		if len(response.Choices) > 0 {
			log.Printf("Multimodal Chat Response Model: %s", response.Model)
			if contentStr, ok := response.Choices[0].Message.Content.(string); ok {
				log.Printf("Multimodal Chat Response Content: %s", contentStr)
			} else {
				log.Printf("Multimodal Chat Response Content: (non-string type received)")
			}
		} else {
			log.Println("Multimodal Chat Response: Received no choices.")
		}
	}
	log.Println("--- Multimodal Chat Completion Example Finished ---")
}

func exampleStructuredOutputChat(ctx context.Context, orClient *client.Client) {
	log.Println("--- Running Structured Output (JSON Schema) Example ---")

	// Define the desired JSON schema using map[string]interface{}
	recipeSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"recipe_name": map[string]interface{}{
				"type":        "string",
				"description": "The name of the recipe",
			},
			"ingredients": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{"type": "string"},
						"quantity": map[string]interface{}{
							"type":   "number",
							"format": "float", // Indicate float if necessary
						},
						"unit": map[string]interface{}{"type": "string"},
					},
					"required": []string{"name", "quantity", "unit"},
				},
				"description": "List of ingredients",
			},
			"steps": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Cooking instructions",
			},
		},
		"required":             []string{"recipe_name", "ingredients", "steps"},
		"additionalProperties": false, // Often good practice
	}

	request := client.ChatCompletionRequest{
		Model: "openai/gpt-4o", // Model must support structured outputs / json_schema
		Messages: []client.Message{
			{Role: "system", Content: "You are a helpful cooking assistant that outputs recipes in the specified JSON format."},
			{Role: "user", Content: "Please provide a simple recipe for chocolate chip cookies using the provided JSON schema."},
		},
		ResponseFormat: &client.ResponseFormat{
			Type: "json_schema",
			JSONSchema: &client.JSONSchemaFormat{
				Name:   "recipe_schema", // Name required by spec
				Strict: true,            // Enforce schema strictly
				Schema: recipeSchema,
			},
		},
		Temperature: PtrFloat64(0.5), // Lower temp might help adherence
		// Ensure provider routing doesn't pick a provider that doesn't support it
		Provider: &client.ProviderPreferences{
			RequireParameters: PtrBool(true), // Filter for providers supporting response_format:json_schema
		},
	}

	response, err := orClient.Chat.Create(ctx, request)
	handleError("Structured Output Chat", err)
	if err == nil && len(response.Choices) > 0 {
		log.Printf("Structured Output Response Model: %s", response.Model)
		if contentStr, ok := response.Choices[0].Message.Content.(string); ok {
			log.Printf("Structured Output Raw Content:\n%s", contentStr)
			// Try to unmarshal the result to verify it's valid JSON matching the structure (basic check)
			var recipeData map[string]interface{}
			if jsonErr := json.Unmarshal([]byte(contentStr), &recipeData); jsonErr != nil {
				log.Printf("Warning: Failed to unmarshal structured output response: %v", jsonErr)
			} else {
				log.Printf("Successfully unmarshaled structured output.")
				// Optionally print parts of the decoded data
				// if name, ok := recipeData["recipe_name"].(string); ok {
				// 	log.Printf(" Parsed Recipe Name: %s", name)
				// }
			}
		} else {
			log.Printf("Structured Output Content: (non-string type received)")
		}
	}
	log.Println("--- Structured Output (JSON Schema) Example Finished ---")
}

func exampleJsonObjectChat(ctx context.Context, orClient *client.Client) {
	log.Println("--- Running JSON Object Response Format Example ---")

	request := client.ChatCompletionRequest{
		Model: "openai/gpt-4o-mini", // Model must support response_format:json_object
		Messages: []client.Message{
			{Role: "system", Content: "You are an information extraction assistant. Respond ONLY with valid JSON."},
			{Role: "user", Content: "Extract the name and city from the following text: 'User John Doe lives in New York City.'"},
		},
		ResponseFormat: &client.ResponseFormat{
			Type: "json_object", // Request JSON object output
		},
		Temperature: PtrFloat64(0.2),
		Provider: &client.ProviderPreferences{
			RequireParameters: PtrBool(true), // Filter for providers supporting response_format:json_object
		},
	}

	response, err := orClient.Chat.Create(ctx, request)
	handleError("JSON Object Chat", err)
	if err == nil && len(response.Choices) > 0 {
		log.Printf("JSON Object Response Model: %s", response.Model)
		if contentStr, ok := response.Choices[0].Message.Content.(string); ok {
			log.Printf("JSON Object Raw Content:\n%s", contentStr)
			// Try to unmarshal the result to verify it's valid JSON
			var jsonData map[string]interface{}
			if jsonErr := json.Unmarshal([]byte(contentStr), &jsonData); jsonErr != nil {
				log.Printf("Warning: Failed to unmarshal JSON object response: %v", jsonErr)
			} else {
				log.Printf("Successfully unmarshaled JSON object response: %+v", jsonData)
			}
		} else {
			log.Printf("JSON Object Content: (non-string type received)")
		}
	}
	log.Println("--- JSON Object Response Format Example Finished ---")
}

func exampleRoutingChat(ctx context.Context, orClient *client.Client) {
	log.Println("--- Running Provider Routing / Fallback Example ---")

	// --- Example 1: Using `Models` for Fallback ---
	log.Println(" Attempting fallback using 'Models' field...")
	requestFallback := client.ChatCompletionRequest{
		// Primary model likely doesn't exist or could fail
		Model: "nonexistent/model-v1",
		// Define fallback models in preferred order
		Models: []string{"google/gemini-flash-1.5", "openai/gpt-4o-mini"},
		Messages: []client.Message{
			{Role: "user", Content: "Tell me a very short joke."},
		},
		MaxTokens: PtrInt(50),
	}
	responseFallback, err := orClient.Chat.Create(ctx, requestFallback)
	handleError("Routing Chat (Models Fallback)", err)
	if err == nil && len(responseFallback.Choices) > 0 {
		log.Printf(" Fallback Response: Model Used: %s", responseFallback.Model)
		if contentStr, ok := responseFallback.Choices[0].Message.Content.(string); ok {
			log.Printf(" Fallback Response Content: %s", contentStr)
		}
	}

	// --- Example 2: Using `Provider` preferences ---
	log.Println("\n Attempting routing using 'Provider' preferences (order Anthropic, OpenAI)...")
	requestProvider := client.ChatCompletionRequest{
		Model: "mistralai/mixtral-8x7b-instruct", // A model available on multiple providers
		Messages: []client.Message{
			{Role: "user", Content: "Why is the sky blue?"},
		},
		MaxTokens: PtrInt(100),
		Provider: &client.ProviderPreferences{
			// Try these providers first, in this order
			Order:          []string{"Anthropic", "OpenAI", "Together"},
			AllowFallbacks: PtrBool(true), // Allow falling back to others if these fail
		},
	}
	responseProvider, err := orClient.Chat.Create(ctx, requestProvider)
	handleError("Routing Chat (Provider Order)", err)
	if err == nil && len(responseProvider.Choices) > 0 {
		log.Printf(" Provider Order Response: Model Used: %s", responseProvider.Model)
		// To know WHICH provider was used, you'd need to call GetGeneration with the response ID
		log.Printf(" Provider Order Response ID: %s (Use GetGeneration to see which provider was used)", responseProvider.ID)
		if contentStr, ok := responseProvider.Choices[0].Message.Content.(string); ok {
			log.Printf(" Provider Order Response Content: %s", contentStr)
		}
	}
	log.Println("--- Provider Routing / Fallback Example Finished ---")
}

func exampleReasoningChat(ctx context.Context, orClient *client.Client) {
	log.Println("--- Running Reasoning Tokens Example (Anthropic) ---")

	request := client.ChatCompletionRequest{
		Model: "anthropic/claude-3.5-sonnet", // Use a model known to support reasoning
		Messages: []client.Message{
			{Role: "user", Content: "Solve this logic puzzle: A father and son have a car crash. The father dies. The son is rushed to the hospital, but the surgeon says, 'I can't operate on this boy, he is my son!' How is this possible?"},
		},
		MaxTokens: PtrInt(500), // Ensure enough tokens for reasoning + answer
		Reasoning: &client.ReasoningConfig{
			// Specify either Effort or MaxTokens, not both. Let's use Effort.
			Effort: PtrString("high"), // Request high reasoning effort
			// MaxTokens: PtrInt(500), // Alternative: specify token budget
			Exclude: PtrBool(false), // Ensure reasoning is included (default is false)
		},
		// Filter for providers supporting reasoning (good practice)
		Provider: &client.ProviderPreferences{
			// This check might not work perfectly as reasoning isn't a standard listed param
			// Better to rely on model choice and provider implementation for now
			// RequireParameters: PtrBool(true),
			Order: []string{"Anthropic"}, // Explicitly target Anthropic
		},
	}

	response, err := orClient.Chat.Create(ctx, request)
	handleError("Reasoning Chat", err)
	if err == nil && len(response.Choices) > 0 {
		log.Printf("Reasoning Chat Response Model: %s", response.Model)
		message := response.Choices[0].Message
		// Check if Reasoning field is present and not nil/empty
		if message.Reasoning != nil && *message.Reasoning != "" {
			log.Printf("Reasoning Steps:\n---\n%s\n---", *message.Reasoning)
		} else {
			log.Println("Reasoning Steps: <Not provided or empty>")
		}
		// Log the final content
		if contentStr, ok := message.Content.(string); ok {
			log.Printf("Final Content: %s", contentStr)
		} else {
			log.Printf("Final Content: (non-string type received)")
		}
	}
	log.Println("--- Reasoning Tokens Example Finished ---")
}

func exampleLogprobsChat(ctx context.Context, orClient *client.Client) {
	log.Println("--- Running Logprobs Example ---")
	request := client.ChatCompletionRequest{
		Model: "openai/gpt-4o-mini", // Many models support logprobs
		Messages: []client.Message{
			{Role: "user", Content: "Say 'Hello World'"},
		},
		MaxTokens:   PtrInt(10),
		Logprobs:    PtrBool(true), // Request log probabilities
		TopLogprobs: PtrInt(2),     // Request top 2 alternatives per token position
	}

	response, err := orClient.Chat.Create(ctx, request)
	handleError("Logprobs Chat", err)
	if err == nil && len(response.Choices) > 0 {
		log.Printf("Logprobs Response Model: %s", response.Model)
		choice := response.Choices[0]
		if contentStr, ok := choice.Message.Content.(string); ok {
			log.Printf("Logprobs Response Content: %s", contentStr)
		}
		// Check if logprobs data is present
		if choice.Logprobs != nil && len(choice.Logprobs.Content) > 0 {
			log.Println("Log Probability Details:")
			for i, lp := range choice.Logprobs.Content {
				log.Printf("  Token %d: '%s', Logprob: %.4f", i, lp.Token, lp.Logprob)
				// Check if top_logprobs were returned
				if len(lp.TopLogprobs) > 0 {
					log.Printf("    Top Alternatives:")
					for _, topLp := range lp.TopLogprobs {
						log.Printf("      - '%s': %.4f", topLp.Token, topLp.Logprob)
					}
				}
			}
		} else {
			log.Println("Logprobs data not found in response.")
		}
	}
	log.Println("--- Logprobs Example Finished ---")
}

func exampleAnthropicCacheChat(ctx context.Context, orClient *client.Client) {
	log.Println("--- Running Anthropic Prompt Caching Example Structure ---")
	log.Println("NOTE: This example demonstrates the *request structure* for caching.")
	log.Println("Actual cost savings must be verified via GetGeneration or Activity Page after making *repeated identical requests* within ~5 minutes.")

	// Example: Caching a large system prompt
	largeSystemPrompt := "You are an expert historian specializing in the detailed political landscape of the late Roman Republic. You have encyclopedic knowledge of key figures like Caesar, Pompey, Cicero, Crassus, Cato, and the various factions, laws, and events between 100 BCE and 44 BCE. Analyze any provided questions with meticulous detail, referencing specific events, dates, and political maneuvers. [Imagine several K tokens of detailed context here...]"
	shortUserQuery := "Briefly explain the significance of the First Triumvirate."

	request := client.ChatCompletionRequest{
		Model: "anthropic/claude-3.5-sonnet", // Must be an Anthropic model supporting caching
		Messages: []client.Message{
			{
				Role: "system",
				// Construct content as a slice for caching part
				Content: []client.ContentPart{
					client.TextContentPart{
						Type: "text",
						Text: "System Preamble: ", // Part before cache
					},
					client.CacheControlContentPart{ // The part to cache
						Type: "text",
						Text: largeSystemPrompt,
						CacheControl: &client.CacheControl{
							Type: "ephemeral", // Required for Anthropic
						},
					},
					// Optional: Text *after* the cached part within the same message
					// client.TextContentPart{
					// 	Type: "text",
					// 	Text: " End System Preamble.",
					// },
				},
			},
			{
				Role:    "user",
				Content: shortUserQuery,
			},
		},
		MaxTokens: PtrInt(500),
		// Ensure routing targets Anthropic if possible
		Provider: &client.ProviderPreferences{
			Order: []string{"Anthropic"},
			// RequireParameters: PtrBool(true), // Caching isn't a standard 'parameter' like response_format
		},
	}

	log.Println("Sending request 1 (cache write or read)...")
	response1, err1 := orClient.Chat.Create(ctx, request)
	handleError("Anthropic Cache Chat - Request 1", err1)
	genID1 := ""
	if err1 == nil && len(response1.Choices) > 0 {
		genID1 = response1.ID // Save ID to check costs later
		log.Printf(" Request 1 ID: %s (Check cost later)", genID1)
		if contentStr, ok := response1.Choices[0].Message.Content.(string); ok {
			log.Printf(" Request 1 Content: %s", contentStr)
		}
	}

	// Simulate sending the *exact same request* again shortly after
	log.Println("\nSending request 2 (potential cache read)...")
	// Create a new context in case the first one timed out, but reuse request object
	ctx2, cancel2 := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel2()
	response2, err2 := orClient.Chat.Create(ctx2, request) // Reuse the *exact same request*
	handleError("Anthropic Cache Chat - Request 2", err2)
	genID2 := ""
	if err2 == nil && len(response2.Choices) > 0 {
		genID2 = response2.ID
		log.Printf(" Request 2 ID: %s (Check cost later - compare with Req 1)", genID2)
		if contentStr, ok := response2.Choices[0].Message.Content.(string); ok {
			log.Printf(" Request 2 Content: %s", contentStr)
		}
	}

	log.Println("\nTo verify caching: Use GetGeneration for IDs or check Activity Page.")
	log.Printf(" Generation ID 1: %s", genID1)
	log.Printf(" Generation ID 2: %s", genID2)
	log.Println(" Expect Req 2 cost to be lower if cache hit occurred.")

	log.Println("--- Anthropic Prompt Caching Example Structure Finished ---")
}

// --- Helpers ---

// handleError is a simple error logging helper.
func handleError(operation string, err error) {
	if err != nil {
		var apiErr *client.APIError
		if errors.As(err, &apiErr) {
			// It's an APIError from the client
			log.Printf("Error during '%s': Status %d - %s\n", operation, apiErr.StatusCode, apiErr.Error())
			if apiErr.Details != nil {
				log.Printf("  API Error Details: Code=%d, Message=%s\n", apiErr.Details.Code, apiErr.Details.Message)
				if apiErr.Details.Metadata != nil {
					log.Printf("  API Error Metadata: %+v\n", apiErr.Details.Metadata)
					// Try parsing specific metadata types
					if modMeta, ok := client.TryGetModerationMetadata(apiErr); ok {
						log.Printf("  Parsed Moderation Metadata: Reasons=%v, Input=%s\n", modMeta.Reasons, modMeta.FlaggedInput)
					}
					if provMeta, ok := client.TryGetProviderErrorMetadata(apiErr); ok {
						log.Printf("  Parsed Provider Metadata: Provider=%s, Raw=%v\n", provMeta.ProviderName, provMeta.Raw)
					}
				}
			}
			// Optional: Log raw body for debugging if needed
			// if apiErr.RawBody != nil {
			// 	log.Printf(" Raw Error Body: %s\n", string(apiErr.RawBody))
			// }
		} else {
			// It's some other error (network, context cancellation, etc.)
			log.Printf("Error during '%s': %v\n", operation, err)
		}
	}
}

// Helpers to get pointers for optional fields in request structs
func PtrInt(i int) *int             { return &i }
func PtrFloat64(f float64) *float64 { return &f }
func PtrBool(b bool) *bool          { return &b }
func PtrString(s string) *string    { return &s }

// Helper to safely dereference optional string pointers for logging
func derefString(s *string) string {
	if s != nil {
		return *s
	}
	return "<nil>"
}
