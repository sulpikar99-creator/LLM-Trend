package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sulpikar99-creator/LLM-Trend/config"
)

// handleGetConfig handles getting configuration
func (s *Server) handleGetConfig(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Return current config (without sensitive data)
	successResponse(c, gin.H{
		"ai_models": s.app.Config.AIModels,
		"risk_limits": s.app.Config.RiskLimits,
		"default_strategy_prompt": s.app.Config.DefaultStrategyPrompt,
	})
}

// UpdateConfigRequest represents config update request
type UpdateConfigRequest struct {
	AIModels              []config.AIModel    `json:"ai_models"`
	RiskLimits            *config.RiskLimits  `json:"risk_limits"`
	DefaultStrategyPrompt string              `json:"default_strategy_prompt"`
}

// handleUpdateConfig handles updating configuration
func (s *Server) handleUpdateConfig(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Update AI models if provided
	if len(req.AIModels) > 0 {
		s.app.Config.AIModels = req.AIModels
	}

	// Update risk limits if provided
	if req.RiskLimits != nil {
		s.app.Config.RiskLimits = *req.RiskLimits
	}

	// Update strategy prompt if provided
	if req.DefaultStrategyPrompt != "" {
		s.app.Config.DefaultStrategyPrompt = req.DefaultStrategyPrompt
	}

	// Save config to file
	if err := config.SaveConfig(s.app.Config); err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to save config: "+err.Error())
		return
	}

	successResponse(c, gin.H{
		"message": "Configuration updated successfully",
		"config": gin.H{
			"ai_models":                s.app.Config.AIModels,
			"risk_limits":              s.app.Config.RiskLimits,
			"default_strategy_prompt":  s.app.Config.DefaultStrategyPrompt,
		},
	})
}
