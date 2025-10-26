# Paygent SDK for Go

A Go SDK for integrating with the Paygent API to track usage and costs for AI models.

## Installation

```bash
go get github.com/paygent/paygent-sdk-go
```

## Usage

### Model Constants

The SDK provides constants for all supported model names to ensure type safety and avoid typos:

```go
import "github.com/paygent/paygent-sdk-go"

// OpenAI Models
paygent.GPT5
paygent.GPT4O
paygent.GPT35Turbo
// ... and many more

// Anthropic Models
paygent.Sonnet45
paygent.Haiku45
paygent.Opus41
// ... and more

// Google DeepMind Models
paygent.Gemini25Pro
paygent.Gemini25Flash
// ... and more

// Meta Models
paygent.Llama4Maverick
paygent.Llama4Scout
// ... and more

// AWS Models
paygent.AmazonNovaMicro
paygent.AmazonNovaLite
paygent.AmazonNovaPro

// Mistral AI Models
paygent.Mistral7BInstruct
paygent.MistralLarge
// ... and more

// Cohere Models
paygent.CommandR7B
paygent.CommandR
// ... and more

// DeepSeek Models
paygent.DeepSeekChat
paygent.DeepSeekReasoner
// ... and more
```

### Service Provider Constants

The SDK also provides constants for service provider names:

```go
import "github.com/paygent/paygent-sdk-go"

// Service Provider Constants
paygent.OpenAI           // "OpenAI"
paygent.Anthropic        // "Anthropic"
paygent.GoogleDeepMind   // "Google DeepMind"
paygent.Meta             // "Meta"
paygent.AWS              // "AWS"
paygent.MistralAI        // "Mistral AI"
paygent.Cohere           // "Cohere"
paygent.DeepSeek         // "DeepSeek"
paygent.Custom           // "Custom"
```

### Basic Usage

```go
package main

import (
    "log"
    "github.com/paygent/paygent-sdk-go"
)

func main() {
    // Create a new client with your API key
    client := paygent.NewClient("your-paygent-api-key")
    
    // Set log level (optional)
    client.SetLogLevel(logrus.InfoLevel)
    
    // Define usage data using constants
    usageData := paygent.UsageData{
        ServiceProvider:  paygent.Meta,
        Model:            paygent.Llama38BInstructLite,
        PromptTokens:     756,
        CompletionTokens: 244,
        TotalTokens:      1000,
    }
    
    // Send usage data
    err := client.SendUsage("agent-123", "customer-456", "email-sent", usageData)
    if err != nil {
        log.Fatalf("Failed to send usage: %v", err)
    }
    
    log.Println("Usage data sent successfully!")
}
```

### Using SendUsageWithTokenString

```go
package main

import (
    "log"
    "github.com/paygent/paygent-sdk-go"
    "github.com/sirupsen/logrus"
)

func main() {
    // Create a new client
    client := paygent.NewClientWithURL("your-api-key", "http://localhost:8080")
    client.SetLogLevel(logrus.InfoLevel)
    
    // Define usage data with prompt and output strings using constants
    usageData := paygent.UsageDataWithStrings{
        ServiceProvider: paygent.OpenAI,
        Model:           paygent.GPT4O,
        PromptString:    "What is the capital of France? Please provide a detailed explanation.",
        OutputString:    "The capital of France is Paris. Paris is located in the north-central part of France and is the country's largest city and economic center.",
    }
    
    // Send usage data (tokens will be automatically counted)
    err := client.SendUsageWithTokenString("agent-123", "customer-456", "question-answer", usageData)
    if err != nil {
        log.Fatalf("Failed to send usage: %v", err)
    }
    
    log.Println("Usage data sent successfully!")
}
```

### Advanced Usage

```go
package main

import (
    "log"
    "github.com/paygent/paygent-sdk-go"
    "github.com/sirupsen/logrus"
)

func main() {
    // Create client with custom base URL
    client := paygent.NewClientWithURL("your-api-key", "https://custom-api.paygent.com")
    
    // Set debug logging
    client.SetLogLevel(logrus.DebugLevel)
    
    // Get logger for custom logging
    logger := client.GetLogger()
    logger.Info("Starting usage tracking...")
    
    // Send usage data
    usageData := paygent.UsageData{
        Model:            "gpt-4",
        PromptTokens:     1000,
        CompletionTokens: 500,
        TotalTokens:      1500,
    }
    
    err := client.SendUsage("agent-789", "customer-101", "chat-completion", usageData)
    if err != nil {
        logger.Errorf("Failed to send usage: %v", err)
        return
    }
    
    logger.Info("Usage data sent successfully!")
}
```

## API Reference

### Client

#### `NewClient(apiKey string) *Client`
Creates a new Paygent SDK client with the default API URL.

#### `NewClientWithURL(apiKey, baseURL string) *Client`
Creates a new Paygent SDK client with a custom base URL.

#### `SendUsage(agentID, customerID, indicator string, usageData UsageData) error`
Sends usage data to the Paygent API. Returns an error if the request fails.

#### `SendUsageWithTokenString(agentID, customerID, indicator string, usageData UsageDataWithStrings) error`
Sends usage data to the Paygent API using prompt and output strings. The function automatically counts tokens using proper tokenizers for each model provider and calculates costs. Returns an error if the request fails.

#### `SetLogLevel(level logrus.Level)`
Sets the logging level for the client.

#### `GetLogger() *logrus.Logger`
Returns the logger instance for custom logging.

### Types

#### `UsageData`
```go
type UsageData struct {
    ServiceProvider  string `json:"service_provider"`
    Model            string `json:"model"`
    PromptTokens     int    `json:"prompt_tokens"`
    CompletionTokens int    `json:"completion_tokens"`
    TotalTokens      int    `json:"total_tokens"`
}
```

#### `UsageDataWithStrings`
```go
type UsageDataWithStrings struct {
    ServiceProvider string `json:"service_provider"`
    Model           string `json:"model"`
    PromptString    string `json:"prompt_string"`
    OutputString    string `json:"output_string"`
}
```

## Supported Models

The SDK includes built-in pricing for models from the following providers:

### OpenAI
- `gpt-5` - $0.00125 prompt, $0.01 completion (per 1000 tokens)
- `gpt-5-mini` - $0.00025 prompt, $0.002 completion (per 1000 tokens)
- `gpt-5-nano` - $0.00005 prompt, $0.0004 completion (per 1000 tokens)
- `gpt-5-pro` - $0.015 prompt, $0.12 completion (per 1000 tokens)
- `gpt-4.1` - $0.002 prompt, $0.008 completion (per 1000 tokens)
- `gpt-4.1-mini` - $0.0004 prompt, $0.0016 completion (per 1000 tokens)
- `gpt-4o` - $0.0025 prompt, $0.01 completion (per 1000 tokens)
- `gpt-4o-mini` - $0.00015 prompt, $0.0006 completion (per 1000 tokens)
- `o1` - $0.015 prompt, $0.06 completion (per 1000 tokens)
- `o1-pro` - $0.15 prompt, $0.6 completion (per 1000 tokens)
- `o3-pro` - $0.02 prompt, $0.08 completion (per 1000 tokens)
- `o3` - $0.002 prompt, $0.008 completion (per 1000 tokens)
- `gpt-3.5-turbo` - $0.0005 prompt, $0.0015 completion (per 1000 tokens)
- `gpt-4-0613` - $0.03 prompt, $0.06 completion (per 1000 tokens)
- `davinci-002` - $0.002 prompt, $0.002 completion (per 1000 tokens)
- `babbage-002` - $0.0004 prompt, $0.0004 completion (per 1000 tokens)
- *... and many more specialized models including realtime, audio, search, and preview variants*

### Anthropic
- `sonnet-4.5` - $0.003 prompt, $0.015 completion (per 1000 tokens)
- `haiku-4.5` - $0.001 prompt, $0.005 completion (per 1000 tokens)
- `opus-4.1` - $0.015 prompt, $0.075 completion (per 1000 tokens)
- `sonnet-4` - $0.003 prompt, $0.015 completion (per 1000 tokens)
- `opus-4` - $0.015 prompt, $0.075 completion (per 1000 tokens)
- `sonnet-3.7` - $0.003 prompt, $0.015 completion (per 1000 tokens)
- `haiku-3.5` - $0.0008 prompt, $0.004 completion (per 1000 tokens)
- `opus-3` - $0.015 prompt, $0.075 completion (per 1000 tokens)
- `haiku-3` - $0.00025 prompt, $0.00125 completion (per 1000 tokens)

### Google DeepMind
- `gemini-2.5-pro` - $0.00125 prompt, $0.01 completion (per 1000 tokens)
- `gemini-2.5-flash` - $0.00015 prompt, $0.0006 completion (per 1000 tokens)
- `gemini-2.5-flash-preview` - $0.30 prompt, $2.50 completion (per 1000 tokens)
- `gemini-2.5-flash-lite` - $0.0001 prompt, $0.0004 completion (per 1000 tokens)
- `gemini-2.5-flash-lite-preview` - $0.0001 prompt, $0.0004 completion (per 1000 tokens)
- `gemini-2.5-flash-native-audio` - $0.0005 prompt, $0.002 completion (per 1000 tokens)
- `gemini-2.5-flash-image` - $0.0003 prompt, $0.03 completion (per 1000 tokens)
- `gemini-2.5-flash-preview-tts` - $0.0005 prompt, $0.01 completion (per 1000 tokens)
- `gemini-2.5-pro-preview-tts` - $0.001 prompt, $0.02 completion (per 1000 tokens)
- `gemini-2.5-computer-use-preview` - $0.00125 prompt, $0.01 completion (per 1000 tokens)

### Meta
- `llama-2-7b` - $0.10 per 1000 tokens
- `llama-2-13b` - $0.20 per 1000 tokens
- `llama-2-70b` - $0.70 per 1000 tokens
- `llama-3-8b` - $0.10 per 1000 tokens
- `llama-3-70b` - $0.70 per 1000 tokens

### AWS
- `amazon-nova-micro` - $0.035 prompt, $0.14 completion (per 1000 tokens)
- `amazon-nova-lite` - $0.06 prompt, $0.24 completion (per 1000 tokens)
- `amazon-nova-pro` - $0.8 prompt, $3.2 completion (per 1000 tokens)

### Mistral AI
- `mistral-7b-instruct` - $0.028 prompt, $0.054 completion (per 1000 tokens)
- `mistral-large` - $2.00 prompt, $6.00 completion (per 1000 tokens)
- `mistral-small` - $0.20 prompt, $0.60 completion (per 1000 tokens)
- `mistral-medium` - $0.40 prompt, $2.00 completion (per 1000 tokens)

### Cohere
- `command-r7b` - $0.0000375 prompt, $0.00015 completion (per 1000 tokens)
- `command-r` - $0.00015 prompt, $0.0006 completion (per 1000 tokens)
- `command-r-plus` - $0.00250 prompt, $0.01 completion (per 1000 tokens)
- `command-a` - $0.001 prompt, $0.002 completion (per 1000 tokens)
- `aya-expanse-8b` - $0.00050 prompt, $0.00150 completion (per 1000 tokens)
- `aya-expanse-32b` - $0.00050 prompt, $0.00150 completion (per 1000 tokens)

### DeepSeek
- `deepseek-chat` - $0.00007 prompt, $0.00027 completion (per 1000 tokens)
- `deepseek-reasoner` - $0.00014 prompt, $0.00219 completion (per 1000 tokens)
- `deepseek-r1-global` - $0.00135 prompt, $0.0054 completion (per 1000 tokens)
- `deepseek-r1-datazone` - $0.001485 prompt, $0.00594 completion (per 1000 tokens)
- `deepseek-v3.2-exp` - $0.000028 prompt, $0.00042 completion (per 1000 tokens)

For unknown models, the SDK will use default pricing of $0.10 per 1000 tokens.

## Token Counting

The SDK uses accurate token counting for different model providers:

- **OpenAI GPT models**: Uses the official tiktoken library with model-specific encodings
- **Anthropic Claude models**: Uses cl100k_base encoding (same as GPT-4)
- **Google Gemini models**: Uses cl100k_base encoding as approximation
- **Meta Llama models**: Uses cl100k_base encoding as approximation
- **Mistral models**: Uses cl100k_base encoding as approximation
- **Cohere models**: Uses cl100k_base encoding as approximation
- **DeepSeek models**: Uses cl100k_base encoding as approximation
- **AWS Titan models**: Uses cl100k_base encoding as approximation
- **Unknown models**: Falls back to word-based estimation (1.3 tokens per word)

The token counting is performed automatically when using `SendUsageWithTokenString()`.

## API Request Format

Both `SendUsage` and `SendUsageWithTokenString` functions send HTTP POST requests to your API endpoint with the following JSON format:

```json
{
  "agentId": "agent-123",
  "customerId": "customer-456", 
  "indicator": "question-answer",
  "amount": 0.045,
  "inputToken": 15,
  "outputToken": 8,
  "model": "gpt-4",
  "serviceProvider": "OpenAI"
}
```

**Headers:**
```
Content-Type: application/json
paygent-api-key: your-api-key
```

## Logging

The SDK uses structured logging with the `logrus` library. You can control the log level and access the logger for custom logging.

## Error Handling

The SDK returns detailed errors for various failure scenarios:

- Invalid usage data
- Network errors
- API errors
- Cost calculation errors

## License

MIT


