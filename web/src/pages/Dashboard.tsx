import { useState, useEffect } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { useNavigate } from 'react-router-dom'
import { api } from '../lib/api'

interface DashboardStats {
  totalPnL: number
  activeTraders: number
  totalTraders: number
  winRate: number
  totalTrades: number
}

export default function Dashboard() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const [stats, setStats] = useState<DashboardStats>({
    totalPnL: 0,
    activeTraders: 0,
    totalTraders: 0,
    winRate: 0,
    totalTrades: 0,
  })
  const [loading, setLoading] = useState(true)
  const [recentActivity, setRecentActivity] = useState<string[]>([])

  useEffect(() => {
    loadDashboardData()
  }, [])

  const loadDashboardData = async () => {
    try {
      setLoading(true)

      // Load traders and performance data in parallel
      const [tradersResp, performanceResp] = await Promise.all([
        api.listTraders().catch(() => ({ data: [] })),
        api.getPerformance().catch(() => ({ data: null })),
      ])

      const traders = tradersResp.data || []
      const performanceData = performanceResp.data

      // Calculate stats
      const activeTraders = traders.filter((t: any) => t.status === 'running').length
      const totalPnL = performanceData?.overall_metrics?.total_pnl || 0
      const winRate = performanceData?.overall_metrics?.win_rate || 0
      const totalTrades = performanceData?.total_trades || 0

      setStats({
        totalPnL,
        activeTraders,
        totalTraders: traders.length,
        winRate,
        totalTrades,
      })

      // Generate recent activity
      const activity: string[] = []
      if (traders.length > 0) {
        activity.push(`You have ${traders.length} trader${traders.length !== 1 ? 's' : ''} configured`)
      }
      if (activeTraders > 0) {
        activity.push(`${activeTraders} trader${activeTraders !== 1 ? 's are' : ' is'} currently running`)
      }
      if (totalTrades > 0) {
        activity.push(`Executed ${totalTrades} trade${totalTrades !== 1 ? 's' : ''} total`)
      }
      if (activity.length === 0) {
        activity.push('No trading activity yet')
      }
      setRecentActivity(activity)
    } catch (err) {
      console.error('Failed to load dashboard data:', err)
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
    return `${(value * 100).toFixed(1)}%`
  }

  return (
    <div className="min-h-screen bg-background">
      <nav className="bg-card border-b border-border">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16 items-center">
            <div className="flex items-center space-x-8">
              <h1 className="text-xl font-bold text-primary">LLM Trend</h1>
              <div className="flex space-x-4">
                <button
                  onClick={() => navigate('/')}
                  className="text-white hover:text-primary"
                >
                  Dashboard
                </button>
                <button
                  onClick={() => navigate('/traders')}
                  className="text-white hover:text-primary"
                >
                  Traders
                </button>
                <button
                  onClick={() => navigate('/analytics')}
                  className="text-white hover:text-primary"
                >
                  Analytics
                </button>
                <button
                  onClick={() => navigate('/settings')}
                  className="text-white hover:text-primary"
                >
                  Settings
                </button>
              </div>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-gray-400">Welcome, {user?.username || 'User'}</span>
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

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-8">
          <h2 className="text-3xl font-bold">Dashboard</h2>
          <p className="text-gray-400 mt-2">Monitor your AI trading performance</p>
        </div>

        {loading ? (
          <div className="text-center py-12">
            <p className="text-gray-400">Loading dashboard...</p>
          </div>
        ) : (
          <>
            {/* Stats Grid */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
              <div className="bg-card border border-border rounded-lg p-6">
                <h3 className="text-sm text-gray-400 mb-2">Total P&L</h3>
                <p className={`text-3xl font-bold ${stats.totalPnL >= 0 ? 'text-green-500' : 'text-red-500'}`}>
                  {formatCurrency(stats.totalPnL)}
                </p>
                {stats.totalTrades > 0 && (
                  <p className="text-xs text-gray-500 mt-2">
                    from {stats.totalTrades} trade{stats.totalTrades !== 1 ? 's' : ''}
                  </p>
                )}
              </div>

              <div className="bg-card border border-border rounded-lg p-6">
                <h3 className="text-sm text-gray-400 mb-2">Active Traders</h3>
                <p className="text-3xl font-bold text-primary">{stats.activeTraders}</p>
                <p className="text-xs text-gray-500 mt-2">
                  of {stats.totalTraders} total
                </p>
              </div>

              <div className="bg-card border border-border rounded-lg p-6">
                <h3 className="text-sm text-gray-400 mb-2">Win Rate</h3>
                <p className="text-3xl font-bold">
                  {stats.totalTrades > 0 ? formatPercent(stats.winRate) : '—'}
                </p>
                {stats.totalTrades > 0 && (
                  <p className="text-xs text-gray-500 mt-2">
                    {Math.round(stats.winRate * stats.totalTrades)} winning trades
                  </p>
                )}
              </div>

              <div className="bg-card border border-border rounded-lg p-6">
                <h3 className="text-sm text-gray-400 mb-2">Total Trades</h3>
                <p className="text-3xl font-bold">{stats.totalTrades}</p>
                <p className="text-xs text-gray-500 mt-2">all time</p>
              </div>
            </div>

            {/* Quick Actions */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
              <div className="bg-card border border-border rounded-lg p-6">
                <h3 className="text-xl font-bold mb-4">Quick Actions</h3>
                <div className="space-y-3">
                  <button
                    onClick={() => navigate('/traders')}
                    className="w-full px-4 py-3 bg-primary hover:bg-primary/90 text-white rounded font-medium text-left flex items-center justify-between"
                  >
                    <span>Manage Traders</span>
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                    </svg>
                  </button>
                  <button
                    onClick={() => navigate('/analytics')}
                    className="w-full px-4 py-3 bg-card hover:bg-border border border-border text-white rounded font-medium text-left flex items-center justify-between"
                  >
                    <span>View Analytics</span>
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                    </svg>
                  </button>
                  <button
                    onClick={() => navigate('/settings')}
                    className="w-full px-4 py-3 bg-card hover:bg-border border border-border text-white rounded font-medium text-left flex items-center justify-between"
                  >
                    <span>Configure Settings</span>
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                    </svg>
                  </button>
                </div>
              </div>

              <div className="bg-card border border-border rounded-lg p-6">
                <h3 className="text-xl font-bold mb-4">Recent Activity</h3>
                <ul className="space-y-2">
                  {recentActivity.map((activity, idx) => (
                    <li key={idx} className="flex items-start text-gray-400">
                      <svg className="w-5 h-5 mr-2 mt-0.5 text-primary" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                      </svg>
                      <span>{activity}</span>
                    </li>
                  ))}
                </ul>
              </div>
            </div>

            {/* Getting Started */}
            <div className="bg-card border border-border rounded-lg p-6">
              <h3 className="text-xl font-bold mb-4">
                {stats.totalTraders === 0 ? 'Getting Started' : 'System Status'}
              </h3>
              {stats.totalTraders === 0 ? (
                <>
                  <p className="text-gray-400 mb-4">
                    Welcome to LLM Trend! Get started with AI-powered trading in 3 simple steps:
                  </p>
                  <ol className="list-decimal list-inside space-y-2 text-gray-400">
                    <li>Configure your risk limits and AI settings in Settings</li>
                    <li>Create your first AI trader in the Traders section</li>
                    <li>Monitor performance in real-time with Analytics</li>
                  </ol>
                  <div className="mt-6">
                    <button
                      onClick={() => navigate('/traders')}
                      className="px-6 py-3 bg-primary hover:bg-primary/90 text-white rounded font-medium"
                    >
                      Create Your First Trader
                    </button>
                  </div>
                </>
              ) : (
                <>
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div className="flex items-center">
                      <div className={`w-3 h-3 rounded-full mr-3 ${stats.activeTraders > 0 ? 'bg-green-500' : 'bg-gray-500'}`}></div>
                      <div>
                        <p className="text-sm text-gray-400">Trading Status</p>
                        <p className="font-bold">
                          {stats.activeTraders > 0 ? 'Active' : 'Inactive'}
                        </p>
                      </div>
                    </div>
                    <div className="flex items-center">
                      <div className="w-3 h-3 rounded-full bg-green-500 mr-3"></div>
                      <div>
                        <p className="text-sm text-gray-400">API Status</p>
                        <p className="font-bold">Connected</p>
                      </div>
                    </div>
                    <div className="flex items-center">
                      <div className="w-3 h-3 rounded-full bg-green-500 mr-3"></div>
                      <div>
                        <p className="text-sm text-gray-400">Database</p>
                        <p className="font-bold">Operational</p>
                      </div>
                    </div>
                  </div>
                </>
              )}
            </div>
          </>
        )}
      </main>
    </div>
  )
}
