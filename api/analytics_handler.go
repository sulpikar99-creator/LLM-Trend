package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sulpikar99-creator/LLM-Trend/analytics"
)

// handleGetDrawdown handles drawdown analysis
func (s *Server) handleGetDrawdown(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Get trader_id from query params (optional)
	traderID := c.Query("trader_id")

	// Get date range (optional)
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// Build query
	query := `
		SELECT timestamp, equity, balance
		FROM performance_records
		WHERE trader_id IN (SELECT id FROM traders WHERE user_id = ?)
	`
	args := []interface{}{userID}

	if traderID != "" {
		query = `
			SELECT timestamp, equity, balance
			FROM performance_records
			WHERE trader_id = ?
		`
		args = []interface{}{traderID}

		// Verify ownership
		var ownerID string
		err := s.app.Database.DB.QueryRow("SELECT user_id FROM traders WHERE id = ?", traderID).Scan(&ownerID)
		if err != nil || ownerID != userID {
			errorResponse(c, http.StatusForbidden, "Access denied")
			return
		}
	}

	if startDateStr != "" && endDateStr != "" {
		query += " AND timestamp BETWEEN ? AND ?"
		args = append(args, startDateStr, endDateStr)
	}

	query += " ORDER BY timestamp ASC"

	// Fetch equity curve from database
	rows, err := s.app.Database.DB.Query(query, args...)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}
	defer rows.Close()

	equityPoints := []analytics.EquityPoint{}
	for rows.Next() {
		var point analytics.EquityPoint
		var timestamp string
		err := rows.Scan(&timestamp, &point.Equity, &point.Balance)
		if err != nil {
			continue
		}

		// Parse timestamp
		t, err := time.Parse("2006-01-02 15:04:05", timestamp)
		if err != nil {
			t, _ = time.Parse(time.RFC3339, timestamp)
		}
		point.Timestamp = t

		equityPoints = append(equityPoints, point)
	}

	// If no data, return empty result
	if len(equityPoints) == 0 {
		successResponse(c, gin.H{
			"message":           "No performance data available",
			"trader_id":         traderID,
			"drawdown_analysis": nil,
			"underwater_chart":  nil,
			"equity_points":     []analytics.EquityPoint{},
		})
		return
	}

	// Calculate drawdown
	ddAnalysis, err := analytics.CalculateDrawdown(equityPoints)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to calculate drawdown: "+err.Error())
		return
	}

	// Calculate underwater chart
	underwater := analytics.CalculateUnderwaterPeriod(equityPoints)

	response := gin.H{
		"drawdown_analysis": ddAnalysis,
		"underwater_chart":  underwater,
		"equity_points":     equityPoints,
		"trader_id":         traderID,
	}

	successResponse(c, response)
}

// handleGetMonteCarlo handles Monte Carlo simulation
func (s *Server) handleGetMonteCarlo(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	traderID := c.Query("trader_id")

	// Parse query parameters
	initialBalance := 10000.0
	if bal := c.Query("initial_balance"); bal != "" {
		if parsed, err := strconv.ParseFloat(bal, 64); err == nil {
			initialBalance = parsed
		}
	}

	numSimulations := 1000
	if sims := c.Query("num_simulations"); sims != "" {
		if parsed, err := strconv.Atoi(sims); err == nil && parsed > 0 {
			numSimulations = parsed
		}
	}

	numPeriods := 100
	if periods := c.Query("num_periods"); periods != "" {
		if parsed, err := strconv.Atoi(periods); err == nil && parsed > 0 {
			numPeriods = parsed
		}
	}

	// Get historical returns from real performance data
	historicalReturns := []float64{}

	query := `
		SELECT equity
		FROM performance_records
		WHERE trader_id IN (SELECT id FROM traders WHERE user_id = ?)
		ORDER BY timestamp ASC
	`
	args := []interface{}{userID}

	if traderID != "" {
		// Verify ownership
		var ownerID string
		err := s.app.Database.DB.QueryRow("SELECT user_id FROM traders WHERE id = ?", traderID).Scan(&ownerID)
		if err != nil || ownerID != userID {
			errorResponse(c, http.StatusForbidden, "Access denied")
			return
		}

		query = `
			SELECT equity
			FROM performance_records
			WHERE trader_id = ?
			ORDER BY timestamp ASC
		`
		args = []interface{}{traderID}
	}

	rows, err := s.app.Database.DB.Query(query, args...)
	if err == nil {
		defer rows.Close()

		var prevEquity float64
		first := true

		for rows.Next() {
			var equity float64
			if err := rows.Scan(&equity); err != nil {
				continue
			}

			if !first && prevEquity > 0 {
				ret := (equity - prevEquity) / prevEquity
				historicalReturns = append(historicalReturns, ret)
			}

			prevEquity = equity
			first = false
		}
	}

	// Use default if no historical data
	if len(historicalReturns) == 0 {
		historicalReturns = []float64{0.02, -0.01, 0.03, -0.015, 0.025, 0.01, -0.02, 0.04}
	}

	// Estimate parameters from history
	params := analytics.EstimateParametersFromHistory(historicalReturns)
	params.InitialBalance = initialBalance
	params.NumSimulations = numSimulations
	params.NumPeriods = numPeriods
	params.UseHistorical = true

	// Allow override via query params
	if mean := c.Query("mean_return"); mean != "" {
		if parsed, err := strconv.ParseFloat(mean, 64); err == nil {
			params.MeanReturn = parsed
		}
	}

	if stdDev := c.Query("std_dev"); stdDev != "" {
		if parsed, err := strconv.ParseFloat(stdDev, 64); err == nil {
			params.StdDevReturn = parsed
		}
	}

	// Run simulation
	result, err := analytics.RunMonteCarloSimulation(params)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Simulation failed: "+err.Error())
		return
	}

	// Return results (limit paths for response size)
	maxPaths := 100
	paths := result.Paths
	if len(paths) > maxPaths {
		paths = paths[:maxPaths]
	}

	response := gin.H{
		"config":               params,
		"paths":                paths,
		"percentiles":          result.Percentiles,
		"mean":                 result.Mean,
		"std_dev":              result.StdDev,
		"median_return":        result.MedianReturn,
		"probability_profit":   result.ProbabilityProfit,
		"probability_loss":     result.ProbabilityLoss,
		"var_95":               result.ValueAtRisk95,
		"var_99":               result.ValueAtRisk99,
		"cvar_95":              result.ConditionalVaR95,
		"avg_max_drawdown":     result.AvgMaxDrawdown,
		"best_case":            result.BestCaseScenario,
		"worst_case":           result.WorstCaseScenario,
		"confidence_intervals": result.ConfidenceIntervals,
		"summary":              result.GetSummary(),
		"historical_samples":   len(historicalReturns),
	}

	successResponse(c, response)
}

// handleGetCorrelation handles correlation matrix
func (s *Server) handleGetCorrelation(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Get symbols from query (comma-separated)
	symbolsStr := c.DefaultQuery("symbols", "BTCUSDT,ETHUSDT,BNBUSDT")
	symbols := strings.Split(symbolsStr, ",")

	// Trim whitespace
	for i := range symbols {
		symbols[i] = strings.TrimSpace(symbols[i])
	}

	// Note: For real correlation, we would need historical price data
	// This could come from:
	// 1. Kline data stored in database (not implemented yet)
	// 2. Real-time fetch from Binance API
	// 3. Decision records with entry prices

	// For now, fetch from decision records as proxy
	priceData := make([]*analytics.PriceData, 0)

	for _, symbol := range symbols {
		query := `
			SELECT decision_json
			FROM decision_records
			WHERE trader_id IN (SELECT id FROM traders WHERE user_id = ?)
			AND decision_json LIKE ?
			ORDER BY timestamp DESC
			LIMIT 100
		`

		rows, err := s.app.Database.DB.Query(query, userID, "%"+symbol+"%")
		if err != nil {
			continue
		}

		prices := []float64{}

		for rows.Next() {
			var decisionJSON sql.NullString
			if err := rows.Scan(&decisionJSON); err != nil {
				continue
			}

			if !decisionJSON.Valid {
				continue
			}

			// Parse decision JSON to extract price
			var decision map[string]interface{}
			if err := json.Unmarshal([]byte(decisionJSON.String), &decision); err != nil {
				continue
			}

			// Extract price if available
			if priceVal, ok := decision["entry_price"].(float64); ok {
				prices = append(prices, priceVal)
			}
		}
		rows.Close()

		if len(prices) > 0 {
			priceData = append(priceData, &analytics.PriceData{
				Symbol: symbol,
				Prices: prices,
			})
		}
	}

	// If no real data, return informative error
	if len(priceData) < 2 {
		successResponse(c, gin.H{
			"message": "Insufficient price data for correlation analysis. Need historical data for at least 2 symbols.",
			"symbols": symbols,
			"note":    "Correlation analysis requires historical trading data. Start trading or provide more symbols.",
		})
		return
	}

	// Calculate correlation matrix
	corrMatrix, err := analytics.CalculateCorrelationMatrix(priceData)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Failed to calculate correlation: "+err.Error())
		return
	}

	// Get top N correlations
	topN := 10
	if n := c.Query("top_n"); n != "" {
		if parsed, err := strconv.Atoi(n); err == nil && parsed > 0 {
			topN = parsed
		}
	}

	topCorrelations := corrMatrix.GetTopCorrelations(topN)

	response := gin.H{
		"symbols":           corrMatrix.Symbols,
		"matrix":            corrMatrix.Matrix,
		"heatmap":           corrMatrix.Heatmap,
		"pairs":             corrMatrix.Pairs,
		"top_correlations":  topCorrelations,
		"avg_correlation":   corrMatrix.AvgCorr,
		"max_correlation":   corrMatrix.MaxCorr,
		"min_correlation":   corrMatrix.MinCorr,
		"summary":           corrMatrix.Summary,
		"requested_symbols": symbolsStr,
	}

	successResponse(c, response)
}

// handleGetPerformance handles performance attribution
func (s *Server) handleGetPerformance(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Get trader_id from query (optional)
	traderID := c.Query("trader_id")

	// Fetch trade history from decision_records
	query := `
		SELECT
			id, trader_id, cycle_number, timestamp,
			decision_json, account_state, execution_logs
		FROM decision_records
		WHERE trader_id IN (SELECT id FROM traders WHERE user_id = ?)
		AND decision_json IS NOT NULL
		ORDER BY timestamp DESC
		LIMIT 1000
	`
	args := []interface{}{userID}

	if traderID != "" {
		// Verify ownership
		var ownerID string
		err := s.app.Database.DB.QueryRow("SELECT user_id FROM traders WHERE id = ?", traderID).Scan(&ownerID)
		if err != nil || ownerID != userID {
			errorResponse(c, http.StatusForbidden, "Access denied")
			return
		}

		query = `
			SELECT
				id, trader_id, cycle_number, timestamp,
				decision_json, account_state, execution_logs
			FROM decision_records
			WHERE trader_id = ?
			AND decision_json IS NOT NULL
			ORDER BY timestamp DESC
			LIMIT 1000
		`
		args = []interface{}{traderID}
	}

	rows, err := s.app.Database.DB.Query(query, args...)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}
	defer rows.Close()

	trades := []*analytics.TradeRecord{}

	for rows.Next() {
		var id, traderIDStr string
		var cycleNumber int
		var timestamp string
		var decisionJSON, accountState, executionLogs sql.NullString

		if err := rows.Scan(&id, &traderIDStr, &cycleNumber, &timestamp, &decisionJSON, &accountState, &executionLogs); err != nil {
			continue
		}

		if !decisionJSON.Valid {
			continue
		}

		// Parse decision JSON
		var decision map[string]interface{}
		if err := json.Unmarshal([]byte(decisionJSON.String), &decision); err != nil {
			continue
		}

		// Parse execution logs for PnL if available
		var execLogs map[string]interface{}
		if executionLogs.Valid {
			json.Unmarshal([]byte(executionLogs.String), &execLogs)
		}

		// Extract trade information
		action, _ := decision["action"].(string)
		if action != "open_long" && action != "open_short" && action != "close" {
			continue
		}

		symbol, _ := decision["symbol"].(string)
		entryPrice, _ := decision["entry_price"].(float64)
		size, _ := decision["size"].(float64)

		// Determine side
		side := "LONG"
		if action == "open_short" {
			side = "SHORT"
		}

		// Try to calculate PnL from execution logs or account state
		pnl := 0.0
		if execLogs != nil {
			if pnlVal, ok := execLogs["pnl"].(float64); ok {
				pnl = pnlVal
			}
		}

		// Parse timestamp
		t, _ := time.Parse("2006-01-02 15:04:05", timestamp)

		// Create trade record
		trade := &analytics.TradeRecord{
			Symbol:      symbol,
			Side:        side,
			EntryPrice:  entryPrice,
			ExitPrice:   entryPrice, // Would need to track actual exit
			Size:        size,
			EntryTime:   t,
			ExitTime:    t.Add(1 * time.Hour), // Approximate
			PnL:         pnl,
			PnLPercent:  (pnl / (entryPrice * size)) * 100,
			Commission:  entryPrice * size * 0.0004 * 2,
			NetPnL:      pnl - (entryPrice * size * 0.0004 * 2),
			Duration:    1.0, // Approximate
			Strategy:    "AI Strategy",
			TradeID:     id,
		}

		trades = append(trades, trade)
	}

	// If no trades, return empty result
	if len(trades) == 0 {
		successResponse(c, gin.H{
			"message":      "No trade data available",
			"trader_id":    traderID,
			"total_trades": 0,
		})
		return
	}

	// Calculate performance metrics
	metrics := analytics.CalculatePerformanceMetrics(trades)

	// Calculate performance attribution
	attribution := analytics.CalculatePerformanceAttribution(trades)

	// Get top/worst performers
	topN := 5
	if n := c.Query("top_n"); n != "" {
		if parsed, err := strconv.Atoi(n); err == nil && parsed > 0 {
			topN = parsed
		}
	}

	topPerformers := attribution.GetTopPerformers(topN)
	worstPerformers := attribution.GetWorstPerformers(topN)

	response := gin.H{
		"overall_metrics":  metrics,
		"by_symbol":        attribution.BySymbol,
		"by_strategy":      attribution.ByStrategy,
		"by_side":          attribution.BySide,
		"by_timeframe":     attribution.ByTimeframe,
		"top_performers":   topPerformers,
		"worst_performers": worstPerformers,
		"trader_id":        traderID,
		"summary":          metrics.GenerateSummaryReport(),
		"total_trades":     len(trades),
	}

	successResponse(c, response)
}
