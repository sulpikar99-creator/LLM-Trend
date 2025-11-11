package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the application configuration
type Config struct {
	AIModels              []AIModel      `json:"ai_models"`
	Exchanges             []Exchange     `json:"exchanges"`
	DefaultStrategyPrompt string         `json:"default_strategy_prompt"`
	RiskLimits            RiskLimits     `json:"risk_limits"`
}

// AIModel represents an AI model configuration
type AIModel struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Provider         string  `json:"provider"`
	APIKeyEncrypted  string  `json:"api_key_encrypted"`
	BaseURL          string  `json:"base_url"`
	MaxTokens        int     `json:"max_tokens"`
	Temperature      float64 `json:"temperature"`
}

// Exchange represents an exchange configuration
type Exchange struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Type              string `json:"type"`
	APIKeyEncrypted   string `json:"api_key_encrypted"`
	APISecretEncrypted string `json:"api_secret_encrypted"`
	Testnet           bool   `json:"testnet"`
}

// RiskLimits represents risk management limits
type RiskLimits struct {
	MaxDrawdownPercent  float64 `json:"max_drawdown_percent"`
	MaxPositionSizeUSD  float64 `json:"max_position_size_usd"`
	MaxLeverage         float64 `json:"max_leverage"`
	DailyLossLimitUSD   float64 `json:"daily_loss_limit_usd"`
}

// LoadConfig loads configuration from config.json
func LoadConfig() (*Config, error) {
	configPath := "config.json"

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		return createDefaultConfig(configPath)
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves configuration to config.json
func SaveConfig(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile("config.json", data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// createDefaultConfig creates a default configuration file
func createDefaultConfig(path string) (*Config, error) {
	config := &Config{
		AIModels: []AIModel{
			{
				ID:          "deepseek-chat",
				Name:        "DeepSeek Chat",
				Provider:    "deepseek",
				BaseURL:     "https://api.deepseek.com",
				MaxTokens:   4000,
				Temperature: 0.7,
			},
		},
		Exchanges: []Exchange{},
		DefaultStrategyPrompt: `You are an expert cryptocurrency trader with advanced knowledge of technical analysis and market dynamics.

Analyze the current market conditions and make informed trading decisions based on:
- Price action and trends
- Technical indicators (RSI, MACD, Bollinger Bands)
- Support and resistance levels
- Current positions and risk exposure

Respond with your decision in JSON format:
{
  "action": "open_long" | "open_short" | "close_position" | "hold",
  "symbol": "BTCUSDT",
  "size": 0.001,
  "leverage": 5,
  "stop_loss": 40000,
  "take_profit": 45000,
  "reasoning": "Explain your decision..."
}`,
		RiskLimits: RiskLimits{
			MaxDrawdownPercent: 20.0,
			MaxPositionSizeUSD: 10000.0,
			MaxLeverage:        10.0,
			DailyLossLimitUSD:  1000.0,
		},
	}

	// Save default config
	if err := SaveConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}
