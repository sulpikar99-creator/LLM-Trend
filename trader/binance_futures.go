package trader

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const (
	BinanceFuturesBaseURL        = "https://fapi.binance.com"
	BinanceFuturesTestnetBaseURL = "https://testnet.binancefuture.com"
)

// BinanceFuturesTrader implements Trader interface for Binance Futures
type BinanceFuturesTrader struct {
	name       string
	apiKey     string
	apiSecret  string
	baseURL    string
	httpClient *http.Client
	connected  bool
	testnet    bool
	mu         sync.RWMutex
}

// NewBinanceFuturesTrader creates a new Binance Futures trader
func NewBinanceFuturesTrader(name, apiKey, apiSecret string, testnet bool) *BinanceFuturesTrader {
	baseURL := BinanceFuturesBaseURL
	if testnet {
		baseURL = BinanceFuturesTestnetBaseURL
	}

	return &BinanceFuturesTrader{
		name:      name,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		baseURL:   baseURL,
		testnet:   testnet,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		connected: false,
	}
}

// GetName returns the trader name
func (b *BinanceFuturesTrader) GetName() string {
	return b.name
}

// GetExchangeType returns the exchange type
func (b *BinanceFuturesTrader) GetExchangeType() string {
	return "binance_futures"
}

// IsTestnet reports whether the trader is configured to use the Binance testnet
func (b *BinanceFuturesTrader) IsTestnet() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.testnet
}

// Connect establishes connection to Binance Futures
func (b *BinanceFuturesTrader) Connect(ctx context.Context) error {
	// Test connection by fetching server time
	resp, err := b.httpClient.Get(b.baseURL + "/fapi/v1/time")
	if err != nil {
		return fmt.Errorf("failed to connect to Binance Futures: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Binance Futures returned status %d", resp.StatusCode)
	}

	b.mu.Lock()
	b.connected = true
	b.mu.Unlock()

	return nil
}

// Disconnect closes connection
func (b *BinanceFuturesTrader) Disconnect() error {
	b.mu.Lock()
	b.connected = false
	b.mu.Unlock()
	return nil
}

// IsConnected returns connection status
func (b *BinanceFuturesTrader) IsConnected() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.connected
}

// signRequest signs the request with HMAC SHA256
func (b *BinanceFuturesTrader) signRequest(params url.Values) string {
	mac := hmac.New(sha256.New, []byte(b.apiSecret))
	mac.Write([]byte(params.Encode()))
	return hex.EncodeToString(mac.Sum(nil))
}

// doRequest performs an HTTP request with signature
func (b *BinanceFuturesTrader) doRequest(ctx context.Context, method, endpoint string, params url.Values, signed bool) ([]byte, error) {
	if !b.IsConnected() {
		return nil, fmt.Errorf("trader not connected")
	}

	if params == nil {
		params = url.Values{}
	}

	if signed {
		params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
		params.Set("signature", b.signRequest(params))
	}

	reqURL := b.baseURL + endpoint
	if method == http.MethodGet || method == http.MethodDelete {
		reqURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-MBX-APIKEY", b.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// GetBalance fetches account balance
func (b *BinanceFuturesTrader) GetBalance(ctx context.Context) (*Balance, error) {
	data, err := b.doRequest(ctx, http.MethodGet, "/fapi/v2/balance", nil, true)
	if err != nil {
		return nil, err
	}

	var balances []struct {
		AccountAlias       string `json:"accountAlias"`
		Asset              string `json:"asset"`
		Balance            string `json:"balance"`
		CrossWalletBalance string `json:"crossWalletBalance"`
		CrossUnPnl         string `json:"crossUnPnl"`
		AvailableBalance   string `json:"availableBalance"`
		MaxWithdrawAmount  string `json:"maxWithdrawAmount"`
	}

	if err := json.Unmarshal(data, &balances); err != nil {
		return nil, fmt.Errorf("failed to parse balance: %w", err)
	}

	// Find USDT balance (most common for futures)
	for _, bal := range balances {
		if bal.Asset == "USDT" {
			balance, _ := strconv.ParseFloat(bal.Balance, 64)
			available, _ := strconv.ParseFloat(bal.AvailableBalance, 64)
			crossWallet, _ := strconv.ParseFloat(bal.CrossWalletBalance, 64)
			unrealizedPnl, _ := strconv.ParseFloat(bal.CrossUnPnl, 64)

			return &Balance{
				Asset:              bal.Asset,
				Balance:            balance,
				AvailableBalance:   available,
				CrossWalletBalance: crossWallet,
				UnrealizedProfit:   unrealizedPnl,
				WalletBalance:      balance,
				MarginBalance:      crossWallet,
				UpdateTime:         time.Now(),
			}, nil
		}
	}

	return nil, fmt.Errorf("USDT balance not found")
}

// GetPositions fetches all open positions
func (b *BinanceFuturesTrader) GetPositions(ctx context.Context) ([]*Position, error) {
	data, err := b.doRequest(ctx, http.MethodGet, "/fapi/v2/positionRisk", nil, true)
	if err != nil {
		return nil, err
	}

	var positionsData []struct {
		Symbol           string `json:"symbol"`
		PositionAmt      string `json:"positionAmt"`
		EntryPrice       string `json:"entryPrice"`
		MarkPrice        string `json:"markPrice"`
		UnRealizedProfit string `json:"unRealizedProfit"`
		Leverage         string `json:"leverage"`
		MarginType       string `json:"marginType"`
		IsolatedMargin   string `json:"isolatedMargin"`
		Notional         string `json:"notional"`
		PositionSide     string `json:"positionSide"`
		UpdateTime       int64  `json:"updateTime"`
	}

	if err := json.Unmarshal(data, &positionsData); err != nil {
		return nil, fmt.Errorf("failed to parse positions: %w", err)
	}

	positions := []*Position{}
	for _, pos := range positionsData {
		posAmt, _ := strconv.ParseFloat(pos.PositionAmt, 64)

		// Skip positions with zero size
		if posAmt == 0 {
			continue
		}

		entryPrice, _ := strconv.ParseFloat(pos.EntryPrice, 64)
		markPrice, _ := strconv.ParseFloat(pos.MarkPrice, 64)
		unrealizedPnl, _ := strconv.ParseFloat(pos.UnRealizedProfit, 64)
		leverage, _ := strconv.Atoi(pos.Leverage)
		isolatedMargin, _ := strconv.ParseFloat(pos.IsolatedMargin, 64)
		notional, _ := strconv.ParseFloat(pos.Notional, 64)

		// Determine position side
		var side PositionSide
		if posAmt > 0 {
			side = PositionSideLong
		} else {
			side = PositionSideShort
			posAmt = -posAmt // Make size positive
		}

		position := &Position{
			Symbol:           pos.Symbol,
			Side:             side,
			Size:             posAmt,
			EntryPrice:       entryPrice,
			MarkPrice:        markPrice,
			UnrealizedProfit: unrealizedPnl,
			Leverage:         leverage,
			MarginType:       pos.MarginType,
			IsolatedMargin:   isolatedMargin,
			Notional:         notional,
			UpdateTime:       time.Unix(pos.UpdateTime/1000, 0),
		}

		positions = append(positions, position)
	}

	return positions, nil
}

// GetPosition fetches a specific position
func (b *BinanceFuturesTrader) GetPosition(ctx context.Context, symbol string) (*Position, error) {
	positions, err := b.GetPositions(ctx)
	if err != nil {
		return nil, err
	}

	for _, pos := range positions {
		if pos.Symbol == symbol {
			return pos, nil
		}
	}

	return nil, fmt.Errorf("position not found for symbol %s", symbol)
}

// PlaceOrder executes a new order
func (b *BinanceFuturesTrader) PlaceOrder(ctx context.Context, req *OrderRequest) (*Order, error) {
	params := url.Values{}
	params.Set("symbol", req.Symbol)
	params.Set("side", string(req.Side))
	params.Set("type", string(req.Type))
	params.Set("quantity", strconv.FormatFloat(req.Quantity, 'f', -1, 64))

	if req.Price > 0 {
		params.Set("price", strconv.FormatFloat(req.Price, 'f', -1, 64))
	}

	if req.StopPrice > 0 {
		params.Set("stopPrice", strconv.FormatFloat(req.StopPrice, 'f', -1, 64))
	}

	if req.ReduceOnly {
		params.Set("reduceOnly", "true")
	}

	if req.TimeInForce != "" {
		params.Set("timeInForce", req.TimeInForce)
	} else if req.Type == OrderTypeLimit {
		params.Set("timeInForce", "GTC") // Good Till Cancel
	}

	data, err := b.doRequest(ctx, http.MethodPost, "/fapi/v1/order", params, true)
	if err != nil {
		return nil, err
	}

	var orderResp struct {
		OrderID       int64  `json:"orderId"`
		ClientOrderID string `json:"clientOrderId"`
		Symbol        string `json:"symbol"`
		Type          string `json:"type"`
		Side          string `json:"side"`
		Price         string `json:"price"`
		OrigQty       string `json:"origQty"`
		ExecutedQty   string `json:"executedQty"`
		AvgPrice      string `json:"avgPrice"`
		Status        string `json:"status"`
		UpdateTime    int64  `json:"updateTime"`
	}

	if err := json.Unmarshal(data, &orderResp); err != nil {
		return nil, fmt.Errorf("failed to parse order response: %w", err)
	}

	price, _ := strconv.ParseFloat(orderResp.Price, 64)
	quantity, _ := strconv.ParseFloat(orderResp.OrigQty, 64)
	executedQty, _ := strconv.ParseFloat(orderResp.ExecutedQty, 64)
	avgPrice, _ := strconv.ParseFloat(orderResp.AvgPrice, 64)

	order := &Order{
		OrderID:       strconv.FormatInt(orderResp.OrderID, 10),
		ClientOrderID: orderResp.ClientOrderID,
		Symbol:        orderResp.Symbol,
		Type:          OrderType(orderResp.Type),
		Side:          OrderSide(orderResp.Side),
		Price:         price,
		Quantity:      quantity,
		ExecutedQty:   executedQty,
		AvgPrice:      avgPrice,
		Status:        OrderStatus(orderResp.Status),
		UpdateTime:    time.Unix(orderResp.UpdateTime/1000, 0),
	}

	return order, nil
}

// CancelOrder cancels an existing order
func (b *BinanceFuturesTrader) CancelOrder(ctx context.Context, symbol, orderID string) error {
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("orderId", orderID)

	_, err := b.doRequest(ctx, http.MethodDelete, "/fapi/v1/order", params, true)
	return err
}

// GetOrder fetches order details
func (b *BinanceFuturesTrader) GetOrder(ctx context.Context, symbol, orderID string) (*Order, error) {
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("orderId", orderID)

	data, err := b.doRequest(ctx, http.MethodGet, "/fapi/v1/order", params, true)
	if err != nil {
		return nil, err
	}

	var orderResp struct {
		OrderID       int64  `json:"orderId"`
		ClientOrderID string `json:"clientOrderId"`
		Symbol        string `json:"symbol"`
		Type          string `json:"type"`
		Side          string `json:"side"`
		Price         string `json:"price"`
		OrigQty       string `json:"origQty"`
		ExecutedQty   string `json:"executedQty"`
		AvgPrice      string `json:"avgPrice"`
		Status        string `json:"status"`
		Time          int64  `json:"time"`
		UpdateTime    int64  `json:"updateTime"`
	}

	if err := json.Unmarshal(data, &orderResp); err != nil {
		return nil, fmt.Errorf("failed to parse order: %w", err)
	}

	price, _ := strconv.ParseFloat(orderResp.Price, 64)
	quantity, _ := strconv.ParseFloat(orderResp.OrigQty, 64)
	executedQty, _ := strconv.ParseFloat(orderResp.ExecutedQty, 64)
	avgPrice, _ := strconv.ParseFloat(orderResp.AvgPrice, 64)

	order := &Order{
		OrderID:       strconv.FormatInt(orderResp.OrderID, 10),
		ClientOrderID: orderResp.ClientOrderID,
		Symbol:        orderResp.Symbol,
		Type:          OrderType(orderResp.Type),
		Side:          OrderSide(orderResp.Side),
		Price:         price,
		Quantity:      quantity,
		ExecutedQty:   executedQty,
		AvgPrice:      avgPrice,
		Status:        OrderStatus(orderResp.Status),
		CreateTime:    time.Unix(orderResp.Time/1000, 0),
		UpdateTime:    time.Unix(orderResp.UpdateTime/1000, 0),
	}

	return order, nil
}

// SetLeverage sets leverage for a symbol
func (b *BinanceFuturesTrader) SetLeverage(ctx context.Context, symbol string, leverage int) error {
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("leverage", strconv.Itoa(leverage))

	_, err := b.doRequest(ctx, http.MethodPost, "/fapi/v1/leverage", params, true)
	return err
}

// SetMarginType sets margin type for a symbol
func (b *BinanceFuturesTrader) SetMarginType(ctx context.Context, symbol, marginType string) error {
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("marginType", marginType)

	_, err := b.doRequest(ctx, http.MethodPost, "/fapi/v1/marginType", params, true)
	return err
}
