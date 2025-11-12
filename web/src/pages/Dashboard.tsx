import { useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { motion } from 'framer-motion'
import { TrendingUp, TrendingDown, Bot, DollarSign, Activity, AlertCircle } from 'lucide-react'
import { api } from '@/lib/api'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { DecisionLog } from '@/components/DecisionLog'

export default function Dashboard() {
  const [stats, setStats] = useState({
    totalPnL: 0,
    activeTraders: 0,
    totalTraders: 0,
    winRate: 0,
  })

  const { data: tradersData } = useQuery({
    queryKey: ['traders'],
    queryFn: () => api.listTraders(),
  })

  const { data: performanceData } = useQuery({
    queryKey: ['performance'],
    queryFn: () => api.getPerformance(),
  })

  useEffect(() => {
    if (tradersData?.data && performanceData?.data) {
      const traders = Array.isArray(tradersData.data) ? tradersData.data : []
      const perf = performanceData.data

      setStats({
        totalPnL: perf.overall_metrics?.total_pnl || 0,
        activeTraders: traders.filter((t) => t.status === 'running').length,
        totalTraders: traders.length,
        winRate: perf.overall_metrics?.win_rate || 0,
      })
    }
  }, [tradersData, performanceData])

  const statCards = [
    {
      title: 'Total P&L',
      value: `$${stats.totalPnL.toFixed(2)}`,
      icon: DollarSign,
      color: stats.totalPnL >= 0 ? 'text-success' : 'text-danger',
      bgColor: stats.totalPnL >= 0 ? 'bg-success/10' : 'bg-danger/10',
      trend: stats.totalPnL >= 0 ? <TrendingUp className="w-4 h-4" /> : <TrendingDown className="w-4 h-4" />,
    },
    {
      title: 'Active Traders',
      value: `${stats.activeTraders}/${stats.totalTraders}`,
      icon: Bot,
      color: 'text-primary',
      bgColor: 'bg-primary/10',
      trend: <Activity className="w-4 h-4" />,
    },
    {
      title: 'Win Rate',
      value: `${(stats.winRate * 100).toFixed(1)}%`,
      icon: TrendingUp,
      color: 'text-success',
      bgColor: 'bg-success/10',
      trend: null,
    },
  ]

  return (
    <div className="space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-4xl font-bold mb-2">Dashboard</h1>
        <p className="text-gray-400">Monitor your AI trading performance in real-time</p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {statCards.map((stat, index) => {
          const Icon = stat.icon
          return (
            <motion.div
              key={stat.title}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.1 }}
            >
              <Card hover>
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm text-gray-400 mb-1">{stat.title}</p>
                    <p className={`text-3xl font-bold ${stat.color}`}>{stat.value}</p>
                  </div>
                  <div className={`p-4 rounded-xl ${stat.bgColor}`}>
                    <Icon className={`w-8 h-8 ${stat.color}`} />
                  </div>
                </div>
                {stat.trend && (
                  <div className={`mt-4 flex items-center space-x-2 ${stat.color}`}>
                    {stat.trend}
                    <span className="text-sm">Live</span>
                  </div>
                )}
              </Card>
            </motion.div>
          )
        })}
      </div>

      {/* Quick Info */}
      {stats.totalTraders === 0 && (
        <Card>
          <div className="flex items-center space-x-4">
            <div className="p-3 bg-primary/10 rounded-xl">
              <AlertCircle className="w-6 h-6 text-primary" />
            </div>
            <div>
              <h3 className="font-semibold mb-1">Get Started</h3>
              <p className="text-gray-400 text-sm">
                Create your first AI trader to start automated trading
              </p>
            </div>
          </div>
        </Card>
      )}

      {/* Traders List */}
      {tradersData?.data && Array.isArray(tradersData.data) && tradersData.data.length > 0 && (
        <div>
          <h2 className="text-2xl font-bold mb-4">Your Traders</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {tradersData.data.map((trader) => (
              <Card key={trader.id} hover>
                <div className="flex justify-between items-start mb-4">
                  <div>
                    <h3 className="font-bold text-lg">{trader.name}</h3>
                    <p className="text-sm text-gray-400">{trader.symbol} • {trader.interval}</p>
                  </div>
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
                </div>
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-gray-400">Exchange</span>
                    <span className="font-medium">{trader.exchange_type}</span>
                  </div>
                  {trader.exchange_config?.testnet && (
                    <Badge variant="warning" size="sm">Testnet</Badge>
                  )}
                </div>
              </Card>
            ))}
          </div>
        </div>
      )}

      {/* AI Decision Log */}
      {tradersData?.data && Array.isArray(tradersData.data) && tradersData.data.length > 0 && (
        <div>
          <h2 className="text-2xl font-bold mb-4">Recent AI Decisions</h2>
          <DecisionLog
            traderId={tradersData.data[0].id}
            limit={5}
          />
          {tradersData.data.length > 1 && (
            <p className="text-sm text-gray-500 mt-2 text-center">
              Showing decisions from {tradersData.data[0].name}. Go to Traders page to see decisions for other traders.
            </p>
          )}
        </div>
      )}
    </div>
  )
}
