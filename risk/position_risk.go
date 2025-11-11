package risk

import (
	"fmt"
	"sync"
)

// PositionRiskConfig contains position-level risk parameters
type PositionRiskConfig struct {
	// Stop-loss settings
	UseStopLoss          bool    `json:"use_stop_loss"`
	DefaultStopLossPercent float64 `json:"default_stop_loss_percent"` // e.g., 2.0 for 2%
	MinStopLossPercent   float64 `json:"min_stop_loss_percent"`      // Minimum allowed SL
	MaxStopLossPercent   float64 `json:"max_stop_loss_percent"`      // Maximum allowed SL

	// Take-profit settings
	UseTakeProfit          bool    `json:"use_take_profit"`
	DefaultTakeProfitPercent float64 `json:"default_take_profit_percent"` // e.g., 4.0 for 4%
	MinTakeProfitPercent   float64 `json:"min_take_profit_percent"`
	MaxTakeProfitPercent   float64 `json:"max_take_profit_percent"`

	// Trailing stop settings
	UseTrailingStop        bool    `json:"use_trailing_stop"`
	TrailingStopPercent    float64 `json:"trailing_stop_percent"`     // e.g., 1.5 for 1.5%
	TrailingStopActivation float64 `json:"trailing_stop_activation"`  // Activate after X% profit

	// Risk/Reward settings
	MinRiskRewardRatio float64 `json:"min_risk_reward_ratio"` // e.g., 1.5 for 1:1.5
	MaxRiskRewardRatio float64 `json:"max_risk_reward_ratio"` // e.g., 5.0 for 1:5

	// Position sizing
	UseFixedSize       bool    `json:"use_fixed_size"`
	FixedSizeUSD       float64 `json:"fixed_size_usd"`
	UseRiskBasedSizing bool    `json:"use_risk_based_sizing"`
}

// PositionRisk contains risk parameters for a specific position
type PositionRisk struct {
	Symbol            string  `json:"symbol"`
	Side              string  `json:"side"` // "LONG" or "SHORT"
	EntryPrice        float64 `json:"entry_price"`
	Size              float64 `json:"size"`
	Leverage          int     `json:"leverage"`
	StopLoss          float64 `json:"stop_loss"`
	TakeProfit        float64 `json:"take_profit"`
	TrailingStopPrice float64 `json:"trailing_stop_price"`
	HighestPrice      float64 `json:"highest_price"` // For long positions
	LowestPrice       float64 `json:"lowest_price"`  // For short positions
	RiskAmount        float64 `json:"risk_amount"`
	RewardAmount      float64 `json:"reward_amount"`
	RiskRewardRatio   float64 `json:"risk_reward_ratio"`
	IsTrailingActive  bool    `json:"is_trailing_active"`
	mu                sync.RWMutex
}

// PositionRiskManager manages position-level risk
type PositionRiskManager struct {
	config    *PositionRiskConfig
	positions map[string]*PositionRisk // key: symbol
	mu        sync.RWMutex
}

// NewPositionRiskManager creates a new position risk manager
func NewPositionRiskManager(config *PositionRiskConfig) *PositionRiskManager {
	if config == nil {
		config = DefaultPositionRiskConfig()
	}

	return &PositionRiskManager{
		config:    config,
		positions: make(map[string]*PositionRisk),
	}
}

// DefaultPositionRiskConfig returns sensible defaults
func DefaultPositionRiskConfig() *PositionRiskConfig {
	return &PositionRiskConfig{
		UseStopLoss:              true,
		DefaultStopLossPercent:   2.0,
		MinStopLossPercent:       0.5,
		MaxStopLossPercent:       10.0,
		UseTakeProfit:            true,
		DefaultTakeProfitPercent: 4.0,
		MinTakeProfitPercent:     1.0,
		MaxTakeProfitPercent:     20.0,
		UseTrailingStop:          false,
		TrailingStopPercent:      1.5,
		TrailingStopActivation:   2.0,
		MinRiskRewardRatio:       1.5,
		MaxRiskRewardRatio:       5.0,
		UseFixedSize:             false,
		FixedSizeUSD:             1000.0,
		UseRiskBasedSizing:       true,
	}
}

// CalculateStopLoss calculates stop-loss price
func (prm *PositionRiskManager) CalculateStopLoss(entryPrice float64, side string, customPercent float64) (float64, error) {
	prm.mu.RLock()
	defer prm.mu.RUnlock()

	if entryPrice <= 0 {
		return 0, fmt.Errorf("invalid entry price")
	}

	stopLossPercent := prm.config.DefaultStopLossPercent
	if customPercent > 0 {
		// Validate custom percent
		if customPercent < prm.config.MinStopLossPercent || customPercent > prm.config.MaxStopLossPercent {
			return 0, fmt.Errorf("stop-loss percent %.2f%% is outside allowed range (%.2f%% - %.2f%%)",
				customPercent, prm.config.MinStopLossPercent, prm.config.MaxStopLossPercent)
		}
		stopLossPercent = customPercent
	}

	stopLoss := 0.0
	if side == "LONG" {
		// For long, SL is below entry
		stopLoss = entryPrice * (1.0 - stopLossPercent/100.0)
	} else if side == "SHORT" {
		// For short, SL is above entry
		stopLoss = entryPrice * (1.0 + stopLossPercent/100.0)
	} else {
		return 0, fmt.Errorf("invalid side: %s", side)
	}

	return stopLoss, nil
}

// CalculateTakeProfit calculates take-profit price
func (prm *PositionRiskManager) CalculateTakeProfit(entryPrice float64, side string, customPercent float64) (float64, error) {
	prm.mu.RLock()
	defer prm.mu.RUnlock()

	if entryPrice <= 0 {
		return 0, fmt.Errorf("invalid entry price")
	}

	takeProfitPercent := prm.config.DefaultTakeProfitPercent
	if customPercent > 0 {
		// Validate custom percent
		if customPercent < prm.config.MinTakeProfitPercent || customPercent > prm.config.MaxTakeProfitPercent {
			return 0, fmt.Errorf("take-profit percent %.2f%% is outside allowed range (%.2f%% - %.2f%%)",
				customPercent, prm.config.MinTakeProfitPercent, prm.config.MaxTakeProfitPercent)
		}
		takeProfitPercent = customPercent
	}

	takeProfit := 0.0
	if side == "LONG" {
		// For long, TP is above entry
		takeProfit = entryPrice * (1.0 + takeProfitPercent/100.0)
	} else if side == "SHORT" {
		// For short, TP is below entry
		takeProfit = entryPrice * (1.0 - takeProfitPercent/100.0)
	} else {
		return 0, fmt.Errorf("invalid side: %s", side)
	}

	return takeProfit, nil
}

// ValidateRiskReward validates risk/reward ratio
func (prm *PositionRiskManager) ValidateRiskReward(entryPrice, stopLoss, takeProfit float64, side string) (bool, float64, string) {
	prm.mu.RLock()
	defer prm.mu.RUnlock()

	if entryPrice <= 0 || stopLoss <= 0 || takeProfit <= 0 {
		return false, 0, "Invalid prices"
	}

	risk := 0.0
	reward := 0.0

	if side == "LONG" {
		risk = entryPrice - stopLoss
		reward = takeProfit - entryPrice
	} else if side == "SHORT" {
		risk = stopLoss - entryPrice
		reward = entryPrice - takeProfit
	} else {
		return false, 0, "Invalid side"
	}

	if risk <= 0 {
		return false, 0, "Stop-loss must be on the opposite side of entry"
	}

	if reward <= 0 {
		return false, 0, "Take-profit must be profitable"
	}

	rrRatio := reward / risk

	if rrRatio < prm.config.MinRiskRewardRatio {
		return false, rrRatio, fmt.Sprintf("Risk/reward ratio %.2f is below minimum %.2f",
			rrRatio, prm.config.MinRiskRewardRatio)
	}

	if rrRatio > prm.config.MaxRiskRewardRatio {
		return false, rrRatio, fmt.Sprintf("Risk/reward ratio %.2f exceeds maximum %.2f",
			rrRatio, prm.config.MaxRiskRewardRatio)
	}

	return true, rrRatio, ""
}

// CreatePositionRisk creates risk parameters for a new position
func (prm *PositionRiskManager) CreatePositionRisk(symbol, side string, entryPrice, size float64, leverage int,
	customSLPercent, customTPPercent float64) (*PositionRisk, error) {

	prm.mu.Lock()
	defer prm.mu.Unlock()

	// Calculate stop-loss
	stopLoss := 0.0
	var err error
	if prm.config.UseStopLoss {
		stopLoss, err = prm.calculateStopLossInternal(entryPrice, side, customSLPercent)
		if err != nil {
			return nil, err
		}
	}

	// Calculate take-profit
	takeProfit := 0.0
	if prm.config.UseTakeProfit {
		takeProfit, err = prm.calculateTakeProfitInternal(entryPrice, side, customTPPercent)
		if err != nil {
			return nil, err
		}
	}

	// Validate risk/reward if both SL and TP are set
	rrRatio := 0.0
	if stopLoss > 0 && takeProfit > 0 {
		valid, ratio, msg := prm.validateRiskRewardInternal(entryPrice, stopLoss, takeProfit, side)
		if !valid {
			return nil, fmt.Errorf("risk/reward validation failed: %s", msg)
		}
		rrRatio = ratio
	}

	// Calculate risk and reward amounts
	riskAmount := 0.0
	rewardAmount := 0.0

	if side == "LONG" {
		if stopLoss > 0 {
			riskAmount = (entryPrice - stopLoss) * size
		}
		if takeProfit > 0 {
			rewardAmount = (takeProfit - entryPrice) * size
		}
	} else if side == "SHORT" {
		if stopLoss > 0 {
			riskAmount = (stopLoss - entryPrice) * size
		}
		if takeProfit > 0 {
			rewardAmount = (entryPrice - takeProfit) * size
		}
	}

	posRisk := &PositionRisk{
		Symbol:            symbol,
		Side:              side,
		EntryPrice:        entryPrice,
		Size:              size,
		Leverage:          leverage,
		StopLoss:          stopLoss,
		TakeProfit:        takeProfit,
		TrailingStopPrice: 0,
		HighestPrice:      entryPrice,
		LowestPrice:       entryPrice,
		RiskAmount:        riskAmount,
		RewardAmount:      rewardAmount,
		RiskRewardRatio:   rrRatio,
		IsTrailingActive:  false,
	}

	prm.positions[symbol] = posRisk
	return posRisk, nil
}

// Internal methods (must be called with lock held)

func (prm *PositionRiskManager) calculateStopLossInternal(entryPrice float64, side string, customPercent float64) (float64, error) {
	stopLossPercent := prm.config.DefaultStopLossPercent
	if customPercent > 0 {
		stopLossPercent = customPercent
	}

	if side == "LONG" {
		return entryPrice * (1.0 - stopLossPercent/100.0), nil
	} else if side == "SHORT" {
		return entryPrice * (1.0 + stopLossPercent/100.0), nil
	}
	return 0, fmt.Errorf("invalid side: %s", side)
}

func (prm *PositionRiskManager) calculateTakeProfitInternal(entryPrice float64, side string, customPercent float64) (float64, error) {
	takeProfitPercent := prm.config.DefaultTakeProfitPercent
	if customPercent > 0 {
		takeProfitPercent = customPercent
	}

	if side == "LONG" {
		return entryPrice * (1.0 + takeProfitPercent/100.0), nil
	} else if side == "SHORT" {
		return entryPrice * (1.0 - takeProfitPercent/100.0), nil
	}
	return 0, fmt.Errorf("invalid side: %s", side)
}

func (prm *PositionRiskManager) validateRiskRewardInternal(entryPrice, stopLoss, takeProfit float64, side string) (bool, float64, string) {
	risk := 0.0
	reward := 0.0

	if side == "LONG" {
		risk = entryPrice - stopLoss
		reward = takeProfit - entryPrice
	} else if side == "SHORT" {
		risk = stopLoss - entryPrice
		reward = entryPrice - takeProfit
	}

	if risk <= 0 || reward <= 0 {
		return false, 0, "Invalid risk/reward calculation"
	}

	rrRatio := reward / risk

	if rrRatio < prm.config.MinRiskRewardRatio {
		return false, rrRatio, fmt.Sprintf("R/R ratio %.2f below minimum", rrRatio)
	}

	return true, rrRatio, ""
}

// UpdatePositionPrice updates position with current price and manages trailing stop
func (prm *PositionRiskManager) UpdatePositionPrice(symbol string, currentPrice float64) (*PositionRisk, bool) {
	prm.mu.Lock()
	defer prm.mu.Unlock()

	pos, exists := prm.positions[symbol]
	if !exists {
		return nil, false
	}

	pos.mu.Lock()
	defer pos.mu.Unlock()

	// Update highest/lowest prices
	if currentPrice > pos.HighestPrice {
		pos.HighestPrice = currentPrice
	}
	if currentPrice < pos.LowestPrice {
		pos.LowestPrice = currentPrice
	}

	// Handle trailing stop for LONG positions
	if prm.config.UseTrailingStop && pos.Side == "LONG" {
		profitPercent := ((currentPrice - pos.EntryPrice) / pos.EntryPrice) * 100

		// Activate trailing stop if profit threshold reached
		if !pos.IsTrailingActive && profitPercent >= prm.config.TrailingStopActivation {
			pos.IsTrailingActive = true
			pos.TrailingStopPrice = currentPrice * (1.0 - prm.config.TrailingStopPercent/100.0)
		}

		// Update trailing stop if active
		if pos.IsTrailingActive {
			newTrailingStop := pos.HighestPrice * (1.0 - prm.config.TrailingStopPercent/100.0)
			if newTrailingStop > pos.TrailingStopPrice {
				pos.TrailingStopPrice = newTrailingStop
			}

			// Check if trailing stop hit
			if currentPrice <= pos.TrailingStopPrice {
				return pos, true // Should close position
			}
		}
	}

	// Handle trailing stop for SHORT positions
	if prm.config.UseTrailingStop && pos.Side == "SHORT" {
		profitPercent := ((pos.EntryPrice - currentPrice) / pos.EntryPrice) * 100

		// Activate trailing stop if profit threshold reached
		if !pos.IsTrailingActive && profitPercent >= prm.config.TrailingStopActivation {
			pos.IsTrailingActive = true
			pos.TrailingStopPrice = currentPrice * (1.0 + prm.config.TrailingStopPercent/100.0)
		}

		// Update trailing stop if active
		if pos.IsTrailingActive {
			newTrailingStop := pos.LowestPrice * (1.0 + prm.config.TrailingStopPercent/100.0)
			if newTrailingStop < pos.TrailingStopPrice {
				pos.TrailingStopPrice = newTrailingStop
			}

			// Check if trailing stop hit
			if currentPrice >= pos.TrailingStopPrice {
				return pos, true // Should close position
			}
		}
	}

	// Check regular stop-loss
	if pos.StopLoss > 0 {
		if (pos.Side == "LONG" && currentPrice <= pos.StopLoss) ||
			(pos.Side == "SHORT" && currentPrice >= pos.StopLoss) {
			return pos, true // Should close position
		}
	}

	// Check take-profit
	if pos.TakeProfit > 0 {
		if (pos.Side == "LONG" && currentPrice >= pos.TakeProfit) ||
			(pos.Side == "SHORT" && currentPrice <= pos.TakeProfit) {
			return pos, true // Should close position
		}
	}

	return pos, false
}

// RemovePosition removes a position from tracking
func (prm *PositionRiskManager) RemovePosition(symbol string) {
	prm.mu.Lock()
	defer prm.mu.Unlock()

	delete(prm.positions, symbol)
}

// GetPosition returns position risk data
func (prm *PositionRiskManager) GetPosition(symbol string) (*PositionRisk, bool) {
	prm.mu.RLock()
	defer prm.mu.RUnlock()

	pos, exists := prm.positions[symbol]
	return pos, exists
}

// GetAllPositions returns all tracked positions
func (prm *PositionRiskManager) GetAllPositions() map[string]*PositionRisk {
	prm.mu.RLock()
	defer prm.mu.RUnlock()

	// Return a copy to prevent external modification
	positions := make(map[string]*PositionRisk, len(prm.positions))
	for k, v := range prm.positions {
		positions[k] = v
	}
	return positions
}

// UpdateConfig updates position risk configuration
func (prm *PositionRiskManager) UpdateConfig(config *PositionRiskConfig) {
	prm.mu.Lock()
	defer prm.mu.Unlock()

	prm.config = config
}

// GetConfig returns current configuration
func (prm *PositionRiskManager) GetConfig() *PositionRiskConfig {
	prm.mu.RLock()
	defer prm.mu.RUnlock()

	return prm.config
}
