package main

import (
	"log"

	"github.com/paylm/paylm-sdk-go"
	"github.com/sirupsen/logrus"
)

func main() {
	// Create a new client with custom base URL for localhost testing
	client := paylm.NewClientWithURL("pk_e0ea0d11bb7f0d174caf578d665454acff97bdb1f85c235af547ccd9a733ef35", "http://localhost:8080")

	// Set log level to see detailed logs
	client.SetLogLevel(logrus.InfoLevel)

	// Example 1: Basic usage with Llama model
	log.Println("=== Example 1: Llama Model ===")
	usageData1 := paylm.UsageData{
		ServiceProvider:  "Meta",
		Model:            "llama-3-8b",
		PromptTokens:     756,
		CompletionTokens: 244,
		TotalTokens:      1000,
	}

	err := client.SendUsage("agent-123", "customer-456", "email-sent", usageData1)
	if err != nil {
		log.Printf("Failed to send usage: %v", err)
	} else {
		log.Println("Usage data sent successfully!")
	}

	// Example 2: GPT-4 model
	log.Println("\n=== Example 2: GPT-4 Model ===")
	usageData2 := paylm.UsageData{
		ServiceProvider:  "OpenAI",
		Model:            "gpt-4",
		PromptTokens:     1000,
		CompletionTokens: 500,
		TotalTokens:      1500,
	}

	err = client.SendUsage("agent-789", "customer-101", "chat-completion", usageData2)
	if err != nil {
		log.Printf("Failed to send usage: %v", err)
	} else {
		log.Println("Usage data sent successfully!")
	}

	// Example 3: Unknown model (will use default pricing)
	log.Println("\n=== Example 3: Unknown Model ===")
	usageData3 := paylm.UsageData{
		ServiceProvider:  "Custom",
		Model:            "custom-model",
		PromptTokens:     200,
		CompletionTokens: 100,
		TotalTokens:      300,
	}

	err = client.SendUsage("agent-999", "customer-888", "custom-task", usageData3)
	if err != nil {
		log.Printf("Failed to send usage: %v", err)
	} else {
		log.Println("Usage data sent successfully!")
	}

	// Example 4: Using custom logger
	log.Println("\n=== Example 4: Custom Logging ===")
	logger := client.GetLogger()
	logger.Info("This is a custom log message from the application")

	// Example 5: SendUsageWithTokenString - using prompt and output strings
	log.Println("\n=== Example 5: SendUsageWithTokenString ===")
	usageDataWithStrings := paylm.UsageDataWithStrings{
		ServiceProvider: "OpenAI",
		Model:           "gpt-4",
		PromptString:    "What is the capital of France? Please provide a detailed explanation.",
		OutputString:    "The capital of France is Paris. Paris is located in the north-central part of France and is the country's largest city and economic center.",
	}

	err = client.SendUsageWithTokenString("agent-555", "customer-777", "question-answer", usageDataWithStrings)
	if err != nil {
		log.Printf("Failed to send usage with strings: %v", err)
	} else {
		log.Println("Usage data with strings sent successfully!")
	}

	// Example 6: Different model providers
	log.Println("\n=== Example 6: Different Model Providers ===")

	// Anthropic Claude
	claudeUsage := paylm.UsageDataWithStrings{
		ServiceProvider: "Anthropic",
		Model:           "claude-3-sonnet",
		PromptString:    "Explain quantum computing in simple terms.",
		OutputString:    "Quantum computing is a revolutionary approach to computation that leverages the principles of quantum mechanics to process information in ways that classical computers cannot.",
	}

	err = client.SendUsageWithTokenString("agent-888", "customer-999", "explanation", claudeUsage)
	if err != nil {
		log.Printf("Failed to send Claude usage: %v", err)
	} else {
		log.Println("Claude usage data sent successfully!")
	}

	// GPT-5 example
	log.Println("\n=== GPT-5 Example ===")
	gpt5Usage := paylm.UsageDataWithStrings{
		ServiceProvider: "OpenAI",
		Model:           "gpt-5",
		PromptString:    "Analyze the following complex business scenario and provide strategic recommendations.",
		OutputString:    "Based on the analysis, I recommend focusing on three key strategic areas: market expansion, operational efficiency, and customer retention.",
	}

	err = client.SendUsageWithTokenString("agent-101", "customer-202", "business-analysis", gpt5Usage)
	if err != nil {
		log.Printf("Failed to send GPT-5 usage: %v", err)
	} else {
		log.Println("GPT-5 usage data sent successfully!")
	}

	// Example 7: Error handling
	log.Println("\n=== Example 7: Error Handling ===")
	// This will fail because we're using a dummy API key
	clientWithInvalidKey := paylm.NewClient("invalid-api-key")
	err = clientWithInvalidKey.SendUsage("agent-123", "customer-456", "test", usageData1)
	if err != nil {
		log.Printf("Expected error with invalid API key: %v", err)
	}
}
