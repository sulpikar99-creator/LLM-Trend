package api

import (
	"net/http"
	"strconv"
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

	// For demo purposes, generate sample equity curve
	// In production, this would come from database
	equityPoints := generateSampleEquityCurve(100, 10000.0)

	// Filter by date range if provided
	if startDateStr != "" && endDateStr != "" {
		startDate, err1 := time.Parse("2006-01-02", startDateStr)
		endDate, err2 := time.Parse("2006-01-02", endDateStr)
		if err1 == nil && err2 == nil {
			filtered := []analytics.EquityPoint{}
			for _, point := range equityPoints {
				if point.Timestamp.After(startDate) && point.Timestamp.Before(endDate) {
					filtered = append(filtered, point)
				}
			}
			equityPoints = filtered
		}
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

	// Get historical returns to estimate parameters (demo data)
	historicalReturns := []float64{0.02, -0.01, 0.03, -0.015, 0.025, 0.01, -0.02, 0.04}

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
		"config":              params,
		"paths":               paths,
		"percentiles":         result.Percentiles,
		"mean":                result.Mean,
		"std_dev":             result.StdDev,
		"median_return":       result.MedianReturn,
		"probability_profit":  result.ProbabilityProfit,
		"probability_loss":    result.ProbabilityLoss,
		"var_95":              result.ValueAtRisk95,
		"var_99":              result.ValueAtRisk99,
		"cvar_95":             result.ConditionalVaR95,
		"avg_max_drawdown":    result.AvgMaxDrawdown,
		"best_case":           result.BestCaseScenario,
		"worst_case":          result.WorstCaseScenario,
		"confidence_intervals": result.ConfidenceIntervals,
		"summary":             result.GetSummary(),
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

	// Parse symbols
	// In production, fetch actual price data from database/exchange
	// For now, generate sample data
	symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT", "SOLUSDT"}
	priceData := make([]*analytics.PriceData, len(symbols))

	for i, symbol := range symbols {
		prices := generateSamplePrices(100, 50000.0+float64(i*10000))
		priceData[i] = &analytics.PriceData{
			Symbol: symbol,
			Prices: prices,
		}
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

	// In production, fetch trade history from database
	// For now, generate sample trades
	trades := generateSampleTrades(50)

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
		"overall_metrics":   metrics,
		"by_symbol":         attribution.BySymbol,
		"by_strategy":       attribution.ByStrategy,
		"by_side":           attribution.BySide,
		"by_timeframe":      attribution.ByTimeframe,
		"top_performers":    topPerformers,
		"worst_performers":  worstPerformers,
		"trader_id":         traderID,
		"summary":           metrics.GenerateSummaryReport(),
		"total_trades":      len(trades),
	}

	successResponse(c, response)
}

// Helper functions for generating sample data
// In production, these would fetch from database

func generateSampleEquityCurve(points int, initialEquity float64) []analytics.EquityPoint {
	result := make([]analytics.EquityPoint, points)
	equity := initialEquity
	baseTime := time.Now().AddDate(0, 0, -points)

	for i := 0; i < points; i++ {
		// Random walk with slight upward bias
		change := (float64(i%10) - 4.5) * 100
		equity += change

		result[i] = analytics.EquityPoint{
			Timestamp: baseTime.Add(time.Duration(i) * time.Hour),
			Equity:    equity,
			Balance:   equity,
		}
	}

	return result
}

func generateSamplePrices(points int, initialPrice float64) []float64 {
	prices := make([]float64, points)
	price := initialPrice

	for i := 0; i < points; i++ {
		// Random walk
		change := (float64(i%20) - 10) * price * 0.001
		price += change

		if price < initialPrice*0.5 {
			price = initialPrice * 0.5
		}
		if price > initialPrice*1.5 {
			price = initialPrice * 1.5
		}

		prices[i] = price
	}

	return prices
}

func generateSampleTrades(count int) []*analytics.TradeRecord {
	trades := make([]*analytics.TradeRecord, count)
	symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT", "SOLUSDT"}
	sides := []string{"LONG", "SHORT"}
	baseTime := time.Now().AddDate(0, 0, -30)

	for i := 0; i < count; i++ {
		symbol := symbols[i%len(symbols)]
		side := sides[i%len(sides)]
		entryPrice := 50000.0 + float64(i*100)
		size := 0.1 + float64(i%10)*0.01

		// Generate win/loss (60% win rate)
		isWin := (i%10) < 6
		exitPrice := entryPrice

		if isWin {
			if side == "LONG" {
				exitPrice = entryPrice * 1.02 // 2% profit
			} else {
				exitPrice = entryPrice * 0.98
			}
		} else {
			if side == "LONG" {
				exitPrice = entryPrice * 0.99 // 1% loss
			} else {
				exitPrice = entryPrice * 1.01
			}
		}

		pnl := 0.0
		if side == "LONG" {
			pnl = (exitPrice - entryPrice) * size
		} else {
			pnl = (entryPrice - exitPrice) * size
		}

		pnlPercent := (pnl / (entryPrice * size)) * 100
		commission := entryPrice * size * 0.0004 * 2 // 0.04% * 2 (entry + exit)
		netPnL := pnl - commission

		entryTime := baseTime.Add(time.Duration(i) * time.Hour * 12)
		exitTime := entryTime.Add(time.Duration(2+i%10) * time.Hour)

		trades[i] = &analytics.TradeRecord{
			Symbol:      symbol,
			Side:        side,
			EntryPrice:  entryPrice,
			ExitPrice:   exitPrice,
			Size:        size,
			EntryTime:   entryTime,
			ExitTime:    exitTime,
			PnL:         pnl,
			PnLPercent:  pnlPercent,
			Commission:  commission,
			NetPnL:      netPnL,
			Duration:    exitTime.Sub(entryTime).Hours(),
			Strategy:    "AI Strategy",
			TradeID:     strconv.Itoa(i + 1),
		}
	}

	return trades
}
