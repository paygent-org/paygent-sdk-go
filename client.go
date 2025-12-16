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
	GPT5: {
		PromptTokensCost:     0.00125, // $0.00125 per 1000 tokens
		CompletionTokensCost: 0.01,    // $0.01 per 1000 tokens
	},
	GPT5Mini: {
		PromptTokensCost:     0.00025, // $0.00025 per 1000 tokens
		CompletionTokensCost: 0.002,   // $0.002 per 1000 tokens
	},
	GPT5Nano: {
		PromptTokensCost:     0.00005, // $0.00005 per 1000 tokens
		CompletionTokensCost: 0.0004,  // $0.0004 per 1000 tokens
	},
	GPT5ChatLatest: {
		PromptTokensCost:     0.00125, // $0.00125 per 1000 tokens
		CompletionTokensCost: 0.01,    // $0.01 per 1000 tokens
	},
	GPT5Codex: {
		PromptTokensCost:     0.00125, // $0.00125 per 1000 tokens
		CompletionTokensCost: 0.01,    // $0.01 per 1000 tokens
	},
	GPT5Pro: {
		PromptTokensCost:     0.015, // $0.015 per 1000 tokens
		CompletionTokensCost: 0.12,  // $0.12 per 1000 tokens
	},
	GPT5SearchAPI: {
		PromptTokensCost:     0.00125, // $0.00125 per 1000 tokens
		CompletionTokensCost: 0.01,    // $0.01 per 1000 tokens
	},
	GPT41: {
		PromptTokensCost:     0.002, // $0.002 per 1000 tokens
		CompletionTokensCost: 0.008, // $0.008 per 1000 tokens
	},
	GPT41Mini: {
		PromptTokensCost:     0.0004, // $0.0004 per 1000 tokens
		CompletionTokensCost: 0.0016, // $0.0016 per 1000 tokens
	},
	GPT41Nano: {
		PromptTokensCost:     0.0001, // $0.0001 per 1000 tokens
		CompletionTokensCost: 0.0004, // $0.0004 per 1000 tokens
	},
	GPT4O: {
		PromptTokensCost:     0.0025, // $0.0025 per 1000 tokens
		CompletionTokensCost: 0.01,   // $0.01 per 1000 tokens
	},
	GPT4O20240513: {
		PromptTokensCost:     0.005, // $0.005 per 1000 tokens
		CompletionTokensCost: 0.015, // $0.015 per 1000 tokens
	},
	GPT4OMini: {
		PromptTokensCost:     0.00015, // $0.00015 per 1000 tokens
		CompletionTokensCost: 0.0006,  // $0.0006 per 1000 tokens
	},
	GPTRealtime: {
		PromptTokensCost:     0.004, // $0.004 per 1000 tokens
		CompletionTokensCost: 0.016, // $0.016 per 1000 tokens
	},
	GPTRealtimeMini: {
		PromptTokensCost:     0.0006, // $0.0006 per 1000 tokens
		CompletionTokensCost: 0.0024, // $0.0024 per 1000 tokens
	},
	GPT4ORealtimePreview: {
		PromptTokensCost:     0.005, // $0.005 per 1000 tokens
		CompletionTokensCost: 0.02,  // $0.02 per 1000 tokens
	},
	GPT4OMiniRealtimePreview: {
		PromptTokensCost:     0.0006, // $0.0006 per 1000 tokens
		CompletionTokensCost: 0.0024, // $0.0024 per 1000 tokens
	},
	GPTAudio: {
		PromptTokensCost:     0.0025, // $0.0025 per 1000 tokens
		CompletionTokensCost: 0.01,   // $0.01 per 1000 tokens
	},
	GPTAudioMini: {
		PromptTokensCost:     0.0006, // $0.0006 per 1000 tokens
		CompletionTokensCost: 0.0024, // $0.0024 per 1000 tokens
	},
	GPT4OAudioPreview: {
		PromptTokensCost:     0.0025, // $0.0025 per 1000 tokens
		CompletionTokensCost: 0.01,   // $0.01 per 1000 tokens
	},
	GPT4OMiniAudioPreview: {
		PromptTokensCost:     0.00015, // $0.00015 per 1000 tokens
		CompletionTokensCost: 0.0006,  // $0.0006 per 1000 tokens
	},
	O1: {
		PromptTokensCost:     0.015, // $0.015 per 1000 tokens
		CompletionTokensCost: 0.06,  // $0.06 per 1000 tokens
	},
	O1Pro: {
		PromptTokensCost:     0.15, // $0.15 per 1000 tokens
		CompletionTokensCost: 0.6,  // $0.6 per 1000 tokens
	},
	O3Pro: {
		PromptTokensCost:     0.02, // $0.02 per 1000 tokens
		CompletionTokensCost: 0.08, // $0.08 per 1000 tokens
	},
	O3: {
		PromptTokensCost:     0.002, // $0.002 per 1000 tokens
		CompletionTokensCost: 0.008, // $0.008 per 1000 tokens
	},
	O3DeepResearch: {
		PromptTokensCost:     0.01, // $0.01 per 1000 tokens
		CompletionTokensCost: 0.04, // $0.04 per 1000 tokens
	},
	O4Mini: {
		PromptTokensCost:     0.0011, // $0.0011 per 1000 tokens
		CompletionTokensCost: 0.0044, // $0.0044 per 1000 tokens
	},
	O4MiniDeepResearch: {
		PromptTokensCost:     0.002, // $0.002 per 1000 tokens
		CompletionTokensCost: 0.008, // $0.008 per 1000 tokens
	},
	O3Mini: {
		PromptTokensCost:     0.0011, // $0.0011 per 1000 tokens
		CompletionTokensCost: 0.0044, // $0.0044 per 1000 tokens
	},
	O1Mini: {
		PromptTokensCost:     0.0011, // $0.0011 per 1000 tokens
		CompletionTokensCost: 0.0044, // $0.0044 per 1000 tokens
	},
	CodexMiniLatest: {
		PromptTokensCost:     0.0015, // $0.0015 per 1000 tokens
		CompletionTokensCost: 0.006,  // $0.006 per 1000 tokens
	},
	GPT4OMiniSearchPreview: {
		PromptTokensCost:     0.00015, // $0.00015 per 1000 tokens
		CompletionTokensCost: 0.0006,  // $0.0006 per 1000 tokens
	},
	GPT4OSearchPreview: {
		PromptTokensCost:     0.0025, // $0.0025 per 1000 tokens
		CompletionTokensCost: 0.01,   // $0.01 per 1000 tokens
	},
	ComputerUsePreview: {
		PromptTokensCost:     0.003, // $0.003 per 1000 tokens
		CompletionTokensCost: 0.012, // $0.012 per 1000 tokens
	},
	ChatGPT4OLatest: {
		PromptTokensCost:     0.005, // $0.005 per 1000 tokens
		CompletionTokensCost: 0.015, // $0.015 per 1000 tokens
	},
	GPT4Turbo20240409: {
		PromptTokensCost:     0.01, // $0.01 per 1000 tokens
		CompletionTokensCost: 0.03, // $0.03 per 1000 tokens
	},
	GPT40125Preview: {
		PromptTokensCost:     0.01, // $0.01 per 1000 tokens
		CompletionTokensCost: 0.03, // $0.03 per 1000 tokens
	},
	GPT41106Preview: {
		PromptTokensCost:     0.01, // $0.01 per 1000 tokens
		CompletionTokensCost: 0.03, // $0.03 per 1000 tokens
	},
	GPT41106VisionPreview: {
		PromptTokensCost:     0.01, // $0.01 per 1000 tokens
		CompletionTokensCost: 0.03, // $0.03 per 1000 tokens
	},
	GPT40613: {
		PromptTokensCost:     0.03, // $0.03 per 1000 tokens
		CompletionTokensCost: 0.06, // $0.06 per 1000 tokens
	},
	GPT40314: {
		PromptTokensCost:     0.03, // $0.03 per 1000 tokens
		CompletionTokensCost: 0.06, // $0.06 per 1000 tokens
	},
	GPT432K: {
		PromptTokensCost:     0.06, // $0.06 per 1000 tokens
		CompletionTokensCost: 0.12, // $0.12 per 1000 tokens
	},
	GPT35Turbo: {
		PromptTokensCost:     0.0005, // $0.0005 per 1000 tokens
		CompletionTokensCost: 0.0015, // $0.0015 per 1000 tokens
	},
	GPT35Turbo0125: {
		PromptTokensCost:     0.0005, // $0.0005 per 1000 tokens
		CompletionTokensCost: 0.0015, // $0.0015 per 1000 tokens
	},
	GPT35Turbo1106: {
		PromptTokensCost:     0.001, // $0.001 per 1000 tokens
		CompletionTokensCost: 0.002, // $0.002 per 1000 tokens
	},
	GPT35Turbo0613: {
		PromptTokensCost:     0.0015, // $0.0015 per 1000 tokens
		CompletionTokensCost: 0.002,  // $0.002 per 1000 tokens
	},
	GPT350301: {
		PromptTokensCost:     0.0015, // $0.0015 per 1000 tokens
		CompletionTokensCost: 0.002,  // $0.002 per 1000 tokens
	},
	GPT35TurboInstruct: {
		PromptTokensCost:     0.0015, // $0.0015 per 1000 tokens
		CompletionTokensCost: 0.002,  // $0.002 per 1000 tokens
	},
	GPT35Turbo16K0613: {
		PromptTokensCost:     0.003, // $0.003 per 1000 tokens
		CompletionTokensCost: 0.004, // $0.004 per 1000 tokens
	},
	Davinci002: {
		PromptTokensCost:     0.002, // $0.002 per 1000 tokens
		CompletionTokensCost: 0.002, // $0.002 per 1000 tokens
	},
	Babbage002: {
		PromptTokensCost:     0.0004, // $0.0004 per 1000 tokens
		CompletionTokensCost: 0.0004, // $0.0004 per 1000 tokens
	},

	// Anthropic Models (pricing per 1000 tokens)
	Sonnet45: {
		PromptTokensCost:     0.003, // $0.003 per 1000 tokens
		CompletionTokensCost: 0.015, // $0.015 per 1000 tokens
	},
	Haiku45: {
		PromptTokensCost:     0.001, // $0.001 per 1000 tokens
		CompletionTokensCost: 0.005, // $0.005 per 1000 tokens
	},
	Opus41: {
		PromptTokensCost:     0.015, // $0.015 per 1000 tokens
		CompletionTokensCost: 0.075, // $0.075 per 1000 tokens
	},
	Sonnet4: {
		PromptTokensCost:     0.003, // $0.003 per 1000 tokens
		CompletionTokensCost: 0.015, // $0.015 per 1000 tokens
	},
	Opus4: {
		PromptTokensCost:     0.015, // $0.015 per 1000 tokens
		CompletionTokensCost: 0.075, // $0.075 per 1000 tokens
	},
	Sonnet37: {
		PromptTokensCost:     0.003, // $0.003 per 1000 tokens
		CompletionTokensCost: 0.015, // $0.015 per 1000 tokens
	},
	Haiku35: {
		PromptTokensCost:     0.0008, // $0.0008 per 1000 tokens
		CompletionTokensCost: 0.004,  // $0.004 per 1000 tokens
	},
	Opus3: {
		PromptTokensCost:     0.015, // $0.015 per 1000 tokens
		CompletionTokensCost: 0.075, // $0.075 per 1000 tokens
	},
	Haiku3: {
		PromptTokensCost:     0.00025, // $0.00025 per 1000 tokens
		CompletionTokensCost: 0.00125, // $0.00125 per 1000 tokens
	},

	// Google DeepMind Models (pricing per 1000 tokens)
	Gemini25Pro: {
		PromptTokensCost:     0.00125, // $0.00125 per 1000 tokens
		CompletionTokensCost: 0.01,    // $0.01 per 1000 tokens
	},
	Gemini25Flash: {
		PromptTokensCost:     0.00015, // $0.00015 per 1000 tokens
		CompletionTokensCost: 0.0006,  // $0.0006 per 1000 tokens
	},
	Gemini25FlashPreview: {
		PromptTokensCost:     0.3, // $0.30 per 1000 tokens
		CompletionTokensCost: 2.5, // $2.50 per 1000 tokens
	},
	Gemini25FlashLite: {
		PromptTokensCost:     0.0001, // $0.0001 per 1000 tokens
		CompletionTokensCost: 0.0004, // $0.0004 per 1000 tokens
	},
	Gemini25FlashLitePreview: {
		PromptTokensCost:     0.0001, // $0.0001 per 1000 tokens
		CompletionTokensCost: 0.0004, // $0.0004 per 1000 tokens
	},
	Gemini25FlashNativeAudio: {
		PromptTokensCost:     0.0005, // $0.0005 per 1000 tokens
		CompletionTokensCost: 0.002,  // $0.002 per 1000 tokens
	},
	Gemini25FlashImage: {
		PromptTokensCost:     0.0003, // $0.0003 per 1000 tokens
		CompletionTokensCost: 0.03,   // $0.03 per 1000 tokens
	},
	Gemini25FlashPreviewTTS: {
		PromptTokensCost:     0.0005, // $0.0005 per 1000 tokens
		CompletionTokensCost: 0.01,   // $0.01 per 1000 tokens
	},
	Gemini25ProPreviewTTS: {
		PromptTokensCost:     0.001, // $0.001 per 1000 tokens
		CompletionTokensCost: 0.02,  // $0.02 per 1000 tokens
	},
	Gemini25ComputerUsePreview: {
		PromptTokensCost:     0.00125, // $0.00125 per 1000 tokens
		CompletionTokensCost: 0.01,    // $0.01 per 1000 tokens
	},

	// Meta Models (pricing per 1000 tokens)
	Llama4Maverick: {
		PromptTokensCost:     0.00027, // $0.00027 per 1000 tokens
		CompletionTokensCost: 0.00085, // $0.00085 per 1000 tokens
	},
	Llama4Scout: {
		PromptTokensCost:     0.00018, // $0.00018 per 1000 tokens
		CompletionTokensCost: 0.00059, // $0.00059 per 1000 tokens
	},
	Llama3370BInstructTurbo: {
		PromptTokensCost:     0.00088, // $0.00088 per 1000 tokens
		CompletionTokensCost: 0.00088, // $0.00088 per 1000 tokens
	},
	Llama323BInstructTurbo: {
		PromptTokensCost:     0.00006, // $0.00006 per 1000 tokens
		CompletionTokensCost: 0.00006, // $0.00006 per 1000 tokens
	},
	Llama31405BInstructTurbo: {
		PromptTokensCost:     0.0035, // $0.0035 per 1000 tokens
		CompletionTokensCost: 0.0035, // $0.0035 per 1000 tokens
	},
	Llama3170BInstructTurbo: {
		PromptTokensCost:     0.00088, // $0.00088 per 1000 tokens
		CompletionTokensCost: 0.00088, // $0.00088 per 1000 tokens
	},
	Llama318BInstructTurbo: {
		PromptTokensCost:     0.00018, // $0.00018 per 1000 tokens
		CompletionTokensCost: 0.00018, // $0.00018 per 1000 tokens
	},
	Llama370BInstructTurbo: {
		PromptTokensCost:     0.00088, // $0.00088 per 1000 tokens
		CompletionTokensCost: 0.00088, // $0.00088 per 1000 tokens
	},
	Llama370BInstructReference: {
		PromptTokensCost:     0.00088, // $0.00088 per 1000 tokens
		CompletionTokensCost: 0.00088, // $0.00088 per 1000 tokens
	},
	Llama38BInstructLite: {
		PromptTokensCost:     0.0001, // $0.0001 per 1000 tokens
		CompletionTokensCost: 0.0001, // $0.0001 per 1000 tokens
	},
	LLaMA2: {
		PromptTokensCost:     0.0009, // $0.0009 per 1000 tokens
		CompletionTokensCost: 0.0009, // $0.0009 per 1000 tokens
	},
	LlamaGuard412B: {
		PromptTokensCost:     0.0002, // $0.0002 per 1000 tokens
		CompletionTokensCost: 0.0002, // $0.0002 per 1000 tokens
	},
	LlamaGuard311BVisionTurbo: {
		PromptTokensCost:     0.00018, // $0.00018 per 1000 tokens
		CompletionTokensCost: 0.00018, // $0.00018 per 1000 tokens
	},
	LlamaGuard38B: {
		PromptTokensCost:     0.0002, // $0.0002 per 1000 tokens
		CompletionTokensCost: 0.0002, // $0.0002 per 1000 tokens
	},
	LlamaGuard28B: {
		PromptTokensCost:     0.0002, // $0.0002 per 1000 tokens
		CompletionTokensCost: 0.0002, // $0.0002 per 1000 tokens
	},
	SalesforceLlamaRankV18B: {
		PromptTokensCost:     0.0001, // $0.0001 per 1000 tokens
		CompletionTokensCost: 0.0001, // $0.0001 per 1000 tokens
	},

	// AWS Models (pricing per 1000 tokens)
	AmazonNovaMicro: {
		PromptTokensCost:     0.035, // $0.035 per 1000 tokens
		CompletionTokensCost: 0.14,  // $0.14 per 1000 tokens
	},
	AmazonNovaLite: {
		PromptTokensCost:     0.06, // $0.06 per 1000 tokens
		CompletionTokensCost: 0.24, // $0.24 per 1000 tokens
	},
	AmazonNovaPro: {
		PromptTokensCost:     0.8, // $0.8 per 1000 tokens
		CompletionTokensCost: 3.2, // $3.2 per 1000 tokens
	},

	// Mistral AI Models (pricing per 1000 tokens)
	Mistral7BInstruct: {
		PromptTokensCost:     0.028, // $0.028 per 1000 tokens
		CompletionTokensCost: 0.054, // $0.054 per 1000 tokens
	},
	MistralLarge: {
		PromptTokensCost:     2.0, // $2.00 per 1000 tokens
		CompletionTokensCost: 6.0, // $6.00 per 1000 tokens
	},
	MistralSmall: {
		PromptTokensCost:     0.2, // $0.20 per 1000 tokens
		CompletionTokensCost: 0.6, // $0.60 per 1000 tokens
	},
	MistralMedium: {
		PromptTokensCost:     0.4, // $0.40 per 1000 tokens
		CompletionTokensCost: 2.0, // $2.00 per 1000 tokens
	},

	// Cohere Models (pricing per 1000 tokens)
	CommandR7B: {
		PromptTokensCost:     0.0000375, // $0.0000375 per 1000 tokens
		CompletionTokensCost: 0.00015,   // $0.00015 per 1000 tokens
	},
	CommandR: {
		PromptTokensCost:     0.00015, // $0.00015 per 1000 tokens
		CompletionTokensCost: 0.0006,  // $0.0006 per 1000 tokens
	},
	CommandRPlus: {
		PromptTokensCost:     0.00250, // $0.00250 per 1000 tokens
		CompletionTokensCost: 0.01,    // $0.01 per 1000 tokens
	},
	CommandA: {
		PromptTokensCost:     0.001, // $0.001 per 1000 tokens
		CompletionTokensCost: 0.002, // $0.002 per 1000 tokens
	},
	AyaExpanse8B32B: {
		PromptTokensCost:     0.00050, // $0.00050 per 1000 tokens
		CompletionTokensCost: 0.00150, // $0.00150 per 1000 tokens
	},

	// DeepSeek Models (pricing per 1000 tokens)
	DeepSeekChat: {
		PromptTokensCost:     0.00007, // $0.00007 per 1000 tokens
		CompletionTokensCost: 0.00027, // $0.00027 per 1000 tokens
	},
	DeepSeekReasoner: {
		PromptTokensCost:     0.00014, // $0.00014 per 1000 tokens
		CompletionTokensCost: 0.00219, // $0.00219 per 1000 tokens
	},
	DeepSeekR1Global: {
		PromptTokensCost:     0.00135, // $0.00135 per 1000 tokens
		CompletionTokensCost: 0.0054,  // $0.0054 per 1000 tokens
	},
	DeepSeekR1DataZone: {
		PromptTokensCost:     0.001485, // $0.001485 per 1000 tokens
		CompletionTokensCost: 0.00594,  // $0.00594 per 1000 tokens
	},
	DeepSeekV32Exp: {
		PromptTokensCost:     0.000028, // $0.000028 per 1000 tokens
		CompletionTokensCost: 0.00042,  // $0.00042 per 1000 tokens
	},
}

// NewClient creates a new Paygent SDK client
func NewClient(apiKey string) *Client {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return &Client{
		apiKey:  apiKey,
		baseURL: "http://13.201.118.45:8080",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// NewClientWithURL creates a new Paygent SDK client with custom base URL
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

// SendUsage sends usage data to the Paygent API
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
	req.Header.Set("paygent-api-key", c.apiKey)

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

// SendUsageWithTokenString sends usage data to the Paygent API using prompt and output strings
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
	url := fmt.Sprintf("%s/api/v1/usage", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		c.logger.Errorf("Failed to create HTTP request: %v", err)
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("paygent-api-key", c.apiKey)

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
