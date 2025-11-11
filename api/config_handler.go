package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
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

// handleUpdateConfig handles updating configuration
func (s *Server) handleUpdateConfig(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Check if user is admin
	role := getUserRole(c)
	if role != "admin" {
		errorResponse(c, http.StatusForbidden, "Admin access required")
		return
	}

	// TODO: Implement config update
	successResponse(c, gin.H{
		"message": "Config update - to be implemented",
	})
}
