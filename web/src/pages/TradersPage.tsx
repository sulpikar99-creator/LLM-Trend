import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { api } from '../lib/api'

interface Trader {
  id: string
  name: string
  exchange_type: string
  status: string
  created_at: string
}

export default function TradersPage() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const [traders, setTraders] = useState<Trader[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showCreateModal, setShowCreateModal] = useState(false)

  // Form state
  const [traderName, setTraderName] = useState('')
  const [exchangeType, setExchangeType] = useState('binance_futures')
  const [apiKey, setApiKey] = useState('')
  const [apiSecret, setApiSecret] = useState('')
  const [useTestnet, setUseTestnet] = useState(true)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    loadTraders()
  }, [])

  const loadTraders = async () => {
    try {
      setLoading(true)
      const response = await api.listTraders()
      setTraders(response.data || [])
    } catch (err: any) {
      setError(err.message || 'Failed to load traders')
    } finally {
      setLoading(false)
    }
  }

  const handleCreateTrader = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    setError('')

    try {
      await api.createTrader({
        name: traderName,
        exchange_type: exchangeType,
        exchange_config: {
          api_key: apiKey,
          api_secret: apiSecret,
          testnet: useTestnet
        }
      })

      // Clear error on success
      setError('')
      setShowCreateModal(false)
      setTraderName('')
      setApiKey('')
      setApiSecret('')
      setUseTestnet(true)
      loadTraders()
    } catch (err: any) {
      setError(err.message || 'Failed to create trader')
    } finally {
      setSubmitting(false)
    }
  }

  const handleStartTrader = async (id: string) => {
    try {
      await api.startTrader(id)
      setError('') // Clear error on success
      loadTraders()
    } catch (err: any) {
      setError(err.message || 'Failed to start trader')
    }
  }

  const handleStopTrader = async (id: string) => {
    try {
      await api.stopTrader(id)
      setError('') // Clear error on success
      loadTraders()
    } catch (err: any) {
      setError(err.message || 'Failed to stop trader')
    }
  }

  const handleDeleteTrader = async (id: string) => {
    if (!confirm('Are you sure you want to delete this trader?')) return

    try {
      await api.deleteTrader(id)
      setError('') // Clear error on success
      loadTraders()
    } catch (err: any) {
      setError(err.message || 'Failed to delete trader')
    }
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
                  className="text-primary font-semibold"
                >
                  Traders
                </button>
                <button
                  onClick={() => navigate('/analytics')}
                  className="text-gray-400 hover:text-white"
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

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-8 flex justify-between items-center">
          <div>
            <h2 className="text-3xl font-bold">Traders</h2>
            <p className="text-gray-400 mt-2">Manage your AI trading bots</p>
          </div>
          <button
            onClick={() => setShowCreateModal(true)}
            className="px-6 py-3 bg-primary hover:bg-primary/90 text-white rounded font-medium"
          >
            + Create Trader
          </button>
        </div>

        {error && (
          <div className="mb-6 bg-danger/10 border border-danger text-danger px-4 py-3 rounded">
            {error}
          </div>
        )}

        {loading ? (
          <div className="text-center py-12">
            <p className="text-gray-400">Loading traders...</p>
          </div>
        ) : traders.length === 0 ? (
          <div className="bg-card border border-border rounded-lg p-12 text-center">
            <p className="text-gray-400 mb-4">No traders yet. Create your first AI trading bot!</p>
            <button
              onClick={() => setShowCreateModal(true)}
              className="px-6 py-3 bg-primary hover:bg-primary/90 text-white rounded font-medium"
            >
              Create Trader
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {traders.map((trader) => (
              <div key={trader.id} className="bg-card border border-border rounded-lg p-6">
                <div className="flex justify-between items-start mb-4">
                  <div>
                    <h3 className="text-xl font-bold">{trader.name}</h3>
                    <p className="text-sm text-gray-400">{trader.exchange_type}</p>
                  </div>
                  <span
                    className={`px-3 py-1 rounded text-sm ${
                      trader.status === 'running'
                        ? 'bg-green-500/20 text-green-500'
                        : 'bg-gray-500/20 text-gray-400'
                    }`}
                  >
                    {trader.status}
                  </span>
                </div>

                <div className="flex space-x-2 mt-4">
                  {trader.status === 'running' ? (
                    <button
                      onClick={() => handleStopTrader(trader.id)}
                      className="flex-1 px-4 py-2 bg-yellow-500 hover:bg-yellow-600 text-white rounded"
                    >
                      Stop
                    </button>
                  ) : (
                    <button
                      onClick={() => handleStartTrader(trader.id)}
                      className="flex-1 px-4 py-2 bg-green-500 hover:bg-green-600 text-white rounded"
                    >
                      Start
                    </button>
                  )}
                  <button
                    onClick={() => handleDeleteTrader(trader.id)}
                    className="px-4 py-2 bg-danger hover:bg-danger/90 text-white rounded"
                  >
                    Delete
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Create Trader Modal */}
        {showCreateModal && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
            <div className="bg-card border border-border rounded-lg p-8 max-w-md w-full">
              <h3 className="text-2xl font-bold mb-6">Create New Trader</h3>

              <form onSubmit={handleCreateTrader} className="space-y-4">
                <div>
                  <label className="block text-sm font-medium mb-2">Trader Name</label>
                  <input
                    type="text"
                    required
                    value={traderName}
                    onChange={(e) => setTraderName(e.target.value)}
                    className="w-full px-3 py-2 border border-border rounded bg-background text-white focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder="My AI Trader"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium mb-2">Exchange</label>
                  <select
                    value={exchangeType}
                    onChange={(e) => setExchangeType(e.target.value)}
                    className="w-full px-3 py-2 border border-border rounded bg-background text-white focus:outline-none focus:ring-2 focus:ring-primary"
                  >
                    <option value="binance_futures">Binance Futures</option>
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium mb-2">API Key</label>
                  <input
                    type="text"
                    required
                    value={apiKey}
                    onChange={(e) => setApiKey(e.target.value)}
                    className="w-full px-3 py-2 border border-border rounded bg-background text-white focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder="Your Binance API Key"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium mb-2">API Secret</label>
                  <input
                    type="password"
                    required
                    value={apiSecret}
                    onChange={(e) => setApiSecret(e.target.value)}
                    className="w-full px-3 py-2 border border-border rounded bg-background text-white focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder="Your Binance API Secret"
                  />
                </div>

                <div className="border border-border rounded p-4">
                  <label className="flex items-center space-x-2 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={useTestnet}
                      onChange={(e) => setUseTestnet(e.target.checked)}
                      className="w-4 h-4 text-primary bg-background border-border rounded focus:ring-2 focus:ring-primary"
                    />
                    <span className="text-sm font-medium">
                      Use Testnet Mode (Recommended for testing)
                    </span>
                  </label>
                  <p className="text-xs text-gray-400 mt-2 ml-6">
                    Testnet allows you to practice trading without risking real money
                  </p>
                </div>

                {useTestnet ? (
                  <div className="bg-yellow-500/10 border border-yellow-500/50 text-yellow-500 px-4 py-3 rounded text-sm">
                    <strong>🔶 Testnet Mode:</strong> Get your testnet API keys from{' '}
                    <a
                      href="https://testnet.binancefuture.com"
                      target="_blank"
                      rel="noopener noreferrer"
                      className="underline font-semibold"
                    >
                      testnet.binancefuture.com
                    </a>
                    <br />
                    <span className="text-xs">No real money will be used. All trades are simulated.</span>
                  </div>
                ) : (
                  <div className="bg-red-500/10 border border-red-500 text-red-500 px-4 py-3 rounded text-sm">
                    <strong>⚠️ REAL TRADING MODE:</strong> You are using your REAL Binance account.
                    <br />
                    <strong>Real money will be traded!</strong> Make sure you understand the risks.
                    <br />
                    <span className="text-xs">Get your API keys from{' '}
                      <a
                        href="https://www.binance.com/en/my/settings/api-management"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="underline font-semibold"
                      >
                        Binance API Management
                      </a>
                    </span>
                  </div>
                )}

                <div className="flex space-x-3 mt-6">
                  <button
                    type="button"
                    onClick={() => setShowCreateModal(false)}
                    className="flex-1 px-4 py-2 border border-border rounded text-gray-400 hover:text-white"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    disabled={submitting}
                    className="flex-1 px-4 py-2 bg-primary hover:bg-primary/90 text-white rounded disabled:opacity-50"
                  >
                    {submitting ? 'Creating...' : 'Create'}
                  </button>
                </div>
              </form>
            </div>
          </div>
        )}
      </main>
    </div>
  )
}
