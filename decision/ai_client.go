package decision

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// AIProvider represents different AI providers
type AIProvider string

const (
	ProviderOpenAI    AIProvider = "openai"
	ProviderDeepSeek  AIProvider = "deepseek"
	ProviderClaude    AIProvider = "claude"
	ProviderQwen      AIProvider = "qwen"
)

// AIConfig represents AI model configuration
type AIConfig struct {
	Provider    AIProvider `json:"provider"`
	Model       string     `json:"model"`
	APIKey      string     `json:"api_key"`
	BaseURL     string     `json:"base_url"`
	MaxTokens   int        `json:"max_tokens"`
	Temperature float64    `json:"temperature"`
}

// AIMessage represents a chat message
type AIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AIRequest represents a request to AI provider
type AIRequest struct {
	Model       string      `json:"model"`
	Messages    []AIMessage `json:"messages"`
	MaxTokens   int         `json:"max_tokens,omitempty"`
	Temperature float64     `json:"temperature,omitempty"`
	Stream      bool        `json:"stream,omitempty"`
}

// AIResponse represents a response from AI provider
type AIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		Delta struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// AIClient handles communication with AI providers
type AIClient struct {
	config     *AIConfig
	httpClient *http.Client
}

// GetDefaultAIConfig returns default AI configuration from environment variables
func GetDefaultAIConfig() *AIConfig {
	// Try to get AI provider from environment
	provider := os.Getenv("AI_PROVIDER")
	if provider == "" {
		provider = "deepseek" // Default to DeepSeek
	}

	// Get API key
	apiKey := os.Getenv("AI_API_KEY")
	if apiKey == "" {
		// Try provider-specific keys
		switch provider {
		case "openai":
			apiKey = os.Getenv("OPENAI_API_KEY")
		case "deepseek":
			apiKey = os.Getenv("DEEPSEEK_API_KEY")
		case "claude":
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
		case "qwen":
			apiKey = os.Getenv("QWEN_API_KEY")
		}
	}

	// If still no API key, return nil
	if apiKey == "" {
		return nil
	}

	// Get model
	model := os.Getenv("AI_MODEL")
	if model == "" {
		// Provider-specific defaults
		switch provider {
		case "openai":
			model = "gpt-4o-mini"
		case "deepseek":
			model = "deepseek-chat"
		case "claude":
			model = "claude-3-5-sonnet-20241022"
		case "qwen":
			model = "qwen-turbo"
		default:
			model = "deepseek-chat"
		}
	}

	// Get base URL
	baseURL := os.Getenv("AI_BASE_URL")
	if baseURL == "" {
		// Provider-specific defaults
		switch provider {
		case "openai":
			baseURL = "https://api.openai.com/v1"
		case "deepseek":
			baseURL = "https://api.deepseek.com/v1"
		case "claude":
			baseURL = "https://api.anthropic.com/v1"
		case "qwen":
			baseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
		default:
			baseURL = "https://api.deepseek.com/v1"
		}
	}

	// Get max tokens
	maxTokens := 4000
	if maxTokensStr := os.Getenv("AI_MAX_TOKENS"); maxTokensStr != "" {
		if val, err := strconv.Atoi(maxTokensStr); err == nil {
			maxTokens = val
		}
	}

	// Get temperature
	temperature := 0.7
	if tempStr := os.Getenv("AI_TEMPERATURE"); tempStr != "" {
		if val, err := strconv.ParseFloat(tempStr, 64); err == nil {
			temperature = val
		}
	}

	return &AIConfig{
		Provider:    AIProvider(provider),
		Model:       model,
		APIKey:      apiKey,
		BaseURL:     baseURL,
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}
}

// NewAIClient creates a new AI client
func NewAIClient(config *AIConfig) *AIClient {
	return &AIClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GetCompletion gets a completion from the AI provider (non-streaming)
func (c *AIClient) GetCompletion(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	messages := []AIMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	req := AIRequest{
		Model:       c.config.Model,
		Messages:    messages,
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
		Stream:      false,
	}

	response, err := c.sendRequest(ctx, req)
	if err != nil {
		return "", err
	}

	// Extract content from response
	if len(response.Choices) > 0 {
		return response.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no response from AI")
}

// GetCompletionStream gets a streaming completion from the AI provider
func (c *AIClient) GetCompletionStream(ctx context.Context, systemPrompt, userPrompt string) (<-chan string, <-chan error) {
	contentChan := make(chan string, 100)
	errorChan := make(chan error, 1)

	go func() {
		defer close(contentChan)
		defer close(errorChan)

		messages := []AIMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		}

		req := AIRequest{
			Model:       c.config.Model,
			Messages:    messages,
			MaxTokens:   c.config.MaxTokens,
			Temperature: c.config.Temperature,
			Stream:      true,
		}

		if err := c.streamRequest(ctx, req, contentChan); err != nil {
			errorChan <- err
		}
	}()

	return contentChan, errorChan
}

// sendRequest sends a non-streaming request
func (c *AIClient) sendRequest(ctx context.Context, req AIRequest) (*AIResponse, error) {
	endpoint := c.getEndpoint()

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var response AIResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// streamRequest sends a streaming request
func (c *AIClient) streamRequest(ctx context.Context, req AIRequest, contentChan chan<- string) error {
	endpoint := c.getEndpoint()

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("AI API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Read streaming response
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if line == "" {
			continue
		}

		// SSE format: "data: {...}"
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			// Check for stream end
			if data == "[DONE]" {
				break
			}

			// Parse JSON
			var streamResp AIResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				continue // Skip malformed chunks
			}

			// Extract delta content
			if len(streamResp.Choices) > 0 {
				delta := streamResp.Choices[0].Delta.Content
				if delta != "" {
					select {
					case contentChan <- delta:
					case <-ctx.Done():
						return ctx.Err()
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	return nil
}

// getEndpoint returns the API endpoint based on provider
func (c *AIClient) getEndpoint() string {
	if c.config.BaseURL != "" {
		return c.config.BaseURL + "/v1/chat/completions"
	}

	switch c.config.Provider {
	case ProviderOpenAI:
		return "https://api.openai.com/v1/chat/completions"
	case ProviderDeepSeek:
		return "https://api.deepseek.com/v1/chat/completions"
	case ProviderClaude:
		return "https://api.anthropic.com/v1/messages"
	case ProviderQwen:
		return "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"
	default:
		return c.config.BaseURL + "/v1/chat/completions"
	}
}

// setHeaders sets request headers based on provider
func (c *AIClient) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")

	switch c.config.Provider {
	case ProviderClaude:
		req.Header.Set("x-api-key", c.config.APIKey)
		req.Header.Set("anthropic-version", "2023-06-01")
	default:
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}
}

// ExtractJSON extracts JSON from AI response text
// Handles cases where AI adds markdown code blocks around JSON
func ExtractJSON(text string) string {
	// Remove markdown code blocks if present
	text = strings.TrimSpace(text)

	// Try to find JSON between ```json and ```
	if strings.Contains(text, "```json") {
		start := strings.Index(text, "```json") + 7
		end := strings.LastIndex(text, "```")
		if start > 0 && end > start {
			return strings.TrimSpace(text[start:end])
		}
	}

	// Try to find JSON between ``` and ```
	if strings.Contains(text, "```") {
		start := strings.Index(text, "```") + 3
		end := strings.LastIndex(text, "```")
		if start > 0 && end > start {
			return strings.TrimSpace(text[start:end])
		}
	}

	// Look for JSON object
	startIdx := strings.Index(text, "{")
	endIdx := strings.LastIndex(text, "}")
	if startIdx >= 0 && endIdx > startIdx {
		return text[startIdx : endIdx+1]
	}

	return text
}
