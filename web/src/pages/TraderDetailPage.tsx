import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { motion } from 'framer-motion'
import { ArrowLeft, Activity, Brain, TrendingUp, BarChart3, Settings } from 'lucide-react'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { LivePositionMonitor } from '@/components/LivePositionMonitor'
import { DecisionLog } from '@/components/DecisionLog'
import { StrategyEditor } from '@/components/StrategyEditor'
import { api } from '@/lib/api'

type TabType = 'positions' | 'decisions' | 'strategy' | 'performance'

export default function TraderDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [activeTab, setActiveTab] = useState<TabType>('positions')

  const { data: traderData, isLoading } = useQuery({
    queryKey: ['trader', id],
    queryFn: () => api.getTrader(id!),
    enabled: !!id,
    refetchInterval: 5000, // Refresh every 5 seconds
  })

  const trader = traderData?.data

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <Activity className="w-12 h-12 text-primary mx-auto mb-4 animate-pulse" />
          <p className="text-gray-400">Loading trader...</p>
        </div>
      </div>
    )
  }

  if (!trader) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <p className="text-gray-400">Trader not found</p>
          <Button onClick={() => navigate('/traders')} className="mt-4">
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Traders
          </Button>
        </div>
      </div>
    )
  }

  const tabs = [
    { id: 'positions' as TabType, label: 'Live Positions', icon: Activity },
    { id: 'decisions' as TabType, label: 'AI Decisions', icon: Brain },
    { id: 'strategy' as TabType, label: 'Strategy', icon: Settings },
    { id: 'performance' as TabType, label: 'Performance', icon: TrendingUp },
  ]

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <Button
            variant="secondary"
            onClick={() => navigate('/traders')}
            className="p-2"
          >
            <ArrowLeft className="w-5 h-5" />
          </Button>
          <div>
            <div className="flex items-center space-x-3">
              <h1 className="text-3xl font-bold">{trader.name}</h1>
              <Badge
                variant={
                  trader.status === 'running'
                    ? 'success'
                    : trader.status === 'error'
                    ? 'danger'
                    : 'default'
                }
              >
                {trader.status}
              </Badge>
              {trader.exchange_config?.testnet && (
                <Badge variant="warning">Testnet</Badge>
              )}
            </div>
            <p className="text-gray-400 mt-1">
              {trader.symbol} • {trader.interval} • {trader.exchange_type}
            </p>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <Card>
        <div className="flex space-x-2 overflow-x-auto">
          {tabs.map((tab) => {
            const Icon = tab.icon
            return (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`flex items-center space-x-2 px-6 py-3 rounded-lg transition-all whitespace-nowrap ${
                  activeTab === tab.id
                    ? 'bg-primary text-white'
                    : 'bg-background text-gray-400 hover:bg-background/50'
                }`}
              >
                <Icon className="w-4 h-4" />
                <span className="font-medium">{tab.label}</span>
              </button>
            )
          })}
        </div>
      </Card>

      {/* Tab Content */}
      <motion.div
        key={activeTab}
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
      >
        {activeTab === 'positions' && (
          <LivePositionMonitor traderId={trader.id} traderName={trader.name} />
        )}

        {activeTab === 'decisions' && (
          <div>
            <DecisionLog traderId={trader.id} limit={20} />
          </div>
        )}

        {activeTab === 'strategy' && (
          <StrategyEditor
            traderId={trader.id}
            currentStrategy={trader.strategy_prompt || ''}
            traderName={trader.name}
          />
        )}

        {activeTab === 'performance' && (
          <div className="space-y-6">
            <Card>
              <div className="text-center py-12">
                <BarChart3 className="w-16 h-16 text-gray-600 mx-auto mb-4" />
                <h3 className="text-xl font-semibold mb-2">Performance Analytics</h3>
                <p className="text-gray-400 mb-4">
                  Detailed performance metrics and charts for {trader.name}
                </p>
                <p className="text-sm text-gray-500">
                  This section will display advanced charts, risk metrics, backtest results, and trade history.
                </p>
              </div>
            </Card>

            {/* Performance Metrics Grid */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <Card>
                <div className="space-y-2">
                  <p className="text-sm text-gray-400">Total Trades</p>
                  <p className="text-2xl font-bold">--</p>
                  <p className="text-xs text-gray-500">Coming soon</p>
                </div>
              </Card>
              <Card>
                <div className="space-y-2">
                  <p className="text-sm text-gray-400">Win Rate</p>
                  <p className="text-2xl font-bold">--</p>
                  <p className="text-xs text-gray-500">Coming soon</p>
                </div>
              </Card>
              <Card>
                <div className="space-y-2">
                  <p className="text-sm text-gray-400">Sharpe Ratio</p>
                  <p className="text-2xl font-bold">--</p>
                  <p className="text-xs text-gray-500">Coming soon</p>
                </div>
              </Card>
            </div>
          </div>
        )}
      </motion.div>
    </div>
  )
}
