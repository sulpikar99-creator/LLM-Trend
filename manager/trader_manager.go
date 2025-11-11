package manager

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/sulpikar99-creator/LLM-Trend/market"
	"github.com/sulpikar99-creator/LLM-Trend/trader"
)

// TraderStatus represents the status of a trader
type TraderStatus string

const (
	TraderStatusStopped TraderStatus = "stopped"
	TraderStatusStarting TraderStatus = "starting"
	TraderStatusRunning  TraderStatus = "running"
	TraderStatusStopping TraderStatus = "stopping"
	TraderStatusError    TraderStatus = "error"
)

// ManagedTrader wraps a trader with management capabilities
type ManagedTrader struct {
	ID            string
	UserID        string
	Trader        trader.Trader
	KlineMonitor  *market.KlineMonitor
	Status        TraderStatus
	ErrorMessage  string
	ctx           context.Context
	cancel        context.CancelFunc
	mu            sync.RWMutex
}

// TraderManager manages multiple traders
type TraderManager struct {
	traders map[string]*ManagedTrader
	mu      sync.RWMutex
}

// NewTraderManager creates a new trader manager
func NewTraderManager() *TraderManager {
	return &TraderManager{
		traders: make(map[string]*ManagedTrader),
	}
}

// AddTrader adds a trader to the manager
func (tm *TraderManager) AddTrader(id, userID string, t trader.Trader) (*ManagedTrader, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Check if trader already exists
	if _, exists := tm.traders[id]; exists {
		return nil, fmt.Errorf("trader with ID %s already exists", id)
	}

	ctx, cancel := context.WithCancel(context.Background())

	mt := &ManagedTrader{
		ID:     id,
		UserID: userID,
		Trader: t,
		Status: TraderStatusStopped,
		ctx:    ctx,
		cancel: cancel,
	}

	tm.traders[id] = mt

	log.Printf("Trader %s added for user %s", id, userID)
	return mt, nil
}

// RemoveTrader removes a trader from the manager
func (tm *TraderManager) RemoveTrader(id string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	mt, exists := tm.traders[id]
	if !exists {
		return fmt.Errorf("trader not found: %s", id)
	}

	// Stop trader if running
	if mt.Status == TraderStatusRunning {
		mt.cancel()
		if mt.KlineMonitor != nil {
			mt.KlineMonitor.Stop()
		}
		mt.Trader.Disconnect()
	}

	delete(tm.traders, id)
	log.Printf("Trader %s removed", id)
	return nil
}

// GetTrader returns a managed trader by ID
func (tm *TraderManager) GetTrader(id string) (*ManagedTrader, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	mt, exists := tm.traders[id]
	if !exists {
		return nil, fmt.Errorf("trader not found: %s", id)
	}

	return mt, nil
}

// ListTraders returns all traders for a user
func (tm *TraderManager) ListTraders(userID string) []*ManagedTrader {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	result := []*ManagedTrader{}
	for _, mt := range tm.traders {
		if mt.UserID == userID {
			result = append(result, mt)
		}
	}

	return result
}

// ListAllTraders returns all traders
func (tm *TraderManager) ListAllTraders() []*ManagedTrader {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	result := make([]*ManagedTrader, 0, len(tm.traders))
	for _, mt := range tm.traders {
		result = append(result, mt)
	}

	return result
}

// StartTrader starts a trader
func (tm *TraderManager) StartTrader(id, symbol, interval string) error {
	mt, err := tm.GetTrader(id)
	if err != nil {
		return err
	}

	mt.mu.Lock()
	defer mt.mu.Unlock()

	if mt.Status == TraderStatusRunning {
		return fmt.Errorf("trader is already running")
	}

	mt.Status = TraderStatusStarting
	mt.ErrorMessage = ""

	// Connect to exchange
	if err := mt.Trader.Connect(mt.ctx); err != nil {
		mt.Status = TraderStatusError
		mt.ErrorMessage = fmt.Sprintf("Failed to connect: %v", err)
		return err
	}

	// Start kline monitor if symbol specified
	if symbol != "" && interval != "" {
		// Determine if testnet based on trader type
		testnet := false // TODO: get from trader config
		mt.KlineMonitor = market.NewKlineMonitor(symbol, interval, testnet, 500)

		if err := mt.KlineMonitor.Start(); err != nil {
			mt.Status = TraderStatusError
			mt.ErrorMessage = fmt.Sprintf("Failed to start kline monitor: %v", err)
			mt.Trader.Disconnect()
			return err
		}
	}

	mt.Status = TraderStatusRunning
	log.Printf("Trader %s started", id)
	return nil
}

// StopTrader stops a trader
func (tm *TraderManager) StopTrader(id string) error {
	mt, err := tm.GetTrader(id)
	if err != nil {
		return err
	}

	mt.mu.Lock()
	defer mt.mu.Unlock()

	if mt.Status != TraderStatusRunning {
		return fmt.Errorf("trader is not running")
	}

	mt.Status = TraderStatusStopping

	// Stop kline monitor
	if mt.KlineMonitor != nil {
		if err := mt.KlineMonitor.Stop(); err != nil {
			log.Printf("Error stopping kline monitor: %v", err)
		}
		mt.KlineMonitor = nil
	}

	// Disconnect trader
	if err := mt.Trader.Disconnect(); err != nil {
		log.Printf("Error disconnecting trader: %v", err)
	}

	mt.Status = TraderStatusStopped
	log.Printf("Trader %s stopped", id)
	return nil
}

// GetTraderStatus returns the current status of a trader
func (mt *ManagedTrader) GetStatus() TraderStatus {
	mt.mu.RLock()
	defer mt.mu.RUnlock()
	return mt.Status
}

// GetTraderInfo returns trader information
func (mt *ManagedTrader) GetInfo() map[string]interface{} {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	info := map[string]interface{}{
		"id":        mt.ID,
		"user_id":   mt.UserID,
		"name":      mt.Trader.GetName(),
		"exchange":  mt.Trader.GetExchangeType(),
		"status":    string(mt.Status),
		"connected": mt.Trader.IsConnected(),
	}

	if mt.ErrorMessage != "" {
		info["error"] = mt.ErrorMessage
	}

	return info
}

// GetBalance gets trader balance
func (mt *ManagedTrader) GetBalance(ctx context.Context) (*trader.Balance, error) {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	if mt.Status != TraderStatusRunning {
		return nil, fmt.Errorf("trader is not running")
	}

	return mt.Trader.GetBalance(ctx)
}

// GetPositions gets trader positions
func (mt *ManagedTrader) GetPositions(ctx context.Context) ([]*trader.Position, error) {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	if mt.Status != TraderStatusRunning {
		return nil, fmt.Errorf("trader is not running")
	}

	return mt.Trader.GetPositions(ctx)
}

// PlaceOrder places an order
func (mt *ManagedTrader) PlaceOrder(ctx context.Context, req *trader.OrderRequest) (*trader.Order, error) {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	if mt.Status != TraderStatusRunning {
		return nil, fmt.Errorf("trader is not running")
	}

	return mt.Trader.PlaceOrder(ctx, req)
}

// GetMarketData gets current market data with indicators
func (mt *ManagedTrader) GetMarketData() (*market.MarketData, error) {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	if mt.KlineMonitor == nil {
		return nil, fmt.Errorf("kline monitor not initialized")
	}

	latestKline := mt.KlineMonitor.GetLatestKline()
	if latestKline == nil {
		return nil, fmt.Errorf("no kline data available")
	}

	indicators, err := mt.KlineMonitor.GetIndicators()
	if err != nil {
		return nil, fmt.Errorf("failed to calculate indicators: %w", err)
	}

	// Calculate 24h stats
	klines := mt.KlineMonitor.GetKlines()
	var highPrice24h, lowPrice24h float64
	if len(klines) > 0 {
		highPrice24h = klines[0].High
		lowPrice24h = klines[0].Low

		for _, k := range klines {
			if k.High > highPrice24h {
				highPrice24h = k.High
			}
			if k.Low < lowPrice24h {
				lowPrice24h = k.Low
			}
		}
	}

	priceChange24h := 0.0
	if len(klines) > 0 {
		priceChange24h = market.CalculatePriceChange(klines[0].Open, latestKline.Close)
	}

	md := &market.MarketData{
		Symbol:             "unknown", // TODO: get from kline monitor
		Price:              latestKline.Close,
		PriceChange24h:     latestKline.Close - klines[0].Open,
		PriceChangePercent: priceChange24h,
		Volume24h:          indicators.Volume24h,
		HighPrice24h:       highPrice24h,
		LowPrice24h:        lowPrice24h,
		LastUpdateTime:     latestKline.CloseTime,
		Indicators:         indicators,
	}

	return md, nil
}
