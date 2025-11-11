package risk

import (
	"fmt"
	"sync"
	"time"
)

// AccountRiskConfig contains account-level risk parameters
type AccountRiskConfig struct {
	// Drawdown limits
	MaxDrawdownPercent float64 `json:"max_drawdown_percent"` // e.g., 20.0 for 20%
	MaxDailyLossPercent float64 `json:"max_daily_loss_percent"` // e.g., 5.0 for 5%
	MaxDailyLossAmount  float64 `json:"max_daily_loss_amount"`  // Absolute amount

	// Position limits
	MaxPositionSizeUSD float64 `json:"max_position_size_usd"` // Max size per position
	MaxTotalPositionUSD float64 `json:"max_total_position_usd"` // Max total exposure
	MaxOpenPositions   int     `json:"max_open_positions"`    // Max number of positions

	// Leverage limits
	MaxLeverage     int     `json:"max_leverage"`      // Global max leverage
	MaxLeveragePerSymbol map[string]int `json:"max_leverage_per_symbol,omitempty"` // Symbol-specific limits

	// Trading limits
	MaxTradesPerDay   int     `json:"max_trades_per_day"`   // Daily trade limit
	MinTimeBetweenTrades float64 `json:"min_time_between_trades_minutes"` // Cooldown period

	// Risk percentage
	RiskPerTradePercent float64 `json:"risk_per_trade_percent"` // e.g., 2.0 for 2%

	// Auto-stop
	AutoStopOnMaxDrawdown bool `json:"auto_stop_on_max_drawdown"`
	AutoStopOnDailyLimit  bool `json:"auto_stop_on_daily_limit"`
}

// AccountRiskMonitor monitors account-level risk metrics
type AccountRiskMonitor struct {
	config            *AccountRiskConfig
	initialBalance    float64
	currentBalance    float64
	peakBalance       float64
	dailyStartBalance float64
	dailyPnL          float64
	dailyTradeCount   int
	lastTradeTime     time.Time
	currentDay        string
	violations        []RiskViolation
	isStopped         bool
	mu                sync.RWMutex
}

// RiskViolation represents a risk rule violation
type RiskViolation struct {
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"`        // "max_drawdown", "daily_loss", etc.
	Message     string    `json:"message"`
	Severity    string    `json:"severity"`    // "warning", "critical"
	Value       float64   `json:"value"`
	Limit       float64   `json:"limit"`
	AutoStopped bool      `json:"auto_stopped"`
}

// RiskStatus contains current risk metrics
type RiskStatus struct {
	CurrentDrawdown        float64          `json:"current_drawdown"`
	CurrentDrawdownPercent float64          `json:"current_drawdown_percent"`
	DailyPnL               float64          `json:"daily_pnl"`
	DailyPnLPercent        float64          `json:"daily_pnl_percent"`
	DailyTradeCount        int              `json:"daily_trade_count"`
	TimeSinceLastTrade     float64          `json:"time_since_last_trade_minutes"`
	OpenPositions          int              `json:"open_positions"`
	TotalExposureUSD       float64          `json:"total_exposure_usd"`
	IsHealthy              bool             `json:"is_healthy"`
	Violations             []RiskViolation  `json:"violations"`
	IsStopped              bool             `json:"is_stopped"`
	Limits                 *AccountRiskConfig `json:"limits"`
}

// NewAccountRiskMonitor creates a new account risk monitor
func NewAccountRiskMonitor(config *AccountRiskConfig, initialBalance float64) *AccountRiskMonitor {
	if config == nil {
		config = DefaultAccountRiskConfig()
	}

	now := time.Now()
	return &AccountRiskMonitor{
		config:            config,
		initialBalance:    initialBalance,
		currentBalance:    initialBalance,
		peakBalance:       initialBalance,
		dailyStartBalance: initialBalance,
		dailyPnL:          0,
		dailyTradeCount:   0,
		lastTradeTime:     now,
		currentDay:        now.Format("2006-01-02"),
		violations:        []RiskViolation{},
		isStopped:         false,
	}
}

// DefaultAccountRiskConfig returns sensible defaults
func DefaultAccountRiskConfig() *AccountRiskConfig {
	return &AccountRiskConfig{
		MaxDrawdownPercent:      20.0, // 20% max drawdown
		MaxDailyLossPercent:     5.0,  // 5% daily loss limit
		MaxDailyLossAmount:      0,    // No absolute limit by default
		MaxPositionSizeUSD:      10000.0,
		MaxTotalPositionUSD:     50000.0,
		MaxOpenPositions:        5,
		MaxLeverage:             10,
		MaxLeveragePerSymbol:    make(map[string]int),
		MaxTradesPerDay:         50,
		MinTimeBetweenTrades:    1.0, // 1 minute cooldown
		RiskPerTradePercent:     2.0,  // 2% risk per trade
		AutoStopOnMaxDrawdown:   true,
		AutoStopOnDailyLimit:    true,
	}
}

// UpdateBalance updates the current balance and checks for violations
func (arm *AccountRiskMonitor) UpdateBalance(newBalance float64) error {
	arm.mu.Lock()
	defer arm.mu.Unlock()

	arm.currentBalance = newBalance

	// Update peak
	if newBalance > arm.peakBalance {
		arm.peakBalance = newBalance
	}

	// Check if new day
	today := time.Now().Format("2006-01-02")
	if today != arm.currentDay {
		arm.resetDailyMetrics(newBalance)
		arm.currentDay = today
	}

	// Update daily P&L
	arm.dailyPnL = newBalance - arm.dailyStartBalance

	// Check for violations
	violations := arm.checkViolations()
	if len(violations) > 0 {
		arm.violations = append(arm.violations, violations...)

		// Check if we should auto-stop
		for _, v := range violations {
			if v.AutoStopped {
				arm.isStopped = true
				return fmt.Errorf("trading stopped due to %s violation: %s", v.Type, v.Message)
			}
		}
	}

	return nil
}

// resetDailyMetrics resets daily tracking metrics
func (arm *AccountRiskMonitor) resetDailyMetrics(balance float64) {
	arm.dailyStartBalance = balance
	arm.dailyPnL = 0
	arm.dailyTradeCount = 0
}

// checkViolations checks for risk violations
func (arm *AccountRiskMonitor) checkViolations() []RiskViolation {
	var violations []RiskViolation
	now := time.Now()

	// Check max drawdown
	if arm.peakBalance > 0 {
		drawdown := arm.peakBalance - arm.currentBalance
		drawdownPercent := (drawdown / arm.peakBalance) * 100

		if drawdownPercent > arm.config.MaxDrawdownPercent {
			violations = append(violations, RiskViolation{
				Timestamp:   now,
				Type:        "max_drawdown",
				Message:     fmt.Sprintf("Max drawdown exceeded: %.2f%% (limit: %.2f%%)", drawdownPercent, arm.config.MaxDrawdownPercent),
				Severity:    "critical",
				Value:       drawdownPercent,
				Limit:       arm.config.MaxDrawdownPercent,
				AutoStopped: arm.config.AutoStopOnMaxDrawdown,
			})
		}
	}

	// Check daily loss
	if arm.dailyStartBalance > 0 {
		dailyLossPercent := (arm.dailyPnL / arm.dailyStartBalance) * 100

		if dailyLossPercent < -arm.config.MaxDailyLossPercent {
			violations = append(violations, RiskViolation{
				Timestamp:   now,
				Type:        "daily_loss_percent",
				Message:     fmt.Sprintf("Daily loss limit exceeded: %.2f%% (limit: %.2f%%)", dailyLossPercent, arm.config.MaxDailyLossPercent),
				Severity:    "critical",
				Value:       dailyLossPercent,
				Limit:       -arm.config.MaxDailyLossPercent,
				AutoStopped: arm.config.AutoStopOnDailyLimit,
			})
		}

		// Check absolute daily loss
		if arm.config.MaxDailyLossAmount > 0 && arm.dailyPnL < -arm.config.MaxDailyLossAmount {
			violations = append(violations, RiskViolation{
				Timestamp:   now,
				Type:        "daily_loss_amount",
				Message:     fmt.Sprintf("Daily loss amount exceeded: $%.2f (limit: $%.2f)", arm.dailyPnL, arm.config.MaxDailyLossAmount),
				Severity:    "critical",
				Value:       arm.dailyPnL,
				Limit:       -arm.config.MaxDailyLossAmount,
				AutoStopped: arm.config.AutoStopOnDailyLimit,
			})
		}
	}

	// Check daily trade count
	if arm.dailyTradeCount >= arm.config.MaxTradesPerDay {
		violations = append(violations, RiskViolation{
			Timestamp:   now,
			Type:        "max_trades_per_day",
			Message:     fmt.Sprintf("Daily trade limit reached: %d (limit: %d)", arm.dailyTradeCount, arm.config.MaxTradesPerDay),
			Severity:    "warning",
			Value:       float64(arm.dailyTradeCount),
			Limit:       float64(arm.config.MaxTradesPerDay),
			AutoStopped: false,
		})
	}

	return violations
}

// CanTrade checks if trading is allowed based on risk rules
func (arm *AccountRiskMonitor) CanTrade() (bool, string) {
	arm.mu.RLock()
	defer arm.mu.RUnlock()

	if arm.isStopped {
		return false, "Trading stopped due to risk violation"
	}

	// Check daily trade limit
	if arm.dailyTradeCount >= arm.config.MaxTradesPerDay {
		return false, fmt.Sprintf("Daily trade limit reached (%d/%d)", arm.dailyTradeCount, arm.config.MaxTradesPerDay)
	}

	// Check cooldown period
	timeSinceLastTrade := time.Since(arm.lastTradeTime).Minutes()
	if timeSinceLastTrade < arm.config.MinTimeBetweenTrades {
		return false, fmt.Sprintf("Cooldown period: %.1f minutes remaining", arm.config.MinTimeBetweenTrades-timeSinceLastTrade)
	}

	return true, ""
}

// RecordTrade records a new trade
func (arm *AccountRiskMonitor) RecordTrade() {
	arm.mu.Lock()
	defer arm.mu.Unlock()

	arm.dailyTradeCount++
	arm.lastTradeTime = time.Now()
}

// CanOpenPosition checks if a new position can be opened
func (arm *AccountRiskMonitor) CanOpenPosition(positionSizeUSD float64, currentPositions int, totalExposureUSD float64) (bool, string) {
	arm.mu.RLock()
	defer arm.mu.RUnlock()

	// Check position count
	if currentPositions >= arm.config.MaxOpenPositions {
		return false, fmt.Sprintf("Max open positions reached (%d/%d)", currentPositions, arm.config.MaxOpenPositions)
	}

	// Check position size
	if positionSizeUSD > arm.config.MaxPositionSizeUSD {
		return false, fmt.Sprintf("Position size exceeds limit: $%.2f (max: $%.2f)", positionSizeUSD, arm.config.MaxPositionSizeUSD)
	}

	// Check total exposure
	newTotalExposure := totalExposureUSD + positionSizeUSD
	if newTotalExposure > arm.config.MaxTotalPositionUSD {
		return false, fmt.Sprintf("Total exposure would exceed limit: $%.2f (max: $%.2f)", newTotalExposure, arm.config.MaxTotalPositionUSD)
	}

	return true, ""
}

// ValidateLeverage validates leverage for a symbol
func (arm *AccountRiskMonitor) ValidateLeverage(symbol string, leverage int) (bool, string) {
	arm.mu.RLock()
	defer arm.mu.RUnlock()

	// Check symbol-specific limit
	if symbolLimit, ok := arm.config.MaxLeveragePerSymbol[symbol]; ok {
		if leverage > symbolLimit {
			return false, fmt.Sprintf("Leverage exceeds symbol limit: %dx (max: %dx for %s)", leverage, symbolLimit, symbol)
		}
	}

	// Check global limit
	if leverage > arm.config.MaxLeverage {
		return false, fmt.Sprintf("Leverage exceeds global limit: %dx (max: %dx)", leverage, arm.config.MaxLeverage)
	}

	return true, ""
}

// CalculatePositionSize calculates safe position size based on risk parameters
func (arm *AccountRiskMonitor) CalculatePositionSize(entryPrice, stopLossPrice float64, leverage int) float64 {
	arm.mu.RLock()
	defer arm.mu.RUnlock()

	if arm.currentBalance <= 0 || entryPrice <= 0 || stopLossPrice <= 0 {
		return 0
	}

	// Calculate risk amount (e.g., 2% of balance)
	riskAmount := arm.currentBalance * (arm.config.RiskPerTradePercent / 100.0)

	// Calculate price difference percentage
	priceDiff := entryPrice - stopLossPrice
	if priceDiff <= 0 {
		return 0
	}

	priceDiffPercent := (priceDiff / entryPrice) * 100

	// Calculate position size
	// Risk Amount = Position Size * (Price Diff % / 100)
	positionSize := riskAmount / (priceDiffPercent / 100.0)

	// Apply leverage
	if leverage > 1 {
		positionSize = positionSize / float64(leverage)
	}

	// Cap at max position size
	if positionSize > arm.config.MaxPositionSizeUSD {
		positionSize = arm.config.MaxPositionSizeUSD
	}

	return positionSize
}

// GetStatus returns current risk status
func (arm *AccountRiskMonitor) GetStatus(openPositions int, totalExposure float64) *RiskStatus {
	arm.mu.RLock()
	defer arm.mu.RUnlock()

	drawdown := 0.0
	drawdownPercent := 0.0
	if arm.peakBalance > 0 {
		drawdown = arm.peakBalance - arm.currentBalance
		drawdownPercent = (drawdown / arm.peakBalance) * 100
	}

	dailyPnLPercent := 0.0
	if arm.dailyStartBalance > 0 {
		dailyPnLPercent = (arm.dailyPnL / arm.dailyStartBalance) * 100
	}

	timeSinceLastTrade := time.Since(arm.lastTradeTime).Minutes()

	// Check if healthy
	isHealthy := !arm.isStopped &&
		drawdownPercent < arm.config.MaxDrawdownPercent &&
		dailyPnLPercent > -arm.config.MaxDailyLossPercent &&
		arm.dailyTradeCount < arm.config.MaxTradesPerDay

	return &RiskStatus{
		CurrentDrawdown:        drawdown,
		CurrentDrawdownPercent: drawdownPercent,
		DailyPnL:               arm.dailyPnL,
		DailyPnLPercent:        dailyPnLPercent,
		DailyTradeCount:        arm.dailyTradeCount,
		TimeSinceLastTrade:     timeSinceLastTrade,
		OpenPositions:          openPositions,
		TotalExposureUSD:       totalExposure,
		IsHealthy:              isHealthy,
		Violations:             arm.violations,
		IsStopped:              arm.isStopped,
		Limits:                 arm.config,
	}
}

// Reset resets the stopped state (use with caution)
func (arm *AccountRiskMonitor) Reset() {
	arm.mu.Lock()
	defer arm.mu.Unlock()

	arm.isStopped = false
	arm.violations = []RiskViolation{}
}

// UpdateConfig updates risk configuration
func (arm *AccountRiskMonitor) UpdateConfig(config *AccountRiskConfig) {
	arm.mu.Lock()
	defer arm.mu.Unlock()

	arm.config = config
}

// GetConfig returns current configuration
func (arm *AccountRiskMonitor) GetConfig() *AccountRiskConfig {
	arm.mu.RLock()
	defer arm.mu.RUnlock()

	return arm.config
}
