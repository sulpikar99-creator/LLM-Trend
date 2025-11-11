package config

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// UserConfig represents user-specific configuration
type UserConfig struct {
	UserID                 string                 `json:"user_id"`
	AIProvider             string                 `json:"ai_provider"`
	AIModel                string                 `json:"ai_model"`
	AIBaseURL              string                 `json:"ai_base_url,omitempty"`
	AIAPIKey               string                 `json:"-"` // Never expose in JSON
	DefaultSystemPrompt    string                 `json:"default_system_prompt,omitempty"`
	DefaultStrategyPrompt  string                 `json:"default_strategy_prompt,omitempty"`
	ExchangeConfigs        map[string]interface{} `json:"exchange_configs,omitempty"`
	RiskLimits             map[string]interface{} `json:"risk_limits,omitempty"`
	NotificationSettings   map[string]interface{} `json:"notification_settings,omitempty"`
	CreatedAt              time.Time              `json:"created_at"`
	UpdatedAt              time.Time              `json:"updated_at"`
}

// UserConfigManager manages user configurations
type UserConfigManager struct {
	db *sql.DB
}

// NewUserConfigManager creates a new user config manager
func NewUserConfigManager(db *sql.DB) *UserConfigManager {
	return &UserConfigManager{db: db}
}

// CreateDefaultConfig creates default configuration for a new user
func (ucm *UserConfigManager) CreateDefaultConfig(userID string) error {
	query := `
		INSERT INTO user_configs (
			user_id, ai_provider, ai_model,
			default_system_prompt, default_strategy_prompt
		) VALUES (?, ?, ?, ?, ?)
	`

	defaultSystemPrompt := "You are an expert cryptocurrency trading AI assistant."
	defaultStrategyPrompt := "Analyze the market data and make informed trading decisions based on technical indicators and risk management principles."

	_, err := ucm.db.Exec(query, userID, "openai", "gpt-4", defaultSystemPrompt, defaultStrategyPrompt)
	if err != nil {
		return fmt.Errorf("failed to create default config: %w", err)
	}

	return nil
}

// GetUserConfig retrieves user configuration
func (ucm *UserConfigManager) GetUserConfig(userID string) (*UserConfig, error) {
	config := &UserConfig{UserID: userID}

	var exchangeConfigsJSON, riskLimitsJSON, notificationSettingsJSON sql.NullString
	var aiBaseURL, aiAPIKey, systemPrompt, strategyPrompt sql.NullString

	query := `
		SELECT
			ai_provider, ai_model, ai_base_url, ai_api_key,
			default_system_prompt, default_strategy_prompt,
			exchange_configs, risk_limits, notification_settings,
			created_at, updated_at
		FROM user_configs
		WHERE user_id = ?
	`

	err := ucm.db.QueryRow(query, userID).Scan(
		&config.AIProvider,
		&config.AIModel,
		&aiBaseURL,
		&aiAPIKey,
		&systemPrompt,
		&strategyPrompt,
		&exchangeConfigsJSON,
		&riskLimitsJSON,
		&notificationSettingsJSON,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Create default config if not exists
		if err := ucm.CreateDefaultConfig(userID); err != nil {
			return nil, err
		}
		// Retry
		return ucm.GetUserConfig(userID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user config: %w", err)
	}

	// Parse nullable strings
	if aiBaseURL.Valid {
		config.AIBaseURL = aiBaseURL.String
	}
	if aiAPIKey.Valid {
		config.AIAPIKey = aiAPIKey.String
	}
	if systemPrompt.Valid {
		config.DefaultSystemPrompt = systemPrompt.String
	}
	if strategyPrompt.Valid {
		config.DefaultStrategyPrompt = strategyPrompt.String
	}

	// Parse JSON fields
	if exchangeConfigsJSON.Valid && exchangeConfigsJSON.String != "" {
		json.Unmarshal([]byte(exchangeConfigsJSON.String), &config.ExchangeConfigs)
	}
	if riskLimitsJSON.Valid && riskLimitsJSON.String != "" {
		json.Unmarshal([]byte(riskLimitsJSON.String), &config.RiskLimits)
	}
	if notificationSettingsJSON.Valid && notificationSettingsJSON.String != "" {
		json.Unmarshal([]byte(notificationSettingsJSON.String), &config.NotificationSettings)
	}

	return config, nil
}

// UpdateUserConfig updates user configuration
func (ucm *UserConfigManager) UpdateUserConfig(userID string, updates map[string]interface{}) error {
	// Build dynamic update query
	query := "UPDATE user_configs SET updated_at = CURRENT_TIMESTAMP"
	args := []interface{}{}

	if provider, ok := updates["ai_provider"].(string); ok {
		query += ", ai_provider = ?"
		args = append(args, provider)
	}

	if model, ok := updates["ai_model"].(string); ok {
		query += ", ai_model = ?"
		args = append(args, model)
	}

	if baseURL, ok := updates["ai_base_url"].(string); ok {
		query += ", ai_base_url = ?"
		args = append(args, baseURL)
	}

	if apiKey, ok := updates["ai_api_key"].(string); ok {
		query += ", ai_api_key = ?"
		args = append(args, apiKey)
	}

	if systemPrompt, ok := updates["default_system_prompt"].(string); ok {
		query += ", default_system_prompt = ?"
		args = append(args, systemPrompt)
	}

	if strategyPrompt, ok := updates["default_strategy_prompt"].(string); ok {
		query += ", default_strategy_prompt = ?"
		args = append(args, strategyPrompt)
	}

	if exchangeConfigs, ok := updates["exchange_configs"]; ok {
		data, _ := json.Marshal(exchangeConfigs)
		query += ", exchange_configs = ?"
		args = append(args, string(data))
	}

	if riskLimits, ok := updates["risk_limits"]; ok {
		data, _ := json.Marshal(riskLimits)
		query += ", risk_limits = ?"
		args = append(args, string(data))
	}

	if notificationSettings, ok := updates["notification_settings"]; ok {
		data, _ := json.Marshal(notificationSettings)
		query += ", notification_settings = ?"
		args = append(args, string(data))
	}

	query += " WHERE user_id = ?"
	args = append(args, userID)

	result, err := ucm.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update user config: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user config not found")
	}

	return nil
}

// DeleteUserConfig deletes user configuration
func (ucm *UserConfigManager) DeleteUserConfig(userID string) error {
	query := `DELETE FROM user_configs WHERE user_id = ?`

	_, err := ucm.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user config: %w", err)
	}

	return nil
}

// SetAIConfig sets AI configuration for a user
func (ucm *UserConfigManager) SetAIConfig(userID, provider, model, baseURL, apiKey string) error {
	updates := map[string]interface{}{
		"ai_provider": provider,
		"ai_model":    model,
	}

	if baseURL != "" {
		updates["ai_base_url"] = baseURL
	}

	if apiKey != "" {
		updates["ai_api_key"] = apiKey
	}

	return ucm.UpdateUserConfig(userID, updates)
}

// SetDefaultPrompts sets default prompts for a user
func (ucm *UserConfigManager) SetDefaultPrompts(userID, systemPrompt, strategyPrompt string) error {
	updates := map[string]interface{}{
		"default_system_prompt":   systemPrompt,
		"default_strategy_prompt": strategyPrompt,
	}

	return ucm.UpdateUserConfig(userID, updates)
}

// GetAIConfig retrieves AI configuration (for use in decision engine)
func (ucm *UserConfigManager) GetAIConfig(userID string) (provider, model, baseURL, apiKey string, err error) {
	query := `
		SELECT ai_provider, ai_model, COALESCE(ai_base_url, ''), COALESCE(ai_api_key, '')
		FROM user_configs
		WHERE user_id = ?
	`

	err = ucm.db.QueryRow(query, userID).Scan(&provider, &model, &baseURL, &apiKey)
	if err == sql.ErrNoRows {
		// Return defaults
		return "openai", "gpt-4", "", "", nil
	}

	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to get AI config: %w", err)
	}

	return provider, model, baseURL, apiKey, nil
}

// GetDefaultPrompts retrieves default prompts
func (ucm *UserConfigManager) GetDefaultPrompts(userID string) (systemPrompt, strategyPrompt string, err error) {
	query := `
		SELECT COALESCE(default_system_prompt, ''), COALESCE(default_strategy_prompt, '')
		FROM user_configs
		WHERE user_id = ?
	`

	err = ucm.db.QueryRow(query, userID).Scan(&systemPrompt, &strategyPrompt)
	if err == sql.ErrNoRows {
		return "", "", nil
	}

	if err != nil {
		return "", "", fmt.Errorf("failed to get default prompts: %w", err)
	}

	return systemPrompt, strategyPrompt, nil
}
