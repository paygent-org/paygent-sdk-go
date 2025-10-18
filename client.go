package paylm

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

// Client represents the Paylm SDK client
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
}

// UsageDataWithStrings represents the usage data structure with prompt and output strings
type UsageDataWithStrings struct {
	ServiceProvider string `json:"service_provider"`
	Model           string `json:"model"`
	PromptString    string `json:"prompt_string"`
	OutputString    string `json:"output_string"`
}

// APIRequest represents the request body for the API call
type APIRequest struct {
	AgentID         string  `json:"agentId"`
	CustomerID      string  `json:"customerId"`
	Indicator       string  `json:"indicator"`
	Amount          float64 `json:"amount"`
	InputToken      int     `json:"inputToken"`
	OutputToken     int     `json:"outputToken"`
	Model           string  `json:"model"`
	ServiceProvider string  `json:"serviceProvider"`
}

// ModelPricing represents pricing information for different models
type ModelPricing struct {
	PromptTokensCost     float64
	CompletionTokensCost float64
}

// Default model pricing (cost per 1000 tokens in USD)
var modelPricing = map[string]ModelPricing{
	// OpenAI Models (pricing per 1000 tokens)
	"gpt-3.5-turbo": {
		PromptTokensCost:     1.5, // $1.50 per 1000 tokens
		CompletionTokensCost: 2.0, // $2.00 per 1000 tokens
	},
	"gpt-3.5-turbo-16k": {
		PromptTokensCost:     3.0, // $3.00 per 1000 tokens
		CompletionTokensCost: 4.0, // $4.00 per 1000 tokens
	},
	"gpt-4": {
		PromptTokensCost:     30.0, // $30.00 per 1000 tokens
		CompletionTokensCost: 60.0, // $60.00 per 1000 tokens
	},
	"gpt-4-turbo": {
		PromptTokensCost:     10.0, // $10.00 per 1000 tokens
		CompletionTokensCost: 30.0, // $30.00 per 1000 tokens
	},
	"gpt-4o": {
		PromptTokensCost:     5.0,  // $5.00 per 1000 tokens
		CompletionTokensCost: 15.0, // $15.00 per 1000 tokens
	},
	"gpt-4o-mini": {
		PromptTokensCost:     0.15, // $0.15 per 1000 tokens
		CompletionTokensCost: 0.6,  // $0.60 per 1000 tokens
	},
	"gpt-5": {
		PromptTokensCost:     10.0, // $10.00 per 1000 tokens (estimated)
		CompletionTokensCost: 30.0, // $30.00 per 1000 tokens (estimated)
	},

	// Anthropic Models (pricing per 1000 tokens)
	"claude-3-haiku": {
		PromptTokensCost:     0.25, // $0.25 per 1000 tokens
		CompletionTokensCost: 1.25, // $1.25 per 1000 tokens
	},
	"claude-3-sonnet": {
		PromptTokensCost:     3.0,  // $3.00 per 1000 tokens
		CompletionTokensCost: 15.0, // $15.00 per 1000 tokens
	},
	"claude-3-opus": {
		PromptTokensCost:     15.0, // $15.00 per 1000 tokens
		CompletionTokensCost: 75.0, // $75.00 per 1000 tokens
	},
	"claude-3.5-sonnet": {
		PromptTokensCost:     3.0,  // $3.00 per 1000 tokens
		CompletionTokensCost: 15.0, // $15.00 per 1000 tokens
	},

	// Google DeepMind Models (pricing per 1000 tokens)
	"gemini-pro": {
		PromptTokensCost:     0.5, // $0.50 per 1000 tokens
		CompletionTokensCost: 1.5, // $1.50 per 1000 tokens
	},
	"gemini-pro-vision": {
		PromptTokensCost:     0.5, // $0.50 per 1000 tokens
		CompletionTokensCost: 1.5, // $1.50 per 1000 tokens
	},
	"gemini-1.5-pro": {
		PromptTokensCost:     1.25, // $1.25 per 1000 tokens
		CompletionTokensCost: 5.0,  // $5.00 per 1000 tokens
	},
	"gemini-1.5-flash": {
		PromptTokensCost:     0.075, // $0.075 per 1000 tokens
		CompletionTokensCost: 0.3,   // $0.30 per 1000 tokens
	},

	// Meta Models (pricing per 1000 tokens)
	"llama-2-7b": {
		PromptTokensCost:     0.1, // $0.10 per 1000 tokens
		CompletionTokensCost: 0.1, // $0.10 per 1000 tokens
	},
	"llama-2-13b": {
		PromptTokensCost:     0.2, // $0.20 per 1000 tokens
		CompletionTokensCost: 0.2, // $0.20 per 1000 tokens
	},
	"llama-2-70b": {
		PromptTokensCost:     0.7, // $0.70 per 1000 tokens
		CompletionTokensCost: 0.7, // $0.70 per 1000 tokens
	},
	"llama-3-8b": {
		PromptTokensCost:     0.1, // $0.10 per 1000 tokens
		CompletionTokensCost: 0.1, // $0.10 per 1000 tokens
	},
	"llama-3-70b": {
		PromptTokensCost:     0.7, // $0.70 per 1000 tokens
		CompletionTokensCost: 0.7, // $0.70 per 1000 tokens
	},

	// AWS Models (pricing per 1000 tokens)
	"claude-3-haiku-aws": {
		PromptTokensCost:     0.25, // $0.25 per 1000 tokens
		CompletionTokensCost: 1.25, // $1.25 per 1000 tokens
	},
	"claude-3-sonnet-aws": {
		PromptTokensCost:     3.0,  // $3.00 per 1000 tokens
		CompletionTokensCost: 15.0, // $15.00 per 1000 tokens
	},
	"titan-text-express": {
		PromptTokensCost:     0.8, // $0.80 per 1000 tokens
		CompletionTokensCost: 1.6, // $1.60 per 1000 tokens
	},
	"titan-text-lite": {
		PromptTokensCost:     0.3, // $0.30 per 1000 tokens
		CompletionTokensCost: 0.4, // $0.40 per 1000 tokens
	},

	// Mistral AI Models (pricing per 1000 tokens)
	"mistral-7b": {
		PromptTokensCost:     0.1, // $0.10 per 1000 tokens
		CompletionTokensCost: 0.1, // $0.10 per 1000 tokens
	},
	"mistral-8x7b": {
		PromptTokensCost:     0.2, // $0.20 per 1000 tokens
		CompletionTokensCost: 0.2, // $0.20 per 1000 tokens
	},
	"mistral-nemo": {
		PromptTokensCost:     0.1, // $0.10 per 1000 tokens
		CompletionTokensCost: 0.1, // $0.10 per 1000 tokens
	},
	"mistral-large": {
		PromptTokensCost:     2.0, // $2.00 per 1000 tokens
		CompletionTokensCost: 6.0, // $6.00 per 1000 tokens
	},

	// Cohere Models (pricing per 1000 tokens)
	"command": {
		PromptTokensCost:     1.5, // $1.50 per 1000 tokens
		CompletionTokensCost: 2.0, // $2.00 per 1000 tokens
	},
	"command-light": {
		PromptTokensCost:     0.3, // $0.30 per 1000 tokens
		CompletionTokensCost: 0.6, // $0.60 per 1000 tokens
	},
	"command-r": {
		PromptTokensCost:     0.5, // $0.50 per 1000 tokens
		CompletionTokensCost: 1.5, // $1.50 per 1000 tokens
	},
	"command-r-plus": {
		PromptTokensCost:     3.0,  // $3.00 per 1000 tokens
		CompletionTokensCost: 15.0, // $15.00 per 1000 tokens
	},

	// DeepSeek Models (pricing per 1000 tokens)
	"deepseek-chat": {
		PromptTokensCost:     0.1, // $0.10 per 1000 tokens
		CompletionTokensCost: 0.2, // $0.20 per 1000 tokens
	},
	"deepseek-coder": {
		PromptTokensCost:     0.1, // $0.10 per 1000 tokens
		CompletionTokensCost: 0.2, // $0.20 per 1000 tokens
	},
}

// NewClient creates a new Paylm SDK client
func NewClient(apiKey string) *Client {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return &Client{
		apiKey:  apiKey,
		baseURL: "https://api.paylm.com", // Replace with actual API URL
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// NewClientWithURL creates a new Paylm SDK client with custom base URL
func NewClientWithURL(apiKey, baseURL string) *Client {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// calculateCost calculates the cost based on model and usage data
func (c *Client) calculateCost(model string, usageData UsageData) (float64, error) {
	pricing, exists := modelPricing[model]
	if !exists {
		c.logger.Warnf("Unknown model '%s', using default pricing", model)
		// Use default pricing for unknown models (per 1000 tokens)
		pricing = ModelPricing{
			PromptTokensCost:     0.1, // $0.10 per 1000 tokens
			CompletionTokensCost: 0.1, // $0.10 per 1000 tokens
		}
	}

	// Calculate cost per 1000 tokens
	promptCost := (float64(usageData.PromptTokens) / 1000.0) * pricing.PromptTokensCost
	completionCost := (float64(usageData.CompletionTokens) / 1000.0) * pricing.CompletionTokensCost
	totalCost := promptCost + completionCost

	c.logger.Debugf("Cost calculation for model '%s': prompt_tokens=%d (%.6f), completion_tokens=%d (%.6f), total=%.6f",
		model, usageData.PromptTokens, promptCost, usageData.CompletionTokens, completionCost, totalCost)

	return totalCost, nil
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

// calculateCostFromStrings calculates the cost based on model and text strings
func (c *Client) calculateCostFromStrings(model string, usageData UsageDataWithStrings) (float64, error) {
	pricing, exists := modelPricing[model]
	if !exists {
		c.logger.Warnf("Unknown model '%s', using default pricing", model)
		// Use default pricing for unknown models (per 1000 tokens)
		pricing = ModelPricing{
			PromptTokensCost:     0.1, // $0.10 per 1000 tokens
			CompletionTokensCost: 0.1, // $0.10 per 1000 tokens
		}
	}

	// Count tokens from strings using proper tokenization
	promptTokens := c.getTokenCount(usageData.Model, usageData.PromptString)
	completionTokens := c.getTokenCount(usageData.Model, usageData.OutputString)

	// Calculate cost per 1000 tokens
	promptCost := (float64(promptTokens) / 1000.0) * pricing.PromptTokensCost
	completionCost := (float64(completionTokens) / 1000.0) * pricing.CompletionTokensCost
	totalCost := promptCost + completionCost

	c.logger.Debugf("Cost calculation for model '%s' from strings: prompt_tokens=%d (%.6f), completion_tokens=%d (%.6f), total=%.6f",
		model, promptTokens, promptCost, completionTokens, completionCost, totalCost)

	return totalCost, nil
}

// SendUsage sends usage data to the Paylm API
func (c *Client) SendUsage(agentID, customerID, indicator string, usageData UsageData) error {
	c.logger.Infof("Starting sendUsage for agentID=%s, customerID=%s, indicator=%s, model=%s",
		agentID, customerID, indicator, usageData.Model)

	// Calculate cost
	cost, err := c.calculateCost(usageData.Model, usageData)
	if err != nil {
		c.logger.Errorf("Failed to calculate cost: %v", err)
		return fmt.Errorf("failed to calculate cost: %w", err)
	}

	c.logger.Infof("Calculated cost: %.6f for model %s", cost, usageData.Model)

	// Prepare API request
	apiRequest := APIRequest{
		AgentID:         agentID,
		CustomerID:      customerID,
		Indicator:       indicator,
		Amount:          cost,
		InputToken:      usageData.PromptTokens,
		OutputToken:     usageData.CompletionTokens,
		Model:           usageData.Model,
		ServiceProvider: usageData.ServiceProvider,
	}

	// Marshal request body
	requestBody, err := json.Marshal(apiRequest)
	if err != nil {
		c.logger.Errorf("Failed to marshal request body: %v", err)
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	c.logger.Debugf("API request body: %s", string(requestBody))

	// Create HTTP request
	url := fmt.Sprintf("%s/api/v1/usage", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		c.logger.Errorf("Failed to create HTTP request: %v", err)
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("paylm-api-key", c.apiKey)

	c.logger.Debugf("Making HTTP POST request to: %s", url)

	// Make HTTP request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Errorf("HTTP request failed: %v", err)
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Errorf("Failed to read response body: %v", err)
		return fmt.Errorf("failed to read response body: %w", err)
	}

	c.logger.Debugf("API response status: %d, body: %s", resp.StatusCode, string(responseBody))

	// Check response status
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		c.logger.Infof("Successfully sent usage data for agentID=%s, customerID=%s, cost=%.6f",
			agentID, customerID, cost)
		return nil
	}

	// Handle error response
	c.logger.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
}

// SendUsageWithTokenString sends usage data to the Paylm API using prompt and output strings
func (c *Client) SendUsageWithTokenString(agentID, customerID, indicator string, usageData UsageDataWithStrings) error {
	c.logger.Infof("Starting sendUsageWithTokenString for agentID=%s, customerID=%s, indicator=%s, serviceProvider=%s, model=%s",
		agentID, customerID, indicator, usageData.ServiceProvider, usageData.Model)

	// Calculate cost from strings
	cost, err := c.calculateCostFromStrings(usageData.Model, usageData)
	if err != nil {
		c.logger.Errorf("Failed to calculate cost from strings: %v", err)
		return fmt.Errorf("failed to calculate cost from strings: %w", err)
	}

	c.logger.Infof("Calculated cost: %.6f for model %s from strings", cost, usageData.Model)

	// Count tokens from strings using proper tokenization
	promptTokens := c.getTokenCount(usageData.Model, usageData.PromptString)
	completionTokens := c.getTokenCount(usageData.Model, usageData.OutputString)

	// Prepare API request
	apiRequest := APIRequest{
		AgentID:         agentID,
		CustomerID:      customerID,
		Indicator:       indicator,
		Amount:          cost,
		InputToken:      promptTokens,
		OutputToken:     completionTokens,
		Model:           usageData.Model,
		ServiceProvider: usageData.ServiceProvider,
	}

	// Marshal request body
	requestBody, err := json.Marshal(apiRequest)
	if err != nil {
		c.logger.Errorf("Failed to marshal request body: %v", err)
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	c.logger.Debugf("API request body: %s", string(requestBody))

	// Create HTTP request
	url := fmt.Sprintf("%s/v1/usage", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		c.logger.Errorf("Failed to create HTTP request: %v", err)
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("paylm-api-key", c.apiKey)

	c.logger.Debugf("Making HTTP POST request to: %s", url)

	// Make HTTP request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Errorf("HTTP request failed: %v", err)
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Errorf("Failed to read response body: %v", err)
		return fmt.Errorf("failed to read response body: %w", err)
	}

	c.logger.Debugf("API response status: %d, body: %s", resp.StatusCode, string(responseBody))

	// Check response status
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		c.logger.Infof("Successfully sent usage data from strings for agentID=%s, customerID=%s, cost=%.6f",
			agentID, customerID, cost)
		return nil
	}

	// Handle error response
	c.logger.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
}

// SetLogLevel sets the logging level for the client
func (c *Client) SetLogLevel(level logrus.Level) {
	c.logger.SetLevel(level)
}

// GetLogger returns the logger instance for custom logging
func (c *Client) GetLogger() *logrus.Logger {
	return c.logger
}
