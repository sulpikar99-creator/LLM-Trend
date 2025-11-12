package market

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients by trader ID
	clients map[string]map[*Client]bool

	// Inbound messages from the clients
	broadcast chan Message

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// Client is a middleman between the websocket connection and the hub
type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan Message
	TraderID string
	UserID   string
}

// Message represents a WebSocket message
type Message struct {
	TraderID  string      `json:"trader_id"`
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// Message types
const (
	MessageTypePositionUpdate = "position_update"
	MessageTypeOrderUpdate    = "order_update"
	MessageTypePnLUpdate      = "pnl_update"
	MessageTypeTradeExecuted  = "trade_executed"
	MessageTypeBalanceUpdate  = "balance_update"
	MessageTypeError          = "error"
	MessageTypePing           = "ping"
	MessageTypePong           = "pong"
)

// Position update data
type PositionUpdate struct {
	Symbol           string  `json:"symbol"`
	Side             string  `json:"side"`
	Size             float64 `json:"size"`
	EntryPrice       float64 `json:"entry_price"`
	MarkPrice        float64 `json:"mark_price"`
	UnrealizedPnL    float64 `json:"unrealized_pnl"`
	LiquidationPrice float64 `json:"liquidation_price"`
	Leverage         float64 `json:"leverage"`
}

// Order update data
type OrderUpdate struct {
	OrderID       string  `json:"order_id"`
	Symbol        string  `json:"symbol"`
	Side          string  `json:"side"`
	Type          string  `json:"type"`
	Quantity      float64 `json:"quantity"`
	Price         float64 `json:"price"`
	FilledQty     float64 `json:"filled_qty"`
	AvgPrice      float64 `json:"avg_price"`
	Status        string  `json:"status"`
	TimeInForce   string  `json:"time_in_force"`
	ExecutionTime int64   `json:"execution_time"`
}

// PnL update data
type PnLUpdate struct {
	TotalPnL        float64 `json:"total_pnl"`
	RealizedPnL     float64 `json:"realized_pnl"`
	UnrealizedPnL   float64 `json:"unrealized_pnl"`
	TotalEquity     float64 `json:"total_equity"`
	AvailableMargin float64 `json:"available_margin"`
	UsedMargin      float64 `json:"used_margin"`
}

// Trade executed data
type TradeExecuted struct {
	Symbol   string  `json:"symbol"`
	Side     string  `json:"side"`
	Quantity float64 `json:"quantity"`
	Price    float64 `json:"price"`
	Fee      float64 `json:"fee"`
	Time     int64   `json:"time"`
}

// Balance update data
type BalanceUpdate struct {
	Asset     string  `json:"asset"`
	Free      float64 `json:"free"`
	Locked    float64 `json:"locked"`
	Total     float64 `json:"total"`
	UpdatedAt int64   `json:"updated_at"`
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]map[*Client]bool),
	}
}

// Run starts the hub's main event loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.TraderID] == nil {
				h.clients[client.TraderID] = make(map[*Client]bool)
			}
			h.clients[client.TraderID][client] = true
			h.mu.Unlock()
			log.Printf("WebSocket client registered for trader %s (user %s)", client.TraderID, client.UserID)

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.TraderID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.clients, client.TraderID)
					}
					log.Printf("WebSocket client unregistered for trader %s", client.TraderID)
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			if clients, ok := h.clients[message.TraderID]; ok {
				for client := range clients {
					select {
					case client.Send <- message:
					default:
						// Client is slow, close the connection
						close(client.Send)
						delete(clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// RegisterClient registers a new client with the hub
func (h *Hub) RegisterClient(client *Client, conn *websocket.Conn) {
	client.Hub = h
	client.Conn = conn
	client.Send = make(chan Message, 256)
	h.register <- client
}

// BroadcastPositionUpdate sends position update to all clients of a trader
func (h *Hub) BroadcastPositionUpdate(traderID string, position PositionUpdate) {
	h.broadcast <- Message{
		TraderID:  traderID,
		Type:      MessageTypePositionUpdate,
		Data:      position,
		Timestamp: time.Now(),
	}
}

// BroadcastOrderUpdate sends order update to all clients of a trader
func (h *Hub) BroadcastOrderUpdate(traderID string, order OrderUpdate) {
	h.broadcast <- Message{
		TraderID:  traderID,
		Type:      MessageTypeOrderUpdate,
		Data:      order,
		Timestamp: time.Now(),
	}
}

// BroadcastPnLUpdate sends PnL update to all clients of a trader
func (h *Hub) BroadcastPnLUpdate(traderID string, pnl PnLUpdate) {
	h.broadcast <- Message{
		TraderID:  traderID,
		Type:      MessageTypePnLUpdate,
		Data:      pnl,
		Timestamp: time.Now(),
	}
}

// BroadcastTradeExecuted sends trade execution notification
func (h *Hub) BroadcastTradeExecuted(traderID string, trade TradeExecuted) {
	h.broadcast <- Message{
		TraderID:  traderID,
		Type:      MessageTypeTradeExecuted,
		Data:      trade,
		Timestamp: time.Now(),
	}
}

// BroadcastBalanceUpdate sends balance update
func (h *Hub) BroadcastBalanceUpdate(traderID string, balance BalanceUpdate) {
	h.broadcast <- Message{
		TraderID:  traderID,
		Type:      MessageTypeBalanceUpdate,
		Data:      balance,
		Timestamp: time.Now(),
	}
}

// BroadcastError sends error message to clients
func (h *Hub) BroadcastError(traderID string, errorMsg string) {
	h.broadcast <- Message{
		TraderID:  traderID,
		Type:      MessageTypeError,
		Data:      map[string]string{"message": errorMsg},
		Timestamp: time.Now(),
	}
}

// GetClientCount returns the number of connected clients for a trader
func (h *Hub) GetClientCount(traderID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if clients, ok := h.clients[traderID]; ok {
		return len(clients)
	}
	return 0
}

// GetTotalClientCount returns total number of connected clients
func (h *Hub) GetTotalClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	count := 0
	for _, clients := range h.clients {
		count += len(clients)
	}
	return count
}

// Constants for WebSocket configuration
const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

// ReadPump pumps messages from the websocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming messages (e.g., pong, subscribe requests)
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err == nil {
			if msgType, ok := msg["type"].(string); ok && msgType == "ping" {
				// Respond to ping with pong
				pongMsg := Message{
					TraderID:  c.TraderID,
					Type:      MessageTypePong,
					Data:      map[string]string{"status": "alive"},
					Timestamp: time.Now(),
				}
				c.Send <- pongMsg
			}
		}
	}
}

// WritePump pumps messages from the hub to the websocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			// Marshal message to JSON
			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				continue
			}

			w.Write(data)

			// Add queued messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				msg := <-c.Send
				data, err := json.Marshal(msg)
				if err != nil {
					log.Printf("Error marshaling queued message: %v", err)
					continue
				}
				w.Write(data)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
