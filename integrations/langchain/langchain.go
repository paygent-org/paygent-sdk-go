package langchain

import (
	"context"
	"strings"

	"github.com/paygent-org/paygent-sdk-go"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// PaygentLangChainCallback implements LangChain callback handler for automatic usage tracking
type PaygentLangChainCallback struct {
	paygentClient      *paygent.Client
	indicator          string
	externalAgentID    string
	externalCustomerID string
	runInfo            map[string]*runInfo
	lastMessages       []llms.MessageContent // Store last messages for fallback
}

type runInfo struct {
	provider  string
	modelName string
}

// NewPaygentLangChainCallback creates a new LangChain callback handler
func NewPaygentLangChainCallback(
	paygentClient *paygent.Client,
	indicator string,
	externalAgentID string,
	externalCustomerID string,
) *PaygentLangChainCallback {
	return &PaygentLangChainCallback{
		paygentClient:      paygentClient,
		indicator:          indicator,
		externalAgentID:    externalAgentID,
		externalCustomerID: externalCustomerID,
		runInfo:            make(map[string]*runInfo),
		lastMessages:       []llms.MessageContent{},
	}
}

// extractProvider extracts the service provider from model type
func (c *PaygentLangChainCallback) extractProvider(modelType string) string {
	switch modelType {
	case "openai", "openai-chat":
		return paygent.OpenAI
	case "anthropic":
		return paygent.Anthropic
	case "mistral":
		return paygent.MistralAI
	case "google", "gemini":
		return paygent.GoogleDeepMind
	case "cohere":
		return paygent.Cohere
	default:
		return "unknown"
	}
}

// HandleLLMStart is called when an LLM starts running
func (c *PaygentLangChainCallback) HandleLLMStart(ctx context.Context, prompts []string) {
	// LangChain Go doesn't provide detailed metadata in callbacks
	// We'll extract usage in HandleLLMGenerateContentEnd
}

// HandleLLMGenerateContentStart is called when content generation starts
func (c *PaygentLangChainCallback) HandleLLMGenerateContentStart(ctx context.Context, ms []llms.MessageContent) {
	// Store messages for potential fallback tracking
	c.lastMessages = ms
}

// HandleLLMGenerateContentEnd is called when content generation ends
func (c *PaygentLangChainCallback) HandleLLMGenerateContentEnd(ctx context.Context, res *llms.ContentResponse) {
	if res == nil || len(res.Choices) == 0 {
		return
	}

	// Extract model info from generation info
	provider := "unknown"
	modelName := "unknown"
	
	// Try to get model name from first choice's generation info
	if res.Choices[0].GenerationInfo != nil {
		if model, ok := res.Choices[0].GenerationInfo["model"].(string); ok {
			modelName = model
		}
		if model, ok := res.Choices[0].GenerationInfo["model_name"].(string); ok {
			modelName = model
		}
	}

	// Try to determine provider from model name
	if modelName != "unknown" {
		switch {
		case strings.Contains(strings.ToLower(modelName), "gpt"):
			provider = paygent.OpenAI
		case strings.Contains(strings.ToLower(modelName), "claude"):
			provider = paygent.Anthropic
		case strings.Contains(strings.ToLower(modelName), "mistral"):
			provider = paygent.MistralAI
		case strings.Contains(strings.ToLower(modelName), "gemini"):
			provider = paygent.GoogleDeepMind
		}
	}

	// Extract usage tokens from generation info
	promptTokens := 0
	completionTokens := 0
	totalTokens := 0

	if res.Choices[0].GenerationInfo != nil {
		// Try different usage field names
		if usage, ok := res.Choices[0].GenerationInfo["usage"].(map[string]interface{}); ok {
			if pt, ok := usage["prompt_tokens"].(int); ok {
				promptTokens = pt
			} else if pt, ok := usage["input_tokens"].(int); ok {
				promptTokens = pt
			}
			if ct, ok := usage["completion_tokens"].(int); ok {
				completionTokens = ct
			} else if ct, ok := usage["output_tokens"].(int); ok {
				completionTokens = ct
			}
			if tt, ok := usage["total_tokens"].(int); ok {
				totalTokens = tt
			} else {
				totalTokens = promptTokens + completionTokens
			}
		}
		
		// Try token_usage field
		if tokenUsage, ok := res.Choices[0].GenerationInfo["token_usage"].(map[string]interface{}); ok {
			if pt, ok := tokenUsage["prompt_tokens"].(int); ok {
				promptTokens = pt
			}
			if ct, ok := tokenUsage["completion_tokens"].(int); ok {
				completionTokens = ct
			}
			if tt, ok := tokenUsage["total_tokens"].(int); ok {
				totalTokens = tt
			}
		}
	}

	// Send usage data if we have token information
	if totalTokens > 0 {
		usageData := paygent.UsageData{
			ServiceProvider:  provider,
			Model:            modelName,
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      totalTokens,
		}

		err := c.paygentClient.SendUsage(
			c.externalAgentID,
			c.externalCustomerID,
			c.indicator,
			usageData,
		)
		if err != nil {
			c.paygentClient.GetLogger().Errorf("Failed to send usage data: %v", err)
		}
	} else {
		// Fallback: use string-based tracking only if we have messages
		if len(c.lastMessages) == 0 {
			// No usage data and no messages - skip tracking
			c.paygentClient.GetLogger().Warnf("No usage data available for LangChain call, skipping tracking")
			return
		}
		
		// Build prompt string from messages
		promptString := ""
		for _, msg := range c.lastMessages {
			for _, part := range msg.Parts {
				if textPart, ok := part.(llms.TextContent); ok {
					promptString += textPart.Text + " "
				}
			}
		}
		
		outputString := ""
		// Extract text from choices
		for _, choice := range res.Choices {
			if choice.Content != "" {
				outputString += choice.Content
			}
		}

		usageDataWithStrings := paygent.UsageDataWithStrings{
			ServiceProvider: provider,
			Model:           modelName,
			PromptString:    promptString,
			OutputString:    outputString,
		}

		err := c.paygentClient.SendUsageWithTokenString(
			c.externalAgentID,
			c.externalCustomerID,
			c.indicator,
			usageDataWithStrings,
		)
		if err != nil {
			c.paygentClient.GetLogger().Errorf("Failed to send usage data with token string: %v", err)
		}
	}
}

// HandleLLMError is called when an LLM encounters an error
func (c *PaygentLangChainCallback) HandleLLMError(ctx context.Context, err error) {
	c.paygentClient.GetLogger().Errorf("LLM error: %v", err)
}

// HandleChainStart is called when a chain starts
func (c *PaygentLangChainCallback) HandleChainStart(ctx context.Context, inputs map[string]any) {
	// Not used for LLM usage tracking
}

// HandleChainEnd is called when a chain ends
func (c *PaygentLangChainCallback) HandleChainEnd(ctx context.Context, outputs map[string]any) {
	// Not used for LLM usage tracking
}

// HandleChainError is called when a chain encounters an error
func (c *PaygentLangChainCallback) HandleChainError(ctx context.Context, err error) {
	// Not used for LLM usage tracking
}

// HandleToolStart is called when a tool starts
func (c *PaygentLangChainCallback) HandleToolStart(ctx context.Context, input string) {
	// Not used for LLM usage tracking
}

// HandleToolEnd is called when a tool ends
func (c *PaygentLangChainCallback) HandleToolEnd(ctx context.Context, output string) {
	// Not used for LLM usage tracking
}

// HandleToolError is called when a tool encounters an error
func (c *PaygentLangChainCallback) HandleToolError(ctx context.Context, err error) {
	// Not used for LLM usage tracking
}

// HandleText is called when text is generated
func (c *PaygentLangChainCallback) HandleText(ctx context.Context, text string) {
	// Not used for LLM usage tracking
}

// HandleAgentAction is called when an agent takes an action
func (c *PaygentLangChainCallback) HandleAgentAction(ctx context.Context, action schema.AgentAction) {
	// Not used for LLM usage tracking
}

// HandleAgentFinish is called when an agent finishes
func (c *PaygentLangChainCallback) HandleAgentFinish(ctx context.Context, finish schema.AgentFinish) {
	// Not used for LLM usage tracking
}

// HandleRetrieverStart is called when a retriever starts
func (c *PaygentLangChainCallback) HandleRetrieverStart(ctx context.Context, query string) {
	// Not used for LLM usage tracking
}

// HandleRetrieverEnd is called when a retriever ends
func (c *PaygentLangChainCallback) HandleRetrieverEnd(ctx context.Context, query string, documents []schema.Document) {
	// Not used for LLM usage tracking
}

// HandleStreamingFunc is called for streaming responses
func (c *PaygentLangChainCallback) HandleStreamingFunc(ctx context.Context, chunk []byte) {
	// Not used for LLM usage tracking
}

// Ensure PaygentLangChainCallback implements callbacks.Handler
var _ callbacks.Handler = (*PaygentLangChainCallback)(nil)
