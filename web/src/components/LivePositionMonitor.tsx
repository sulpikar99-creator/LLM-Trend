import { useState, useEffect } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { TrendingUp, TrendingDown, Activity, DollarSign, Target, Zap, Clock, Brain } from 'lucide-react'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { useWebSocket } from '@/hooks/useWebSocket'

interface LivePositionMonitorProps {
  traderId: string
  traderName: string
}

export function LivePositionMonitor({ traderId, traderName }: LivePositionMonitorProps) {
  const [positions, setPositions] = useState<any[]>([])
  const [balance, setBalance] = useState<any>(null)
  const [lastUpdate, setLastUpdate] = useState<Date | null>(null)
  const [recentTrades, setRecentTrades] = useState<any[]>([])
  const [aiStatus, setAiStatus] = useState<'idle' | 'analyzing' | 'executing'>('idle')

  const { isConnected, lastMessage } = useWebSocket({
    traderId,
    onMessage: (message) => {
      setLastUpdate(new Date())

      switch (message.type) {
        case 'position_update':
          setPositions(message.data.positions || [])
          break
        case 'balance_update':
          setBalance(message.data)
          break
        case 'trade_executed':
          setRecentTrades((prev) => [message.data, ...prev].slice(0, 10))
          setAiStatus('idle')
          break
        case 'market_update':
          // Update live price
          break
        case 'pnl_update':
          // Update PnL
          break
      }
    },
  })

  // Simulate AI status based on trading activity
  useEffect(() => {
    if (lastMessage?.type === 'trade_executed') {
      setAiStatus('executing')
      setTimeout(() => setAiStatus('idle'), 2000)
    }
  }, [lastMessage])

  const totalPnL = positions.reduce((sum, pos) => sum + (pos.unrealized_pnl || 0), 0)
  const totalPositionValue = positions.reduce((sum, pos) => sum + (pos.notional || 0), 0)

  return (
    <div className="space-y-4">
      {/* Header with Connection Status */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-3">
          <Activity className={`w-5 h-5 ${isConnected ? 'text-success' : 'text-danger'}`} />
          <div>
            <h3 className="font-bold">{traderName}</h3>
            <p className="text-xs text-gray-500">
              {isConnected ? 'Connected • Live Updates' : 'Disconnected'}
            </p>
          </div>
        </div>
        <div className="flex items-center space-x-2">
          {aiStatus === 'analyzing' && (
            <Badge variant="warning">
              <Brain className="w-3 h-3 mr-1 animate-pulse" />
              AI Analyzing
            </Badge>
          )}
          {aiStatus === 'executing' && (
            <Badge variant="success">
              <Zap className="w-3 h-3 mr-1" />
              Executing Trade
            </Badge>
          )}
          {aiStatus === 'idle' && (
            <Badge variant="default">
              <Clock className="w-3 h-3 mr-1" />
              Monitoring
            </Badge>
          )}
        </div>
      </div>

      {/* Balance Overview */}
      {balance && (
        <Card>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div>
              <p className="text-xs text-gray-500 mb-1">Total Balance</p>
              <p className="text-xl font-bold">${balance.balance?.toFixed(2) || '0.00'}</p>
            </div>
            <div>
              <p className="text-xs text-gray-500 mb-1">Available</p>
              <p className="text-lg font-semibold text-success">
                ${balance.available_balance?.toFixed(2) || '0.00'}
              </p>
            </div>
            <div>
              <p className="text-xs text-gray-500 mb-1">Unrealized PnL</p>
              <p className={`text-lg font-semibold ${balance.unrealized_profit >= 0 ? 'text-success' : 'text-danger'}`}>
                {balance.unrealized_profit >= 0 ? '+' : ''}${balance.unrealized_profit?.toFixed(2) || '0.00'}
              </p>
            </div>
            <div>
              <p className="text-xs text-gray-500 mb-1">Margin Ratio</p>
              <p className="text-lg font-semibold">
                {balance.margin_ratio ? `${(balance.margin_ratio * 100).toFixed(2)}%` : 'N/A'}
              </p>
            </div>
          </div>
        </Card>
      )}

      {/* Active Positions */}
      <Card>
        <h4 className="font-bold mb-4 flex items-center">
          <Target className="w-4 h-4 mr-2" />
          Active Positions ({positions.length})
        </h4>

        {positions.length === 0 ? (
          <div className="text-center py-8 text-gray-500">
            <Activity className="w-12 h-12 mx-auto mb-2 opacity-50" />
            <p>No active positions</p>
            <p className="text-sm">AI is monitoring the market</p>
          </div>
        ) : (
          <div className="space-y-3">
            <AnimatePresence>
              {positions.map((position, index) => (
                <motion.div
                  key={position.symbol || index}
                  initial={{ opacity: 0, x: -20 }}
                  animate={{ opacity: 1, x: 0 }}
                  exit={{ opacity: 0, x: 20 }}
                  transition={{ delay: index * 0.1 }}
                >
                  <div className="p-4 border border-border rounded-lg hover:border-primary/50 transition-colors">
                    <div className="flex items-start justify-between mb-3">
                      <div className="flex items-center space-x-2">
                        {position.side === 'LONG' ? (
                          <TrendingUp className="w-5 h-5 text-success" />
                        ) : (
                          <TrendingDown className="w-5 h-5 text-danger" />
                        )}
                        <div>
                          <h5 className="font-bold">{position.symbol}</h5>
                          <Badge variant={position.side === 'LONG' ? 'success' : 'danger'} size="sm">
                            {position.side} {position.leverage}x
                          </Badge>
                        </div>
                      </div>
                      <div className="text-right">
                        <p className={`text-lg font-bold ${position.unrealized_pnl >= 0 ? 'text-success' : 'text-danger'}`}>
                          {position.unrealized_pnl >= 0 ? '+' : ''}${position.unrealized_pnl?.toFixed(2) || '0.00'}
                        </p>
                        <p className="text-xs text-gray-500">
                          {position.pnl_percentage ? `${position.pnl_percentage.toFixed(2)}%` : ''}
                        </p>
                      </div>
                    </div>

                    <div className="grid grid-cols-2 md:grid-cols-4 gap-3 text-sm">
                      <div>
                        <p className="text-gray-500 text-xs">Size</p>
                        <p className="font-medium">{position.size?.toFixed(4) || '0.0000'}</p>
                      </div>
                      <div>
                        <p className="text-gray-500 text-xs">Entry Price</p>
                        <p className="font-medium">${position.entry_price?.toFixed(2) || '0.00'}</p>
                      </div>
                      <div>
                        <p className="text-gray-500 text-xs">Mark Price</p>
                        <p className="font-medium">${position.mark_price?.toFixed(2) || '0.00'}</p>
                      </div>
                      <div>
                        <p className="text-gray-500 text-xs">Liquidation</p>
                        <p className="font-medium text-danger">
                          ${position.liquidation_price?.toFixed(2) || 'N/A'}
                        </p>
                      </div>
                    </div>
                  </div>
                </motion.div>
              ))}
            </AnimatePresence>
          </div>
        )}
      </Card>

      {/* Recent AI Trades */}
      {recentTrades.length > 0 && (
        <Card>
          <h4 className="font-bold mb-4 flex items-center">
            <Zap className="w-4 h-4 mr-2" />
            Recent AI Trades
          </h4>
          <div className="space-y-2">
            {recentTrades.map((trade, index) => (
              <motion.div
                key={index}
                initial={{ opacity: 0, y: -10 }}
                animate={{ opacity: 1, y: 0 }}
                className="flex items-center justify-between p-3 bg-background/30 rounded-lg"
              >
                <div className="flex items-center space-x-3">
                  {trade.side === 'BUY' ? (
                    <TrendingUp className="w-4 h-4 text-success" />
                  ) : (
                    <TrendingDown className="w-4 h-4 text-danger" />
                  )}
                  <div>
                    <p className="text-sm font-medium">
                      {trade.side} {trade.symbol}
                    </p>
                    <p className="text-xs text-gray-500">
                      {trade.quantity} @ ${trade.price?.toFixed(2)}
                    </p>
                  </div>
                </div>
                <div className="text-right">
                  <p className="text-sm font-medium">
                    ${trade.total?.toFixed(2)}
                  </p>
                  <p className="text-xs text-gray-500">
                    {new Date(trade.timestamp).toLocaleTimeString()}
                  </p>
                </div>
              </motion.div>
            ))}
          </div>
        </Card>
      )}

      {/* Summary Stats */}
      <div className="grid grid-cols-2 gap-4">
        <Card>
          <div className="flex items-center space-x-3">
            <div className="p-3 bg-success/10 rounded-xl">
              <DollarSign className="w-6 h-6 text-success" />
            </div>
            <div>
              <p className="text-xs text-gray-500">Total PnL</p>
              <p className={`text-xl font-bold ${totalPnL >= 0 ? 'text-success' : 'text-danger'}`}>
                {totalPnL >= 0 ? '+' : ''}${totalPnL.toFixed(2)}
              </p>
            </div>
          </div>
        </Card>
        <Card>
          <div className="flex items-center space-x-3">
            <div className="p-3 bg-primary/10 rounded-xl">
              <Activity className="w-6 h-6 text-primary" />
            </div>
            <div>
              <p className="text-xs text-gray-500">Position Value</p>
              <p className="text-xl font-bold">${totalPositionValue.toFixed(2)}</p>
            </div>
          </div>
        </Card>
      </div>

      {/* Last Update Time */}
      {lastUpdate && (
        <p className="text-xs text-gray-500 text-center">
          Last update: {lastUpdate.toLocaleTimeString()}
        </p>
      )}
    </div>
  )
}
