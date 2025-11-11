package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// handleGetDecisions handles getting decision logs for a trader
func (s *Server) handleGetDecisions(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	traderID := c.Param("trader_id")

	// TODO: Implement get decisions
	successResponse(c, gin.H{
		"trader_id": traderID,
		"decisions": []gin.H{},
		"message": "Get decisions - to be implemented",
	})
}

// handleGetDecisionByCycle handles getting a specific decision by cycle number
func (s *Server) handleGetDecisionByCycle(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	traderID := c.Param("trader_id")
	cycle := c.Param("cycle")

	// TODO: Implement get decision by cycle
	successResponse(c, gin.H{
		"trader_id": traderID,
		"cycle": cycle,
		"decision": gin.H{},
		"message": "Get decision by cycle - to be implemented",
	})
}
