package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "strings"
    "time"
)

type OpenAIClient struct {
    BaseURL    string
    APIKey     string
    HTTPClient *http.Client
    Model      string
}

func NewOpenAIClient(baseURL, apiKey string) *OpenAIClient {
    model := os.Getenv("OPENAI_MODEL")
    if strings.TrimSpace(model) == "" {
        model = "gpt-4o-mini"
    }
    return &OpenAIClient{
        BaseURL: baseURL,
        APIKey:  apiKey,
        HTTPClient: &http.Client{
            Timeout: 30 * time.Second,
        },
        Model: model,
    }
}

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

type SummaryResponse struct {
    Summary string `json:"summary"`
}

type ClassificationLabel struct {
    Label string  `json:"label"`
    Score float64 `json:"score"`
}

type ClassifyResponse struct {
    Labels []ClassificationLabel `json:"labels"`
}

type DraftResponse struct {
    Draft string `json:"draft"`
}

type APIError struct {
    Message string `json:"message"`
    Code    int    `json:"code"`
}

func (e *APIError) Error() string {
    return fmt.Sprintf("API error %d: %s", e.Code, e.Message)
}

func (c *OpenAIClient) makeRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
    url := fmt.Sprintf("%s%s", c.BaseURL, endpoint)
    req, err := http.NewRequest(method, url, body)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
    return c.HTTPClient.Do(req)
}

func (c *OpenAIClient) SummarizeEmail(content string) (*SummaryResponse, error) {
    reqBody := chatRequest{
        Model: c.Model,
        Messages: []chatMessage{
            {Role: "system", Content: "You are an assistant that summarizes emails. Return a concise summary in plain text."},
            {Role: "user", Content: fmt.Sprintf("Summarize this email (HTML allowed):\n\n%s", content)},
        },
    }
    raw, _ := json.Marshal(reqBody)
    resp, err := c.makeRequest("POST", "/v1/chat/completions", bytes.NewReader(raw))
    if err != nil {
        return nil, fmt.Errorf("failed to make request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        bodyBytes, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(bodyBytes))
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

func (c *OpenAIClient) ClassifyEmail(content string) (*ClassifyResponse, error) {
    reqBody := chatRequest{
        Model: c.Model,
        Messages: []chatMessage{
            {Role: "system", Content: "Classify the email into labels. Output strict JSON: {\"labels\":[{\"label\":string,\"score\":number}]} with no extra text."},
            {Role: "user", Content: fmt.Sprintf("Classify this email (HTML allowed):\n\n%s", content)},
        },
    }
    raw, _ := json.Marshal(reqBody)
    resp, err := c.makeRequest("POST", "/v1/chat/completions", bytes.NewReader(raw))
    if err != nil {
        return nil, fmt.Errorf("failed to make request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        bodyBytes, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(bodyBytes))
    }

    var cr chatResponse
    if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
        return nil, fmt.Errorf("failed to decode chat response: %w", err)
    }
    if len(cr.Choices) == 0 {
        return nil, fmt.Errorf("no choices returned from model")
    }
    var out ClassifyResponse
    if err := json.Unmarshal([]byte(cr.Choices[0].Message.Content), &out); err != nil {
        return nil, fmt.Errorf("model did not return valid JSON for classification: %w", err)
    }
    return &out, nil
}

func (c *OpenAIClient) DraftReply(content string) (*DraftResponse, error) {
    reqBody := chatRequest{
        Model: c.Model,
        Messages: []chatMessage{
            {Role: "system", Content: "Write a polite, concise reply to the user's email. Output only the reply text."},
            {Role: "user", Content: fmt.Sprintf("Write a reply to this email (HTML allowed):\n\n%s", content)},
        },
    }
    raw, _ := json.Marshal(reqBody)
    resp, err := c.makeRequest("POST", "/v1/chat/completions", bytes.NewReader(raw))
    if err != nil {
        return nil, fmt.Errorf("failed to make request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        bodyBytes, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(bodyBytes))
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


