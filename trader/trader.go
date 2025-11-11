package trader

import (
	"context"
	"time"
)

// OrderType represents the type of order
type OrderType string

const (
	OrderTypeMarket     OrderType = "MARKET"
	OrderTypeLimit      OrderType = "LIMIT"
	OrderTypeStopLoss   OrderType = "STOP_LOSS"
	OrderTypeTakeProfit OrderType = "TAKE_PROFIT"
)

// OrderSide represents the side of an order
type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

// PositionSide represents the side of a position
type PositionSide string

const (
	PositionSideLong  PositionSide = "LONG"
	PositionSideShort PositionSide = "SHORT"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusNew      OrderStatus = "NEW"
	OrderStatusFilled   OrderStatus = "FILLED"
	OrderStatusCanceled OrderStatus = "CANCELED"
	OrderStatusRejected OrderStatus = "REJECTED"
)

// Balance represents account balance information
type Balance struct {
	Asset              string    `json:"asset"`
	Balance            float64   `json:"balance"`
	AvailableBalance   float64   `json:"available_balance"`
	UnrealizedProfit   float64   `json:"unrealized_profit"`
	MarginBalance      float64   `json:"margin_balance"`
	WalletBalance      float64   `json:"wallet_balance"`
	CrossWalletBalance float64   `json:"cross_wallet_balance"`
	UpdateTime         time.Time `json:"update_time"`
}

// Position represents an open position
type Position struct {
	Symbol           string       `json:"symbol"`
	Side             PositionSide `json:"side"`
	Size             float64      `json:"size"`
	EntryPrice       float64      `json:"entry_price"`
	MarkPrice        float64      `json:"mark_price"`
	UnrealizedProfit float64      `json:"unrealized_profit"`
	Leverage         int          `json:"leverage"`
	MarginType       string       `json:"margin_type"`
	IsolatedMargin   float64      `json:"isolated_margin"`
	Notional         float64      `json:"notional"`
	UpdateTime       time.Time    `json:"update_time"`
}

// Order represents an order
type Order struct {
	OrderID       string      `json:"order_id"`
	ClientOrderID string      `json:"client_order_id"`
	Symbol        string      `json:"symbol"`
	Type          OrderType   `json:"type"`
	Side          OrderSide   `json:"side"`
	Price         float64     `json:"price"`
	Quantity      float64     `json:"quantity"`
	StopPrice     float64     `json:"stop_price"`
	Status        OrderStatus `json:"status"`
	ExecutedQty   float64     `json:"executed_qty"`
	AvgPrice      float64     `json:"avg_price"`
	CreateTime    time.Time   `json:"create_time"`
	UpdateTime    time.Time   `json:"update_time"`
}

// OrderRequest represents a request to place an order
type OrderRequest struct {
	Symbol      string    `json:"symbol"`
	Side        OrderSide `json:"side"`
	Type        OrderType `json:"type"`
	Quantity    float64   `json:"quantity"`
	Price       float64   `json:"price,omitempty"`
	StopPrice   float64   `json:"stop_price,omitempty"`
	ReduceOnly  bool      `json:"reduce_only,omitempty"`
	TimeInForce string    `json:"time_in_force,omitempty"`
}

// Trader is the unified interface for all exchange implementations
type Trader interface {
	// GetName returns the trader name
	GetName() string

	// GetExchangeType returns the exchange type (e.g., "binance", "okx")
	GetExchangeType() string

	// Connect establishes connection to the exchange
	Connect(ctx context.Context) error

	// Disconnect closes connection to the exchange
	Disconnect() error

	// GetBalance fetches account balance
	GetBalance(ctx context.Context) (*Balance, error)

	// GetPositions fetches all open positions
	GetPositions(ctx context.Context) ([]*Position, error)

	// GetPosition fetches a specific position for a symbol
	GetPosition(ctx context.Context, symbol string) (*Position, error)

	// PlaceOrder executes a new order
	PlaceOrder(ctx context.Context, req *OrderRequest) (*Order, error)

	// CancelOrder cancels an existing order
	CancelOrder(ctx context.Context, symbol, orderID string) error

	// GetOrder fetches order details
	GetOrder(ctx context.Context, symbol, orderID string) (*Order, error)

	// SetLeverage sets leverage for a symbol
	SetLeverage(ctx context.Context, symbol string, leverage int) error

	// SetMarginType sets margin type (ISOLATED/CROSSED) for a symbol
	SetMarginType(ctx context.Context, symbol, marginType string) error

	// IsConnected returns connection status
	IsConnected() bool
}
