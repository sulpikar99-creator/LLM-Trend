// User types
export interface User {
  id: string
  username: string
  email: string
  role: string
  created_at: string
}

export interface LoginRequest {
  username: string
  password: string
}

export interface RegisterRequest {
  username: string
  email: string
  password: string
  beta_code?: string
}

export interface AuthResponse {
  success: boolean
  data: {
    token: string
    user_id: string
    username: string
    role: string
  }
}

// Trader types
export interface Trader {
  id: string
  user_id: string
  name: string
  exchange_type: string
  symbol: string
  interval: string
  status: 'stopped' | 'running' | 'error'
  created_at: string
  updated_at: string
  exchange_config?: {
    testnet: boolean
  }
}

export interface CreateTraderRequest {
  name: string
  exchange_type: string
  symbol: string
  interval: string
  api_key: string
  api_secret: string
  testnet: boolean
  strategy_prompt: string
}

// Analytics types
export interface PerformanceMetrics {
  total_pnl: number
  total_trades: number
  win_rate: number
  profit_factor: number
  sharpe_ratio: number
  avg_win: number
  avg_loss: number
  max_consecutive_wins: number
  max_consecutive_losses: number
}

export interface DrawdownAnalysis {
  max_drawdown: number
  current_drawdown: number
  max_drawdown_duration: number
  recovery_time: number
}

export interface MonteCarloResult {
  probability_profit: number
  probability_loss: number
  var_95: number
  var_99: number
  median_return: number
  best_case: number
  worst_case: number
}

// WebSocket types
export interface WebSocketMessage {
  type: 'position_update' | 'order_update' | 'pnl_update' | 'trade_executed' | 'balance_update'
  trader_id: string
  data: any
  timestamp: string
}

// Config types
export interface AIConfig {
  provider: string
  model_id: string
  api_key: string
  base_url: string
  max_tokens: number
  temperature: number
}

export interface RiskLimits {
  max_drawdown_percent: number
  max_position_size_usd: number
  max_leverage: number
  daily_loss_limit_usd: number
}

export interface Config {
  ai_models: AIConfig[]
  risk_limits: RiskLimits
  default_strategy_prompt: string
}

// Decision types
export interface AIDecision {
  cycle_number: number
  timestamp: string
  decision: {
    action: string
    confidence: number
    reasoning: string
    entry_price?: number
    stop_loss?: number
    take_profit?: number
    position_size_pct?: number
    symbol?: string
    leverage?: number
  }
  market_analysis?: {
    price: number
    trend: string
    volatility: string
    indicators: any
  }
  execution_logs?: any
  error?: string
}

export interface DecisionResponse {
  trader_id: string
  decisions: AIDecision[]
  count: number
  total: number
}

// API Response types
export interface ApiResponse<T> {
  success: boolean
  data: T
  error?: string
  message?: string
}
