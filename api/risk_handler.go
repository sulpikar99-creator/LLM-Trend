package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sulpikar99-creator/LLM-Trend/risk"
)

// handleGetRiskConfig gets current risk configuration for a trader
func (s *Server) handleGetRiskConfig(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	traderID := c.Param("trader_id")
	if traderID == "" {
		errorResponse(c, http.StatusBadRequest, "Trader ID required")
		return
	}

	// Get trader
	trader, err := s.app.TraderManager.GetTrader(traderID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "Trader not found")
		return
	}

	// Check ownership
	if trader.UserID != userID {
		errorResponse(c, http.StatusForbidden, "Access denied")
		return
	}

	response := gin.H{
		"trader_id": traderID,
	}

	// Get account risk config
	if trader.AccountRisk != nil {
		response["account_risk"] = trader.AccountRisk.GetConfig()
	} else {
		response["account_risk"] = risk.DefaultAccountRiskConfig()
	}

	// Get position risk config
	if trader.PositionRisk != nil {
		response["position_risk"] = trader.PositionRisk.GetConfig()
	} else {
		response["position_risk"] = risk.DefaultPositionRiskConfig()
	}

	successResponse(c, response)
}

// handleUpdateRiskConfig updates risk configuration for a trader
func (s *Server) handleUpdateRiskConfig(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	traderID := c.Param("trader_id")
	if traderID == "" {
		errorResponse(c, http.StatusBadRequest, "Trader ID required")
		return
	}

	// Parse request body
	var req struct {
		AccountRisk  *risk.AccountRiskConfig  `json:"account_risk"`
		PositionRisk *risk.PositionRiskConfig `json:"position_risk"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Get trader
	trader, err := s.app.TraderManager.GetTrader(traderID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "Trader not found")
		return
	}

	// Check ownership
	if trader.UserID != userID {
		errorResponse(c, http.StatusForbidden, "Access denied")
		return
	}

	// Update account risk config
	if req.AccountRisk != nil {
		if trader.AccountRisk != nil {
			trader.AccountRisk.UpdateConfig(req.AccountRisk)
		} else {
			// Initialize if not exists
			balance, _ := trader.GetBalance(c.Request.Context())
			initialBalance := 10000.0
			if balance != nil {
				initialBalance = balance.Balance
			}
			trader.AccountRisk = risk.NewAccountRiskMonitor(req.AccountRisk, initialBalance)
		}
	}

	// Update position risk config
	if req.PositionRisk != nil {
		if trader.PositionRisk != nil {
			trader.PositionRisk.UpdateConfig(req.PositionRisk)
		} else {
			trader.PositionRisk = risk.NewPositionRiskManager(req.PositionRisk)
		}
	}

	response := gin.H{
		"trader_id":     traderID,
		"account_risk":  trader.AccountRisk.GetConfig(),
		"position_risk": trader.PositionRisk.GetConfig(),
		"message":       "Risk configuration updated successfully",
	}

	successResponse(c, response)
}

// handleGetRiskStatus gets current risk status for a trader
func (s *Server) handleGetRiskStatus(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	traderID := c.Param("trader_id")
	if traderID == "" {
		errorResponse(c, http.StatusBadRequest, "Trader ID required")
		return
	}

	// Get trader
	trader, err := s.app.TraderManager.GetTrader(traderID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "Trader not found")
		return
	}

	// Check ownership
	if trader.UserID != userID {
		errorResponse(c, http.StatusForbidden, "Access denied")
		return
	}

	response := gin.H{
		"trader_id": traderID,
	}

	// Get account risk status
	if trader.AccountRisk != nil {
		// Get current positions for status
		positions, err := trader.GetPositions(c.Request.Context())
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, "Failed to get positions: "+err.Error())
			return
		}

		totalExposure := 0.0
		for _, pos := range positions {
			totalExposure += pos.Notional
		}

		status := trader.AccountRisk.GetStatus(len(positions), totalExposure)
		response["account_risk_status"] = status
	} else {
		response["account_risk_status"] = gin.H{
			"message": "Risk monitoring not enabled",
		}
	}

	// Get position risk status
	if trader.PositionRisk != nil {
		positions := trader.PositionRisk.GetAllPositions()
		response["position_risk_status"] = gin.H{
			"tracked_positions": positions,
			"count":             len(positions),
		}
	} else {
		response["position_risk_status"] = gin.H{
			"message": "Position risk management not enabled",
		}
	}

	successResponse(c, response)
}

// handleResetRiskMonitor resets risk monitor (clears violations and stopped state)
func (s *Server) handleResetRiskMonitor(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	traderID := c.Param("trader_id")
	if traderID == "" {
		errorResponse(c, http.StatusBadRequest, "Trader ID required")
		return
	}

	// Get trader
	trader, err := s.app.TraderManager.GetTrader(traderID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "Trader not found")
		return
	}

	// Check ownership
	if trader.UserID != userID {
		errorResponse(c, http.StatusForbidden, "Access denied")
		return
	}

	// Reset account risk monitor
	if trader.AccountRisk != nil {
		trader.AccountRisk.Reset()
	}

	response := gin.H{
		"trader_id": traderID,
		"message":   "Risk monitor reset successfully",
	}

	successResponse(c, response)
}

// handleCalculatePositionSize calculates safe position size based on risk parameters
func (s *Server) handleCalculatePositionSize(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	traderID := c.Param("trader_id")
	if traderID == "" {
		errorResponse(c, http.StatusBadRequest, "Trader ID required")
		return
	}

	// Parse request
	var req struct {
		EntryPrice    float64 `json:"entry_price" binding:"required"`
		StopLossPrice float64 `json:"stop_loss_price" binding:"required"`
		Leverage      int     `json:"leverage" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Get trader
	trader, err := s.app.TraderManager.GetTrader(traderID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "Trader not found")
		return
	}

	// Check ownership
	if trader.UserID != userID {
		errorResponse(c, http.StatusForbidden, "Access denied")
		return
	}

	if trader.AccountRisk == nil {
		errorResponse(c, http.StatusBadRequest, "Risk management not enabled for this trader")
		return
	}

	// Calculate position size
	positionSize := trader.AccountRisk.CalculatePositionSize(
		req.EntryPrice,
		req.StopLossPrice,
		req.Leverage,
	)

	response := gin.H{
		"trader_id":       traderID,
		"entry_price":     req.EntryPrice,
		"stop_loss_price": req.StopLossPrice,
		"leverage":        req.Leverage,
		"position_size":   positionSize,
		"position_size_usd": positionSize * req.EntryPrice,
	}

	successResponse(c, response)
}

// handleCalculateStopLossTakeProfit calculates SL/TP prices
func (s *Server) handleCalculateStopLossTakeProfit(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	traderID := c.Param("trader_id")
	if traderID == "" {
		errorResponse(c, http.StatusBadRequest, "Trader ID required")
		return
	}

	// Parse request
	var req struct {
		EntryPrice     float64 `json:"entry_price" binding:"required"`
		Side           string  `json:"side" binding:"required"` // "LONG" or "SHORT"
		StopLossPercent float64 `json:"stop_loss_percent"`
		TakeProfitPercent float64 `json:"take_profit_percent"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Get trader
	trader, err := s.app.TraderManager.GetTrader(traderID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "Trader not found")
		return
	}

	// Check ownership
	if trader.UserID != userID {
		errorResponse(c, http.StatusForbidden, "Access denied")
		return
	}

	if trader.PositionRisk == nil {
		errorResponse(c, http.StatusBadRequest, "Position risk management not enabled")
		return
	}

	// Calculate stop-loss
	stopLoss, err := trader.PositionRisk.CalculateStopLoss(req.EntryPrice, req.Side, req.StopLossPercent)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Failed to calculate stop-loss: "+err.Error())
		return
	}

	// Calculate take-profit
	takeProfit, err := trader.PositionRisk.CalculateTakeProfit(req.EntryPrice, req.Side, req.TakeProfitPercent)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Failed to calculate take-profit: "+err.Error())
		return
	}

	// Validate risk/reward
	valid, rrRatio, msg := trader.PositionRisk.ValidateRiskReward(req.EntryPrice, stopLoss, takeProfit, req.Side)

	response := gin.H{
		"trader_id":         traderID,
		"entry_price":       req.EntryPrice,
		"side":              req.Side,
		"stop_loss":         stopLoss,
		"take_profit":       takeProfit,
		"risk_reward_ratio": rrRatio,
		"is_valid":          valid,
		"message":           msg,
	}

	successResponse(c, response)
}
