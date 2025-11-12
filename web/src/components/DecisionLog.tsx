import { useQuery } from '@tanstack/react-query'
import { motion } from 'framer-motion'
import { Brain, TrendingUp, TrendingDown, Activity, Clock, Target } from 'lucide-react'
import { api } from '@/lib/api'
import { Card } from '@/components/ui/Card'
import type { AIDecision } from '@/types'

interface DecisionLogProps {
  traderId: string
  limit?: number
}

export function DecisionLog({ traderId, limit = 10 }: DecisionLogProps) {
  const { data, isLoading, error } = useQuery({
    queryKey: ['decisions', traderId, limit],
    queryFn: () => api.getDecisions(traderId, limit),
    refetchInterval: 30000, // Refresh every 30 seconds
  })

  const decisions = data?.data?.decisions || []

  if (isLoading) {
    return (
      <Card>
        <div className="text-center py-8">
          <Brain className="w-12 h-12 text-primary mx-auto mb-4 animate-pulse" />
          <p className="text-gray-400">Loading AI decisions...</p>
        </div>
      </Card>
    )
  }

  if (error) {
    return (
      <Card>
        <div className="text-center py-8">
          <p className="text-red-400">Failed to load decisions</p>
        </div>
      </Card>
    )
  }

  if (decisions.length === 0) {
    return (
      <Card>
        <div className="text-center py-8">
          <Brain className="w-12 h-12 text-gray-600 mx-auto mb-4" />
          <p className="text-gray-400 mb-2">No AI decisions yet</p>
          <p className="text-gray-500 text-sm">
            The AI will make decisions once the trader starts monitoring the market
          </p>
        </div>
      </Card>
    )
  }

  const getActionIcon = (action: string) => {
    switch (action.toLowerCase()) {
      case 'buy':
      case 'long':
        return <TrendingUp className="w-5 h-5 text-success" />
      case 'sell':
      case 'short':
        return <TrendingDown className="w-5 h-5 text-danger" />
      case 'hold':
        return <Activity className="w-5 h-5 text-warning" />
      default:
        return <Target className="w-5 h-5 text-gray-400" />
    }
  }

  const getActionColor = (action: string) => {
    switch (action.toLowerCase()) {
      case 'buy':
      case 'long':
        return 'text-success border-success/20 bg-success/5'
      case 'sell':
      case 'short':
        return 'text-danger border-danger/20 bg-danger/5'
      case 'hold':
        return 'text-warning border-warning/20 bg-warning/5'
      default:
        return 'text-gray-400 border-gray-600/20 bg-gray-800/5'
    }
  }

  const getConfidenceColor = (confidence: number) => {
    if (confidence >= 0.7) return 'text-success'
    if (confidence >= 0.5) return 'text-warning'
    return 'text-danger'
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center space-x-2">
          <Brain className="w-5 h-5 text-primary" />
          <h3 className="text-lg font-semibold">AI Decision Log</h3>
        </div>
        <span className="text-sm text-gray-400">{decisions.length} decisions</span>
      </div>

      <div className="space-y-3">
        {decisions.map((decision: AIDecision, index: number) => {
          const decisionData = decision.decision || {}
          const action = decisionData.action || 'unknown'
          const confidence = decisionData.confidence || 0
          const reasoning = decisionData.reasoning || 'No reasoning provided'

          return (
            <motion.div
              key={decision.cycle_number || index}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.05 }}
            >
              <Card hover className={`border ${getActionColor(action)}`}>
                <div className="space-y-3">
                  {/* Header */}
                  <div className="flex items-start justify-between">
                    <div className="flex items-center space-x-3">
                      {getActionIcon(action)}
                      <div>
                        <div className="flex items-center space-x-2">
                          <span className="font-semibold uppercase">{action}</span>
                          <span className={`text-sm ${getConfidenceColor(confidence)}`}>
                            {(confidence * 100).toFixed(0)}% confidence
                          </span>
                        </div>
                        <div className="flex items-center space-x-2 text-xs text-gray-500 mt-1">
                          <Clock className="w-3 h-3" />
                          <span>{new Date(decision.timestamp).toLocaleString()}</span>
                          {decision.cycle_number && (
                            <span className="ml-2">Cycle #{decision.cycle_number}</span>
                          )}
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* Reasoning */}
                  <div className="pl-8">
                    <p className="text-sm text-gray-300 italic">"{reasoning}"</p>
                  </div>

                  {/* Trade Details */}
                  {(decisionData.entry_price || decisionData.stop_loss || decisionData.take_profit) && (
                    <div className="pl-8 grid grid-cols-2 md:grid-cols-4 gap-3 pt-2 border-t border-border/30">
                      {decisionData.entry_price && (
                        <div>
                          <p className="text-xs text-gray-500">Entry</p>
                          <p className="text-sm font-medium">${decisionData.entry_price.toFixed(2)}</p>
                        </div>
                      )}
                      {decisionData.stop_loss && (
                        <div>
                          <p className="text-xs text-gray-500">Stop Loss</p>
                          <p className="text-sm font-medium text-danger">${decisionData.stop_loss.toFixed(2)}</p>
                        </div>
                      )}
                      {decisionData.take_profit && (
                        <div>
                          <p className="text-xs text-gray-500">Take Profit</p>
                          <p className="text-sm font-medium text-success">${decisionData.take_profit.toFixed(2)}</p>
                        </div>
                      )}
                      {decisionData.position_size_pct && (
                        <div>
                          <p className="text-xs text-gray-500">Position Size</p>
                          <p className="text-sm font-medium">{decisionData.position_size_pct.toFixed(1)}%</p>
                        </div>
                      )}
                    </div>
                  )}

                  {/* Market Analysis */}
                  {decision.market_analysis && (
                    <div className="pl-8 pt-2 border-t border-border/30">
                      <p className="text-xs text-gray-500 mb-2">Market Analysis</p>
                      <div className="grid grid-cols-3 gap-2 text-sm">
                        {decision.market_analysis.price && (
                          <div>
                            <span className="text-gray-500">Price: </span>
                            <span className="font-medium">${decision.market_analysis.price.toFixed(2)}</span>
                          </div>
                        )}
                        {decision.market_analysis.trend && (
                          <div>
                            <span className="text-gray-500">Trend: </span>
                            <span className="font-medium">{decision.market_analysis.trend}</span>
                          </div>
                        )}
                        {decision.market_analysis.volatility && (
                          <div>
                            <span className="text-gray-500">Volatility: </span>
                            <span className="font-medium">{decision.market_analysis.volatility}</span>
                          </div>
                        )}
                      </div>
                    </div>
                  )}

                  {/* Error Display */}
                  {decision.error && (
                    <div className="pl-8 pt-2 border-t border-border/30">
                      <p className="text-xs text-red-400">Error: {decision.error}</p>
                    </div>
                  )}
                </div>
              </Card>
            </motion.div>
          )
        })}
      </div>
    </div>
  )
}
