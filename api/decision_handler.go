package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sulpikar99-creator/LLM-Trend/logger"
)

// handleGetDecisions handles getting decision logs for a trader
func (s *Server) handleGetDecisions(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	traderID := c.Param("trader_id")

	// Verify ownership
	mt, err := s.app.TraderManager.GetTrader(traderID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "Trader not found")
		return
	}

	if mt.UserID != userID {
		errorResponse(c, http.StatusForbidden, "Access denied")
		return
	}

	// Get query parameters
	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	// Get date filter if provided
	dateStr := c.Query("date")
	var decisions []map[string]interface{}

	decisionLogger := logger.NewDecisionLogger("decision_logs")

	if dateStr != "" {
		// Parse date
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid date format (use YYYY-MM-DD)")
			return
		}

		decisions, err = decisionLogger.GetDecisionByDate(traderID, date)
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, "Failed to retrieve decisions")
			return
		}
	} else {
		// Get latest decisions
		decisions, err = decisionLogger.GetLatestDecisions(traderID, limit)
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, "Failed to retrieve decisions")
			return
		}
	}

	// Get total count
	totalCount, _ := decisionLogger.GetDecisionCount(traderID)

	successResponse(c, gin.H{
		"trader_id": traderID,
		"decisions": decisions,
		"count":     len(decisions),
		"total":     totalCount,
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
	cycleStr := c.Param("cycle")

	// Verify ownership
	mt, err := s.app.TraderManager.GetTrader(traderID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "Trader not found")
		return
	}

	if mt.UserID != userID {
		errorResponse(c, http.StatusForbidden, "Access denied")
		return
	}

	cycleNum, err := strconv.Atoi(cycleStr)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid cycle number")
		return
	}

	// Get all decisions and find the one with matching cycle
	decisionLogger := logger.NewDecisionLogger("decision_logs")
	decisions, err := decisionLogger.GetLatestDecisions(traderID, 1000)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to retrieve decisions")
		return
	}

	// Find decision with matching cycle number
	for _, decision := range decisions {
		if cycle, ok := decision["cycle_number"]; ok {
			if cycleFloat, ok := cycle.(float64); ok && int(cycleFloat) == cycleNum {
				successResponse(c, gin.H{
					"trader_id": traderID,
					"cycle":     cycleNum,
					"decision":  decision,
				})
				return
			}
		}
	}

	errorResponse(c, http.StatusNotFound, "Decision not found for cycle "+cycleStr)
}

// handleDeleteOldDecisions handles deletion of old decision logs
func (s *Server) handleDeleteOldDecisions(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	traderID := c.Param("trader_id")

	// Verify ownership
	mt, err := s.app.TraderManager.GetTrader(traderID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "Trader not found")
		return
	}

	if mt.UserID != userID {
		errorResponse(c, http.StatusForbidden, "Access denied")
		return
	}

	// Get days parameter
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		errorResponse(c, http.StatusBadRequest, "Invalid days parameter")
		return
	}

	decisionLogger := logger.NewDecisionLogger("decision_logs")
	deletedCount, err := decisionLogger.DeleteOldDecisions(traderID, days)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to delete decisions")
		return
	}

	successResponse(c, gin.H{
		"trader_id": traderID,
		"deleted":   deletedCount,
		"message":   "Old decisions deleted successfully",
	})
}
