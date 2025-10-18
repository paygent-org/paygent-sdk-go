# Paylm SDK for Go

A Go SDK for integrating with the Paylm API to track usage and costs for AI models.

## Installation

```bash
go get github.com/paylm/paylm-sdk-go
```

## Usage

### Basic Usage

```go
package main

import (
    "log"
    "github.com/paylm/paylm-sdk-go"
)

func main() {
    // Create a new client with your API key
    client := paylm.NewClient("your-paylm-api-key")
    
    // Set log level (optional)
    client.SetLogLevel(logrus.InfoLevel)
    
    // Define usage data
    usageData := paylm.UsageData{
        ServiceProvider:  "Meta",
        Model:            "llama-3-8b",
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
    "github.com/paylm/paylm-sdk-go"
    "github.com/sirupsen/logrus"
)

func main() {
    // Create a new client
    client := paylm.NewClientWithURL("your-api-key", "http://localhost:8080")
    client.SetLogLevel(logrus.InfoLevel)
    
    // Define usage data with prompt and output strings
    usageData := paylm.UsageDataWithStrings{
        ServiceProvider: "OpenAI",
        Model:           "gpt-4",
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
    "github.com/paylm/paylm-sdk-go"
    "github.com/sirupsen/logrus"
)

func main() {
    // Create client with custom base URL
    client := paylm.NewClientWithURL("your-api-key", "https://custom-api.paylm.com")
    
    // Set debug logging
    client.SetLogLevel(logrus.DebugLevel)
    
    // Get logger for custom logging
    logger := client.GetLogger()
    logger.Info("Starting usage tracking...")
    
    // Send usage data
    usageData := paylm.UsageData{
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
Creates a new Paylm SDK client with the default API URL.

#### `NewClientWithURL(apiKey, baseURL string) *Client`
Creates a new Paylm SDK client with a custom base URL.

#### `SendUsage(agentID, customerID, indicator string, usageData UsageData) error`
Sends usage data to the Paylm API. Returns an error if the request fails.

#### `SendUsageWithTokenString(agentID, customerID, indicator string, usageData UsageDataWithStrings) error`
Sends usage data to the Paylm API using prompt and output strings. The function automatically counts tokens using proper tokenizers for each model provider and calculates costs. Returns an error if the request fails.

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
- `gpt-3.5-turbo` - $1.50 prompt, $2.00 completion (per 1000 tokens)
- `gpt-3.5-turbo-16k` - $3.00 prompt, $4.00 completion (per 1000 tokens)
- `gpt-4` - $30.00 prompt, $60.00 completion (per 1000 tokens)
- `gpt-4-turbo` - $10.00 prompt, $30.00 completion (per 1000 tokens)
- `gpt-4o` - $5.00 prompt, $15.00 completion (per 1000 tokens)
- `gpt-4o-mini` - $0.15 prompt, $0.60 completion (per 1000 tokens)
- `gpt-5` - $10.00 prompt, $30.00 completion (per 1000 tokens, estimated)

### Anthropic
- `claude-3-haiku` - $0.25 prompt, $1.25 completion (per 1000 tokens)
- `claude-3-sonnet` - $3.00 prompt, $15.00 completion (per 1000 tokens)
- `claude-3-opus` - $15.00 prompt, $75.00 completion (per 1000 tokens)
- `claude-3.5-sonnet` - $3.00 prompt, $15.00 completion (per 1000 tokens)

### Google DeepMind
- `gemini-pro` - $0.50 prompt, $1.50 completion (per 1000 tokens)
- `gemini-pro-vision` - $0.50 prompt, $1.50 completion (per 1000 tokens)
- `gemini-1.5-pro` - $1.25 prompt, $5.00 completion (per 1000 tokens)
- `gemini-1.5-flash` - $0.075 prompt, $0.30 completion (per 1000 tokens)

### Meta
- `llama-2-7b` - $0.10 per 1000 tokens
- `llama-2-13b` - $0.20 per 1000 tokens
- `llama-2-70b` - $0.70 per 1000 tokens
- `llama-3-8b` - $0.10 per 1000 tokens
- `llama-3-70b` - $0.70 per 1000 tokens

### AWS
- `claude-3-haiku-aws` - $0.25 prompt, $1.25 completion (per 1000 tokens)
- `claude-3-sonnet-aws` - $3.00 prompt, $15.00 completion (per 1000 tokens)
- `titan-text-express` - $0.80 prompt, $1.60 completion (per 1000 tokens)
- `titan-text-lite` - $0.30 prompt, $0.40 completion (per 1000 tokens)

### Mistral AI
- `mistral-7b` - $0.10 per 1000 tokens
- `mistral-8x7b` - $0.20 per 1000 tokens
- `mistral-nemo` - $0.10 per 1000 tokens
- `mistral-large` - $2.00 prompt, $6.00 completion (per 1000 tokens)

### Cohere
- `command` - $1.50 prompt, $2.00 completion (per 1000 tokens)
- `command-light` - $0.30 prompt, $0.60 completion (per 1000 tokens)
- `command-r` - $0.50 prompt, $1.50 completion (per 1000 tokens)
- `command-r-plus` - $3.00 prompt, $15.00 completion (per 1000 tokens)

### DeepSeek
- `deepseek-chat` - $0.10 prompt, $0.20 completion (per 1000 tokens)
- `deepseek-coder` - $0.10 prompt, $0.20 completion (per 1000 tokens)

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
paylm-api-key: your-api-key
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


