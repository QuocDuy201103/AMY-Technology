package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// DeepseekClient handles communication with the Deepseek API
type DeepseekClient struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	Model      string
}

// NewDeepseekClient creates a new DeepseekClient instance
func NewDeepseekClient(baseURL, apiKey string) *DeepseekClient {
	model := os.Getenv("DEEPSEEK_MODEL")
	if strings.TrimSpace(model) == "" {
		model = "deepseek-chat"
	}
	// Trim API key to remove any whitespace/newlines that might cause header issues
	apiKey = strings.TrimSpace(apiKey)
	return &DeepseekClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Model: model,
	}
}

// SummaryResponse represents the response from the summarize endpoint
type SummaryResponse struct {
	Summary string `json:"summary"`
}

// ClassificationLabel represents a classification label
type ClassificationLabel struct {
	Label string  `json:"label"`
	Score float64 `json:"score"`
}

// ClassifyResponse represents the response from the classify endpoint
type ClassifyResponse struct {
	Labels []ClassificationLabel `json:"labels"`
}

// EmailRequest represents a single email in the batch request
type EmailRequest struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

// BatchClassificationResult represents the classification result for a single email in batch
type BatchClassificationResult struct {
	ID     string                 `json:"id"`
	Labels []ClassificationLabel `json:"labels"`
}

// DraftResponse represents the response from the draft endpoint
type DraftResponse struct {
	Draft string `json:"draft"`
}

// APIError represents an error response from the API
type APIError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.Code, e.Message)
}

// makeRequest performs an HTTP request with retries
func (c *DeepseekClient) makeRequest(method, endpoint string, body io.Reader, maxRetries int) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.BaseURL, endpoint)
	log.Printf("Making request to: %s %s", method, url)

	// Read body content once so we can reuse it on retries
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			time.Sleep(backoff)
		}

		// Create a new reader for each retry attempt
		var bodyReader io.Reader
		if bodyBytes != nil {
			bodyReader = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequest(method, url, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Default to JSON; callers can override with their body if needed
		req.Header.Set("Content-Type", "application/json")
		// Trim API key again before setting header to ensure no invalid characters
		apiKey := strings.TrimSpace(c.APIKey)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request to %s failed: %w", url, err)
			continue
		}

		// Retry on 5xx errors
		if resp.StatusCode >= 500 && resp.StatusCode < 600 && attempt < maxRetries {
			resp.Body.Close()
			lastErr = fmt.Errorf("server error %d from %s", resp.StatusCode, url)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("request to %s failed after %d retries: %w", url, maxRetries, lastErr)
}

// DeepSeek chat request/response (OpenAI compatible shape)
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream,omitempty"`
}

type chatChoice struct {
	Index        int         `json:"index"`
	FinishReason string      `json:"finish_reason"`
	Message      chatMessage `json:"message"`
}

type chatResponse struct {
	Choices []chatChoice `json:"choices"`
}

// SummarizeEmail sends email content to the summarize endpoint
func (c *DeepseekClient) SummarizeEmail(content string) (*SummaryResponse, error) {
	// Build prompt
	reqBody := chatRequest{
		Model: c.Model,
		Messages: []chatMessage{
			{Role: "system", Content: "You are an assistant that summarizes emails. Return a concise summary in plain text."},
			{Role: "user", Content: fmt.Sprintf("Summarize this email (HTML allowed):\n\n%s", content)},
		},
	}
	raw, _ := json.Marshal(reqBody)
	resp, err := c.makeRequest("POST", "/v1/chat/completions", bytes.NewReader(raw), 3)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read response body for error details
		bodyBytes, readErr := io.ReadAll(resp.Body)
		errorMsg := fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
		if readErr == nil && len(bodyBytes) > 0 {
			errorMsg = fmt.Sprintf("unexpected status code: %d, response: %s", resp.StatusCode, string(bodyBytes))
		}

		// Try to decode as APIError
		var apiErr APIError
		if json.Unmarshal(bodyBytes, &apiErr) == nil {
			return nil, &apiErr
		}

		return nil, fmt.Errorf(errorMsg)
	}

	var cr chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return nil, fmt.Errorf("failed to decode chat response: %w", err)
	}
	if len(cr.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from model")
	}
	return &SummaryResponse{Summary: strings.TrimSpace(cr.Choices[0].Message.Content)}, nil
}

// ClassifyEmail sends email content to the classify endpoint
func (c *DeepseekClient) ClassifyEmail(content string) (*ClassifyResponse, error) {
	// Instruct model to output strict JSON
	reqBody := chatRequest{
		Model: c.Model,
		Messages: []chatMessage{
			{Role: "system", Content: "Classify the email into labels. Output strict JSON: {\"labels\":[{\"label\":string,\"score\":number}]} with no extra text."},
			{Role: "user", Content: fmt.Sprintf("Classify this email (HTML allowed):\n\n%s", content)},
		},
	}
	raw, _ := json.Marshal(reqBody)
	resp, err := c.makeRequest("POST", "/v1/chat/completions", bytes.NewReader(raw), 3)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read response body for error details
		bodyBytes, readErr := io.ReadAll(resp.Body)
		errorMsg := fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
		if readErr == nil && len(bodyBytes) > 0 {
			errorMsg = fmt.Sprintf("unexpected status code: %d, response: %s", resp.StatusCode, string(bodyBytes))
		}

		// Try to decode as APIError
		var apiErr APIError
		if json.Unmarshal(bodyBytes, &apiErr) == nil {
			return nil, &apiErr
		}

		return nil, fmt.Errorf(errorMsg)
	}

	var cr chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return nil, fmt.Errorf("failed to decode chat response: %w", err)
	}
	if len(cr.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from model")
	}
	var out ClassifyResponse
	// Try to parse strict JSON from model content
	responseContent := strings.TrimSpace(cr.Choices[0].Message.Content)
	
	// Log raw content for debugging
	log.Printf("DeepSeek API response content: %s", responseContent)
	
	// Try to extract JSON if wrapped in markdown code blocks
	if strings.HasPrefix(responseContent, "```json") {
		responseContent = strings.TrimPrefix(responseContent, "```json")
		responseContent = strings.TrimSuffix(responseContent, "```")
		responseContent = strings.TrimSpace(responseContent)
	} else if strings.HasPrefix(responseContent, "```") {
		responseContent = strings.TrimPrefix(responseContent, "```")
		responseContent = strings.TrimSuffix(responseContent, "```")
		responseContent = strings.TrimSpace(responseContent)
	}
	
	if err := json.Unmarshal([]byte(responseContent), &out); err != nil {
		log.Printf("Failed to parse JSON from model response: %v, content: %s", err, responseContent)
		return nil, fmt.Errorf("model did not return valid JSON for classification: %w, content: %s", err, responseContent)
	}
	
	// Validate that labels are not empty
	if len(out.Labels) == 0 {
		log.Printf("Warning: Model returned empty labels, content: %s", responseContent)
	}
	
	return &out, nil
}

// DraftReply sends email content to the draft endpoint
func (c *DeepseekClient) DraftReply(content string) (*DraftResponse, error) {
	reqBody := chatRequest{
		Model: c.Model,
		Messages: []chatMessage{
			{Role: "system", Content: "Write a polite, concise reply to the user's email. Output only the reply text."},
			{Role: "user", Content: fmt.Sprintf("Write a reply to this email (HTML allowed):\n\n%s", content)},
		},
	}
	raw, _ := json.Marshal(reqBody)
	resp, err := c.makeRequest("POST", "/v1/chat/completions", bytes.NewReader(raw), 3)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read response body for error details
		bodyBytes, readErr := io.ReadAll(resp.Body)
		errorMsg := fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
		if readErr == nil && len(bodyBytes) > 0 {
			errorMsg = fmt.Sprintf("unexpected status code: %d, response: %s", resp.StatusCode, string(bodyBytes))
		}

		// Try to decode as APIError
		var apiErr APIError
		if json.Unmarshal(bodyBytes, &apiErr) == nil {
			return nil, &apiErr
		}

		return nil, fmt.Errorf(errorMsg)
	}

	var cr chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return nil, fmt.Errorf("failed to decode chat response: %w", err)
	}
	if len(cr.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from model")
	}
	return &DraftResponse{Draft: strings.TrimSpace(cr.Choices[0].Message.Content)}, nil
}

// ClassifyEmailsBatch processes multiple emails for classification
func (c *DeepseekClient) ClassifyEmailsBatch(emails []EmailRequest) ([]BatchClassificationResult, error) {
	results := make([]BatchClassificationResult, len(emails))
	
	// Process emails sequentially (can be parallelized if needed)
	for i, email := range emails {
		classification, err := c.ClassifyEmail(email.Content)
		if err != nil {
			// Log error but continue processing other emails
			log.Printf("Error classifying email %s: %v", email.ID, err)
			// Return error result for this email
			results[i] = BatchClassificationResult{
				ID:     email.ID,
				Labels: []ClassificationLabel{},
			}
			continue
		}
		
		results[i] = BatchClassificationResult{
			ID:     email.ID,
			Labels: classification.Labels,
		}
	}
	
	return results, nil
}
