package paygent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// SttUsageData represents the STT usage data structure
type SttUsageData struct {
	ServiceProvider string  `json:"service_provider"`
	Model           string  `json:"model"`
	AudioDuration   float64 `json:"audio_duration"` // Duration in seconds
	Plan            string  `json:"plan"`
}

// TtsUsageData represents the TTS usage data structure
type TtsUsageData struct {
	ServiceProvider string `json:"service_provider"`
	Model           string `json:"model"`
	CharacterCount  int    `json:"character_count"` // Number of characters
	Plan            string `json:"plan"`
}

// InitializeVoiceSession initializes a voice session
func (c *Client) InitializeVoiceSession(sessionID, agentID, customerID string) error {
	requestData := map[string]interface{}{
		"sessionId":  sessionID,
		"agentId":    agentID,
		"customerId": customerID,
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/voice/session", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("paygent-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("api request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	return nil
}

// TrackVoiceSTT tracks STT usage for a voice session.
// Note: data["audioDuration"] should be in minutes.
func (c *Client) TrackVoiceSTT(sessionID string, data map[string]interface{}) error {
	requestData := map[string]interface{}{
		"sessionId":    sessionID,
		"audioMinutes": data["audioDuration"],
		"provider":     data["serviceProvider"],
		"model":        data["model"],
		"plan":         data["plan"],
		"language":     data["language"],
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/voice/stt", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("paygent-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("api request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	return nil
}

// TrackVoiceLLM tracks LLM usage for a voice session
func (c *Client) TrackVoiceLLM(sessionID string, data map[string]interface{}) error {
	requestData := map[string]interface{}{
		"sessionId":        sessionID,
		"provider":         data["serviceProvider"],
		"model":            data["model"],
		"plan":             data["plan"],
		"promptTokens":     data["promptTokens"],
		"completionTokens": data["completionTokens"],
	}
	// Optional: cached prompt tokens (for models with prompt caching, e.g. OpenAI, Anthropic)
	if cached, ok := data["cachedTokens"]; ok && cached != nil {
		requestData["cachedTokens"] = cached
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/voice/llm", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("paygent-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("api request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	return nil
}

// TrackVoiceTTS tracks TTS usage for a voice session
func (c *Client) TrackVoiceTTS(sessionID string, data map[string]interface{}) error {
	requestData := map[string]interface{}{
		"sessionId":  sessionID,
		"provider":   data["serviceProvider"],
		"model":      data["model"],
		"plan":       data["plan"],
		"characters": data["characterCount"],
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/voice/tts", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("paygent-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("api request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	return nil
}

// SetVoiceIndicator sets the voice indicator and duration for a session.
// Note: callDuration should be in minutes for per-minute billing.
func (c *Client) SetVoiceIndicator(sessionID, indicator string, callDuration float64) error {
	requestData := map[string]interface{}{
		"sessionId":     sessionID,
		"indicator":     indicator,
		"totalDuration": callDuration,
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/voice/indicator", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("paygent-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("api request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	return nil
}

// SendSttUsage sends STT usage data to the Paygent API (Legacy)
func (c *Client) SendSttUsage(agentID, customerID string, sttUsageData SttUsageData) error {
	rawUsage := RawUsageData{
		Provider:      sttUsageData.ServiceProvider,
		Model:         sttUsageData.Model,
		AudioDuration: int(sttUsageData.AudioDuration),
		Plan:          sttUsageData.Plan,
	}

	_, err := c.SendUsageV2(agentID, customerID, "stt-usage", rawUsage)
	return err
}

// SendTtsUsage sends TTS usage data to the Paygent API (Legacy)
func (c *Client) SendTtsUsage(agentID, customerID string, ttsUsageData TtsUsageData) error {
	rawUsage := RawUsageData{
		Provider:       ttsUsageData.ServiceProvider,
		Model:          ttsUsageData.Model,
		CharacterCount: ttsUsageData.CharacterCount,
		Plan:           ttsUsageData.Plan,
	}

	_, err := c.SendUsageV2(agentID, customerID, "tts-usage", rawUsage)
	return err
}
