package trader

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBinanceFuturesConnection tests connection to Binance Futures
// This test requires BINANCE_TESTNET_API_KEY and BINANCE_TESTNET_SECRET environment variables
func TestBinanceFuturesConnection(t *testing.T) {
	apiKey := os.Getenv("BINANCE_TESTNET_API_KEY")
	apiSecret := os.Getenv("BINANCE_TESTNET_SECRET")

	if apiKey == "" || apiSecret == "" {
		t.Skip("Skipping integration test: BINANCE_TESTNET_API_KEY or BINANCE_TESTNET_SECRET not set")
	}

	trader := NewBinanceFuturesTrader("test-trader", apiKey, apiSecret, true)
	assert.NotNil(t, trader)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test connection
	err := trader.Connect(ctx)
	assert.NoError(t, err)
	assert.True(t, trader.IsConnected())

	// Cleanup
	trader.Disconnect()
	assert.False(t, trader.IsConnected())
}

// TestBinanceFuturesGetBalance tests balance retrieval
func TestBinanceFuturesGetBalance(t *testing.T) {
	apiKey := os.Getenv("BINANCE_TESTNET_API_KEY")
	apiSecret := os.Getenv("BINANCE_TESTNET_SECRET")

	if apiKey == "" || apiSecret == "" {
		t.Skip("Skipping integration test: credentials not set")
	}

	trader := NewBinanceFuturesTrader("test-trader", apiKey, apiSecret, true)
	require.NotNil(t, trader)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := trader.Connect(ctx)
	require.NoError(t, err)
	defer trader.Disconnect()

	// Get balance
	balance, err := trader.GetBalance(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, balance)
	assert.Equal(t, "USDT", balance.Asset)
	assert.GreaterOrEqual(t, balance.Balance, 0.0)
	assert.GreaterOrEqual(t, balance.AvailableBalance, 0.0)
}

// TestBinanceFuturesGetPositions tests position retrieval
func TestBinanceFuturesGetPositions(t *testing.T) {
	apiKey := os.Getenv("BINANCE_TESTNET_API_KEY")
	apiSecret := os.Getenv("BINANCE_TESTNET_SECRET")

	if apiKey == "" || apiSecret == "" {
		t.Skip("Skipping integration test: credentials not set")
	}

	trader := NewBinanceFuturesTrader("test-trader", apiKey, apiSecret, true)
	require.NotNil(t, trader)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := trader.Connect(ctx)
	require.NoError(t, err)
	defer trader.Disconnect()

	// Get positions
	positions, err := trader.GetPositions(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, positions)

	// Positions can be empty, but should not error
	for _, pos := range positions {
		assert.NotEmpty(t, pos.Symbol)
		assert.Greater(t, pos.Size, 0.0)
		assert.Greater(t, pos.EntryPrice, 0.0)
	}
}

// TestBinanceFuturesSetLeverage tests leverage setting
func TestBinanceFuturesSetLeverage(t *testing.T) {
	apiKey := os.Getenv("BINANCE_TESTNET_API_KEY")
	apiSecret := os.Getenv("BINANCE_TESTNET_SECRET")

	if apiKey == "" || apiSecret == "" {
		t.Skip("Skipping integration test: credentials not set")
	}

	trader := NewBinanceFuturesTrader("test-trader", apiKey, apiSecret, true)
	require.NotNil(t, trader)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := trader.Connect(ctx)
	require.NoError(t, err)
	defer trader.Disconnect()

	// Set leverage for BTCUSDT
	err = trader.SetLeverage(ctx, "BTCUSDT", 5)
	// May succeed or fail depending on account status, but should not panic
	if err != nil {
		t.Logf("SetLeverage returned error (may be expected): %v", err)
	}
}

// TestBinanceFuturesPlaceOrder tests order placement (will not actually execute)
func TestBinanceFuturesPlaceOrder(t *testing.T) {
	t.Skip("Skipping order placement test to avoid actual trades")

	apiKey := os.Getenv("BINANCE_TESTNET_API_KEY")
	apiSecret := os.Getenv("BINANCE_TESTNET_SECRET")

	if apiKey == "" || apiSecret == "" {
		t.Skip("Skipping integration test: credentials not set")
	}

	trader := NewBinanceFuturesTrader("test-trader", apiKey, apiSecret, true)
	require.NotNil(t, trader)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := trader.Connect(ctx)
	require.NoError(t, err)
	defer trader.Disconnect()

	// Place a small market order (testnet only!)
	orderReq := &OrderRequest{
		Symbol:   "BTCUSDT",
		Side:     OrderSideBuy,
		Type:     OrderTypeMarket,
		Quantity: 0.001, // Very small amount
	}

	order, err := trader.PlaceOrder(ctx, orderReq)
	if err != nil {
		t.Logf("PlaceOrder returned error: %v", err)
	} else {
		assert.NotNil(t, order)
		assert.NotEmpty(t, order.OrderID)
		assert.Equal(t, "BTCUSDT", order.Symbol)
	}
}

// TestBinanceFuturesExchangeType tests exchange type identification
func TestBinanceFuturesExchangeType(t *testing.T) {
	trader := NewBinanceFuturesTrader("test", "key", "secret", false)
	assert.Equal(t, "binance_futures", trader.GetExchangeType())

	traderTestnet := NewBinanceFuturesTrader("test", "key", "secret", true)
	assert.Equal(t, "binance_futures", traderTestnet.GetExchangeType())
}

// TestBinanceFuturesGetName tests trader name retrieval
func TestBinanceFuturesGetName(t *testing.T) {
	trader := NewBinanceFuturesTrader("my-trader", "key", "secret", false)
	assert.Equal(t, "my-trader", trader.GetName())
}

// TestBinanceFuturesSignRequest tests HMAC signing
func TestBinanceFuturesSignRequest(t *testing.T) {
	trader := NewBinanceFuturesTrader("test", "test-key", "test-secret", false)

	// Create test params
	params := map[string]string{
		"symbol":    "BTCUSDT",
		"side":      "BUY",
		"type":      "LIMIT",
		"quantity":  "1",
		"price":     "50000",
		"timestamp": "1234567890",
	}

	// Convert to url.Values for signing
	import "net/url"
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}

	// Sign request
	signature := trader.signRequest(values)

	// Signature should be non-empty hex string
	assert.NotEmpty(t, signature)
	assert.Len(t, signature, 64) // SHA256 produces 64 hex chars
}
