import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { api } from '../lib/api'

export default function AnalyticsPage() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [activeTab, setActiveTab] = useState<'performance' | 'drawdown' | 'montecarlo' | 'correlation'>('performance')

  // Performance data
  const [performanceData, setPerformanceData] = useState<any>(null)
  const [drawdownData, setDrawdownData] = useState<any>(null)
  const [monteCarloData, setMonteCarloData] = useState<any>(null)
  const [correlationData, setCorrelationData] = useState<any>(null)

  useEffect(() => {
    loadAnalytics()
  }, [])

  const loadAnalytics = async () => {
    try {
      setLoading(true)
      setError('')

      // Load all analytics data in parallel
      const [perfResp, ddResp, mcResp, corrResp] = await Promise.all([
        api.getPerformance().catch(() => ({ data: null })),
        api.getDrawdown().catch(() => ({ data: null })),
        api.getMonteCarlo().catch(() => ({ data: null })),
        api.getCorrelation().catch(() => ({ data: null })),
      ])

      setPerformanceData(perfResp.data)
      setDrawdownData(ddResp.data)
      setMonteCarloData(mcResp.data)
      setCorrelationData(corrResp.data)
    } catch (err: any) {
      setError(err.message || 'Failed to load analytics')
    } finally {
      setLoading(false)
    }
  }

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(value)
  }

  const formatPercent = (value: number) => {
    return `${(value * 100).toFixed(2)}%`
  }

  return (
    <div className="min-h-screen bg-background">
      {/* Navigation */}
      <nav className="bg-card border-b border-border">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16 items-center">
            <div className="flex items-center space-x-8">
              <h1 className="text-xl font-bold text-primary">LLM Trend</h1>
              <div className="flex space-x-4">
                <button
                  onClick={() => navigate('/')}
                  className="text-gray-400 hover:text-white"
                >
                  Dashboard
                </button>
                <button
                  onClick={() => navigate('/traders')}
                  className="text-gray-400 hover:text-white"
                >
                  Traders
                </button>
                <button
                  onClick={() => navigate('/analytics')}
                  className="text-primary font-semibold"
                >
                  Analytics
                </button>
                <button
                  onClick={() => navigate('/settings')}
                  className="text-gray-400 hover:text-white"
                >
                  Settings
                </button>
              </div>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-gray-400">Welcome, {user?.username}</span>
              <button
                onClick={logout}
                className="px-4 py-2 bg-danger hover:bg-danger/90 text-white rounded"
              >
                Logout
              </button>
            </div>
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-8">
          <h2 className="text-3xl font-bold">Analytics</h2>
          <p className="text-gray-400 mt-2">Comprehensive trading performance analysis</p>
        </div>

        {error && (
          <div className="mb-6 bg-danger/10 border border-danger text-danger px-4 py-3 rounded">
            {error}
          </div>
        )}

        {/* Tabs */}
        <div className="flex space-x-2 mb-6 border-b border-border">
          <button
            onClick={() => setActiveTab('performance')}
            className={`px-4 py-2 font-medium border-b-2 transition-colors ${
              activeTab === 'performance'
                ? 'border-primary text-primary'
                : 'border-transparent text-gray-400 hover:text-white'
            }`}
          >
            Performance
          </button>
          <button
            onClick={() => setActiveTab('drawdown')}
            className={`px-4 py-2 font-medium border-b-2 transition-colors ${
              activeTab === 'drawdown'
                ? 'border-primary text-primary'
                : 'border-transparent text-gray-400 hover:text-white'
            }`}
          >
            Drawdown
          </button>
          <button
            onClick={() => setActiveTab('montecarlo')}
            className={`px-4 py-2 font-medium border-b-2 transition-colors ${
              activeTab === 'montecarlo'
                ? 'border-primary text-primary'
                : 'border-transparent text-gray-400 hover:text-white'
            }`}
          >
            Monte Carlo
          </button>
          <button
            onClick={() => setActiveTab('correlation')}
            className={`px-4 py-2 font-medium border-b-2 transition-colors ${
              activeTab === 'correlation'
                ? 'border-primary text-primary'
                : 'border-transparent text-gray-400 hover:text-white'
            }`}
          >
            Correlation
          </button>
        </div>

        {loading ? (
          <div className="text-center py-12">
            <p className="text-gray-400">Loading analytics...</p>
          </div>
        ) : (
          <>
            {/* Performance Tab */}
            {activeTab === 'performance' && (
              <div className="space-y-6">
                {performanceData && performanceData.overall_metrics ? (
                  <>
                    {/* Overall Metrics */}
                    <div className="bg-card border border-border rounded-lg p-6">
                      <h3 className="text-xl font-bold mb-4">Overall Performance</h3>
                      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                        <div>
                          <p className="text-sm text-gray-400 mb-1">Total P&L</p>
                          <p className={`text-2xl font-bold ${performanceData.overall_metrics.total_pnl >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                            {formatCurrency(performanceData.overall_metrics.total_pnl || 0)}
                          </p>
                        </div>
                        <div>
                          <p className="text-sm text-gray-400 mb-1">Total Trades</p>
                          <p className="text-2xl font-bold">{performanceData.total_trades || 0}</p>
                        </div>
                        <div>
                          <p className="text-sm text-gray-400 mb-1">Win Rate</p>
                          <p className="text-2xl font-bold text-primary">
                            {formatPercent(performanceData.overall_metrics.win_rate || 0)}
                          </p>
                        </div>
                        <div>
                          <p className="text-sm text-gray-400 mb-1">Profit Factor</p>
                          <p className="text-2xl font-bold">
                            {(performanceData.overall_metrics.profit_factor || 0).toFixed(2)}
                          </p>
                        </div>
                        <div>
                          <p className="text-sm text-gray-400 mb-1">Sharpe Ratio</p>
                          <p className="text-2xl font-bold">
                            {(performanceData.overall_metrics.sharpe_ratio || 0).toFixed(2)}
                          </p>
                        </div>
                        <div>
                          <p className="text-sm text-gray-400 mb-1">Avg Win</p>
                          <p className="text-2xl font-bold text-green-500">
                            {formatCurrency(performanceData.overall_metrics.avg_win || 0)}
                          </p>
                        </div>
                        <div>
                          <p className="text-sm text-gray-400 mb-1">Avg Loss</p>
                          <p className="text-2xl font-bold text-red-500">
                            {formatCurrency(performanceData.overall_metrics.avg_loss || 0)}
                          </p>
                        </div>
                        <div>
                          <p className="text-sm text-gray-400 mb-1">Max Streak</p>
                          <p className="text-2xl font-bold">
                            {performanceData.overall_metrics.max_consecutive_wins || 0}W /
                            {performanceData.overall_metrics.max_consecutive_losses || 0}L
                          </p>
                        </div>
                      </div>
                    </div>

                    {/* By Symbol */}
                    {performanceData.by_symbol && Object.keys(performanceData.by_symbol).length > 0 && (
                      <div className="bg-card border border-border rounded-lg p-6">
                        <h3 className="text-xl font-bold mb-4">Performance by Symbol</h3>
                        <div className="overflow-x-auto">
                          <table className="w-full">
                            <thead>
                              <tr className="border-b border-border">
                                <th className="text-left py-2 px-4 text-gray-400">Symbol</th>
                                <th className="text-right py-2 px-4 text-gray-400">Trades</th>
                                <th className="text-right py-2 px-4 text-gray-400">P&L</th>
                                <th className="text-right py-2 px-4 text-gray-400">Win Rate</th>
                              </tr>
                            </thead>
                            <tbody>
                              {Object.entries(performanceData.by_symbol).map(([symbol, data]: [string, any]) => (
                                <tr key={symbol} className="border-b border-border/50">
                                  <td className="py-2 px-4 font-medium">{symbol}</td>
                                  <td className="text-right py-2 px-4">{data.trades || 0}</td>
                                  <td className={`text-right py-2 px-4 font-bold ${data.total_pnl >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                                    {formatCurrency(data.total_pnl || 0)}
                                  </td>
                                  <td className="text-right py-2 px-4">{formatPercent(data.win_rate || 0)}</td>
                                </tr>
                              ))}
                            </tbody>
                          </table>
                        </div>
                      </div>
                    )}

                    {/* Top/Worst Performers */}
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                      {performanceData.top_performers && performanceData.top_performers.length > 0 && (
                        <div className="bg-card border border-border rounded-lg p-6">
                          <h3 className="text-xl font-bold mb-4 text-green-500">Top Performers</h3>
                          <div className="space-y-2">
                            {performanceData.top_performers.map((item: any, idx: number) => (
                              <div key={idx} className="flex justify-between items-center">
                                <span className="text-gray-400">{item.name}</span>
                                <span className="text-green-500 font-bold">{formatCurrency(item.pnl)}</span>
                              </div>
                            ))}
                          </div>
                        </div>
                      )}
                      {performanceData.worst_performers && performanceData.worst_performers.length > 0 && (
                        <div className="bg-card border border-border rounded-lg p-6">
                          <h3 className="text-xl font-bold mb-4 text-red-500">Worst Performers</h3>
                          <div className="space-y-2">
                            {performanceData.worst_performers.map((item: any, idx: number) => (
                              <div key={idx} className="flex justify-between items-center">
                                <span className="text-gray-400">{item.name}</span>
                                <span className="text-red-500 font-bold">{formatCurrency(item.pnl)}</span>
                              </div>
                            ))}
                          </div>
                        </div>
                      )}
                    </div>
                  </>
                ) : (
                  <div className="bg-card border border-border rounded-lg p-12 text-center">
                    <p className="text-gray-400">No performance data available yet. Start trading to see analytics!</p>
                  </div>
                )}
              </div>
            )}

            {/* Drawdown Tab */}
            {activeTab === 'drawdown' && (
              <div className="space-y-6">
                {drawdownData && drawdownData.drawdown_analysis ? (
                  <>
                    <div className="bg-card border border-border rounded-lg p-6">
                      <h3 className="text-xl font-bold mb-4">Drawdown Analysis</h3>
                      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                        <div>
                          <p className="text-sm text-gray-400 mb-1">Max Drawdown</p>
                          <p className="text-2xl font-bold text-red-500">
                            {formatPercent(drawdownData.drawdown_analysis.max_drawdown || 0)}
                          </p>
                        </div>
                        <div>
                          <p className="text-sm text-gray-400 mb-1">Current Drawdown</p>
                          <p className="text-2xl font-bold">
                            {formatPercent(drawdownData.drawdown_analysis.current_drawdown || 0)}
                          </p>
                        </div>
                        <div>
                          <p className="text-sm text-gray-400 mb-1">Max DD Duration</p>
                          <p className="text-2xl font-bold">
                            {(drawdownData.drawdown_analysis.max_drawdown_duration || 0).toFixed(0)} days
                          </p>
                        </div>
                        <div>
                          <p className="text-sm text-gray-400 mb-1">Recovery Time</p>
                          <p className="text-2xl font-bold">
                            {(drawdownData.drawdown_analysis.recovery_time || 0).toFixed(0)} days
                          </p>
                        </div>
                      </div>
                    </div>

                    {drawdownData.equity_points && drawdownData.equity_points.length > 0 && (
                      <div className="bg-card border border-border rounded-lg p-6">
                        <h3 className="text-xl font-bold mb-4">Equity Curve</h3>
                        <p className="text-gray-400">
                          Tracking {drawdownData.equity_points.length} data points
                        </p>
                      </div>
                    )}
                  </>
                ) : (
                  <div className="bg-card border border-border rounded-lg p-12 text-center">
                    <p className="text-gray-400">No drawdown data available yet. Trading history is needed for drawdown analysis.</p>
                  </div>
                )}
              </div>
            )}

            {/* Monte Carlo Tab */}
            {activeTab === 'montecarlo' && (
              <div className="space-y-6">
                {monteCarloData ? (
                  <>
                    <div className="bg-card border border-border rounded-lg p-6">
                      <h3 className="text-xl font-bold mb-4">Monte Carlo Simulation Results</h3>
                      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                        <div>
                          <p className="text-sm text-gray-400 mb-1">Profit Probability</p>
                          <p className="text-2xl font-bold text-green-500">
                            {formatPercent(monteCarloData.probability_profit || 0)}
                          </p>
                        </div>
                        <div>
                          <p className="text-sm text-gray-400 mb-1">Loss Probability</p>
                          <p className="text-2xl font-bold text-red-500">
                            {formatPercent(monteCarloData.probability_loss || 0)}
                          </p>
                        </div>
                        <div>
                          <p className="text-sm text-gray-400 mb-1">VaR (95%)</p>
                          <p className="text-2xl font-bold">
                            {formatCurrency(monteCarloData.var_95 || 0)}
                          </p>
                        </div>
                        <div>
                          <p className="text-sm text-gray-400 mb-1">VaR (99%)</p>
                          <p className="text-2xl font-bold">
                            {formatCurrency(monteCarloData.var_99 || 0)}
                          </p>
                        </div>
                      </div>
                    </div>

                    {monteCarloData.summary && (
                      <div className="bg-card border border-border rounded-lg p-6">
                        <h3 className="text-xl font-bold mb-4">Summary</h3>
                        <pre className="text-gray-400 whitespace-pre-wrap text-sm">
                          {monteCarloData.summary}
                        </pre>
                      </div>
                    )}

                    {monteCarloData.config && (
                      <div className="bg-card border border-border rounded-lg p-6">
                        <h3 className="text-xl font-bold mb-4">Simulation Parameters</h3>
                        <div className="grid grid-cols-2 md:grid-cols-3 gap-4 text-sm">
                          <div>
                            <p className="text-gray-400">Simulations</p>
                            <p className="font-bold">{monteCarloData.config.num_simulations}</p>
                          </div>
                          <div>
                            <p className="text-gray-400">Periods</p>
                            <p className="font-bold">{monteCarloData.config.num_periods}</p>
                          </div>
                          <div>
                            <p className="text-gray-400">Initial Balance</p>
                            <p className="font-bold">{formatCurrency(monteCarloData.config.initial_balance)}</p>
                          </div>
                        </div>
                      </div>
                    )}
                  </>
                ) : (
                  <div className="bg-card border border-border rounded-lg p-12 text-center">
                    <p className="text-gray-400">No Monte Carlo simulation data available.</p>
                  </div>
                )}
              </div>
            )}

            {/* Correlation Tab */}
            {activeTab === 'correlation' && (
              <div className="space-y-6">
                {correlationData && correlationData.matrix ? (
                  <>
                    <div className="bg-card border border-border rounded-lg p-6">
                      <h3 className="text-xl font-bold mb-4">Correlation Statistics</h3>
                      <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
                        <div>
                          <p className="text-sm text-gray-400 mb-1">Avg Correlation</p>
                          <p className="text-2xl font-bold">
                            {(correlationData.avg_correlation || 0).toFixed(3)}
                          </p>
                        </div>
                        <div>
                          <p className="text-sm text-gray-400 mb-1">Max Correlation</p>
                          <p className="text-2xl font-bold text-green-500">
                            {(correlationData.max_correlation || 0).toFixed(3)}
                          </p>
                        </div>
                        <div>
                          <p className="text-sm text-gray-400 mb-1">Min Correlation</p>
                          <p className="text-2xl font-bold text-red-500">
                            {(correlationData.min_correlation || 0).toFixed(3)}
                          </p>
                        </div>
                      </div>
                    </div>

                    {correlationData.top_correlations && correlationData.top_correlations.length > 0 && (
                      <div className="bg-card border border-border rounded-lg p-6">
                        <h3 className="text-xl font-bold mb-4">Top Correlations</h3>
                        <div className="space-y-2">
                          {correlationData.top_correlations.map((corr: any, idx: number) => (
                            <div key={idx} className="flex justify-between items-center border-b border-border/50 pb-2">
                              <span className="text-gray-400">{corr.pair || `Pair ${idx + 1}`}</span>
                              <span className={`font-bold ${corr.value >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                                {(corr.value || 0).toFixed(3)}
                              </span>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}

                    {correlationData.summary && (
                      <div className="bg-card border border-border rounded-lg p-6">
                        <h3 className="text-xl font-bold mb-4">Analysis Summary</h3>
                        <p className="text-gray-400 whitespace-pre-wrap">{correlationData.summary}</p>
                      </div>
                    )}
                  </>
                ) : (
                  <div className="bg-card border border-border rounded-lg p-12 text-center">
                    <p className="text-gray-400 mb-2">
                      {correlationData?.message || 'No correlation data available yet.'}
                    </p>
                    {correlationData?.note && (
                      <p className="text-sm text-gray-500">{correlationData.note}</p>
                    )}
                  </div>
                )}
              </div>
            )}
          </>
        )}
      </main>
    </div>
  )
}
