package mistral

import (
	"context"
	"encoding/json"

	"github.com/gage-technologies/mistral-go"
	"github.com/paygent-org/paygent-sdk-go"
)

// PaygentMistral wraps the Mistral client with automatic usage tracking
type PaygentMistral struct {
	client        *mistral.MistralClient
	paygentClient *paygent.Client
}

// MistralChatParams contains parameters for chat completion with Paygent tracking
type MistralChatParams struct {
	Model              string
	Messages           []mistral.ChatMessage
	Indicator          string
	ExternalAgentID    string
	ExternalCustomerID string
	Temperature        *float64
	TopP               *float64
	MaxTokens          *int
}

// NewPaygentMistral creates a new PaygentMistral wrapper
func NewPaygentMistral(client *mistral.MistralClient, paygentClient *paygent.Client) *PaygentMistral {
	return &PaygentMistral{
		client:        client,
		paygentClient: paygentClient,
	}
}

// ChatComplete creates a chat completion with automatic usage tracking
func (p *PaygentMistral) ChatComplete(ctx context.Context, params MistralChatParams) (*mistral.ChatCompletionResponse, error) {
	// Build optional parameters
	chatParams := &mistral.ChatRequestParams{}
	if params.Temperature != nil {
		chatParams.Temperature = *params.Temperature
	}
	if params.TopP != nil {
		chatParams.TopP = *params.TopP
	}
	if params.MaxTokens != nil {
		chatParams.MaxTokens = *params.MaxTokens
	}

	// Make Mistral API call
	resp, err := p.client.Chat(params.Model, params.Messages, chatParams)
	if err != nil {
		return resp, err
	}

	// Extract usage data from response with robust fallback mechanism
	hasValidUsage := resp.Usage.PromptTokens > 0 && resp.Usage.CompletionTokens > 0

	if hasValidUsage {
		// Primary path: Use usage data from API response
		usageData := paygent.UsageData{
			ServiceProvider:  paygent.MistralAI,
			Model:            params.Model,
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}

		err = p.paygentClient.SendUsage(
			params.ExternalAgentID,
			params.ExternalCustomerID,
			params.Indicator,
			usageData,
		)
		if err != nil {
			p.paygentClient.GetLogger().Errorf("Failed to send usage data: %v", err)
		}
	} else {
		// Fallback path: Calculate tokens from actual strings
		promptString, _ := json.Marshal(params.Messages)
		outputString := ""
		if len(resp.Choices) > 0 {
			outputString = resp.Choices[0].Message.Content
		}

		usageDataWithStrings := paygent.UsageDataWithStrings{
			ServiceProvider: paygent.MistralAI,
			Model:           params.Model,
			PromptString:    string(promptString),
			OutputString:    outputString,
		}

		err = p.paygentClient.SendUsageWithTokenString(
			params.ExternalAgentID,
			params.ExternalCustomerID,
			params.Indicator,
			usageDataWithStrings,
		)
		if err != nil {
			p.paygentClient.GetLogger().Errorf("Failed to send usage data with token string: %v", err)
		}
	}

	return resp, nil
}
