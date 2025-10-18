package paylm

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-api-key")
	if client == nil {
		t.Fatal("Expected client to be created")
	}
	if client.apiKey != "test-api-key" {
		t.Errorf("Expected apiKey to be 'test-api-key', got '%s'", client.apiKey)
	}
	if client.baseURL != "https://api.paylm.com" {
		t.Errorf("Expected baseURL to be 'https://api.paylm.com', got '%s'", client.baseURL)
	}
}

func TestNewClientWithURL(t *testing.T) {
	customURL := "https://custom-api.paylm.com"
	client := NewClientWithURL("test-api-key", customURL)
	if client == nil {
		t.Fatal("Expected client to be created")
	}
	if client.apiKey != "test-api-key" {
		t.Errorf("Expected apiKey to be 'test-api-key', got '%s'", client.apiKey)
	}
	if client.baseURL != customURL {
		t.Errorf("Expected baseURL to be '%s', got '%s'", customURL, client.baseURL)
	}
}

func TestCalculateCost(t *testing.T) {
	client := NewClient("test-api-key")

	tests := []struct {
		name      string
		model     string
		usageData UsageData
		expected  float64
	}{
		{
			name:  "Llama model",
			model: "llama",
			usageData: UsageData{
				PromptTokens:     1000,
				CompletionTokens: 500,
			},
			expected: 0.15, // (1000 + 500) / 1000 * 0.1
		},
		{
			name:  "GPT-4 model",
			model: "gpt-4",
			usageData: UsageData{
				PromptTokens:     1000,
				CompletionTokens: 500,
			},
			expected: 60.0, // 1000 * 0.03 + 500 * 0.06
		},
		{
			name:  "Unknown model",
			model: "unknown-model",
			usageData: UsageData{
				PromptTokens:     1000,
				CompletionTokens: 500,
			},
			expected: 0.15, // (1000 + 500) / 1000 * 0.1 (default pricing)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost, err := client.calculateCost(tt.model, tt.usageData)
			if err != nil {
				t.Errorf("calculateCost() error = %v", err)
				return
			}
			// Use tolerance for floating point comparison
			tolerance := 0.0001
			if cost < tt.expected-tolerance || cost > tt.expected+tolerance {
				t.Errorf("calculateCost() = %v, want %v (tolerance: %v)", cost, tt.expected, tolerance)
			}
		})
	}
}

func TestSetLogLevel(t *testing.T) {
	client := NewClient("test-api-key")
	client.SetLogLevel(logrus.DebugLevel)

	if client.logger.GetLevel() != logrus.DebugLevel {
		t.Errorf("Expected log level to be DebugLevel, got %v", client.logger.GetLevel())
	}
}

func TestGetLogger(t *testing.T) {
	client := NewClient("test-api-key")
	logger := client.GetLogger()

	if logger == nil {
		t.Fatal("Expected logger to be returned")
	}
	if logger != client.logger {
		t.Error("Expected returned logger to be the same as client's logger")
	}
}

func TestGetTokenCount(t *testing.T) {
	client := NewClient("test-api-key")

	tests := []struct {
		name     string
		model    string
		text     string
		expected int
	}{
		{
			name:     "Empty string",
			model:    "gpt-3.5-turbo",
			text:     "",
			expected: 0,
		},
		{
			name:     "Single word GPT",
			model:    "gpt-3.5-turbo",
			text:     "hello",
			expected: 1,
		},
		{
			name:     "Multiple words Claude",
			model:    "claude-3-sonnet",
			text:     "hello world test",
			expected: 3,
		},
		{
			name:     "Longer text Gemini",
			model:    "gemini-pro",
			text:     "This is a longer text with multiple words to test token counting",
			expected: 12,
		},
		{
			name:     "Unknown model fallback",
			model:    "unknown-model",
			text:     "test text",
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.getTokenCount(tt.model, tt.text)
			if result < tt.expected {
				t.Errorf("getTokenCount() = %v, expected at least %v", result, tt.expected)
			}
		})
	}
}

func TestCalculateCostFromStrings(t *testing.T) {
	client := NewClient("test-api-key")

	tests := []struct {
		name      string
		model     string
		usageData UsageDataWithStrings
		expected  float64
	}{
		{
			name:  "GPT-4 model with strings",
			model: "gpt-4",
			usageData: UsageDataWithStrings{
				ServiceProvider: "OpenAI",
				Model:           "gpt-4",
				PromptString:    "What is the capital of France?",
				OutputString:    "The capital of France is Paris.",
			},
			expected: 0.0, // Will be calculated based on token count
		},
		{
			name:  "Unknown model with strings",
			model: "unknown-model",
			usageData: UsageDataWithStrings{
				ServiceProvider: "Unknown",
				Model:           "unknown-model",
				PromptString:    "test prompt",
				OutputString:    "test output",
			},
			expected: 0.0, // Will be calculated based on token count
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost, err := client.calculateCostFromStrings(tt.model, tt.usageData)
			if err != nil {
				t.Errorf("calculateCostFromStrings() error = %v", err)
				return
			}
			if cost < 0 {
				t.Errorf("calculateCostFromStrings() = %v, expected non-negative value", cost)
			}
		})
	}
}
