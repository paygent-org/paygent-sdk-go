package anthropic

import (
	"context"
	"encoding/json"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/paygent-org/paygent-sdk-go"
)

// PaygentAnthropic wraps the Anthropic client with automatic usage tracking
type PaygentAnthropic struct {
	client        *anthropic.Client
	paygentClient *paygent.Client
}

// MessageParams contains parameters for message creation with Paygent tracking
type MessageParams struct {
	Model                string
	Messages             []anthropic.MessageParam
	MaxTokens            int
	Indicator            string
	ExternalAgentID      string
	ExternalCustomerID   string
	System               interface{}
	Temperature          *float64
	TopP                 *float64
	TopK                 *int
	StopSequences        []string
	Metadata             *anthropic.MetadataParam
	ToolChoice           anthropic.ToolChoiceUnionParam
	Tools                []anthropic.ToolParam
}

// NewPaygentAnthropic creates a new PaygentAnthropic wrapper
func NewPaygentAnthropic(client *anthropic.Client, paygentClient *paygent.Client) *PaygentAnthropic {
	return &PaygentAnthropic{
		client:        client,
		paygentClient: paygentClient,
	}
}

// CreateMessage creates a message with automatic usage tracking
func (p *PaygentAnthropic) CreateMessage(ctx context.Context, params MessageParams) (*anthropic.Message, error) {
	// Build Anthropic request
	reqParams := anthropic.MessageNewParams{
		Model:     anthropic.Model(params.Model),
		Messages:  params.Messages,
		MaxTokens: int64(params.MaxTokens),
	}

	// Add optional parameters
	if params.Temperature != nil {
		reqParams.Temperature = anthropic.Float(*params.Temperature)
	}
	if params.TopP != nil {
		reqParams.TopP = anthropic.Float(*params.TopP)
	}
	if len(params.StopSequences) > 0 {
		reqParams.StopSequences = params.StopSequences
	}

	// Make Anthropic API call
	resp, err := p.client.Messages.New(ctx, reqParams)
	if err != nil {
		return resp, err
	}

	// Extract usage data from response with robust fallback mechanism
	hasValidUsage := resp.Usage.InputTokens > 0 && resp.Usage.OutputTokens > 0

	if hasValidUsage {
		// Primary path: Use usage data from API response
		usageData := paygent.UsageData{
			ServiceProvider:  paygent.Anthropic,
			Model:            params.Model,
			PromptTokens:     int(resp.Usage.InputTokens),
			CompletionTokens: int(resp.Usage.OutputTokens),
			TotalTokens:      int(resp.Usage.InputTokens + resp.Usage.OutputTokens),
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
		if len(resp.Content) > 0 {
			// Handle the content block properly using the new structure
			content := resp.Content[0]
			if content.Type == "text" {
				outputString = content.Text
			} else if content.Type == "thinking" {
				outputString = content.Thinking
			}
		}

		usageDataWithStrings := paygent.UsageDataWithStrings{
			ServiceProvider: paygent.Anthropic,
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
