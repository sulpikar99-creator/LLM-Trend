package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sulpikar99-creator/LLM-Trend/market"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, validate origin properly
		origin := r.Header.Get("Origin")
		// Allow localhost origins for development
		return origin == "http://localhost:3000" ||
			origin == "http://localhost:5173" ||
			origin == "http://localhost:8080"
	},
}

// handleWebSocket handles WebSocket connection requests for a specific trader
// @Summary      WebSocket connection for trader updates
// @Description  Establishes WebSocket connection for real-time trader updates
// @Tags         websocket
// @Param        trader_id path string true "Trader ID"
// @Success      101  "Switching Protocols"
// @Failure      401  {object}  ErrorResponse "Unauthorized"
// @Failure      403  {object}  ErrorResponse "Forbidden - trader not owned by user"
// @Failure      404  {object}  ErrorResponse "Trader not found"
// @Security     BearerAuth
// @Router       /api/ws/traders/{trader_id} [get]
func (s *Server) handleWebSocket(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	traderID := c.Param("trader_id")
	if traderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Trader ID is required"})
		return
	}

	// Verify trader belongs to user
	var dbUserID string
	err := s.app.Database.DB.QueryRow(
		"SELECT user_id FROM traders WHERE id = ?",
		traderID,
	).Scan(&dbUserID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trader not found"})
		return
	}

	if dbUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this trader"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade connection"})
		return
	}

	// Create new client
	client := &market.Client{
		TraderID: traderID,
		UserID:   userID,
	}

	// Register client with hub (using reflection to set private fields)
	// Note: This requires the Client struct to have exported fields or we need to modify it
	s.wsHub.RegisterClient(client, conn)

	// Start read and write pumps
	go client.WritePump()
	go client.ReadPump()
}

// handleWebSocketAll handles WebSocket connection for all user's traders
// @Summary      WebSocket connection for all user traders
// @Description  Establishes WebSocket connection for all trader updates of the user
// @Tags         websocket
// @Success      101  "Switching Protocols"
// @Failure      401  {object}  ErrorResponse "Unauthorized"
// @Security     BearerAuth
// @Router       /api/ws/all [get]
func (s *Server) handleWebSocketAll(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get all trader IDs for user
	rows, err := s.app.Database.DB.Query(
		"SELECT id FROM traders WHERE user_id = ?",
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get traders"})
		return
	}
	defer rows.Close()

	var traderIDs []string
	for rows.Next() {
		var traderID string
		if err := rows.Scan(&traderID); err != nil {
			continue
		}
		traderIDs = append(traderIDs, traderID)
	}

	if len(traderIDs) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No traders found"})
		return
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade connection"})
		return
	}

	// Create multi-trader client
	// For simplicity, we'll use the first trader ID as primary
	// In production, you might want a different approach
	client := &market.Client{
		TraderID: traderIDs[0], // Primary trader
		UserID:   userID,
	}

	s.wsHub.RegisterClient(client, conn)

	go client.WritePump()
	go client.ReadPump()
}

// handleWebSocketStats returns WebSocket statistics
// @Summary      Get WebSocket statistics
// @Description  Returns current WebSocket connection statistics
// @Tags         websocket
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /api/ws/stats [get]
func (s *Server) handleWebSocketStats(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Get trader IDs for user
	rows, err := s.app.Database.DB.Query(
		"SELECT id FROM traders WHERE user_id = ?",
		userID,
	)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to get traders")
		return
	}
	defer rows.Close()

	traderStats := make(map[string]int)
	totalClients := 0

	for rows.Next() {
		var traderID string
		if err := rows.Scan(&traderID); err != nil {
			continue
		}
		count := s.wsHub.GetClientCount(traderID)
		traderStats[traderID] = count
		totalClients += count
	}

	successResponse(c, gin.H{
		"total_connections": totalClients,
		"traders":           traderStats,
		"server_total":      s.wsHub.GetTotalClientCount(),
	})
}
