package gemini

import (
	"context"
	"encoding/json"

	"github.com/google/generative-ai-go/genai"
	"github.com/paygent-org/paygent-sdk-go"
)

// PaygentGemini wraps the Gemini client with automatic usage tracking
type PaygentGemini struct {
	client        *genai.Client
	paygentClient *paygent.Client
}

// GeminiGenerateParams contains parameters for content generation with Paygent tracking
type GeminiGenerateParams struct {
	Model              string
	Contents           []*genai.Content
	Indicator          string
	ExternalAgentID    string
	ExternalCustomerID string
	Temperature        *float32
	TopP               *float32
	TopK               *int32
	MaxOutputTokens    *int32
	StopSequences      []string
	SafetySettings     []*genai.SafetySetting
	Tools              []*genai.Tool
	ToolConfig         *genai.ToolConfig
	SystemInstruction  *genai.Content
}

// NewPaygentGemini creates a new PaygentGemini wrapper
func NewPaygentGemini(client *genai.Client, paygentClient *paygent.Client) *PaygentGemini {
	return &PaygentGemini{
		client:        client,
		paygentClient: paygentClient,
	}
}

// GenerateContent generates content with automatic usage tracking
func (p *PaygentGemini) GenerateContent(ctx context.Context, params GeminiGenerateParams) (*genai.GenerateContentResponse, error) {
	// Get the model
	model := p.client.GenerativeModel(params.Model)

	// Configure generation parameters
	if params.Temperature != nil {
		model.Temperature = params.Temperature
	}
	if params.TopP != nil {
		model.TopP = params.TopP
	}
	if params.TopK != nil {
		model.TopK = params.TopK
	}
	if params.MaxOutputTokens != nil {
		model.MaxOutputTokens = params.MaxOutputTokens
	}
	if len(params.StopSequences) > 0 {
		model.StopSequences = params.StopSequences
	}
	if len(params.SafetySettings) > 0 {
		model.SafetySettings = params.SafetySettings
	}
	if len(params.Tools) > 0 {
		model.Tools = params.Tools
	}
	if params.ToolConfig != nil {
		model.ToolConfig = params.ToolConfig
	}
	if params.SystemInstruction != nil {
		model.SystemInstruction = params.SystemInstruction
	}

	// Make Gemini API call - convert Contents to Parts
	var parts []genai.Part
	for _, content := range params.Contents {
		if content != nil && len(content.Parts) > 0 {
			parts = append(parts, content.Parts...)
		}
	}
	
	resp, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		return resp, err
	}

	// Extract usage data from response with robust fallback mechanism
	hasValidUsage := resp.UsageMetadata != nil &&
		(resp.UsageMetadata.PromptTokenCount > 0 || resp.UsageMetadata.CandidatesTokenCount > 0)

	if hasValidUsage {
		// Primary path: Use usage metadata from API response
		usageData := paygent.UsageData{
			ServiceProvider:  params.Model,
			Model:            params.Model,
			PromptTokens:     int(resp.UsageMetadata.PromptTokenCount),
			CompletionTokens: int(resp.UsageMetadata.CandidatesTokenCount),
			TotalTokens:      int(resp.UsageMetadata.TotalTokenCount),
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
		promptString, _ := json.Marshal(params.Contents)
		outputString := ""
		if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
			if len(resp.Candidates[0].Content.Parts) > 0 {
				if textPart, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
					outputString = string(textPart)
				}
			}
		}

		usageDataWithStrings := paygent.UsageDataWithStrings{
			ServiceProvider: params.Model,
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

// StartChat starts a chat session with automatic usage tracking
func (p *PaygentGemini) StartChat(model string, indicator, externalAgentID, externalCustomerID string, history []*genai.Content) *PaygentGeminiChat {
	geminiModel := p.client.GenerativeModel(model)
	chatSession := geminiModel.StartChat()
	chatSession.History = history

	return &PaygentGeminiChat{
		session:            chatSession,
		paygentClient:      p.paygentClient,
		model:              model,
		indicator:          indicator,
		externalAgentID:    externalAgentID,
		externalCustomerID: externalCustomerID,
	}
}

// PaygentGeminiChat wraps a Gemini chat session with automatic usage tracking
type PaygentGeminiChat struct {
	session            *genai.ChatSession
	paygentClient      *paygent.Client
	model              string
	indicator          string
	externalAgentID    string
	externalCustomerID string
}

// SendMessage sends a message in the chat with automatic usage tracking
func (c *PaygentGeminiChat) SendMessage(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
	// Make Gemini API call
	resp, err := c.session.SendMessage(ctx, parts...)
	if err != nil {
		return resp, err
	}

	// Extract usage data from response with robust fallback mechanism
	hasValidUsage := resp.UsageMetadata != nil &&
		(resp.UsageMetadata.PromptTokenCount > 0 || resp.UsageMetadata.CandidatesTokenCount > 0)

	if hasValidUsage {
		// Primary path: Use usage metadata from API response
		usageData := paygent.UsageData{
			ServiceProvider:  c.model,
			Model:            c.model,
			PromptTokens:     int(resp.UsageMetadata.PromptTokenCount),
			CompletionTokens: int(resp.UsageMetadata.CandidatesTokenCount),
			TotalTokens:      int(resp.UsageMetadata.TotalTokenCount),
		}

		err = c.paygentClient.SendUsage(
			c.externalAgentID,
			c.externalCustomerID,
			c.indicator,
			usageData,
		)
		if err != nil {
			c.paygentClient.GetLogger().Errorf("Failed to send usage data: %v", err)
		}
	} else {
		// Fallback path: Calculate tokens from message and response
		promptString, _ := json.Marshal(parts)
		outputString := ""
		if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
			if len(resp.Candidates[0].Content.Parts) > 0 {
				if textPart, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
					outputString = string(textPart)
				}
			}
		}

		usageDataWithStrings := paygent.UsageDataWithStrings{
			ServiceProvider: c.model,
			Model:           c.model,
			PromptString:    string(promptString),
			OutputString:    outputString,
		}

		err = c.paygentClient.SendUsageWithTokenString(
			c.externalAgentID,
			c.externalCustomerID,
			c.indicator,
			usageDataWithStrings,
		)
		if err != nil {
			c.paygentClient.GetLogger().Errorf("Failed to send usage data with token string: %v", err)
		}
	}

	return resp, nil
}

// GetHistory returns the chat history
func (c *PaygentGeminiChat) GetHistory() []*genai.Content {
	return c.session.History
}
