package market

import (
	"fmt"
	"math"
	"time"
)

// Kline represents a candlestick/kline data point
type Kline struct {
	OpenTime             time.Time `json:"open_time"`
	Open                 float64   `json:"open"`
	High                 float64   `json:"high"`
	Low                  float64   `json:"low"`
	Close                float64   `json:"close"`
	Volume               float64   `json:"volume"`
	CloseTime            time.Time `json:"close_time"`
	QuoteAssetVolume     float64   `json:"quote_asset_volume"`
	NumberOfTrades       int64     `json:"number_of_trades"`
	TakerBuyBaseVolume   float64   `json:"taker_buy_base_volume"`
	TakerBuyQuoteVolume  float64   `json:"taker_buy_quote_volume"`
}

// TechnicalIndicators holds calculated technical indicators
type TechnicalIndicators struct {
	RSI            float64   `json:"rsi"`
	MACD           float64   `json:"macd"`
	MACDSignal     float64   `json:"macd_signal"`
	MACDHistogram  float64   `json:"macd_histogram"`
	BollingerUpper float64   `json:"bollinger_upper"`
	BollingerMid   float64   `json:"bollinger_mid"`
	BollingerLower float64   `json:"bollinger_lower"`
	SMA20          float64   `json:"sma_20"`
	SMA50          float64   `json:"sma_50"`
	SMA200         float64   `json:"sma_200"`
	EMA12          float64   `json:"ema_12"`
	EMA26          float64   `json:"ema_26"`
	Volume24h      float64   `json:"volume_24h"`
	PriceChange24h float64   `json:"price_change_24h"`
	Timestamp      time.Time `json:"timestamp"`
}

// MarketData represents complete market data for a symbol
type MarketData struct {
	Symbol             string               `json:"symbol"`
	Price              float64              `json:"price"`
	PriceChange24h     float64              `json:"price_change_24h"`
	PriceChangePercent float64              `json:"price_change_percent"`
	Volume24h          float64              `json:"volume_24h"`
	HighPrice24h       float64              `json:"high_price_24h"`
	LowPrice24h        float64              `json:"low_price_24h"`
	LastUpdateTime     time.Time            `json:"last_update_time"`
	Indicators         *TechnicalIndicators `json:"indicators,omitempty"`
}

// CalculateRSI calculates Relative Strength Index
// Uses safe division to prevent panic
func CalculateRSI(prices []float64, period int) (float64, error) {
	if len(prices) < period+1 {
		return 0, fmt.Errorf("insufficient data: need at least %d prices, got %d", period+1, len(prices))
	}

	gains := 0.0
	losses := 0.0

	// Calculate initial average gain/loss
	for i := 1; i <= period; i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			gains += change
		} else {
			losses += math.Abs(change)
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	// Calculate RSI
	if avgLoss == 0 {
		return 100, nil // All gains, RSI = 100
	}

	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))

	return rsi, nil
}

// CalculateSMA calculates Simple Moving Average
// Uses safe division to prevent panic
func CalculateSMA(prices []float64, period int) (float64, error) {
	if len(prices) < period {
		return 0, fmt.Errorf("insufficient data: need at least %d prices, got %d", period, len(prices))
	}

	sum := 0.0
	for i := len(prices) - period; i < len(prices); i++ {
		sum += prices[i]
	}

	if period == 0 {
		return 0, fmt.Errorf("period cannot be zero")
	}

	return sum / float64(period), nil
}

// CalculateEMA calculates Exponential Moving Average
func CalculateEMA(prices []float64, period int) (float64, error) {
	if len(prices) < period {
		return 0, fmt.Errorf("insufficient data: need at least %d prices, got %d", period, len(prices))
	}

	multiplier := 2.0 / float64(period+1)

	// Start with SMA
	sma, err := CalculateSMA(prices[:period], period)
	if err != nil {
		return 0, err
	}

	ema := sma

	// Calculate EMA for remaining prices
	for i := period; i < len(prices); i++ {
		ema = (prices[i] * multiplier) + (ema * (1 - multiplier))
	}

	return ema, nil
}

// CalculateMACD calculates MACD (Moving Average Convergence Divergence)
func CalculateMACD(prices []float64) (macd, signal, histogram float64, err error) {
	if len(prices) < 26 {
		return 0, 0, 0, fmt.Errorf("insufficient data: need at least 26 prices, got %d", len(prices))
	}

	// Calculate 12-period EMA
	ema12, err := CalculateEMA(prices, 12)
	if err != nil {
		return 0, 0, 0, err
	}

	// Calculate 26-period EMA
	ema26, err := CalculateEMA(prices, 26)
	if err != nil {
		return 0, 0, 0, err
	}

	// MACD line = EMA12 - EMA26
	macd = ema12 - ema26

	// For signal line, we would need to calculate 9-period EMA of MACD
	// Simplified version: using MACD value as approximation
	signal = macd * 0.9 // Simplified

	histogram = macd - signal

	return macd, signal, histogram, nil
}

// CalculateBollingerBands calculates Bollinger Bands
// Uses safe division to prevent panic
func CalculateBollingerBands(prices []float64, period int, stdDevMultiplier float64) (upper, middle, lower float64, err error) {
	if len(prices) < period {
		return 0, 0, 0, fmt.Errorf("insufficient data: need at least %d prices, got %d", period, len(prices))
	}

	// Calculate SMA (middle band)
	middle, err = CalculateSMA(prices, period)
	if err != nil {
		return 0, 0, 0, err
	}

	// Calculate standard deviation
	sum := 0.0
	for i := len(prices) - period; i < len(prices); i++ {
		diff := prices[i] - middle
		sum += diff * diff
	}

	if period == 0 {
		return 0, 0, 0, fmt.Errorf("period cannot be zero")
	}

	variance := sum / float64(period)
	stdDev := math.Sqrt(variance)

	// Calculate upper and lower bands
	upper = middle + (stdDev * stdDevMultiplier)
	lower = middle - (stdDev * stdDevMultiplier)

	return upper, middle, lower, nil
}

// CalculatePriceChange calculates price change percentage
// Uses safe division to prevent panic
func CalculatePriceChange(oldPrice, newPrice float64) float64 {
	if oldPrice == 0 {
		return 0 // Avoid division by zero
	}
	return ((newPrice - oldPrice) / oldPrice) * 100
}

// ExtractClosePrices extracts close prices from klines
func ExtractClosePrices(klines []*Kline) []float64 {
	prices := make([]float64, len(klines))
	for i, k := range klines {
		prices[i] = k.Close
	}
	return prices
}

// CalculateIndicators calculates all technical indicators from klines
func CalculateIndicators(klines []*Kline) (*TechnicalIndicators, error) {
	if len(klines) == 0 {
		return nil, fmt.Errorf("no kline data provided")
	}

	prices := ExtractClosePrices(klines)
	indicators := &TechnicalIndicators{
		Timestamp: klines[len(klines)-1].CloseTime,
	}

	// Calculate RSI
	if rsi, err := CalculateRSI(prices, 14); err == nil {
		indicators.RSI = rsi
	}

	// Calculate MACD
	if macd, signal, histogram, err := CalculateMACD(prices); err == nil {
		indicators.MACD = macd
		indicators.MACDSignal = signal
		indicators.MACDHistogram = histogram
	}

	// Calculate Bollinger Bands
	if upper, middle, lower, err := CalculateBollingerBands(prices, 20, 2.0); err == nil {
		indicators.BollingerUpper = upper
		indicators.BollingerMid = middle
		indicators.BollingerLower = lower
	}

	// Calculate SMAs
	if sma20, err := CalculateSMA(prices, 20); err == nil {
		indicators.SMA20 = sma20
	}
	if sma50, err := CalculateSMA(prices, 50); err == nil {
		indicators.SMA50 = sma50
	}
	if sma200, err := CalculateSMA(prices, 200); err == nil {
		indicators.SMA200 = sma200
	}

	// Calculate EMAs
	if ema12, err := CalculateEMA(prices, 12); err == nil {
		indicators.EMA12 = ema12
	}
	if ema26, err := CalculateEMA(prices, 26); err == nil {
		indicators.EMA26 = ema26
	}

	// Calculate 24h volume
	volume24h := 0.0
	for _, k := range klines {
		volume24h += k.Volume
	}
	indicators.Volume24h = volume24h

	// Calculate 24h price change
	if len(prices) > 0 {
		indicators.PriceChange24h = CalculatePriceChange(prices[0], prices[len(prices)-1])
	}

	return indicators, nil
}
