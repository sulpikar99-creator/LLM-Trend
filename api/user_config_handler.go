package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// handleGetUserConfig gets the current user's configuration
func (s *Server) handleGetUserConfig(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userConfigMgr := s.app.Config.UserConfigManager
	userConfig, err := userConfigMgr.GetUserConfig(userID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "User configuration not found")
		return
	}

	successResponse(c, userConfig)
}

// handleUpdateUserConfig updates the current user's configuration
func (s *Server) handleUpdateUserConfig(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	userConfigMgr := s.app.Config.UserConfigManager
	if err := userConfigMgr.UpdateUserConfig(userID, updates); err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to update configuration: "+err.Error())
		return
	}

	// Get updated config
	userConfig, err := userConfigMgr.GetUserConfig(userID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to retrieve updated configuration")
		return
	}

	successResponse(c, gin.H{
		"message": "Configuration updated successfully",
		"config":  userConfig,
	})
}

// handleUpdateUserAIConfig updates AI-related configuration
func (s *Server) handleUpdateUserAIConfig(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req struct {
		AIProvider string `json:"ai_provider" binding:"required,oneof=openai deepseek claude qwen"`
		AIModel    string `json:"ai_model" binding:"required"`
		AIBaseURL  string `json:"ai_base_url"`
		AIApiKey   string `json:"ai_api_key"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	userConfigMgr := s.app.Config.UserConfigManager
	if err := userConfigMgr.SetAIConfig(userID, req.AIProvider, req.AIModel, req.AIBaseURL, req.AIApiKey); err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to update AI configuration: "+err.Error())
		return
	}

	// Get updated AI config
	provider, model, baseURL, apiKey, err := userConfigMgr.GetAIConfig(userID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to retrieve AI configuration")
		return
	}

	aiConfig := gin.H{
		"ai_provider": provider,
		"ai_model":    model,
		"ai_base_url": baseURL,
		"ai_api_key":  apiKey,
	}

	successResponse(c, gin.H{
		"message":   "AI configuration updated successfully",
		"ai_config": aiConfig,
	})
}

// handleGetUserAIConfig gets the current user's AI configuration
func (s *Server) handleGetUserAIConfig(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userConfigMgr := s.app.Config.UserConfigManager
	provider, model, baseURL, apiKey, err := userConfigMgr.GetAIConfig(userID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "AI configuration not found")
		return
	}

	aiConfig := gin.H{
		"ai_provider": provider,
		"ai_model":    model,
		"ai_base_url": baseURL,
		"ai_api_key":  apiKey,
	}

	successResponse(c, aiConfig)
}

// handleUpdateUserPrompts updates default prompts for the user
func (s *Server) handleUpdateUserPrompts(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req struct {
		SystemPrompt   string `json:"system_prompt"`
		StrategyPrompt string `json:"strategy_prompt"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	userConfigMgr := s.app.Config.UserConfigManager
	if err := userConfigMgr.SetDefaultPrompts(userID, req.SystemPrompt, req.StrategyPrompt); err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to update prompts: "+err.Error())
		return
	}

	// Get updated prompts
	systemPrompt, strategyPrompt, err := userConfigMgr.GetDefaultPrompts(userID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to retrieve prompts")
		return
	}

	prompts := gin.H{
		"system_prompt":   systemPrompt,
		"strategy_prompt": strategyPrompt,
	}

	successResponse(c, gin.H{
		"message": "Prompts updated successfully",
		"prompts": prompts,
	})
}

// handleGetUserPrompts gets the current user's default prompts
func (s *Server) handleGetUserPrompts(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userConfigMgr := s.app.Config.UserConfigManager
	systemPrompt, strategyPrompt, err := userConfigMgr.GetDefaultPrompts(userID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "Prompts not found")
		return
	}

	prompts := gin.H{
		"system_prompt":   systemPrompt,
		"strategy_prompt": strategyPrompt,
	}

	successResponse(c, prompts)
}

// handleUpdateExchangeConfig updates exchange configuration for the user
func (s *Server) handleUpdateExchangeConfig(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var exchangeConfigs map[string]interface{}
	if err := c.ShouldBindJSON(&exchangeConfigs); err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	userConfigMgr := s.app.Config.UserConfigManager
	updates := map[string]interface{}{
		"exchange_configs": exchangeConfigs,
	}

	if err := userConfigMgr.UpdateUserConfig(userID, updates); err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to update exchange configuration: "+err.Error())
		return
	}

	successResponse(c, gin.H{
		"message":          "Exchange configuration updated successfully",
		"exchange_configs": exchangeConfigs,
	})
}

// handleUpdateRiskLimits updates risk limits for the user
func (s *Server) handleUpdateRiskLimits(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var riskLimits map[string]interface{}
	if err := c.ShouldBindJSON(&riskLimits); err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	userConfigMgr := s.app.Config.UserConfigManager
	updates := map[string]interface{}{
		"risk_limits": riskLimits,
	}

	if err := userConfigMgr.UpdateUserConfig(userID, updates); err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to update risk limits: "+err.Error())
		return
	}

	successResponse(c, gin.H{
		"message":     "Risk limits updated successfully",
		"risk_limits": riskLimits,
	})
}

// handleUpdateNotificationSettings updates notification settings for the user
func (s *Server) handleUpdateNotificationSettings(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var notificationSettings map[string]interface{}
	if err := c.ShouldBindJSON(&notificationSettings); err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	userConfigMgr := s.app.Config.UserConfigManager
	updates := map[string]interface{}{
		"notification_settings": notificationSettings,
	}

	if err := userConfigMgr.UpdateUserConfig(userID, updates); err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to update notification settings: "+err.Error())
		return
	}

	successResponse(c, gin.H{
		"message":               "Notification settings updated successfully",
		"notification_settings": notificationSettings,
	})
}

// handleGetAvailableAIProviders returns list of supported AI providers
func (s *Server) handleGetAvailableAIProviders(c *gin.Context) {
	providers := []map[string]interface{}{
		{
			"id":          "openai",
			"name":        "OpenAI",
			"models":      []string{"gpt-4", "gpt-4-turbo", "gpt-3.5-turbo"},
			"description": "OpenAI's GPT models",
		},
		{
			"id":          "deepseek",
			"name":        "DeepSeek",
			"models":      []string{"deepseek-chat", "deepseek-coder"},
			"description": "DeepSeek AI models",
		},
		{
			"id":          "claude",
			"name":        "Anthropic Claude",
			"models":      []string{"claude-3-opus", "claude-3-sonnet", "claude-3-haiku"},
			"description": "Anthropic's Claude models",
		},
		{
			"id":          "qwen",
			"name":        "Qwen",
			"models":      []string{"qwen-turbo", "qwen-plus", "qwen-max"},
			"description": "Alibaba's Qwen models",
		},
	}

	successResponse(c, gin.H{
		"providers": providers,
		"count":     len(providers),
	})
}

// handleResetUserConfig resets user config to defaults
func (s *Server) handleResetUserConfig(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userConfigMgr := s.app.Config.UserConfigManager

	// Delete existing config
	_, err := s.app.Database.DB.Exec("DELETE FROM user_configs WHERE user_id = ?", userID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to delete existing config")
		return
	}

	// Create default config
	if err := userConfigMgr.CreateDefaultConfig(userID); err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to create default config: "+err.Error())
		return
	}

	// Get new config
	userConfig, err := userConfigMgr.GetUserConfig(userID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to retrieve config")
		return
	}

	successResponse(c, gin.H{
		"message": "Configuration reset to defaults successfully",
		"config":  userConfig,
	})
}
