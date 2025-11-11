package manager

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/sulpikar99-creator/LLM-Trend/decision"
	"github.com/sulpikar99-creator/LLM-Trend/logger"
	"github.com/sulpikar99-creator/LLM-Trend/market"
	"github.com/sulpikar99-creator/LLM-Trend/risk"
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
	ID               string
	UserID           string
	Trader           trader.Trader
	KlineMonitor     *market.KlineMonitor
	DecisionEngine   *decision.DecisionEngine
	DecisionLogger   *logger.DecisionLogger
	AccountRisk      *risk.AccountRiskMonitor
	PositionRisk     *risk.PositionRiskManager
	Status           TraderStatus
	ErrorMessage     string
	AutoTrading      bool
	DecisionInterval time.Duration
	SystemPrompt     string
	StrategyPrompt   string
	Symbol           string
	MaxPositionUSD   float64
	MaxLeverage      int
	DB               *sql.DB // Database connection for stats
	ctx              context.Context
	cancel           context.CancelFunc
	mu               sync.RWMutex
}

// TraderManager manages multiple traders
type TraderManager struct {
	traders map[string]*ManagedTrader
	db      *sql.DB
	mu      sync.RWMutex
}

// NewTraderManager creates a new trader manager
func NewTraderManager() *TraderManager {
	return &TraderManager{
		traders: make(map[string]*ManagedTrader),
		db:      nil, // Will be set later if needed
	}
}

// SetDatabase sets the database connection for the trader manager
func (tm *TraderManager) SetDatabase(db *sql.DB) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.db = db

	// Update existing traders
	for _, mt := range tm.traders {
		mt.DB = db
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
		DB:     tm.db,
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
		// Store symbol for later reference
		mt.Symbol = symbol

		// Determine if testnet based on exchange config or trader implementation
		testnet := false

		// Try to detect from BinanceFuturesTrader baseURL
		if bf, ok := mt.Trader.(*trader.BinanceFuturesTrader); ok {
			// Access private field through reflection or use type assertion
			// For now, check exchange type name
			exchType := bf.GetExchangeType()
			testnet = strings.Contains(strings.ToLower(exchType), "testnet")
		}

		// If we have database access, try to get from exchange_config
		if !testnet && mt.DB != nil {
			var exchangeConfigJSON sql.NullString
			err := mt.DB.QueryRow("SELECT exchange_config FROM traders WHERE id = ?", mt.ID).Scan(&exchangeConfigJSON)
			if err == nil && exchangeConfigJSON.Valid {
				var config map[string]interface{}
				if json.Unmarshal([]byte(exchangeConfigJSON.String), &config) == nil {
					if testnetVal, ok := config["testnet"].(bool); ok {
						testnet = testnetVal
					}
				}
			}
		}

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

	// Get symbol from stored value or kline monitor
	symbol := mt.Symbol
	if symbol == "" {
		symbol = "UNKNOWN"
	}

	md := &market.MarketData{
		Symbol:             symbol,
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

// EnableAutoTrading enables automated trading with AI decision engine
func (mt *ManagedTrader) EnableAutoTrading(aiConfig *decision.AIConfig, systemPrompt, strategyPrompt string, decisionInterval time.Duration) error {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	if mt.Status != TraderStatusRunning {
		return fmt.Errorf("trader must be running to enable auto-trading")
	}

	if mt.AutoTrading {
		return fmt.Errorf("auto-trading already enabled")
	}

	// Initialize decision engine
	mt.DecisionEngine = decision.NewDecisionEngine(aiConfig)

	// Initialize decision logger
	mt.DecisionLogger = logger.NewDecisionLogger("decision_logs")

	// Set configuration
	mt.SystemPrompt = systemPrompt
	mt.StrategyPrompt = strategyPrompt
	mt.DecisionInterval = decisionInterval

	// Start decision loop
	mt.AutoTrading = true
	go mt.decisionLoop()

	log.Printf("Auto-trading enabled for trader %s with %v interval", mt.ID, decisionInterval)
	return nil
}

// DisableAutoTrading disables automated trading
func (mt *ManagedTrader) DisableAutoTrading() error {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	if !mt.AutoTrading {
		return fmt.Errorf("auto-trading not enabled")
	}

	mt.AutoTrading = false
	log.Printf("Auto-trading disabled for trader %s", mt.ID)
	return nil
}

// decisionLoop runs the AI decision-making loop
func (mt *ManagedTrader) decisionLoop() {
	ticker := time.NewTicker(mt.DecisionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-mt.ctx.Done():
			return

		case <-ticker.C:
			// Check if auto-trading is still enabled
			mt.mu.RLock()
			enabled := mt.AutoTrading
			mt.mu.RUnlock()

			if !enabled {
				return
			}

			// Make decision
			if err := mt.makeAndExecuteDecision(); err != nil {
				log.Printf("Error in decision cycle for trader %s: %v", mt.ID, err)
			}
		}
	}
}

// makeAndExecuteDecision makes an AI decision and executes it
func (mt *ManagedTrader) makeAndExecuteDecision() error {
	ctx := context.Background()

	// Gather context
	decisionCtx, err := mt.gatherDecisionContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to gather context: %w", err)
	}

	// Make decision
	record, err := mt.DecisionEngine.MakeDecision(ctx, decisionCtx, mt.SystemPrompt, mt.StrategyPrompt)
	if err != nil {
		// Log even failed decisions
		if mt.DecisionLogger != nil {
			mt.DecisionLogger.LogDecision(mt.ID, record)
		}
		return fmt.Errorf("decision failed: %w", err)
	}

	// Log decision
	if mt.DecisionLogger != nil {
		if err := mt.DecisionLogger.LogDecision(mt.ID, record); err != nil {
			log.Printf("Failed to log decision: %v", err)
		}
	}

	// Execute decision
	if record.Decision != nil {
		if err := mt.executeDecision(ctx, record.Decision); err != nil {
			log.Printf("Failed to execute decision: %v", err)
			return err
		}
	}

	return nil
}

// gatherDecisionContext gathers all context needed for decision-making
func (mt *ManagedTrader) gatherDecisionContext(ctx context.Context) (*decision.DecisionContext, error) {
	// Get market data
	marketData, err := mt.GetMarketData()
	if err != nil {
		return nil, fmt.Errorf("failed to get market data: %w", err)
	}

	// Get balance
	balance, err := mt.GetBalance(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	// Get positions
	positions, err := mt.GetPositions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}

	// Get klines
	klines := []*market.Kline{}
	if mt.KlineMonitor != nil {
		klines = mt.KlineMonitor.GetKlines()
	}

	// Get recent decisions (last 10)
	recentDecisions := []*decision.DecisionRecord{}
	if mt.DecisionLogger != nil {
		decisions, err := mt.DecisionLogger.GetLatestDecisions(mt.ID, 10)
		if err == nil {
			// Convert to DecisionRecord (simplified)
			for _, d := range decisions {
				// We can parse these back if needed
				_ = d
			}
		}
	}

	// Calculate TotalTrades and WinRate from database
	totalTrades := 0
	winRate := 0.0

	if mt.DB != nil {
		// Count total trades from decision_records
		var count int
		err := mt.DB.QueryRow(`
			SELECT COUNT(*)
			FROM decision_records
			WHERE trader_id = ?
			AND decision_json IS NOT NULL
			AND decision_json LIKE '%"action":"open_%'
		`, mt.ID).Scan(&count)
		if err == nil {
			totalTrades = count
		}

		// Calculate win rate from execution logs or performance records
		if totalTrades > 0 {
			var wins int
			err = mt.DB.QueryRow(`
				SELECT COUNT(*)
				FROM decision_records
				WHERE trader_id = ?
				AND execution_logs IS NOT NULL
				AND execution_logs LIKE '%"success":true%'
			`, mt.ID).Scan(&wins)
			if err == nil && totalTrades > 0 {
				winRate = float64(wins) / float64(totalTrades) * 100
			}
		}
	}

	return &decision.DecisionContext{
		MarketData:      marketData,
		Klines:          klines,
		Balance:         balance,
		Positions:       positions,
		Symbol:          mt.Symbol,
		MaxPositionUSD:  mt.MaxPositionUSD,
		MaxLeverage:     mt.MaxLeverage,
		RiskPercent:     2.0, // Default 2% risk per trade
		RecentDecisions: recentDecisions,
		TotalTrades:     totalTrades,
		WinRate:         winRate,
		TotalPnL:        balance.UnrealizedProfit,
	}, nil
}

// executeDecision executes a trading decision
func (mt *ManagedTrader) executeDecision(ctx context.Context, dec *decision.TradingDecision) error {
	log.Printf("Executing decision: %s for trader %s", dec.Action, mt.ID)

	switch dec.Action {
	case decision.ActionOpenLong:
		return mt.openPosition(ctx, dec, trader.OrderSideBuy)

	case decision.ActionOpenShort:
		return mt.openPosition(ctx, dec, trader.OrderSideSell)

	case decision.ActionClose:
		return mt.closeAllPositions(ctx)

	case decision.ActionHold:
		log.Printf("Decision: HOLD - no action taken")
		return nil

	case decision.ActionUpdateSLTP:
		return mt.updateStopLossTakeProfit(ctx, dec)

	default:
		return fmt.Errorf("unknown action: %s", dec.Action)
	}
}

// openPosition opens a new position
func (mt *ManagedTrader) openPosition(ctx context.Context, dec *decision.TradingDecision, side trader.OrderSide) error {
	// Risk management checks
	if mt.AccountRisk != nil {
		// Check if trading is allowed
		canTrade, reason := mt.AccountRisk.CanTrade()
		if !canTrade {
			return fmt.Errorf("risk check failed: %s", reason)
		}

		// Get current positions
		positions, err := mt.Trader.GetPositions(ctx)
		if err != nil {
			log.Printf("Warning: failed to get positions for risk check: %v", err)
		}

		totalExposure := 0.0
		for _, pos := range positions {
			totalExposure += pos.Notional
		}

		// Get current market price for size estimation
		balance, err := mt.Trader.GetBalance(ctx)
		if err == nil && mt.AccountRisk != nil {
			mt.AccountRisk.UpdateBalance(balance.Balance)
		}

		// Estimate position size (using decision size as approximation)
		positionSizeUSD := dec.Size * 50000.0 // Rough estimate, will be refined
		canOpen, reason := mt.AccountRisk.CanOpenPosition(positionSizeUSD, len(positions), totalExposure)
		if !canOpen {
			return fmt.Errorf("risk check failed: %s", reason)
		}

		// Validate leverage
		canUseLeverage, reason := mt.AccountRisk.ValidateLeverage(dec.Symbol, dec.Leverage)
		if !canUseLeverage {
			return fmt.Errorf("leverage check failed: %s", reason)
		}

		// Record trade
		mt.AccountRisk.RecordTrade()
	}

	// Set leverage first
	if err := mt.Trader.SetLeverage(ctx, dec.Symbol, dec.Leverage); err != nil {
		log.Printf("Warning: failed to set leverage: %v", err)
	}

	// Place market order
	orderReq := &trader.OrderRequest{
		Symbol:   dec.Symbol,
		Side:     side,
		Type:     trader.OrderTypeMarket,
		Quantity: dec.Size,
	}

	order, err := mt.Trader.PlaceOrder(ctx, orderReq)
	if err != nil {
		return fmt.Errorf("failed to place order: %w", err)
	}

	log.Printf("Order placed: %s %s %.4f @ market, ID: %s", dec.Symbol, side, dec.Size, order.OrderID)

	// Place stop-loss if specified
	if dec.StopLoss > 0 {
		stopSide := trader.OrderSideSell
		if side == trader.OrderSideSell {
			stopSide = trader.OrderSideBuy
		}

		stopReq := &trader.OrderRequest{
			Symbol:     dec.Symbol,
			Side:       stopSide,
			Type:       trader.OrderTypeStopLoss,
			Quantity:   dec.Size,
			StopPrice:  dec.StopLoss,
			ReduceOnly: true,
		}

		if _, err := mt.Trader.PlaceOrder(ctx, stopReq); err != nil {
			log.Printf("Warning: failed to place stop-loss: %v", err)
		}
	}

	// Place take-profit if specified
	if dec.TakeProfit > 0 {
		tpSide := trader.OrderSideSell
		if side == trader.OrderSideSell {
			tpSide = trader.OrderSideBuy
		}

		tpReq := &trader.OrderRequest{
			Symbol:     dec.Symbol,
			Side:       tpSide,
			Type:       trader.OrderTypeTakeProfit,
			Quantity:   dec.Size,
			StopPrice:  dec.TakeProfit,
			ReduceOnly: true,
		}

		if _, err := mt.Trader.PlaceOrder(ctx, tpReq); err != nil {
			log.Printf("Warning: failed to place take-profit: %v", err)
		}
	}

	return nil
}

// closeAllPositions closes all open positions
func (mt *ManagedTrader) closeAllPositions(ctx context.Context) error {
	positions, err := mt.GetPositions(ctx)
	if err != nil {
		return err
	}

	for _, pos := range positions {
		side := trader.OrderSideSell
		if pos.Side == trader.PositionSideShort {
			side = trader.OrderSideBuy
		}

		orderReq := &trader.OrderRequest{
			Symbol:     pos.Symbol,
			Side:       side,
			Type:       trader.OrderTypeMarket,
			Quantity:   pos.Size,
			ReduceOnly: true,
		}

		if _, err := mt.Trader.PlaceOrder(ctx, orderReq); err != nil {
			log.Printf("Failed to close position %s: %v", pos.Symbol, err)
			continue
		}

		log.Printf("Closed position: %s %s %.4f", pos.Symbol, pos.Side, pos.Size)
	}

	return nil
}

// updateStopLossTakeProfit updates SL/TP for existing positions
func (mt *ManagedTrader) updateStopLossTakeProfit(ctx context.Context, dec *decision.TradingDecision) error {
	// This would require more complex order management
	// For now, just log
	log.Printf("Update SL/TP requested: SL=%.2f, TP=%.2f", dec.StopLoss, dec.TakeProfit)
	return nil
}
