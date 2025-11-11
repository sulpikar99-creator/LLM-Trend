// WebSocket client for real-time trader updates

export interface WebSocketMessage {
  trader_id: string;
  type: string;
  data: any;
  timestamp: string;
}

export interface PositionUpdate {
  symbol: string;
  side: string;
  size: number;
  entry_price: number;
  mark_price: number;
  unrealized_pnl: number;
  liquidation_price: number;
  leverage: number;
}

export interface OrderUpdate {
  order_id: string;
  symbol: string;
  side: string;
  type: string;
  quantity: number;
  price: number;
  filled_qty: number;
  avg_price: number;
  status: string;
  time_in_force: string;
  execution_time: number;
}

export interface PnLUpdate {
  total_pnl: number;
  realized_pnl: number;
  unrealized_pnl: number;
  total_equity: number;
  available_margin: number;
  used_margin: number;
}

export interface TradeExecuted {
  symbol: string;
  side: string;
  quantity: number;
  price: number;
  fee: number;
  time: number;
}

export interface BalanceUpdate {
  asset: string;
  free: number;
  locked: number;
  total: number;
  updated_at: number;
}

export type MessageHandler = (message: WebSocketMessage) => void;
export type ErrorHandler = (error: Event) => void;
export type CloseHandler = () => void;
export type OpenHandler = () => void;

export interface WebSocketOptions {
  onMessage: MessageHandler;
  onError?: ErrorHandler;
  onClose?: CloseHandler;
  onOpen?: OpenHandler;
  reconnectAttempts?: number;
  reconnectDelay?: number;
  maxReconnectDelay?: number;
  debug?: boolean;
}

export class TradingWebSocket {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts: number;
  private reconnectDelay: number;
  private maxReconnectDelay: number;
  private reconnectTimer: number | null = null;
  private pingTimer: number | null = null;
  private debug: boolean;
  private isManualClose = false;

  private traderID: string;
  private onMessage: MessageHandler;
  private onError?: ErrorHandler;
  private onClose?: CloseHandler;
  private onOpen?: OpenHandler;

  constructor(traderID: string, options: WebSocketOptions) {
    this.traderID = traderID;
    this.onMessage = options.onMessage;
    this.onError = options.onError;
    this.onClose = options.onClose;
    this.onOpen = options.onOpen;
    this.maxReconnectAttempts = options.reconnectAttempts || 5;
    this.reconnectDelay = options.reconnectDelay || 1000;
    this.maxReconnectDelay = options.maxReconnectDelay || 30000;
    this.debug = options.debug || false;
  }

  /**
   * Connect to WebSocket server
   */
  connect(): void {
    const token = localStorage.getItem('token');
    if (!token) {
      this.log('No authentication token found');
      return;
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const wsUrl = `${protocol}//${host}/api/ws/traders/${this.traderID}`;

    this.log(`Connecting to ${wsUrl}`);

    try {
      this.ws = new WebSocket(wsUrl);
      this.setupEventHandlers();
    } catch (error) {
      this.log('Failed to create WebSocket connection', error);
      this.scheduleReconnect();
    }
  }

  /**
   * Setup WebSocket event handlers
   */
  private setupEventHandlers(): void {
    if (!this.ws) return;

    this.ws.onopen = () => {
      this.log('WebSocket connected');
      this.reconnectAttempts = 0;
      this.startPingInterval();
      this.onOpen?.();
    };

    this.ws.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        this.log('Received message:', message);

        // Handle pong messages (keep-alive)
        if (message.type === 'pong') {
          this.log('Received pong');
          return;
        }

        // Call message handler
        this.onMessage(message);
      } catch (error) {
        this.log('Error parsing message:', error);
      }
    };

    this.ws.onerror = (error) => {
      this.log('WebSocket error:', error);
      this.onError?.(error);
    };

    this.ws.onclose = (event) => {
      this.log(`WebSocket closed (code: ${event.code}, reason: ${event.reason})`);
      this.stopPingInterval();

      if (!this.isManualClose) {
        this.onClose?.();
        this.scheduleReconnect();
      }
    };
  }

  /**
   * Send ping message to keep connection alive
   */
  private sendPing(): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      const pingMessage = JSON.stringify({ type: 'ping' });
      this.ws.send(pingMessage);
      this.log('Sent ping');
    }
  }

  /**
   * Start ping interval timer
   */
  private startPingInterval(): void {
    this.stopPingInterval();
    // Send ping every 30 seconds
    this.pingTimer = setInterval(() => {
      this.sendPing();
    }, 30000);
  }

  /**
   * Stop ping interval timer
   */
  private stopPingInterval(): void {
    if (this.pingTimer) {
      clearInterval(this.pingTimer);
      this.pingTimer = null;
    }
  }

  /**
   * Schedule reconnection attempt
   */
  private scheduleReconnect(): void {
    if (this.isManualClose) {
      return;
    }

    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      this.log(`Max reconnect attempts (${this.maxReconnectAttempts}) reached`);
      return;
    }

    // Clear any existing reconnect timer
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
    }

    // Calculate exponential backoff delay
    const delay = Math.min(
      this.reconnectDelay * Math.pow(2, this.reconnectAttempts),
      this.maxReconnectDelay
    );

    this.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts + 1}/${this.maxReconnectAttempts})`);

    this.reconnectTimer = setTimeout(() => {
      this.reconnectAttempts++;
      this.connect();
    }, delay);
  }

  /**
   * Manually disconnect WebSocket
   */
  disconnect(): void {
    this.log('Manual disconnect');
    this.isManualClose = true;
    this.stopPingInterval();

    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  /**
   * Send message to server
   */
  send(message: any): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      const messageStr = typeof message === 'string' ? message : JSON.stringify(message);
      this.ws.send(messageStr);
      this.log('Sent message:', message);
    } else {
      this.log('Cannot send message - WebSocket not connected');
    }
  }

  /**
   * Check if WebSocket is connected
   */
  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }

  /**
   * Get connection state
   */
  getState(): number {
    return this.ws?.readyState ?? WebSocket.CLOSED;
  }

  /**
   * Log debug messages
   */
  private log(message: string, ...args: any[]): void {
    if (this.debug) {
      console.log(`[TradingWebSocket:${this.traderID}] ${message}`, ...args);
    }
  }
}

/**
 * Hook for using WebSocket in React components
 */
export function useWebSocket(
  traderID: string | null,
  options: WebSocketOptions
) {
  const [ws, setWs] = React.useState<TradingWebSocket | null>(null);
  const [isConnected, setIsConnected] = React.useState(false);

  React.useEffect(() => {
    if (!traderID) return;

    const websocket = new TradingWebSocket(traderID, {
      ...options,
      onOpen: () => {
        setIsConnected(true);
        options.onOpen?.();
      },
      onClose: () => {
        setIsConnected(false);
        options.onClose?.();
      },
    });

    websocket.connect();
    setWs(websocket);

    return () => {
      websocket.disconnect();
    };
  }, [traderID]);

  return { ws, isConnected };
}

// Add React import at the top if not using it in a React context
import * as React from 'react';
