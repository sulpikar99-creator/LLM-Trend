package market

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	BinanceWSBaseURL = "wss://fstream.binance.com"
	BinanceWSTestnet = "wss://stream.binancefuture.com"
)

// KlineMonitor monitors real-time kline data via WebSocket
type KlineMonitor struct {
	symbol    string
	interval  string
	baseURL   string
	conn      *websocket.Conn
	klines    []*Kline
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	onUpdate  func(*Kline)
	maxKlines int
}

// binanceKlineEvent represents WebSocket kline event from Binance
type binanceKlineEvent struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	Kline     struct {
		StartTime            int64  `json:"t"`
		EndTime              int64  `json:"T"`
		Symbol               string `json:"s"`
		Interval             string `json:"i"`
		FirstTradeID         int64  `json:"f"`
		LastTradeID          int64  `json:"L"`
		Open                 string `json:"o"`
		Close                string `json:"c"`
		High                 string `json:"h"`
		Low                  string `json:"l"`
		Volume               string `json:"v"`
		NumberOfTrades       int64  `json:"n"`
		IsClosed             bool   `json:"x"`
		QuoteAssetVolume     string `json:"q"`
		TakerBuyBaseVolume   string `json:"V"`
		TakerBuyQuoteVolume  string `json:"Q"`
	} `json:"k"`
}

// NewKlineMonitor creates a new kline monitor
func NewKlineMonitor(symbol, interval string, testnet bool, maxKlines int) *KlineMonitor {
	baseURL := BinanceWSBaseURL
	if testnet {
		baseURL = BinanceWSTestnet
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &KlineMonitor{
		symbol:    symbol,
		interval:  interval,
		baseURL:   baseURL,
		klines:    make([]*Kline, 0, maxKlines),
		ctx:       ctx,
		cancel:    cancel,
		maxKlines: maxKlines,
	}
}

// SetOnUpdate sets callback for kline updates
func (km *KlineMonitor) SetOnUpdate(callback func(*Kline)) {
	km.mu.Lock()
	defer km.mu.Unlock()
	km.onUpdate = callback
}

// Start starts the WebSocket connection and monitoring
func (km *KlineMonitor) Start() error {
	streamName := fmt.Sprintf("%s@kline_%s", km.symbol, km.interval)
	wsURL := fmt.Sprintf("%s/ws/%s", km.baseURL, streamName)

	var err error
	km.conn, _, err = websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	// Start goroutine to read messages
	go km.readMessages()

	// Start goroutine for periodic ping
	go km.keepAlive()

	log.Printf("Kline monitor started for %s (%s)", km.symbol, km.interval)
	return nil
}

// Stop stops the WebSocket connection
func (km *KlineMonitor) Stop() error {
	km.cancel()

	if km.conn != nil {
		// Send close message
		err := km.conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		)
		if err != nil {
			log.Printf("Error sending close message: %v", err)
		}

		err = km.conn.Close()
		if err != nil {
			return fmt.Errorf("failed to close WebSocket: %w", err)
		}
	}

	log.Printf("Kline monitor stopped for %s", km.symbol)
	return nil
}

// readMessages reads WebSocket messages in a goroutine
func (km *KlineMonitor) readMessages() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in readMessages: %v", r)
		}
	}()

	for {
		select {
		case <-km.ctx.Done():
			return

		default:
			_, message, err := km.conn.ReadMessage()
			if err != nil {
				// Check if context was cancelled
				select {
				case <-km.ctx.Done():
					return
				default:
					log.Printf("WebSocket read error: %v", err)
					// Try to reconnect after delay
					time.Sleep(5 * time.Second)
					if err := km.reconnect(); err != nil {
						log.Printf("Failed to reconnect: %v", err)
						return
					}
					continue
				}
			}

			km.handleMessage(message)
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (km *KlineMonitor) handleMessage(message []byte) {
	var event binanceKlineEvent
	if err := json.Unmarshal(message, &event); err != nil {
		log.Printf("Failed to parse kline event: %v", err)
		return
	}

	// Convert to Kline struct with safe parsing
	kline := km.parseKline(&event)
	if kline == nil {
		return
	}

	// Update klines array
	km.mu.Lock()
	if event.Kline.IsClosed {
		// Kline is closed, add to history
		km.klines = append(km.klines, kline)

		// Keep only last N klines
		if len(km.klines) > km.maxKlines {
			km.klines = km.klines[len(km.klines)-km.maxKlines:]
		}
	} else {
		// Kline is still open, update last one if exists
		if len(km.klines) > 0 {
			km.klines[len(km.klines)-1] = kline
		} else {
			km.klines = append(km.klines, kline)
		}
	}
	callback := km.onUpdate
	km.mu.Unlock()

	// Call callback if set
	if callback != nil {
		callback(kline)
	}
}

// parseKline parses binance kline event to Kline struct with safe type conversions
func (km *KlineMonitor) parseKline(event *binanceKlineEvent) *Kline {
	// Safe float parsing
	open, err := strconv.ParseFloat(event.Kline.Open, 64)
	if err != nil {
		log.Printf("Failed to parse open price: %v", err)
		return nil
	}

	high, err := strconv.ParseFloat(event.Kline.High, 64)
	if err != nil {
		log.Printf("Failed to parse high price: %v", err)
		return nil
	}

	low, err := strconv.ParseFloat(event.Kline.Low, 64)
	if err != nil {
		log.Printf("Failed to parse low price: %v", err)
		return nil
	}

	close, err := strconv.ParseFloat(event.Kline.Close, 64)
	if err != nil {
		log.Printf("Failed to parse close price: %v", err)
		return nil
	}

	volume, err := strconv.ParseFloat(event.Kline.Volume, 64)
	if err != nil {
		log.Printf("Failed to parse volume: %v", err)
		return nil
	}

	quoteVolume, _ := strconv.ParseFloat(event.Kline.QuoteAssetVolume, 64)
	takerBuyBase, _ := strconv.ParseFloat(event.Kline.TakerBuyBaseVolume, 64)
	takerBuyQuote, _ := strconv.ParseFloat(event.Kline.TakerBuyQuoteVolume, 64)

	return &Kline{
		OpenTime:            time.Unix(event.Kline.StartTime/1000, 0),
		Open:                open,
		High:                high,
		Low:                 low,
		Close:               close,
		Volume:              volume,
		CloseTime:           time.Unix(event.Kline.EndTime/1000, 0),
		QuoteAssetVolume:    quoteVolume,
		NumberOfTrades:      event.Kline.NumberOfTrades,
		TakerBuyBaseVolume:  takerBuyBase,
		TakerBuyQuoteVolume: takerBuyQuote,
	}
}

// keepAlive sends periodic ping messages to keep connection alive
func (km *KlineMonitor) keepAlive() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-km.ctx.Done():
			return

		case <-ticker.C:
			if km.conn != nil {
				err := km.conn.WriteMessage(websocket.PingMessage, []byte{})
				if err != nil {
					log.Printf("Failed to send ping: %v", err)
				}
			}
		}
	}
}

// reconnect attempts to reconnect to WebSocket
func (km *KlineMonitor) reconnect() error {
	if km.conn != nil {
		km.conn.Close()
	}

	streamName := fmt.Sprintf("%s@kline_%s", km.symbol, km.interval)
	wsURL := fmt.Sprintf("%s/ws/%s", km.baseURL, streamName)

	var err error
	km.conn, _, err = websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to reconnect: %w", err)
	}

	log.Printf("Reconnected to WebSocket for %s", km.symbol)
	return nil
}

// GetKlines returns current klines (thread-safe)
func (km *KlineMonitor) GetKlines() []*Kline {
	km.mu.RLock()
	defer km.mu.RUnlock()

	// Return a copy to prevent external modification
	klinesCopy := make([]*Kline, len(km.klines))
	copy(klinesCopy, km.klines)
	return klinesCopy
}

// GetLatestKline returns the latest kline (thread-safe)
func (km *KlineMonitor) GetLatestKline() *Kline {
	km.mu.RLock()
	defer km.mu.RUnlock()

	if len(km.klines) == 0 {
		return nil
	}

	return km.klines[len(km.klines)-1]
}

// GetIndicators calculates and returns technical indicators
func (km *KlineMonitor) GetIndicators() (*TechnicalIndicators, error) {
	klines := km.GetKlines()
	if len(klines) == 0 {
		return nil, fmt.Errorf("no kline data available")
	}

	return CalculateIndicators(klines)
}
