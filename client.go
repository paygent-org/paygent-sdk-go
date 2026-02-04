package paygent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkoukk/tiktoken-go"
	"github.com/sirupsen/logrus"
)

// Client represents the Paygent SDK client
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Logger
}

// UsageData represents the usage data structure
type UsageData struct {
	ServiceProvider  string `json:"service_provider"`
	Model            string `json:"model"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	TotalTokens      int    `json:"total_tokens"`
	Plan             string `json:"plan"`
}

// UsageDataWithStrings represents the usage data structure with prompt and output strings
type UsageDataWithStrings struct {
	ServiceProvider string `json:"service_provider"`
	Model           string `json:"model"`
	PromptString    string `json:"prompt_string"`
	OutputString    string `json:"output_string"`
	Plan            string `json:"plan"`
}

// RawUsageData represents raw usage data for V2 API
type RawUsageData struct {
	Provider       string `json:"provider"`
	Model          string `json:"model"`
	InputTokens    int    `json:"inputTokens"`
	OutputTokens   int    `json:"outputTokens"`
	CachedTokens   int    `json:"cachedTokens"`
	AudioDuration  int    `json:"audioDuration"`
	CharacterCount int    `json:"characterCount"`
	Plan           string `json:"plan"`
}

// CostBreakdown represents the cost breakdown in V2 response
type CostBreakdown struct {
	PromptCost     float64 `json:"promptCost"`
	CompletionCost float64 `json:"completionCost"`
	TotalCost      float64 `json:"totalCost"`
}

// SendUsageV2Response represents the response from V2 usage API
type SendUsageV2Response struct {
	CPDataID       string        `json:"cpDataId"`
	CalculatedCost float64       `json:"calculatedCost"`
	Breakdown      CostBreakdown `json:"breakdown"`
}

// Customer represents a customer in Paygent
type Customer struct {
	ID         string `json:"id"`
	ExternalID string `json:"externalId"`
	Name       string `json:"name"`
	Email      string `json:"email,omitempty"`
}

// CustomerCreateOrGetRequest represents the request to create or get a customer
type CustomerCreateOrGetRequest struct {
	ExternalID string `json:"externalId"`
	Name       string `json:"name"`
	Email      string `json:"email,omitempty"`
}


const (
	defaultBaseURL = "https://cp-api.withpaygent.com"
	// defaultBaseURL = "http://localhost:8082"
	defaultTimeout = 30 * time.Second
)

// NewClient creates a new Paygent SDK client with default production URL
func NewClient(apiKey string) *Client {
	logger := logrus.New()
	// Set to ERROR level for minimal logging - only log errors
	logger.SetLevel(logrus.ErrorLevel)

	return &Client{
		apiKey:     apiKey,
		baseURL:    defaultBaseURL,    // Locked configuration
		httpClient: &http.Client{Timeout: defaultTimeout}, // Locked timeout
		logger:     logger,
	}
}
// getTokenCount estimates tokens for a given model and text
// Supports OpenAI, Anthropic, Google, Meta, AWS, Mistral, Cohere, DeepSeek
func (c *Client) getTokenCount(model, text string) int {
	if len(text) == 0 {
		return 0
	}

	modelLower := strings.ToLower(model)

	// OpenAI GPT models
	if strings.HasPrefix(modelLower, "gpt-") {
		encoding, err := tiktoken.EncodingForModel(model)
		if err != nil {
			c.logger.Warnf("Failed to get encoding for model %s, using cl100k_base: %v", model, err)
			encoding, err = tiktoken.GetEncoding("cl100k_base")
			if err != nil {
				c.logger.Errorf("Failed to get cl100k_base encoding: %v", err)
				return c.fallbackTokenCount(text)
			}
		}
		return len(encoding.Encode(text, nil, nil))
	}

	// Anthropic Claude models
	if strings.HasPrefix(modelLower, "claude-") {
		encoding, err := tiktoken.GetEncoding("cl100k_base")
		if err != nil {
			c.logger.Errorf("Failed to get cl100k_base encoding for Claude: %v", err)
			return c.fallbackTokenCount(text)
		}
		return len(encoding.Encode(text, nil, nil))
	}

	// Google DeepMind Gemini models
	if strings.HasPrefix(modelLower, "gemini-") {
		encoding, err := tiktoken.GetEncoding("cl100k_base")
		if err != nil {
			c.logger.Errorf("Failed to get cl100k_base encoding for Gemini: %v", err)
			return c.fallbackTokenCount(text)
		}
		return len(encoding.Encode(text, nil, nil))
	}

	// Meta Llama models
	if strings.HasPrefix(modelLower, "llama") {
		// Use cl100k_base as approximation for Llama models
		encoding, err := tiktoken.GetEncoding("cl100k_base")
		if err != nil {
			c.logger.Errorf("Failed to get cl100k_base encoding for Llama: %v", err)
			return c.fallbackTokenCount(text)
		}
		return len(encoding.Encode(text, nil, nil))
	}

	// Mistral models
	if strings.HasPrefix(modelLower, "mistral") {
		// Use cl100k_base as approximation for Mistral models
		encoding, err := tiktoken.GetEncoding("cl100k_base")
		if err != nil {
			c.logger.Errorf("Failed to get cl100k_base encoding for Mistral: %v", err)
			return c.fallbackTokenCount(text)
		}
		return len(encoding.Encode(text, nil, nil))
	}

	// Cohere models
	if strings.HasPrefix(modelLower, "command") {
		// Use cl100k_base as approximation for Cohere models
		encoding, err := tiktoken.GetEncoding("cl100k_base")
		if err != nil {
			c.logger.Errorf("Failed to get cl100k_base encoding for Cohere: %v", err)
			return c.fallbackTokenCount(text)
		}
		return len(encoding.Encode(text, nil, nil))
	}

	// DeepSeek models
	if strings.HasPrefix(modelLower, "deepseek") {
		// Use cl100k_base as approximation for DeepSeek models
		encoding, err := tiktoken.GetEncoding("cl100k_base")
		if err != nil {
			c.logger.Errorf("Failed to get cl100k_base encoding for DeepSeek: %v", err)
			return c.fallbackTokenCount(text)
		}
		return len(encoding.Encode(text, nil, nil))
	}

	// AWS Titan models
	if strings.HasPrefix(modelLower, "titan-") {
		// Use cl100k_base as approximation for AWS Titan models
		encoding, err := tiktoken.GetEncoding("cl100k_base")
		if err != nil {
			c.logger.Errorf("Failed to get cl100k_base encoding for Titan: %v", err)
			return c.fallbackTokenCount(text)
		}
		return len(encoding.Encode(text, nil, nil))
	}

	// Fallback for unknown models
	c.logger.Warnf("Unknown model '%s', using fallback token counting", model)
	return c.fallbackTokenCount(text)
}

// fallbackTokenCount provides a rough estimate when proper tokenization fails
func (c *Client) fallbackTokenCount(text string) int {
	// Rough estimate: ~4 characters per token for English text
	// This is a conservative estimate
	words := len(strings.Fields(text))
	if words == 0 {
		return 1 // At least 1 token for non-empty text
	}

	// Rough approximation: 1.3 tokens per word on average
	return int(float64(words) * 1.3)
}

// SendUsageV2 sends usage data to the Paygent V2 API
func (c *Client) SendUsageV2(agentID, customerID, indicator string, usageData RawUsageData) (*SendUsageV2Response, error) {
	requestData := map[string]interface{}{
		"agentId":    agentID,
		"customerId": customerID,
		"indicator":  indicator,
		"rawUsage":   usageData,
	}
	
	// Ensure plan has a default
	if usageData, ok := requestData["rawUsage"].(RawUsageData); ok {
		if usageData.Plan == "" {
			usageData.Plan = "default"
		}
		requestData["rawUsage"] = usageData
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		c.logger.Errorf("Failed to marshal request body: %v", err)
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	url := fmt.Sprintf("%s/api/v2/usage", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("paygent-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("api request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	var v2Response SendUsageV2Response
	if err := json.Unmarshal(responseBody, &v2Response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &v2Response, nil
}

// CreateOrGetCustomer creates or gets a customer in Paygent
func (c *Client) CreateOrGetCustomer(customerData CustomerCreateOrGetRequest) (*Customer, error) {
	requestBody, err := json.Marshal(customerData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/customers/create-or-get", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("paygent-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("api request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	var customer Customer
	if err := json.Unmarshal(responseBody, &customer); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &customer, nil
}

// SendUsage sends usage data to the Paygent API (Legacy)
func (c *Client) SendUsage(agentID, customerID, indicator string, usageData UsageData) error {
	rawUsage := RawUsageData{
		Provider:     usageData.ServiceProvider,
		Model:        usageData.Model,
		InputTokens:  usageData.PromptTokens,
		OutputTokens: usageData.CompletionTokens,
		Plan:         usageData.Plan,
	}

	_, err := c.SendUsageV2(agentID, customerID, indicator, rawUsage)
	return err
}

// SendUsageWithTokenString sends usage data to the Paygent API using prompt and output strings (Legacy)
func (c *Client) SendUsageWithTokenString(agentID, customerID, indicator string, usageData UsageDataWithStrings) error {
	promptTokens := c.getTokenCount(usageData.Model, usageData.PromptString)
	completionTokens := c.getTokenCount(usageData.Model, usageData.OutputString)

	rawUsage := RawUsageData{
		Provider:     usageData.ServiceProvider,
		Model:        usageData.Model,
		InputTokens:  promptTokens,
		OutputTokens: completionTokens,
		Plan:         usageData.Plan,
	}

	_, err := c.SendUsageV2(agentID, customerID, indicator, rawUsage)
	return err
}

// SetLogLevel sets the logging level for the client
func (c *Client) SetLogLevel(level logrus.Level) {
	c.logger.SetLevel(level)
}

// GetLogger returns the logger instance for custom logging
func (c *Client) GetLogger() *logrus.Logger {
	return c.logger
}
