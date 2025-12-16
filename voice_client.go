package paygent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// SttUsageData represents the STT usage data structure
type SttUsageData struct {
	ServiceProvider string `json:"service_provider"`
	Model           string `json:"model"`
	AudioDuration   int    `json:"audio_duration"` // Duration in seconds
}

// SttAPIRequest represents the request body for the STT API call
type SttAPIRequest struct {
	AgentID         string  `json:"agentId"`
	CustomerID      string  `json:"customerId"`
	Indicator       string  `json:"indicator"`
	Amount          float64 `json:"amount"`
	AudioDuration   int     `json:"audioDuration"`
	Model           string  `json:"model"`
	ServiceProvider string  `json:"serviceProvider"`
}

// SttModelPricing represents pricing information for STT models (cost per hour in USD)
type SttModelPricing struct {
	CostPerHour float64 // Cost per hour in USD
}

// STT model pricing (cost per hour in USD)
var sttModelPricing = map[string]SttModelPricing{
	// Deepgram Models
	DeepgramFlux: {
		CostPerHour: 0.462, // $0.462 per hour
	},
	DeepgramNova3Monolingual: {
		CostPerHour: 0.462, // $0.462 per hour
	},
	DeepgramNova3Multilingual: {
		CostPerHour: 0.552, // $0.552 per hour
	},
	DeepgramNova1: {
		CostPerHour: 0.348, // $0.348 per hour
	},
	DeepgramNova2: {
		CostPerHour: 0.348, // $0.348 per hour
	},
	DeepgramEnhanced: {
		CostPerHour: 0.99, // $0.99 per hour
	},
	DeepgramBase: {
		CostPerHour: 0.87, // $0.87 per hour
	},
	DeepgramRedaction: {
		CostPerHour: 0.12, // $0.12 per hour (add-on)
	},
	DeepgramKeytermPrompting: {
		CostPerHour: 0.072, // $0.072 per hour (add-on)
	},
	DeepgramSpeakerDiarization: {
		CostPerHour: 0.12, // $0.12 per hour (add-on)
	},

	// Microsoft Azure Speech Service Models
	AzureSpeechStandard: {
		CostPerHour: 1.0, // $1.0 per hour
	},
	AzureSpeechCustom: {
		CostPerHour: 1.2, // $1.2 per hour
	},

	// Google Cloud Speech-to-Text Models
	GoogleCloudSpeechStandard: {
		CostPerHour: 0.96, // $0.96 per hour
	},

	// AssemblyAI Models
	AssemblyAIUniversalStreaming: {
		CostPerHour: 0.15, // $0.15 per hour
	},
	AssemblyAIUniversalStreamingMultilang: {
		CostPerHour: 0.15, // $0.15 per hour
	},
	AssemblyAIKeytermsPrompting: {
		CostPerHour: 0.04, // $0.04 per hour
	},
}

// calculateSttCost calculates the cost based on model and audio duration
func (c *Client) calculateSttCost(model string, audioDurationSeconds int) (float64, error) {
	pricing, exists := sttModelPricing[model]
	if !exists {
		c.logger.Warnf("Unknown STT model '%s', using default pricing", model)
		// Use default pricing for unknown models (per hour)
		pricing = SttModelPricing{
			CostPerHour: 0.5, // $0.50 per hour default
		}
	}

	// Calculate cost: (duration in seconds / 3600) * cost per hour
	// Convert seconds to hours and multiply by cost per hour
	durationHours := float64(audioDurationSeconds) / 3600.0
	totalCost := durationHours * pricing.CostPerHour

	c.logger.Debugf("STT cost calculation for model '%s': duration=%d seconds (%.6f hours), cost_per_hour=%.6f, total=%.6f",
		model, audioDurationSeconds, durationHours, pricing.CostPerHour, totalCost)

	return totalCost, nil
}

// SendSttUsage sends STT usage data to the Paygent API
func (c *Client) SendSttUsage(agentID, customerID string, sttUsageData SttUsageData) error {
	c.logger.Infof("Starting SendSttUsage for agentID=%s, customerID=%s, model=%s, duration=%d seconds",
		agentID, customerID, sttUsageData.Model, sttUsageData.AudioDuration)

	// Calculate cost
	cost, err := c.calculateSttCost(sttUsageData.Model, sttUsageData.AudioDuration)
	if err != nil {
		c.logger.Errorf("Failed to calculate STT cost: %v", err)
		return fmt.Errorf("failed to calculate STT cost: %w", err)
	}

	c.logger.Infof("Calculated STT cost: %.6f for model %s", cost, sttUsageData.Model)

	// Prepare API request
	apiRequest := SttAPIRequest{
		AgentID:         agentID,
		CustomerID:      customerID,
		Indicator:       "stt-usage", // Default indicator for STT usage
		Amount:          cost,
		AudioDuration:   sttUsageData.AudioDuration,
		Model:           sttUsageData.Model,
		ServiceProvider: sttUsageData.ServiceProvider,
	}

	// Marshal request body
	requestBody, err := json.Marshal(apiRequest)
	if err != nil {
		c.logger.Errorf("Failed to marshal STT request body: %v", err)
		return fmt.Errorf("failed to marshal STT request body: %w", err)
	}

	c.logger.Debugf("STT API request body: %s", string(requestBody))

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

	c.logger.Debugf("STT API response status: %d, body: %s", resp.StatusCode, string(responseBody))

	// Check response status
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		c.logger.Infof("Successfully sent STT usage data for agentID=%s, customerID=%s, cost=%.6f",
			agentID, customerID, cost)
		return nil
	}

	// Handle error response
	c.logger.Errorf("STT API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	return fmt.Errorf("STT API request failed with status %d: %s", resp.StatusCode, string(responseBody))
}

// TtsUsageData represents the TTS usage data structure
type TtsUsageData struct {
	ServiceProvider string `json:"service_provider"`
	Model           string `json:"model"`
	CharacterCount  int    `json:"character_count"` // Number of characters
}

// TtsAPIRequest represents the request body for the TTS API call
type TtsAPIRequest struct {
	AgentID         string  `json:"agentId"`
	CustomerID      string  `json:"customerId"`
	Indicator       string  `json:"indicator"`
	Amount          float64 `json:"amount"`
	CharacterCount  int     `json:"characterCount"`
	Model           string  `json:"model"`
	ServiceProvider string  `json:"serviceProvider"`
}

// TtsModelPricing represents pricing information for TTS models (cost per 1 million characters in USD)
type TtsModelPricing struct {
	CostPerMillionCharacters float64 // Cost per 1 million characters in USD
}

// TTS model pricing (cost per 1 million characters in USD)
var ttsModelPricing = map[string]TtsModelPricing{
	// Amazon Polly Models
	PollyStandard: {
		CostPerMillionCharacters: 0.4, // $0.4 per 1 million characters
	},
	PollyNeural: {
		CostPerMillionCharacters: 16.0, // $16 per 1 million characters
	},
	PollyLongForm: {
		CostPerMillionCharacters: 100.0, // $100 per 1 million characters
	},
	PollyGenerative: {
		CostPerMillionCharacters: 30.0, // $30 per 1 million characters
	},

	// Microsoft Azure Speech Service TTS Models
	AzureTtsStandard: {
		CostPerMillionCharacters: 15.0, // $15 per 1 million characters
	},
	AzureTtsCustom: {
		CostPerMillionCharacters: 24.0, // $24 per 1 million characters
	},
	AzureTtsCustomNeuralHD: {
		CostPerMillionCharacters: 48.0, // $48 per 1 million characters
	},

	// Google Cloud Text-to-Speech TTS Models
	GoogleCloudTtsChirp3HD: {
		CostPerMillionCharacters: 30.0, // $30 per 1 million characters
	},
	GoogleCloudTtsInstantCustom: {
		CostPerMillionCharacters: 60.0, // $60 per 1 million characters
	},
	GoogleCloudTtsWaveNet: {
		CostPerMillionCharacters: 4.0, // $4 per 1 million characters
	},
	GoogleCloudTtsStudio: {
		CostPerMillionCharacters: 160.0, // $160 per 1 million characters
	},
	GoogleCloudTtsStandard: {
		CostPerMillionCharacters: 4.0, // $4 per 1 million characters
	},
	GoogleCloudTtsNeural2: {
		CostPerMillionCharacters: 16.0, // $16 per 1 million characters
	},
	GoogleCloudTtsPolyglotPreview: {
		CostPerMillionCharacters: 16.0, // $16 per 1 million characters
	},

	// Deepgram TTS Models
	DeepgramAura2: {
		CostPerMillionCharacters: 30.0, // $30 per 1 million characters
	},
	DeepgramAura1: {
		CostPerMillionCharacters: 15.0, // $15 per 1 million characters
	},
}

// calculateTtsCost calculates the cost based on model and character count
func (c *Client) calculateTtsCost(model string, characterCount int) (float64, error) {
	pricing, exists := ttsModelPricing[model]
	if !exists {
		c.logger.Warnf("Unknown TTS model '%s', using default pricing", model)
		// Use default pricing for unknown models (per 1 million characters)
		pricing = TtsModelPricing{
			CostPerMillionCharacters: 10.0, // $10 per 1 million characters default
		}
	}

	// Calculate cost: (character count / 1,000,000) * cost per 1 million characters
	// Convert character count to millions and multiply by cost per million
	charactersInMillions := float64(characterCount) / 1000000.0
	totalCost := charactersInMillions * pricing.CostPerMillionCharacters

	c.logger.Debugf("TTS cost calculation for model '%s': characters=%d (%.6f millions), cost_per_million=%.6f, total=%.6f",
		model, characterCount, charactersInMillions, pricing.CostPerMillionCharacters, totalCost)

	return totalCost, nil
}

// SendTtsUsage sends TTS usage data to the Paygent API
func (c *Client) SendTtsUsage(agentID, customerID string, ttsUsageData TtsUsageData) error {
	c.logger.Infof("Starting SendTtsUsage for agentID=%s, customerID=%s, model=%s, characters=%d",
		agentID, customerID, ttsUsageData.Model, ttsUsageData.CharacterCount)

	// Calculate cost
	cost, err := c.calculateTtsCost(ttsUsageData.Model, ttsUsageData.CharacterCount)
	if err != nil {
		c.logger.Errorf("Failed to calculate TTS cost: %v", err)
		return fmt.Errorf("failed to calculate TTS cost: %w", err)
	}

	c.logger.Infof("Calculated TTS cost: %.6f for model %s", cost, ttsUsageData.Model)

	// Prepare API request
	apiRequest := TtsAPIRequest{
		AgentID:         agentID,
		CustomerID:      customerID,
		Indicator:       "tts-usage", // Default indicator for TTS usage
		Amount:          cost,
		CharacterCount:  ttsUsageData.CharacterCount,
		Model:           ttsUsageData.Model,
		ServiceProvider: ttsUsageData.ServiceProvider,
	}

	// Marshal request body
	requestBody, err := json.Marshal(apiRequest)
	if err != nil {
		c.logger.Errorf("Failed to marshal TTS request body: %v", err)
		return fmt.Errorf("failed to marshal TTS request body: %w", err)
	}

	c.logger.Debugf("TTS API request body: %s", string(requestBody))

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

	c.logger.Debugf("TTS API response status: %d, body: %s", resp.StatusCode, string(responseBody))

	// Check response status
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		c.logger.Infof("Successfully sent TTS usage data for agentID=%s, customerID=%s, cost=%.6f",
			agentID, customerID, cost)
		return nil
	}

	// Handle error response
	c.logger.Errorf("TTS API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	return fmt.Errorf("TTS API request failed with status %d: %s", resp.StatusCode, string(responseBody))
}
