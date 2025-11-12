package decision

import (
	"fmt"
	"math"
	"time"
)

// PerformanceAnalysis represents analysis of recent trading performance
type PerformanceAnalysis struct {
	BestTrade           *TradeAnalysis
	WorstTrade          *TradeAnalysis
	ConsecutiveLosses   int
	SuccessfulPatterns  []TradingPattern
	FailedPatterns      []TradingPattern
	RecentWinRate       float64
	RecentWins          int
	AvgHoldTime         string
	BestSymbol          string
	WorstSymbol         string
}

// TradeAnalysis represents analysis of a single trade
type TradeAnalysis struct {
	Action     string
	Symbol     string
	PnL        float64
	PnLPercent float64
	Reasoning  string
	Timestamp  time.Time
}

// TradingPattern represents a trading pattern found in history
type TradingPattern struct {
	Description      string
	WinCount         int
	LossCount        int
	AvgProfitPercent float64
	AvgLossPercent   float64
}

// analyzeRecentPerformance analyzes recent trading decisions to provide insights
func (e *DecisionEngine) analyzeRecentPerformance(recentDecisions []*DecisionRecord) *PerformanceAnalysis {
	analysis := &PerformanceAnalysis{
		SuccessfulPatterns: []TradingPattern{},
		FailedPatterns:     []TradingPattern{},
	}

	if len(recentDecisions) == 0 {
		return analysis
	}

	// Track trades with P&L
	var bestPnL, worstPnL float64
	var bestTrade, worstTrade *DecisionRecord
	consecutiveLosses := 0
	currentStreak := 0
	wins := 0
	totalHoldTime := time.Duration(0)
	symbolPnL := make(map[string]float64)

	// Analyze each decision
	for i, record := range recentDecisions {
		if record.Decision == nil {
			continue
		}

		// Calculate P&L from position snapshots
		pnl := 0.0
		pnlPct := 0.0
		if len(record.PositionSnapshots) > 0 {
			for _, pos := range record.PositionSnapshots {
				pnl += pos.UnrealizedPnL
				pnlPct += pos.UnrealizedPnLPct
			}
		}

		// Track best/worst trades
		if bestTrade == nil || pnl > bestPnL {
			bestPnL = pnl
			bestTrade = record
		}
		if worstTrade == nil || pnl < worstPnL {
			worstPnL = pnl
			worstTrade = record
		}

		// Track consecutive losses
		if pnl < 0 {
			currentStreak++
			if currentStreak > consecutiveLosses {
				consecutiveLosses = currentStreak
			}
		} else if pnl > 0 {
			currentStreak = 0
			wins++
		}

		// Track symbol performance
		if record.Decision.Symbol != "" {
			symbolPnL[record.Decision.Symbol] += pnl
		}

		// Calculate hold time
		if i > 0 {
			totalHoldTime += record.Timestamp.Sub(recentDecisions[i-1].Timestamp)
		}
	}

	// Build best/worst trade analysis
	if bestTrade != nil && bestPnL > 0 {
		analysis.BestTrade = &TradeAnalysis{
			Action:     string(bestTrade.Decision.Action),
			Symbol:     bestTrade.Decision.Symbol,
			PnL:        bestPnL,
			PnLPercent: (bestPnL / bestTrade.AccountState.TotalBalance) * 100,
			Reasoning:  bestTrade.Decision.Reasoning,
			Timestamp:  bestTrade.Timestamp,
		}
	}

	if worstTrade != nil && worstPnL < 0 {
		analysis.WorstTrade = &TradeAnalysis{
			Action:     string(worstTrade.Decision.Action),
			Symbol:     worstTrade.Decision.Symbol,
			PnL:        worstPnL,
			PnLPercent: (worstPnL / worstTrade.AccountState.TotalBalance) * 100,
			Reasoning:  worstTrade.Decision.Reasoning,
			Timestamp:  worstTrade.Timestamp,
		}
	}

	// Calculate win rate
	analysis.RecentWins = wins
	analysis.RecentWinRate = (float64(wins) / float64(len(recentDecisions))) * 100
	analysis.ConsecutiveLosses = consecutiveLosses

	// Calculate average hold time
	if len(recentDecisions) > 1 {
		avgHoldTime := totalHoldTime / time.Duration(len(recentDecisions)-1)
		analysis.AvgHoldTime = formatDuration(avgHoldTime)
	} else {
		analysis.AvgHoldTime = "N/A"
	}

	// Find best/worst performing symbols
	var bestSymbolPnL, worstSymbolPnL float64
	for symbol, pnl := range symbolPnL {
		if analysis.BestSymbol == "" || pnl > bestSymbolPnL {
			analysis.BestSymbol = symbol
			bestSymbolPnL = pnl
		}
		if analysis.WorstSymbol == "" || pnl < worstSymbolPnL {
			analysis.WorstSymbol = symbol
			worstSymbolPnL = pnl
		}
	}

	// Analyze patterns
	analysis.SuccessfulPatterns = e.findSuccessfulPatterns(recentDecisions)
	analysis.FailedPatterns = e.findFailedPatterns(recentDecisions)

	return analysis
}

// findSuccessfulPatterns identifies successful trading patterns
func (e *DecisionEngine) findSuccessfulPatterns(decisions []*DecisionRecord) []TradingPattern {
	patterns := []TradingPattern{}

	// Pattern 1: Long positions during uptrends
	longWins := 0
	longProfit := 0.0
	for _, record := range decisions {
		if record.Decision != nil && record.Decision.Action == ActionOpenLong {
			if len(record.PositionSnapshots) > 0 {
				for _, pos := range record.PositionSnapshots {
					if pos.UnrealizedPnL > 0 {
						longWins++
						longProfit += pos.UnrealizedPnLPct
					}
				}
			}
		}
	}
	if longWins >= 3 {
		patterns = append(patterns, TradingPattern{
			Description:      "Long positions with positive momentum",
			WinCount:         longWins,
			AvgProfitPercent: longProfit / float64(longWins),
		})
	}

	// Pattern 2: High confidence decisions
	highConfWins := 0
	highConfProfit := 0.0
	for _, record := range decisions {
		if record.Decision != nil && record.Decision.Confidence >= 0.7 {
			if len(record.PositionSnapshots) > 0 {
				for _, pos := range record.PositionSnapshots {
					if pos.UnrealizedPnL > 0 {
						highConfWins++
						highConfProfit += pos.UnrealizedPnLPct
					}
				}
			}
		}
	}
	if highConfWins >= 3 {
		patterns = append(patterns, TradingPattern{
			Description:      "High confidence decisions (>70%)",
			WinCount:         highConfWins,
			AvgProfitPercent: highConfProfit / float64(highConfWins),
		})
	}

	// Pattern 3: Taking profit at target levels
	tpWins := 0
	tpProfit := 0.0
	for _, record := range decisions {
		if record.Decision != nil && record.Decision.TakeProfit > 0 {
			if len(record.PositionSnapshots) > 0 {
				for _, pos := range record.PositionSnapshots {
					if pos.UnrealizedPnL > 0 {
						tpWins++
						tpProfit += pos.UnrealizedPnLPct
					}
				}
			}
		}
	}
	if tpWins >= 3 {
		patterns = append(patterns, TradingPattern{
			Description:      "Trades with clear take-profit targets",
			WinCount:         tpWins,
			AvgProfitPercent: tpProfit / float64(tpWins),
		})
	}

	return patterns
}

// findFailedPatterns identifies failed trading patterns to avoid
func (e *DecisionEngine) findFailedPatterns(decisions []*DecisionRecord) []TradingPattern {
	patterns := []TradingPattern{}

	// Pattern 1: Short positions during uptrends
	shortLosses := 0
	shortLoss := 0.0
	for _, record := range decisions {
		if record.Decision != nil && record.Decision.Action == ActionOpenShort {
			if len(record.PositionSnapshots) > 0 {
				for _, pos := range record.PositionSnapshots {
					if pos.UnrealizedPnL < 0 {
						shortLosses++
						shortLoss += math.Abs(pos.UnrealizedPnLPct)
					}
				}
			}
		}
	}
	if shortLosses >= 3 {
		patterns = append(patterns, TradingPattern{
			Description:    "Counter-trend short positions",
			LossCount:      shortLosses,
			AvgLossPercent: shortLoss / float64(shortLosses),
		})
	}

	// Pattern 2: Low confidence trades
	lowConfLosses := 0
	lowConfLoss := 0.0
	for _, record := range decisions {
		if record.Decision != nil && record.Decision.Confidence < 0.5 {
			if len(record.PositionSnapshots) > 0 {
				for _, pos := range record.PositionSnapshots {
					if pos.UnrealizedPnL < 0 {
						lowConfLosses++
						lowConfLoss += math.Abs(pos.UnrealizedPnLPct)
					}
				}
			}
		}
	}
	if lowConfLosses >= 3 {
		patterns = append(patterns, TradingPattern{
			Description:    "Low confidence decisions (<50%)",
			LossCount:      lowConfLosses,
			AvgLossPercent: lowConfLoss / float64(lowConfLosses),
		})
	}

	// Pattern 3: No stop loss protection
	noSLLosses := 0
	noSLLoss := 0.0
	for _, record := range decisions {
		if record.Decision != nil && record.Decision.StopLoss == 0 {
			if len(record.PositionSnapshots) > 0 {
				for _, pos := range record.PositionSnapshots {
					if pos.UnrealizedPnL < 0 {
						noSLLosses++
						noSLLoss += math.Abs(pos.UnrealizedPnLPct)
					}
				}
			}
		}
	}
	if noSLLosses >= 2 {
		patterns = append(patterns, TradingPattern{
			Description:    "Trades without stop-loss protection",
			LossCount:      noSLLosses,
			AvgLossPercent: noSLLoss / float64(noSLLosses),
		})
	}

	// Pattern 4: Revenge trading (trading right after losses)
	revengeLosses := 0
	revengeLoss := 0.0
	for i := 1; i < len(decisions); i++ {
		prevRecord := decisions[i-1]
		currRecord := decisions[i]

		if prevRecord.Decision != nil && currRecord.Decision != nil {
			// Check if previous trade was a loss
			prevPnL := 0.0
			if len(prevRecord.PositionSnapshots) > 0 {
				for _, pos := range prevRecord.PositionSnapshots {
					prevPnL += pos.UnrealizedPnL
				}
			}

			// Check if current trade is also a loss and came quickly after
			if prevPnL < 0 && currRecord.Timestamp.Sub(prevRecord.Timestamp) < 30*time.Minute {
				currPnL := 0.0
				currPnLPct := 0.0
				if len(currRecord.PositionSnapshots) > 0 {
					for _, pos := range currRecord.PositionSnapshots {
						currPnL += pos.UnrealizedPnL
						currPnLPct += pos.UnrealizedPnLPct
					}
				}
				if currPnL < 0 {
					revengeLosses++
					revengeLoss += math.Abs(currPnLPct)
				}
			}
		}
	}
	if revengeLosses >= 2 {
		patterns = append(patterns, TradingPattern{
			Description:    "Revenge trading (trading immediately after losses)",
			LossCount:      revengeLosses,
			AvgLossPercent: revengeLoss / float64(revengeLosses),
		})
	}

	return patterns
}

// formatDuration formats duration in human-readable format
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "< 1 minute"
	} else if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%.1f hours", d.Hours())
	} else {
		return fmt.Sprintf("%.1f days", d.Hours()/24)
	}
}
