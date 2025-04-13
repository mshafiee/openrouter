package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// ChatService handles communication with the chat completion related methods of the OpenRouter API.
type ChatService service // Use type alias as established in client.go

// Create sends a request to generate a chat completion based on the provided messages and parameters.
// This is for non-streaming requests. For streaming, use CreateStream.
func (s *ChatService) Create(ctx context.Context, request ChatCompletionRequest) (response ChatCompletionResponse, err error) {
	// Ensure stream is false for this method, although the API might handle it.
	// It's clearer to the user if they use the correct method.
	if request.Stream {
		err = fmt.Errorf("use CreateStream for streaming requests")
		return
	}

	// Ensure either Messages or Prompt is provided (API requirement)
	// Note: The API likely validates this, but a client-side check can be helpful.
	if len(request.Messages) == 0 && request.Prompt == "" {
		err = fmt.Errorf("either Messages or Prompt must be provided in the request")
		return
	}

	path := "chat/completions"
	err = s.client.doRequest(ctx, http.MethodPost, path, request, &response, false) // false = use standard API key
	return
}

// --- Streaming ---

// ChatCompletionStream represents a stream of chat completion chunks.
// It must be closed once completed to release underlying resources.
type ChatCompletionStream struct {
	response   *http.Response
	scanner    *bufio.Scanner
	isFinished bool
	err        error // Stores error encountered during stream processing
	// Store the usage from the last chunk if available
	Usage *ResponseUsage
}

// CreateStream sends a request to generate a chat completion and returns a stream
// for receiving response chunks (Server-Sent Events).
// The caller *must* set `request.Stream = true`.
func (s *ChatService) CreateStream(ctx context.Context, request ChatCompletionRequest) (*ChatCompletionStream, error) {
	if !request.Stream {
		return nil, fmt.Errorf("request.Stream must be true to use CreateStream")
	}
	// Ensure either Messages or Prompt is provided (API requirement)
	if len(request.Messages) == 0 && request.Prompt == "" {
		return nil, fmt.Errorf("either Messages or Prompt must be provided in the request")
	}

	path := "chat/completions"
	resp, err := s.client.doStreamingRequest(ctx, http.MethodPost, path, request)
	if err != nil {
		// doStreamingRequest already wrapped potential API errors
		return nil, err
	}

	return &ChatCompletionStream{
		response: resp,
		scanner:  bufio.NewScanner(resp.Body),
	}, nil
}

// Recv reads the next chunk from the stream.
// It returns io.EOF when the stream is finished.
// Any errors encountered during streaming will be returned here and stored in the stream's Err field.
func (s *ChatCompletionStream) Recv() (response ChatCompletionChunk, err error) {
	if s.isFinished || s.err != nil {
		return ChatCompletionChunk{}, io.EOF
	}

	for s.scanner.Scan() {
		line := s.scanner.Bytes()

		// SSE lines starting with ':' are comments, skip them
		if bytes.HasPrefix(line, []byte(":")) {
			continue
		}
		// Skip empty lines which act as separators between events in SSE
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}

		// Lines should start with "data: "
		if !bytes.HasPrefix(line, []byte("data: ")) {
			// This indicates a malformed SSE event from the server
			s.err = fmt.Errorf("unexpected line format in SSE stream: %s", string(line))
			s.Close() // Attempt cleanup
			return ChatCompletionChunk{}, s.err
		}

		// Extract the JSON data part
		data := bytes.TrimPrefix(line, []byte("data: "))

		// Check for the [DONE] message
		if bytes.Equal(data, []byte("[DONE]")) {
			s.isFinished = true
			// EOF signals the end of the stream to the caller
			return ChatCompletionChunk{}, io.EOF
		}

		// Unmarshal the data into a chunk
		err = json.Unmarshal(data, &response)
		if err != nil {
			s.err = fmt.Errorf("failed to unmarshal stream chunk: %w; data: %s", err, string(data))
			s.Close() // Attempt cleanup
			return ChatCompletionChunk{}, s.err
		}

		// Store usage if present in this chunk
		if response.Usage != nil {
			s.Usage = response.Usage
		}

		// Successfully processed a data chunk, return it
		return response, nil
	}

	// Check for scanner errors after the loop finishes (e.g., connection closed prematurely)
	if scanErr := s.scanner.Err(); scanErr != nil {
		if errors.Is(scanErr, context.Canceled) || errors.Is(scanErr, context.DeadlineExceeded) {
			s.err = scanErr // Propagate context errors correctly
		} else {
			s.err = fmt.Errorf("stream scanner error: %w", scanErr)
		}
		s.Close() // Ensure cleanup on error
		return ChatCompletionChunk{}, s.err
	}

	// If the loop finishes without data and without error, it means the stream ended cleanly
	// without a specific [DONE] message (less common but possible). Treat as EOF.
	s.isFinished = true
	return ChatCompletionChunk{}, io.EOF
}

// Close closes the underlying HTTP response body.
// It should be called by the user when finished with the stream, typically using defer.
func (s *ChatCompletionStream) Close() error {
	if s.response != nil && s.response.Body != nil {
		return s.response.Body.Close()
	}
	return nil
}

// Err returns any error encountered and stored during stream processing by Recv.
// Returns nil if the stream finished cleanly or is still ongoing without errors.
func (s *ChatCompletionStream) Err() error {
	return s.err
}
