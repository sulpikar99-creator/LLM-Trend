package api

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sulpikar99-creator/LLM-Trend/trader"
)

// exchangeConfigPayload represents nested exchange configuration payloads from the frontend
type exchangeConfigPayload struct {
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
	Testnet   *bool  `json:"testnet"`
}

// CreateTraderRequest represents a request to create a trader
type CreateTraderRequest struct {
	Name           string                 `json:"name" binding:"required"`
	ExchangeType   string                 `json:"exchange_type" binding:"required"`
	APIKey         string                 `json:"api_key"`
	APISecret      string                 `json:"api_secret"`
	Testnet        *bool                  `json:"testnet"`
	ExchangeConfig *exchangeConfigPayload `json:"exchange_config"`
	Symbol         string                 `json:"symbol"`
	Interval       string                 `json:"interval"`
	StrategyPrompt string                 `json:"strategy_prompt"`
}

// StartTraderRequest represents a request to start a trader
type StartTraderRequest struct {
	Symbol   string `json:"symbol"`
	Interval string `json:"interval"`
}

const (
	defaultStartSymbol   = "BTCUSDT"
	defaultStartInterval = "1h"
)

// handleListTraders lists all traders for the authenticated user
func (s *Server) handleListTraders(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Get traders from manager
	managedTraders := s.app.TraderManager.ListTraders(userID)

	// Convert to response format
	traders := []gin.H{}
	for _, mt := range managedTraders {
		traders = append(traders, mt.GetInfo())
	}

	successResponse(c, gin.H{
		"traders": traders,
		"count":   len(traders),
	})
}

// handleCreateTrader creates a new trader
func (s *Server) handleCreateTrader(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req CreateTraderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	apiKey := req.APIKey
	apiSecret := req.APISecret
	if req.ExchangeConfig != nil {
		if apiKey == "" {
			apiKey = req.ExchangeConfig.APIKey
		}
		if apiSecret == "" {
			apiSecret = req.ExchangeConfig.APISecret
		}
	}

	if apiKey == "" || apiSecret == "" {
		errorResponse(c, http.StatusBadRequest, "API key and secret are required")
		return
	}

	testnet := false
	if req.ExchangeConfig != nil && req.ExchangeConfig.Testnet != nil {
		testnet = *req.ExchangeConfig.Testnet
	}
	if req.Testnet != nil {
		testnet = *req.Testnet
	}

	// Generate trader ID
	traderID := uuid.New().String()

	// Encrypt API credentials
	apiKeyEncrypted, err := s.app.CryptoService.EncryptAES(apiKey)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to encrypt API key")
		return
	}

	apiSecretEncrypted, err := s.app.CryptoService.EncryptAES(apiSecret)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to encrypt API secret")
		return
	}

	// Create exchange config JSON
	exchangeConfig := map[string]interface{}{
		"api_key_encrypted":    apiKeyEncrypted,
		"api_secret_encrypted": apiSecretEncrypted,
		"testnet":              testnet,
	}
	exchangeConfigJSON, _ := json.Marshal(exchangeConfig)

	// Create AI config JSON
	aiConfig := map[string]interface{}{
		"model_id": "deepseek-chat", // Default
	}
	aiConfigJSON, _ := json.Marshal(aiConfig)

	// Store in database
	_, err = s.app.Database.DB.Exec(`
		INSERT INTO traders (id, user_id, name, exchange_type, exchange_config, ai_config, strategy_prompt, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		traderID, userID, req.Name, req.ExchangeType, string(exchangeConfigJSON),
		string(aiConfigJSON), req.StrategyPrompt, "stopped",
	)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to create trader")
		return
	}

	// Create trader instance
	var t trader.Trader
	switch req.ExchangeType {
	case "binance_futures":
		t = trader.NewBinanceFuturesTrader(req.Name, apiKey, apiSecret, testnet)
	default:
		errorResponse(c, http.StatusBadRequest, "Unsupported exchange type: "+req.ExchangeType)
		return
	}

	// Add to manager
	mt, err := s.app.TraderManager.AddTrader(traderID, userID, t)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to add trader to manager")
		return
	}

	if req.Symbol != "" {
		mt.Symbol = req.Symbol
	}
	mt.Testnet = testnet

	successResponse(c, gin.H{
		"trader_id": traderID,
		"name":      req.Name,
		"exchange":  req.ExchangeType,
		"status":    "stopped",
		"testnet":   testnet,
		"message":   "Trader created successfully",
	})
}

// handleGetTrader gets a specific trader
func (s *Server) handleGetTrader(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	traderID := c.Param("id")

	// Get trader from manager
	mt, err := s.app.TraderManager.GetTrader(traderID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "Trader not found")
		return
	}

	// Check ownership
	if mt.UserID != userID {
		errorResponse(c, http.StatusForbidden, "Access denied")
		return
	}

	// Get additional info
	info := mt.GetInfo()

	// Get balance if running
	if mt.GetStatus() == "running" {
		if balance, err := mt.GetBalance(c.Request.Context()); err == nil {
			info["balance"] = balance
		}

		// Get positions if running
		if positions, err := mt.GetPositions(c.Request.Context()); err == nil {
			info["positions"] = positions
		}

		// Get market data if available
		if marketData, err := mt.GetMarketData(); err == nil {
			info["market_data"] = marketData
		}
	}

	successResponse(c, info)
}

// handleUpdateTrader updates a trader
func (s *Server) handleUpdateTrader(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	traderID := c.Param("id")

	// Verify ownership
	var ownerID string
	err := s.app.Database.DB.QueryRow(
		"SELECT user_id FROM traders WHERE id = ?", traderID,
	).Scan(&ownerID)
	if err == sql.ErrNoRows {
		errorResponse(c, http.StatusNotFound, "Trader not found")
		return
	}
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}
	if ownerID != userID {
		errorResponse(c, http.StatusForbidden, "Access denied")
		return
	}

	var req struct {
		Name           string `json:"name"`
		StrategyPrompt string `json:"strategy_prompt"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Update database
	_, err = s.app.Database.DB.Exec(`
		UPDATE traders
		SET name = COALESCE(NULLIF(?, ''), name),
		    strategy_prompt = COALESCE(NULLIF(?, ''), strategy_prompt),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		req.Name, req.StrategyPrompt, traderID,
	)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to update trader")
		return
	}

	successResponse(c, gin.H{"message": "Trader updated successfully"})
}

// handleDeleteTrader deletes a trader
func (s *Server) handleDeleteTrader(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	traderID := c.Param("id")

	// Verify ownership
	var ownerID string
	err := s.app.Database.DB.QueryRow(
		"SELECT user_id FROM traders WHERE id = ?", traderID,
	).Scan(&ownerID)
	if err == sql.ErrNoRows {
		errorResponse(c, http.StatusNotFound, "Trader not found")
		return
	}
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}
	if ownerID != userID {
		errorResponse(c, http.StatusForbidden, "Access denied")
		return
	}

	// Remove from manager
	if err := s.app.TraderManager.RemoveTrader(traderID); err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to remove trader")
		return
	}

	// Delete from database
	_, err = s.app.Database.DB.Exec("DELETE FROM traders WHERE id = ?", traderID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to delete trader")
		return
	}

	successResponse(c, gin.H{"message": "Trader deleted successfully"})
}

// handleStartTrader starts a trader
func (s *Server) handleStartTrader(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	traderID := c.Param("id")

	var req StartTraderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if err != io.EOF {
			errorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
			return
		}
	}

	if req.Symbol == "" {
		req.Symbol = defaultStartSymbol
	}
	if req.Interval == "" {
		req.Interval = defaultStartInterval
	}

	// Get trader from manager
	mt, err := s.app.TraderManager.GetTrader(traderID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "Trader not found")
		return
	}

	// Check ownership
	if mt.UserID != userID {
		errorResponse(c, http.StatusForbidden, "Access denied")
		return
	}

	// Start trader
	if err := s.app.TraderManager.StartTrader(traderID, req.Symbol, req.Interval); err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to start trader: "+err.Error())
		return
	}

	// Update status in database
	_, err = s.app.Database.DB.Exec(
		"UPDATE traders SET status = 'running', updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		traderID,
	)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to update status")
		return
	}

	successResponse(c, gin.H{
		"trader_id": traderID,
		"status":    "running",
		"message":   "Trader started successfully",
	})
}

// handleStopTrader stops a trader
func (s *Server) handleStopTrader(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	traderID := c.Param("id")

	// Get trader from manager
	mt, err := s.app.TraderManager.GetTrader(traderID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "Trader not found")
		return
	}

	// Check ownership
	if mt.UserID != userID {
		errorResponse(c, http.StatusForbidden, "Access denied")
		return
	}

	// Stop trader
	if err := s.app.TraderManager.StopTrader(traderID); err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to stop trader: "+err.Error())
		return
	}

	// Update status in database
	_, err = s.app.Database.DB.Exec(
		"UPDATE traders SET status = 'stopped', updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		traderID,
	)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to update status")
		return
	}

	successResponse(c, gin.H{
		"trader_id": traderID,
		"status":    "stopped",
		"message":   "Trader stopped successfully",
	})
}
