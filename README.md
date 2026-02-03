# Paygent SDK for Go

A Go SDK for integrating with the Paygent API to track usage and costs for AI models with server-side pricing.

## Installation

```bash
go get github.com/paygent-org/paygent-sdk-go
```

## core features

- **Server-Side Pricing**: No more client-side pricing logic. All costs are calculated and managed on the Paygent server.
- **Voice Session Tracking**: End-to-end tracking for voice applications (STT -> LLM -> TTS).
- **Customer Management**: Easily create or retrieve customers via external IDs.
- **Automatic Token Counting**: Built-in support for accurate tokenization across major providers.
- **V2 API Support**: Native support for the Paygent V2 usage endpoint.

## Usage

### Model Constants

The SDK provides constants for all supported model names:

```go
import "github.com/paygent-org/paygent-sdk-go"

// OpenAI Models
paygent.GPT4O
paygent.O1
// ...

// Anthropic Models
paygent.Sonnet37
paygent.Haiku35
// ...

// Deepgram Models
paygent.DeepgramNova2
paygent.DeepgramAura2
// ...
```

### Basic Usage (V2 API)

```go
package main

import (
    "log"
    "github.com/paygent-org/paygent-sdk-go"
)

func main() {
    client := paygent.NewClient("your-paygent-api-key")
    
    // Track usage with exact token counts
    err := client.SendUsage("agent-123", "customer-456", "chat-pulse", paygent.UsageData{
        ServiceProvider:  paygent.OpenAI,
        Model:            paygent.GPT4O,
        PromptTokens:     150,
        CompletionTokens: 300,
        TotalTokens:      450,
    })
    
    if err != nil {
        log.Fatalf("Failed to send usage: %v", err)
    }
}
```

### Automatic Token Counting

```go
    // Track usage with raw strings - Paygent handles tokenization automatically
    err := client.SendUsageWithTokenString("agent-123", "customer-456", "qa-pulse", paygent.UsageDataWithStrings{
        ServiceProvider: paygent.Anthropic,
        Model:           paygent.Sonnet37,
        PromptString:    "How does server-side pricing work?",
        OutputString:    "It centralizes pricing logic for better consistency.",
    })
```

### Voice Session Tracking

Track the entire lifecycle of a voice interaction:

```go
    // 1. Initialize session
    client.InitializeVoiceSession("session-abc", "agent-123", "customer-456")

    // 2. Track STT Pulse
    client.TrackVoiceSTT("session-abc", map[string]interface{}{
        "serviceProvider": paygent.Deepgram,
        "model":           paygent.DeepgramNova2,
        "audioDuration":   5.2,
    })

    // 3. Track LLM Pulse
    client.TrackVoiceLLM("session-abc", map[string]interface{}{
        "serviceProvider": paygent.OpenAI,
        "model":           paygent.GPT4O,
        "promptTokens":    100,
        "completionTokens": 200,
    })

    // 4. Track TTS Pulse
    client.TrackVoiceTTS("session-abc", map[string]interface{}{
        "serviceProvider": paygent.DeepgramTTS,
        "model":           paygent.DeepgramAura2,
        "characterCount":  150,
    })

    // 5. Set final indicator (revenue/outcome)
    client.SetVoiceIndicator("session-abc", "successful-call", 30.0)
```

### Customer Management

```go
    customer, err := client.CreateOrGetCustomer(paygent.CustomerCreateOrGetRequest{
        ExternalID: "ext_user_123",
        Name:       "John Doe",
        Email:      "john@example.com",
    })
```

## API Reference

### Client Methods

- `NewClient(apiKey string)`: Creates a new client.
- `SendUsageV2(agentID, customerID, indicator, usageData)`: Directly call the V2 API.
- `CreateOrGetCustomer(request)`: Manage customers.
- `InitializeVoiceSession(sessionID, agentID, customerID)`: Start a voice session.
- `TrackVoiceSTT/LLM/TTS(...)`: Pulse tracking for voice.

## License

MIT


