package client

// This file is reserved for potential future expansion of streaming helpers
// or handling different types of streaming endpoints beyond Chat Completions.

// Currently, the primary streaming implementation for Chat Completions
// resides within chat.go, specifically:
//
// 1. ChatService.CreateStream(...): Initiates the streaming request and returns
//    a *ChatCompletionStream.
// 2. ChatCompletionStream struct: Manages the underlying HTTP response and scanner.
// 3. ChatCompletionStream.Recv(): Reads, parses Server-Sent Events (SSE),
//    unmarshals ChatCompletionChunk data, and handles the "[DONE]" message.
// 4. ChatCompletionStream.Close(): Closes the response body.
// 5. ChatCompletionStream.Err(): Reports any errors encountered during streaming.

// If OpenRouter introduces other streaming endpoints (e.g., for different services
// or with different data formats) or if more complex SSE features need handling,
// generic parsing functions or stream management types could be added here.
