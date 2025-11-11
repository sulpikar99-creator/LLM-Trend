package analytics

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// TradeRecord represents a completed trade
type TradeRecord struct {
	Symbol      string    `json:"symbol"`
	Side        string    `json:"side"` // "LONG" or "SHORT"
	EntryPrice  float64   `json:"entry_price"`
	ExitPrice   float64   `json:"exit_price"`
	Size        float64   `json:"size"`
	EntryTime   time.Time `json:"entry_time"`
	ExitTime    time.Time `json:"exit_time"`
	PnL         float64   `json:"pnl"`
	PnLPercent  float64   `json:"pnl_percent"`
	Commission  float64   `json:"commission"`
	NetPnL      float64   `json:"net_pnl"`
	Duration    float64   `json:"duration_hours"`
	Strategy    string    `json:"strategy,omitempty"`
	TradeID     string    `json:"trade_id"`
}

// PerformanceMetrics contains comprehensive performance statistics
type PerformanceMetrics struct {
	// Overall metrics
	TotalTrades        int     `json:"total_trades"`
	WinningTrades      int     `json:"winning_trades"`
	LosingTrades       int     `json:"losing_trades"`
	BreakEvenTrades    int     `json:"break_even_trades"`
	WinRate            float64 `json:"win_rate"`

	// P&L metrics
	TotalPnL           float64 `json:"total_pnl"`
	TotalPnLPercent    float64 `json:"total_pnl_percent"`
	GrossProfitrate    float64 `json:"gross_profit"`
	GrossLoss          float64 `json:"gross_loss"`
	NetProfit          float64 `json:"net_profit"`
	ProfitFactor       float64 `json:"profit_factor"`

	// Trade statistics
	AvgWin             float64 `json:"avg_win"`
	AvgLoss            float64 `json:"avg_loss"`
	AvgWinPercent      float64 `json:"avg_win_percent"`
	AvgLossPercent     float64 `json:"avg_loss_percent"`
	LargestWin         float64 `json:"largest_win"`
	LargestLoss        float64 `json:"largest_loss"`
	AvgTrade           float64 `json:"avg_trade"`

	// Risk metrics
	SharpeRatio        float64 `json:"sharpe_ratio"`
	SortinoRatio       float64 `json:"sortino_ratio"`
	CalmarRatio        float64 `json:"calmar_ratio"`
	MaxDrawdown        float64 `json:"max_drawdown"`
	AvgDrawdown        float64 `json:"avg_drawdown"`

	// Time metrics
	AvgTradeDuration   float64 `json:"avg_trade_duration_hours"`
	AvgWinDuration     float64 `json:"avg_win_duration_hours"`
	AvgLossDuration    float64 `json:"avg_loss_duration_hours"`

	// Streak metrics
	LongestWinStreak   int     `json:"longest_win_streak"`
	LongestLossStreak  int     `json:"longest_loss_streak"`
	CurrentStreak      int     `json:"current_streak"`
	StreakType         string  `json:"streak_type"` // "win" or "loss"

	// Additional metrics
	ExpectancyPerTrade float64 `json:"expectancy_per_trade"`
	RRRatio            float64 `json:"rr_ratio"` // Risk/Reward Ratio
	TotalCommissions   float64 `json:"total_commissions"`
}

// PerformanceAttribution breaks down performance by various dimensions
type PerformanceAttribution struct {
	BySymbol   map[string]*SymbolPerformance   `json:"by_symbol"`
	ByStrategy map[string]*StrategyPerformance `json:"by_strategy"`
	ByTimeframe map[string]*TimeframePerformance `json:"by_timeframe"`
	BySide     map[string]*SidePerformance     `json:"by_side"`
	Overall    *PerformanceMetrics             `json:"overall"`
}

// SymbolPerformance contains performance metrics for a specific symbol
type SymbolPerformance struct {
	Symbol      string              `json:"symbol"`
	Metrics     *PerformanceMetrics `json:"metrics"`
	Trades      []*TradeRecord      `json:"trades,omitempty"`
	Contribution float64            `json:"contribution_pct"` // % of total P&L
}

// StrategyPerformance contains performance metrics for a specific strategy
type StrategyPerformance struct {
	Strategy    string              `json:"strategy"`
	Metrics     *PerformanceMetrics `json:"metrics"`
	Trades      []*TradeRecord      `json:"trades,omitempty"`
	Contribution float64            `json:"contribution_pct"`
}

// TimeframePerformance contains performance metrics for a timeframe
type TimeframePerformance struct {
	Period      string              `json:"period"` // "daily", "weekly", "monthly"
	Metrics     *PerformanceMetrics `json:"metrics"`
	StartDate   time.Time           `json:"start_date"`
	EndDate     time.Time           `json:"end_date"`
}

// SidePerformance contains performance metrics by trade direction
type SidePerformance struct {
	Side        string              `json:"side"` // "LONG" or "SHORT"
	Metrics     *PerformanceMetrics `json:"metrics"`
	Trades      []*TradeRecord      `json:"trades,omitempty"`
}

// CalculatePerformanceMetrics calculates comprehensive performance metrics
func CalculatePerformanceMetrics(trades []*TradeRecord) *PerformanceMetrics {
	if len(trades) == 0 {
		return &PerformanceMetrics{}
	}

	metrics := &PerformanceMetrics{
		TotalTrades: len(trades),
	}

	var (
		totalPnL         = 0.0
		totalPnLPercent  = 0.0
		grossProfit      = 0.0
		grossLoss        = 0.0
		winningTrades    = 0
		losingTrades     = 0
		breakEvenTrades  = 0
		winSum           = 0.0
		lossSum          = 0.0
		winPercentSum    = 0.0
		lossPercentSum   = 0.0
		largestWin       = 0.0
		largestLoss      = 0.0
		winDurationSum   = 0.0
		lossDurationSum  = 0.0
		durationSum      = 0.0
		returns          = make([]float64, 0)
		commissions      = 0.0
		currentStreak    = 0
		longestWinStreak = 0
		longestLossStreak = 0
		lastWasWin       = false
	)

	for _, trade := range trades {
		netPnL := trade.NetPnL
		pnlPercent := trade.PnLPercent

		totalPnL += netPnL
		totalPnLPercent += pnlPercent
		commissions += trade.Commission
		durationSum += trade.Duration
		returns = append(returns, pnlPercent/100) // Convert to decimal

		if netPnL > 0 {
			winningTrades++
			grossProfit += netPnL
			winSum += netPnL
			winPercentSum += pnlPercent
			winDurationSum += trade.Duration

			if netPnL > largestWin {
				largestWin = netPnL
			}

			// Update streak
			if lastWasWin {
				currentStreak++
			} else {
				currentStreak = 1
				lastWasWin = true
			}

			if currentStreak > longestWinStreak {
				longestWinStreak = currentStreak
			}

		} else if netPnL < 0 {
			losingTrades++
			grossLoss += math.Abs(netPnL)
			lossSum += math.Abs(netPnL)
			lossPercentSum += math.Abs(pnlPercent)
			lossDurationSum += trade.Duration

			if netPnL < largestLoss {
				largestLoss = netPnL
			}

			// Update streak
			if !lastWasWin {
				currentStreak++
			} else {
				currentStreak = 1
				lastWasWin = false
			}

			if currentStreak > longestLossStreak {
				longestLossStreak = currentStreak
			}

		} else {
			breakEvenTrades++
		}
	}

	// Calculate averages
	metrics.WinningTrades = winningTrades
	metrics.LosingTrades = losingTrades
	metrics.BreakEvenTrades = breakEvenTrades
	metrics.WinRate = float64(winningTrades) / float64(len(trades))

	metrics.TotalPnL = totalPnL
	metrics.TotalPnLPercent = totalPnLPercent
	metrics.GrossProfitrate = grossProfit
	metrics.GrossLoss = grossLoss
	metrics.NetProfit = totalPnL
	metrics.TotalCommissions = commissions

	if grossLoss > 0 {
		metrics.ProfitFactor = grossProfit / grossLoss
	} else if grossProfit > 0 {
		metrics.ProfitFactor = math.Inf(1)
	}

	if winningTrades > 0 {
		metrics.AvgWin = winSum / float64(winningTrades)
		metrics.AvgWinPercent = winPercentSum / float64(winningTrades)
		metrics.AvgWinDuration = winDurationSum / float64(winningTrades)
	}

	if losingTrades > 0 {
		metrics.AvgLoss = lossSum / float64(losingTrades)
		metrics.AvgLossPercent = lossPercentSum / float64(losingTrades)
		metrics.AvgLossDuration = lossDurationSum / float64(losingTrades)
	}

	metrics.LargestWin = largestWin
	metrics.LargestLoss = largestLoss
	metrics.AvgTrade = totalPnL / float64(len(trades))
	metrics.AvgTradeDuration = durationSum / float64(len(trades))

	// Calculate expectancy
	winProb := float64(winningTrades) / float64(len(trades))
	lossProb := float64(losingTrades) / float64(len(trades))
	metrics.ExpectancyPerTrade = (winProb * metrics.AvgWin) - (lossProb * metrics.AvgLoss)

	// Risk/Reward ratio
	if metrics.AvgLoss > 0 {
		metrics.RRRatio = metrics.AvgWin / metrics.AvgLoss
	}

	// Streaks
	metrics.LongestWinStreak = longestWinStreak
	metrics.LongestLossStreak = longestLossStreak
	metrics.CurrentStreak = currentStreak
	if lastWasWin {
		metrics.StreakType = "win"
	} else {
		metrics.StreakType = "loss"
	}

	// Calculate risk metrics
	if len(returns) > 0 {
		metrics.SharpeRatio = CalculateSharpeRatio(returns, 0)
		metrics.SortinoRatio = CalculateSortinoRatio(returns, 0)
	}

	return metrics
}

// CalculatePerformanceAttribution calculates performance attribution
func CalculatePerformanceAttribution(trades []*TradeRecord) *PerformanceAttribution {
	pa := &PerformanceAttribution{
		BySymbol:   make(map[string]*SymbolPerformance),
		ByStrategy: make(map[string]*StrategyPerformance),
		BySide:     make(map[string]*SidePerformance),
		ByTimeframe: make(map[string]*TimeframePerformance),
	}

	// Overall metrics
	pa.Overall = CalculatePerformanceMetrics(trades)

	// Group trades by symbol
	symbolTrades := make(map[string][]*TradeRecord)
	for _, trade := range trades {
		symbolTrades[trade.Symbol] = append(symbolTrades[trade.Symbol], trade)
	}

	// Calculate per-symbol metrics
	for symbol, symTrades := range symbolTrades {
		metrics := CalculatePerformanceMetrics(symTrades)
		contribution := 0.0
		if pa.Overall.TotalPnL != 0 {
			contribution = (metrics.TotalPnL / pa.Overall.TotalPnL) * 100
		}

		pa.BySymbol[symbol] = &SymbolPerformance{
			Symbol:       symbol,
			Metrics:      metrics,
			Trades:       symTrades,
			Contribution: contribution,
		}
	}

	// Group trades by strategy
	strategyTrades := make(map[string][]*TradeRecord)
	for _, trade := range trades {
		strategy := trade.Strategy
		if strategy == "" {
			strategy = "default"
		}
		strategyTrades[strategy] = append(strategyTrades[strategy], trade)
	}

	// Calculate per-strategy metrics
	for strategy, stratTrades := range strategyTrades {
		metrics := CalculatePerformanceMetrics(stratTrades)
		contribution := 0.0
		if pa.Overall.TotalPnL != 0 {
			contribution = (metrics.TotalPnL / pa.Overall.TotalPnL) * 100
		}

		pa.ByStrategy[strategy] = &StrategyPerformance{
			Strategy:     strategy,
			Metrics:      metrics,
			Trades:       stratTrades,
			Contribution: contribution,
		}
	}

	// Group trades by side
	sideTrades := make(map[string][]*TradeRecord)
	for _, trade := range trades {
		sideTrades[trade.Side] = append(sideTrades[trade.Side], trade)
	}

	// Calculate per-side metrics
	for side, sideTrds := range sideTrades {
		metrics := CalculatePerformanceMetrics(sideTrds)
		pa.BySide[side] = &SidePerformance{
			Side:   side,
			Metrics: metrics,
			Trades: sideTrds,
		}
	}

	// Group trades by timeframe (daily, weekly, monthly)
	pa.calculateTimeframePerformance(trades)

	return pa
}

// calculateTimeframePerformance calculates performance for different timeframes
func (pa *PerformanceAttribution) calculateTimeframePerformance(trades []*TradeRecord) {
	if len(trades) == 0 {
		return
	}

	// Sort trades by exit time
	sortedTrades := make([]*TradeRecord, len(trades))
	copy(sortedTrades, trades)
	sort.Slice(sortedTrades, func(i, j int) bool {
		return sortedTrades[i].ExitTime.Before(sortedTrades[j].ExitTime)
	})

	// Group by day
	dailyTrades := make(map[string][]*TradeRecord)
	for _, trade := range sortedTrades {
		day := trade.ExitTime.Format("2006-01-02")
		dailyTrades[day] = append(dailyTrades[day], trade)
	}

	// Calculate daily metrics (store only if needed)
	// For now, we'll just calculate overall periods

	// Weekly performance (last 4 weeks)
	now := time.Now()
	weekAgo := now.AddDate(0, 0, -7)
	var weekTrades []*TradeRecord
	for _, trade := range sortedTrades {
		if trade.ExitTime.After(weekAgo) {
			weekTrades = append(weekTrades, trade)
		}
	}
	if len(weekTrades) > 0 {
		pa.ByTimeframe["weekly"] = &TimeframePerformance{
			Period:    "weekly",
			Metrics:   CalculatePerformanceMetrics(weekTrades),
			StartDate: weekAgo,
			EndDate:   now,
		}
	}

	// Monthly performance
	monthAgo := now.AddDate(0, -1, 0)
	var monthTrades []*TradeRecord
	for _, trade := range sortedTrades {
		if trade.ExitTime.After(monthAgo) {
			monthTrades = append(monthTrades, trade)
		}
	}
	if len(monthTrades) > 0 {
		pa.ByTimeframe["monthly"] = &TimeframePerformance{
			Period:    "monthly",
			Metrics:   CalculatePerformanceMetrics(monthTrades),
			StartDate: monthAgo,
			EndDate:   now,
		}
	}
}

// GetTopPerformers returns top N symbols by P&L
func (pa *PerformanceAttribution) GetTopPerformers(n int) []*SymbolPerformance {
	symbols := make([]*SymbolPerformance, 0, len(pa.BySymbol))
	for _, perf := range pa.BySymbol {
		symbols = append(symbols, perf)
	}

	sort.Slice(symbols, func(i, j int) bool {
		return symbols[i].Metrics.TotalPnL > symbols[j].Metrics.TotalPnL
	})

	if n > len(symbols) {
		n = len(symbols)
	}

	return symbols[:n]
}

// GetWorstPerformers returns worst N symbols by P&L
func (pa *PerformanceAttribution) GetWorstPerformers(n int) []*SymbolPerformance {
	symbols := make([]*SymbolPerformance, 0, len(pa.BySymbol))
	for _, perf := range pa.BySymbol {
		symbols = append(symbols, perf)
	}

	sort.Slice(symbols, func(i, j int) bool {
		return symbols[i].Metrics.TotalPnL < symbols[j].Metrics.TotalPnL
	})

	if n > len(symbols) {
		n = len(symbols)
	}

	return symbols[:n]
}

// GenerateSummaryReport creates a text summary of performance
func (pm *PerformanceMetrics) GenerateSummaryReport() string {
	return fmt.Sprintf(`Performance Summary:
===================
Total Trades: %d (Wins: %d, Losses: %d, Break-even: %d)
Win Rate: %.2f%%
Net P&L: $%.2f (%.2f%%)
Profit Factor: %.2f
Expectancy: $%.2f per trade

Average Win: $%.2f (%.2f%%)
Average Loss: $%.2f (%.2f%%)
Risk/Reward: %.2f

Largest Win: $%.2f
Largest Loss: $%.2f

Win Streak: %d
Loss Streak: %d
Current Streak: %d %s

Sharpe Ratio: %.3f
Sortino Ratio: %.3f
`,
		pm.TotalTrades, pm.WinningTrades, pm.LosingTrades, pm.BreakEvenTrades,
		pm.WinRate*100,
		pm.NetProfit, pm.TotalPnLPercent,
		pm.ProfitFactor,
		pm.ExpectancyPerTrade,
		pm.AvgWin, pm.AvgWinPercent,
		pm.AvgLoss, pm.AvgLossPercent,
		pm.RRRatio,
		pm.LargestWin,
		pm.LargestLoss,
		pm.LongestWinStreak,
		pm.LongestLossStreak,
		pm.CurrentStreak, pm.StreakType,
		pm.SharpeRatio,
		pm.SortinoRatio,
	)
}
