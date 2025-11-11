import { useAuth } from '../contexts/AuthContext'
import { useNavigate } from 'react-router-dom'

export default function Dashboard() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

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

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-8">
          <h2 className="text-3xl font-bold">Dashboard</h2>
          <p className="text-gray-400 mt-2">Monitor your AI trading performance</p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          <div className="bg-card border border-border rounded-lg p-6">
            <h3 className="text-sm text-gray-400 mb-2">Total P&L</h3>
            <p className="text-3xl font-bold text-primary">$0.00</p>
          </div>
          <div className="bg-card border border-border rounded-lg p-6">
            <h3 className="text-sm text-gray-400 mb-2">Active Traders</h3>
            <p className="text-3xl font-bold">0</p>
          </div>
          <div className="bg-card border border-border rounded-lg p-6">
            <h3 className="text-sm text-gray-400 mb-2">Win Rate</h3>
            <p className="text-3xl font-bold">0%</p>
          </div>
        </div>

        <div className="bg-card border border-border rounded-lg p-6">
          <h3 className="text-xl font-bold mb-4">Getting Started</h3>
          <p className="text-gray-400 mb-4">
            Welcome to LLM Trend! This is the Phase 1 setup of your AI-powered trading platform.
          </p>
          <ul className="list-disc list-inside space-y-2 text-gray-400">
            <li>Backend API is running with authentication</li>
            <li>Frontend is connected and configured</li>
            <li>Ready for Phase 2: Exchange Integration</li>
          </ul>
        </div>
      </main>
    </div>
  )
}
