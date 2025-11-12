import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { motion } from 'framer-motion'
import {
  LineChart,
  Line,
  AreaChart,
  Area,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts'
import { Activity, DollarSign, Target, AlertTriangle } from 'lucide-react'
import { api } from '@/lib/api'
import { Card } from '@/components/ui/Card'
import { Select } from '@/components/ui/Select'
import type { Trader } from '@/types'

export default function AnalyticsPage() {
  const [selectedTrader, setSelectedTrader] = useState<string>('all')
  const [timeRange, setTimeRange] = useState<string>('7d')

  const { data: tradersData } = useQuery({
    queryKey: ['traders'],
    queryFn: () => api.listTraders(),
  })

  const { data: performanceData, isLoading: isPerfLoading } = useQuery({
    queryKey: ['performance', selectedTrader],
    queryFn: () => api.getPerformance(selectedTrader === 'all' ? undefined : selectedTrader),
  })

  const traders = Array.isArray(tradersData?.data) ? tradersData.data : []
  const performance = performanceData?.data

  // Extract metrics
  const metrics = performance?.overall_metrics || {}
  const traderStats = Array.isArray(performance?.trader_stats) ? performance.trader_stats : []
  const pnlHistory = Array.isArray(performance?.pnl_history) ? performance.pnl_history : []
  const drawdownHistory = Array.isArray(performance?.drawdown_history) ? performance.drawdown_history : []

  // Calculate win/loss distribution
  const winLossData = [
    { name: 'Wins', value: metrics.total_wins || 0, color: '#10b981' },
    { name: 'Losses', value: metrics.total_losses || 0, color: '#ef4444' },
  ]

  // Stats cards
  const statCards = [
    {
      title: 'Total P&L',
      value: `$${(metrics.total_pnl || 0).toFixed(2)}`,
      change: metrics.total_pnl >= 0 ? '+' : '',
      icon: DollarSign,
      color: metrics.total_pnl >= 0 ? 'text-success' : 'text-danger',
      bgColor: metrics.total_pnl >= 0 ? 'bg-success/10' : 'bg-danger/10',
    },
    {
      title: 'Win Rate',
      value: `${((metrics.win_rate || 0) * 100).toFixed(1)}%`,
      change: metrics.win_rate >= 0.5 ? 'Above Average' : 'Below Average',
      icon: Target,
      color: metrics.win_rate >= 0.5 ? 'text-success' : 'text-warning',
      bgColor: metrics.win_rate >= 0.5 ? 'bg-success/10' : 'bg-warning/10',
    },
    {
      title: 'Total Trades',
      value: (metrics.total_trades || 0).toString(),
      change: `${metrics.total_wins || 0}W / ${metrics.total_losses || 0}L`,
      icon: Activity,
      color: 'text-primary',
      bgColor: 'bg-primary/10',
    },
    {
      title: 'Max Drawdown',
      value: `${((metrics.max_drawdown || 0) * 100).toFixed(2)}%`,
      change: metrics.max_drawdown < 0.2 ? 'Good' : 'High Risk',
      icon: AlertTriangle,
      color: metrics.max_drawdown < 0.2 ? 'text-success' : 'text-danger',
      bgColor: metrics.max_drawdown < 0.2 ? 'bg-success/10' : 'bg-danger/10',
    },
  ]

  // Format P&L history for chart
  const pnlChartData = pnlHistory.map((item: any) => ({
    time: new Date(item.timestamp).toLocaleDateString(),
    pnl: item.cumulative_pnl,
    daily: item.daily_pnl,
  }))

  // Format drawdown history for chart
  const drawdownChartData = drawdownHistory.map((item: any) => ({
    time: new Date(item.timestamp).toLocaleDateString(),
    drawdown: item.drawdown_percent * 100,
  }))

  // Format trader performance for comparison
  const traderComparisonData = traderStats.map((stat: any) => ({
    name: stat.trader_name || 'Unknown',
    pnl: stat.total_pnl || 0,
    trades: stat.total_trades || 0,
    winRate: (stat.win_rate || 0) * 100,
  }))

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-4xl font-bold mb-2">Analytics</h1>
          <p className="text-gray-400">Deep dive into your trading performance</p>
        </div>
        <div className="flex space-x-4">
          <Select
            value={selectedTrader}
            onChange={(e) => setSelectedTrader(e.target.value)}
            label=""
            options={[
              { value: 'all', label: 'All Traders' },
              ...traders.map((t) => ({ value: t.id, label: t.name })),
            ]}
          />
          <Select
            value={timeRange}
            onChange={(e) => setTimeRange(e.target.value)}
            label=""
            options={[
              { value: '24h', label: 'Last 24 Hours' },
              { value: '7d', label: 'Last 7 Days' },
              { value: '30d', label: 'Last 30 Days' },
              { value: 'all', label: 'All Time' },
            ]}
          />
        </div>
      </div>

      {isPerfLoading ? (
        <div className="text-center py-12">
          <p className="text-gray-400">Loading analytics...</p>
        </div>
      ) : (
        <>
          {/* Stats Grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
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
                    <div className="flex items-center justify-between mb-4">
                      <div className={`p-3 rounded-xl ${stat.bgColor}`}>
                        <Icon className={`w-6 h-6 ${stat.color}`} />
                      </div>
                    </div>
                    <p className="text-sm text-gray-400 mb-1">{stat.title}</p>
                    <p className={`text-2xl font-bold ${stat.color} mb-1`}>{stat.value}</p>
                    <p className="text-xs text-gray-500">{stat.change}</p>
                  </Card>
                </motion.div>
              )
            })}
          </div>

          {/* Cumulative P&L Chart */}
          {pnlChartData.length > 0 && (
            <Card>
              <h3 className="text-xl font-bold mb-6">Cumulative P&L Over Time</h3>
              <ResponsiveContainer width="100%" height={300}>
                <AreaChart data={pnlChartData}>
                  <defs>
                    <linearGradient id="colorPnl" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#6366f1" stopOpacity={0.3} />
                      <stop offset="95%" stopColor="#6366f1" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                  <XAxis dataKey="time" stroke="#9ca3af" />
                  <YAxis stroke="#9ca3af" />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: '#1f2937',
                      border: '1px solid #374151',
                      borderRadius: '8px',
                    }}
                  />
                  <Legend />
                  <Area
                    type="monotone"
                    dataKey="pnl"
                    stroke="#6366f1"
                    fillOpacity={1}
                    fill="url(#colorPnl)"
                    name="Cumulative P&L ($)"
                  />
                </AreaChart>
              </ResponsiveContainer>
            </Card>
          )}

          {/* Drawdown Chart */}
          {drawdownChartData.length > 0 && (
            <Card>
              <h3 className="text-xl font-bold mb-6">Drawdown Analysis</h3>
              <ResponsiveContainer width="100%" height={250}>
                <LineChart data={drawdownChartData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                  <XAxis dataKey="time" stroke="#9ca3af" />
                  <YAxis stroke="#9ca3af" />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: '#1f2937',
                      border: '1px solid #374151',
                      borderRadius: '8px',
                    }}
                  />
                  <Legend />
                  <Line
                    type="monotone"
                    dataKey="drawdown"
                    stroke="#ef4444"
                    strokeWidth={2}
                    name="Drawdown (%)"
                  />
                </LineChart>
              </ResponsiveContainer>
            </Card>
          )}

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Win/Loss Distribution */}
            {winLossData.some((d) => d.value > 0) && (
              <Card>
                <h3 className="text-xl font-bold mb-6">Win/Loss Distribution</h3>
                <ResponsiveContainer width="100%" height={250}>
                  <PieChart>
                    <Pie
                      data={winLossData}
                      cx="50%"
                      cy="50%"
                      labelLine={false}
                      label={({ name, value }) => `${name}: ${value}`}
                      outerRadius={80}
                      fill="#8884d8"
                      dataKey="value"
                    >
                      {winLossData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={entry.color} />
                      ))}
                    </Pie>
                    <Tooltip />
                  </PieChart>
                </ResponsiveContainer>
              </Card>
            )}

            {/* Trader Comparison */}
            {traderComparisonData.length > 0 && (
              <Card>
                <h3 className="text-xl font-bold mb-6">Trader Performance Comparison</h3>
                <ResponsiveContainer width="100%" height={250}>
                  <BarChart data={traderComparisonData}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                    <XAxis dataKey="name" stroke="#9ca3af" />
                    <YAxis stroke="#9ca3af" />
                    <Tooltip
                      contentStyle={{
                        backgroundColor: '#1f2937',
                        border: '1px solid #374151',
                        borderRadius: '8px',
                      }}
                    />
                    <Legend />
                    <Bar dataKey="pnl" fill="#6366f1" name="P&L ($)" />
                  </BarChart>
                </ResponsiveContainer>
              </Card>
            )}
          </div>

          {/* Performance Metrics Table */}
          {traderStats.length > 0 && (
            <Card>
              <h3 className="text-xl font-bold mb-6">Detailed Trader Statistics</h3>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-border">
                      <th className="text-left py-3 px-4 font-semibold text-gray-400">Trader</th>
                      <th className="text-right py-3 px-4 font-semibold text-gray-400">P&L</th>
                      <th className="text-right py-3 px-4 font-semibold text-gray-400">Trades</th>
                      <th className="text-right py-3 px-4 font-semibold text-gray-400">Win Rate</th>
                      <th className="text-right py-3 px-4 font-semibold text-gray-400">Avg Win</th>
                      <th className="text-right py-3 px-4 font-semibold text-gray-400">Avg Loss</th>
                      <th className="text-right py-3 px-4 font-semibold text-gray-400">Sharpe Ratio</th>
                    </tr>
                  </thead>
                  <tbody>
                    {traderStats.map((stat: any, index: number) => (
                      <tr key={index} className="border-b border-border/50 hover:bg-white/5">
                        <td className="py-3 px-4 font-medium">{stat.trader_name || 'Unknown'}</td>
                        <td
                          className={`py-3 px-4 text-right font-semibold ${
                            stat.total_pnl >= 0 ? 'text-success' : 'text-danger'
                          }`}
                        >
                          ${(stat.total_pnl || 0).toFixed(2)}
                        </td>
                        <td className="py-3 px-4 text-right">{stat.total_trades || 0}</td>
                        <td className="py-3 px-4 text-right">
                          {((stat.win_rate || 0) * 100).toFixed(1)}%
                        </td>
                        <td className="py-3 px-4 text-right text-success">
                          ${(stat.avg_win || 0).toFixed(2)}
                        </td>
                        <td className="py-3 px-4 text-right text-danger">
                          ${(stat.avg_loss || 0).toFixed(2)}
                        </td>
                        <td className="py-3 px-4 text-right">
                          {(stat.sharpe_ratio || 0).toFixed(2)}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </Card>
          )}

          {/* No Data State */}
          {!performance || (traderStats.length === 0 && pnlHistory.length === 0) && (
            <Card>
              <div className="text-center py-12">
                <Activity className="w-16 h-16 text-gray-600 mx-auto mb-4" />
                <p className="text-gray-400 mb-2">No trading data available yet</p>
                <p className="text-gray-500 text-sm">
                  Start your traders to see performance analytics and charts
                </p>
              </div>
            </Card>
          )}
        </>
      )}
    </div>
  )
}
