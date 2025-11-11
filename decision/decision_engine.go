package decision

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sulpikar99-creator/LLM-Trend/market"
	"github.com/sulpikar99-creator/LLM-Trend/trader"
)

// DecisionAction represents the action to take
type DecisionAction string

const (
	ActionOpenLong   DecisionAction = "open_long"
	ActionOpenShort  DecisionAction = "open_short"
	ActionClose      DecisionAction = "close_position"
	ActionHold       DecisionAction = "hold"
	ActionUpdateSLTP DecisionAction = "update_sl_tp"
)

// TradingDecision represents an AI trading decision
type TradingDecision struct {
	Action      DecisionAction `json:"action"`
	Symbol      string         `json:"symbol"`
	Size        float64        `json:"size,omitempty"`
	Leverage    int            `json:"leverage,omitempty"`
	StopLoss    float64        `json:"stop_loss,omitempty"`
	TakeProfit  float64        `json:"take_profit,omitempty"`
	Reasoning   string         `json:"reasoning"`
	Confidence  float64        `json:"confidence,omitempty"`
}

// DecisionContext contains context for making decisions
type DecisionContext struct {
	// Market data
	MarketData *market.MarketData
	Klines     []*market.Kline

	// Account state
	Balance   *trader.Balance
	Positions []*trader.Position

	// Configuration
	Symbol         string
	MaxPositionUSD float64
	MaxLeverage    int
	RiskPercent    float64

	// Historical data
	RecentDecisions []*DecisionRecord
	TotalTrades     int
	WinRate         float64
	TotalPnL        float64
}

// DecisionRecord represents a complete decision record for logging
type DecisionRecord struct {
	ID              string                 `json:"id"`
	TraderID        string                 `json:"trader_id"`
	CycleNumber     int                    `json:"cycle_number"`
	Timestamp       time.Time              `json:"timestamp"`
	SystemPrompt    string                 `json:"system_prompt"`
	InputPrompt     string                 `json:"input_prompt"`
	AIResponse      string                 `json:"ai_response"`
	DecisionJSON    string                 `json:"decision_json"`
	Decision        *TradingDecision       `json:"decision,omitempty"`
	AccountState    *AccountSnapshot       `json:"account_state"`
	PositionSnapshots []*PositionSnapshot  `json:"position_snapshots"`
	ExecutionLogs   []string               `json:"execution_logs"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
	AILatencyMs     int64                  `json:"ai_latency_ms"`
}

// AccountSnapshot represents account state at decision time
type AccountSnapshot struct {
	TotalBalance     float64 `json:"total_balance"`
	AvailableBalance float64 `json:"available_balance"`
	UnrealizedPnL    float64 `json:"unrealized_pnl"`
	MarginUsed       float64 `json:"margin_used"`
	MarginRatio      float64 `json:"margin_ratio"`
}

// PositionSnapshot represents a position at decision time
type PositionSnapshot struct {
	Symbol           string  `json:"symbol"`
	Side             string  `json:"side"`
	Size             float64 `json:"size"`
	EntryPrice       float64 `json:"entry_price"`
	CurrentPrice     float64 `json:"current_price"`
	UnrealizedPnL    float64 `json:"unrealized_pnl"`
	UnrealizedPnLPct float64 `json:"unrealized_pnl_pct"`
	Leverage         int     `json:"leverage"`
}

// DecisionEngine makes trading decisions using AI
type DecisionEngine struct {
	aiClient   *AIClient
	cycleCount int
}

// NewDecisionEngine creates a new decision engine
func NewDecisionEngine(aiConfig *AIConfig) *DecisionEngine {
	return &DecisionEngine{
		aiClient:   NewAIClient(aiConfig),
		cycleCount: 0,
	}
}

// MakeDecision makes a trading decision based on context
func (e *DecisionEngine) MakeDecision(ctx context.Context, decisionCtx *DecisionContext, systemPrompt, strategyPrompt string) (*DecisionRecord, error) {
	e.cycleCount++

	record := &DecisionRecord{
		ID:          fmt.Sprintf("decision_%d_%d", time.Now().Unix(), e.cycleCount),
		CycleNumber: e.cycleCount,
		Timestamp:   time.Now(),
		SystemPrompt: systemPrompt,
		ExecutionLogs: []string{},
	}

	// Create account snapshot
	record.AccountState = e.createAccountSnapshot(decisionCtx.Balance)

	// Create position snapshots
	record.PositionSnapshots = e.createPositionSnapshots(decisionCtx.Positions, decisionCtx.MarketData)

	// Build input prompt with context
	inputPrompt := e.buildInputPrompt(decisionCtx, strategyPrompt)
	record.InputPrompt = inputPrompt

	// Call AI
	startTime := time.Now()
	aiResponse, err := e.aiClient.GetCompletion(ctx, systemPrompt, inputPrompt)
	record.AILatencyMs = time.Since(startTime).Milliseconds()
	record.AIResponse = aiResponse

	if err != nil {
		record.ErrorMessage = fmt.Sprintf("AI request failed: %v", err)
		return record, err
	}

	// Extract and parse decision JSON
	decisionJSON := ExtractJSON(aiResponse)
	record.DecisionJSON = decisionJSON

	var decision TradingDecision
	if err := json.Unmarshal([]byte(decisionJSON), &decision); err != nil {
		record.ErrorMessage = fmt.Sprintf("Failed to parse decision JSON: %v", err)
		return record, fmt.Errorf("failed to parse decision: %w", err)
	}

	// Validate decision
	if err := e.validateDecision(&decision, decisionCtx); err != nil {
		record.ErrorMessage = fmt.Sprintf("Decision validation failed: %v", err)
		return record, err
	}

	record.Decision = &decision
	record.ExecutionLogs = append(record.ExecutionLogs, fmt.Sprintf("Decision validated: %s", decision.Action))

	return record, nil
}

// buildInputPrompt builds the input prompt with market context
func (e *DecisionEngine) buildInputPrompt(ctx *DecisionContext, strategyPrompt string) string {
	var sb strings.Builder

	// Strategy instructions
	if strategyPrompt != "" {
		sb.WriteString("## Strategy Instructions\n")
		sb.WriteString(strategyPrompt)
		sb.WriteString("\n\n")
	}

	// Current market data
	sb.WriteString("## Current Market Data\n")
	if ctx.MarketData != nil {
		sb.WriteString(fmt.Sprintf("Symbol: %s\n", ctx.Symbol))
		sb.WriteString(fmt.Sprintf("Current Price: %.2f\n", ctx.MarketData.Price))
		sb.WriteString(fmt.Sprintf("24h Change: %.2f%%\n", ctx.MarketData.PriceChangePercent))
		sb.WriteString(fmt.Sprintf("24h High: %.2f\n", ctx.MarketData.HighPrice24h))
		sb.WriteString(fmt.Sprintf("24h Low: %.2f\n", ctx.MarketData.LowPrice24h))
		sb.WriteString(fmt.Sprintf("24h Volume: %.2f\n", ctx.MarketData.Volume24h))

		if ctx.MarketData.Indicators != nil {
			sb.WriteString("\n### Technical Indicators\n")
			sb.WriteString(fmt.Sprintf("RSI (14): %.2f\n", ctx.MarketData.Indicators.RSI))
			sb.WriteString(fmt.Sprintf("MACD: %.2f\n", ctx.MarketData.Indicators.MACD))
			sb.WriteString(fmt.Sprintf("MACD Signal: %.2f\n", ctx.MarketData.Indicators.MACDSignal))
			sb.WriteString(fmt.Sprintf("MACD Histogram: %.2f\n", ctx.MarketData.Indicators.MACDHistogram))
			sb.WriteString(fmt.Sprintf("Bollinger Upper: %.2f\n", ctx.MarketData.Indicators.BollingerUpper))
			sb.WriteString(fmt.Sprintf("Bollinger Mid: %.2f\n", ctx.MarketData.Indicators.BollingerMid))
			sb.WriteString(fmt.Sprintf("Bollinger Lower: %.2f\n", ctx.MarketData.Indicators.BollingerLower))
			sb.WriteString(fmt.Sprintf("SMA 20: %.2f\n", ctx.MarketData.Indicators.SMA20))
			sb.WriteString(fmt.Sprintf("SMA 50: %.2f\n", ctx.MarketData.Indicators.SMA50))
			sb.WriteString(fmt.Sprintf("EMA 12: %.2f\n", ctx.MarketData.Indicators.EMA12))
			sb.WriteString(fmt.Sprintf("EMA 26: %.2f\n", ctx.MarketData.Indicators.EMA26))
		}
	}

	// Account state
	sb.WriteString("\n## Account State\n")
	if ctx.Balance != nil {
		sb.WriteString(fmt.Sprintf("Total Balance: %.2f USDT\n", ctx.Balance.Balance))
		sb.WriteString(fmt.Sprintf("Available Balance: %.2f USDT\n", ctx.Balance.AvailableBalance))
		sb.WriteString(fmt.Sprintf("Unrealized P&L: %.2f USDT\n", ctx.Balance.UnrealizedProfit))
	}

	// Current positions
	sb.WriteString("\n## Current Positions\n")
	if len(ctx.Positions) > 0 {
		for _, pos := range ctx.Positions {
			sb.WriteString(fmt.Sprintf("- %s %s: Size %.4f, Entry %.2f, Current %.2f, P&L %.2f USDT (%.2f%%)\n",
				pos.Symbol, pos.Side, pos.Size, pos.EntryPrice, pos.MarkPrice, pos.UnrealizedProfit,
				e.calculatePnLPercent(pos)))
		}
	} else {
		sb.WriteString("No open positions\n")
	}

	// Risk parameters
	sb.WriteString("\n## Risk Parameters\n")
	sb.WriteString(fmt.Sprintf("Max Position Size: %.2f USDT\n", ctx.MaxPositionUSD))
	sb.WriteString(fmt.Sprintf("Max Leverage: %dx\n", ctx.MaxLeverage))
	sb.WriteString(fmt.Sprintf("Risk Per Trade: %.2f%%\n", ctx.RiskPercent))

	// Performance stats
	if ctx.TotalTrades > 0 {
		sb.WriteString("\n## Performance Statistics\n")
		sb.WriteString(fmt.Sprintf("Total Trades: %d\n", ctx.TotalTrades))
		sb.WriteString(fmt.Sprintf("Win Rate: %.2f%%\n", ctx.WinRate))
		sb.WriteString(fmt.Sprintf("Total P&L: %.2f USDT\n", ctx.TotalPnL))
	}

	sb.WriteString("\n## Your Decision\n")
	sb.WriteString("Based on the above data, provide your trading decision in the following JSON format:\n")
	sb.WriteString("```json\n")
	sb.WriteString("{\n")
	sb.WriteString(`  "action": "open_long" | "open_short" | "close_position" | "hold",` + "\n")
	sb.WriteString(`  "symbol": "BTCUSDT",` + "\n")
	sb.WriteString(`  "size": 0.001,` + "\n")
	sb.WriteString(`  "leverage": 5,` + "\n")
	sb.WriteString(`  "stop_loss": 40000,` + "\n")
	sb.WriteString(`  "take_profit": 45000,` + "\n")
	sb.WriteString(`  "reasoning": "Explain your decision and the key factors..."` + "\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n")

	return sb.String()
}

// createAccountSnapshot creates account snapshot from balance
func (e *DecisionEngine) createAccountSnapshot(balance *trader.Balance) *AccountSnapshot {
	if balance == nil {
		return &AccountSnapshot{}
	}

	marginUsed := balance.Balance - balance.AvailableBalance
	marginRatio := 0.0
	if balance.Balance > 0 {
		marginRatio = (marginUsed / balance.Balance) * 100
	}

	return &AccountSnapshot{
		TotalBalance:     balance.Balance,
		AvailableBalance: balance.AvailableBalance,
		UnrealizedPnL:    balance.UnrealizedProfit,
		MarginUsed:       marginUsed,
		MarginRatio:      marginRatio,
	}
}

// createPositionSnapshots creates position snapshots
func (e *DecisionEngine) createPositionSnapshots(positions []*trader.Position, marketData *market.MarketData) []*PositionSnapshot {
	snapshots := []*PositionSnapshot{}

	for _, pos := range positions {
		currentPrice := pos.MarkPrice
		if marketData != nil && pos.Symbol == marketData.Symbol {
			currentPrice = marketData.Price
		}

		pnlPct := e.calculatePnLPercent(pos)

		snapshot := &PositionSnapshot{
			Symbol:           pos.Symbol,
			Side:             string(pos.Side),
			Size:             pos.Size,
			EntryPrice:       pos.EntryPrice,
			CurrentPrice:     currentPrice,
			UnrealizedPnL:    pos.UnrealizedProfit,
			UnrealizedPnLPct: pnlPct,
			Leverage:         pos.Leverage,
		}

		snapshots = append(snapshots, snapshot)
	}

	return snapshots
}

// calculatePnLPercent calculates P&L percentage
func (e *DecisionEngine) calculatePnLPercent(pos *trader.Position) float64 {
	if pos.EntryPrice == 0 {
		return 0
	}

	priceDiff := pos.MarkPrice - pos.EntryPrice
	if pos.Side == trader.PositionSideShort {
		priceDiff = -priceDiff
	}

	return (priceDiff / pos.EntryPrice) * 100 * float64(pos.Leverage)
}

// validateDecision validates a trading decision
func (e *DecisionEngine) validateDecision(decision *TradingDecision, ctx *DecisionContext) error {
	// Validate action
	switch decision.Action {
	case ActionOpenLong, ActionOpenShort:
		// Check if we have enough balance
		if ctx.Balance != nil && ctx.Balance.AvailableBalance <= 0 {
			return fmt.Errorf("insufficient balance")
		}

		// Validate size
		if decision.Size <= 0 {
			return fmt.Errorf("invalid position size: %.4f", decision.Size)
		}

		// Validate leverage
		if decision.Leverage <= 0 || decision.Leverage > ctx.MaxLeverage {
			return fmt.Errorf("invalid leverage: %d (max: %d)", decision.Leverage, ctx.MaxLeverage)
		}

		// Check position size limit
		notional := decision.Size * ctx.MarketData.Price * float64(decision.Leverage)
		if notional > ctx.MaxPositionUSD {
			return fmt.Errorf("position size %.2f exceeds maximum %.2f", notional, ctx.MaxPositionUSD)
		}

	case ActionClose, ActionHold, ActionUpdateSLTP:
		// These actions don't need extensive validation
	default:
		return fmt.Errorf("unknown action: %s", decision.Action)
	}

	return nil
}
