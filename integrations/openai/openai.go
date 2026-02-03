package openai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/paygent-org/paygent-sdk-go"
	"github.com/sashabaranov/go-openai"
)

// PaygentOpenAI wraps the OpenAI client with automatic usage tracking
type PaygentOpenAI struct {
	client        *openai.Client
	paygentClient *paygent.Client
}

// ChatCompletionParams contains parameters for chat completion with Paygent tracking
type ChatCompletionParams struct {
	Model                string
	Messages             []openai.ChatCompletionMessage
	Indicator            string
	ExternalAgentID      string
	ExternalCustomerID   string
	MaxTokens            int
	Temperature          float32
	TopP                 float32
	N                    int
	Stop                 []string
	PresencePenalty      float32
	FrequencyPenalty     float32
	LogitBias            map[string]int
	User                 string
	ResponseFormat       *openai.ChatCompletionResponseFormat
	Seed                 *int
	Tools                []openai.Tool
	ToolChoice           interface{}
	ParallelToolCalls    interface{}
	Functions            []openai.FunctionDefinition
	FunctionCall         interface{}
}

// NewPaygentOpenAI creates a new PaygentOpenAI wrapper
func NewPaygentOpenAI(client *openai.Client, paygentClient *paygent.Client) *PaygentOpenAI {
	return &PaygentOpenAI{
		client:        client,
		paygentClient: paygentClient,
	}
}

// CreateChatCompletion creates a chat completion with automatic usage tracking
func (p *PaygentOpenAI) CreateChatCompletion(ctx context.Context, params ChatCompletionParams) (openai.ChatCompletionResponse, error) {
	// Build OpenAI request
	req := openai.ChatCompletionRequest{
		Model:    params.Model,
		Messages: params.Messages,
	}

	// Add optional parameters
	if params.MaxTokens > 0 {
		req.MaxTokens = params.MaxTokens
	}
	if params.Temperature != 0 {
		req.Temperature = params.Temperature
	}
	if params.TopP != 0 {
		req.TopP = params.TopP
	}
	if params.N > 0 {
		req.N = params.N
	}
	if len(params.Stop) > 0 {
		req.Stop = params.Stop
	}
	if params.PresencePenalty != 0 {
		req.PresencePenalty = params.PresencePenalty
	}
	if params.FrequencyPenalty != 0 {
		req.FrequencyPenalty = params.FrequencyPenalty
	}
	if params.LogitBias != nil {
		req.LogitBias = params.LogitBias
	}
	if params.User != "" {
		req.User = params.User
	}
	if params.ResponseFormat != nil {
		req.ResponseFormat = params.ResponseFormat
	}
	if params.Seed != nil {
		req.Seed = params.Seed
	}
	if len(params.Tools) > 0 {
		req.Tools = params.Tools
	}
	if params.ToolChoice != nil {
		req.ToolChoice = params.ToolChoice
	}
	if params.ParallelToolCalls != nil {
		req.ParallelToolCalls = params.ParallelToolCalls
	}
	if len(params.Functions) > 0 {
		req.Functions = params.Functions
	}
	if params.FunctionCall != nil {
		req.FunctionCall = params.FunctionCall
	}

	// Make OpenAI API call
	resp, err := p.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return resp, err
	}

	// Extract usage data from response with robust fallback mechanism
	hasValidUsage := resp.Usage.PromptTokens > 0 && resp.Usage.CompletionTokens > 0

	if hasValidUsage {
		// Primary path: Use usage data from API response
		usageData := paygent.UsageData{
			ServiceProvider:  paygent.OpenAI,
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
			// Log error but don't fail the request
			p.paygentClient.GetLogger().Errorf("Failed to send usage data: %v", err)
		}
	} else {
		// Fallback path: Calculate tokens from actual strings
		// This ensures we never lose billing data even if API response format changes
		promptString, _ := json.Marshal(params.Messages)
		outputString := ""
		if len(resp.Choices) > 0 {
			outputString = resp.Choices[0].Message.Content
		}

		usageDataWithStrings := paygent.UsageDataWithStrings{
			ServiceProvider: paygent.OpenAI,
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
			// Log error but don't fail the request
			p.paygentClient.GetLogger().Errorf("Failed to send usage data with token string: %v", err)
		}
	}

	return resp, nil
}

// CreateEmbedding creates embeddings with automatic usage tracking
func (p *PaygentOpenAI) CreateEmbedding(ctx context.Context, model string, input interface{}, indicator, externalAgentID, externalCustomerID string) (openai.EmbeddingResponse, error) {
	// Build OpenAI request
	req := openai.EmbeddingRequest{
		Model: openai.EmbeddingModel(model),
	}

	// Handle different input types
	switch v := input.(type) {
	case string:
		req.Input = v
	case []string:
		req.Input = v
	default:
		return openai.EmbeddingResponse{}, fmt.Errorf("unsupported input type: %T", input)
	}

	// Make OpenAI API call
	resp, err := p.client.CreateEmbeddings(ctx, req)
	if err != nil {
		return resp, err
	}

	// Extract usage data from response with robust fallback mechanism
	hasValidUsage := resp.Usage.PromptTokens > 0

	if hasValidUsage {
		// Primary path: Use usage data from API response
		usageData := paygent.UsageData{
			ServiceProvider:  paygent.OpenAI,
			Model:            model,
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.PromptTokens, // Embeddings don't have completion tokens
			TotalTokens:      resp.Usage.TotalTokens,
		}

		err = p.paygentClient.SendUsage(
			externalAgentID,
			externalCustomerID,
			indicator,
			usageData,
		)
		if err != nil {
			p.paygentClient.GetLogger().Errorf("Failed to send usage data: %v", err)
		}
	} else {
		// Fallback path: Calculate tokens from input text
		inputText := ""
		switch v := input.(type) {
		case string:
			inputText = v
		case []string:
			for _, s := range v {
				inputText += s + " "
			}
		}

		usageDataWithStrings := paygent.UsageDataWithStrings{
			ServiceProvider: paygent.OpenAI,
			Model:           model,
			PromptString:    inputText,
			OutputString:    "", // Embeddings don't have output
		}

		err = p.paygentClient.SendUsageWithTokenString(
			externalAgentID,
			externalCustomerID,
			indicator,
			usageDataWithStrings,
		)
		if err != nil {
			p.paygentClient.GetLogger().Errorf("Failed to send usage data with token string: %v", err)
		}
	}

	return resp, nil
}
