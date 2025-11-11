import { useState, useEffect, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { api } from '../lib/api'

interface AIModel {
  id: string
  name: string
  provider: string
  base_url: string
  max_tokens: number
  temperature: number
  api_key_encrypted?: string
}

interface RiskLimits {
  max_drawdown_percent: number
  max_position_size_usd: number
  max_leverage: number
  daily_loss_limit_usd: number
}

interface Config {
  ai_models: AIModel[]
  risk_limits: RiskLimits
  default_strategy_prompt: string
}

export default function SettingsPage() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const [config, setConfig] = useState<Config | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const [submitting, setSubmitting] = useState(false)

  // Risk limits form state
  const [maxDrawdown, setMaxDrawdown] = useState(20)
  const [maxPositionSize, setMaxPositionSize] = useState(10000)
  const [maxLeverage, setMaxLeverage] = useState(10)
  const [dailyLossLimit, setDailyLossLimit] = useState(1000)

  // AI model form state
  const [aiProvider, setAiProvider] = useState('deepseek')
  const [aiModel, setAiModel] = useState('deepseek-chat')
  const [aiApiKey, setAiApiKey] = useState('')
  const [aiBaseUrl, setAiBaseUrl] = useState('https://api.deepseek.com')
  const [aiMaxTokens, setAiMaxTokens] = useState(4000)
  const [aiTemperature, setAiTemperature] = useState(0.7)

  // Strategy prompt
  const [strategyPrompt, setStrategyPrompt] = useState('')

  // Timeout refs for cleanup
  const successTimeoutRef = useRef<NodeJS.Timeout | null>(null)

  useEffect(() => {
    loadConfig()

    // Cleanup timeout on unmount
    return () => {
      if (successTimeoutRef.current) {
        clearTimeout(successTimeoutRef.current)
      }
    }
  }, [])

  const loadConfig = async () => {
    try {
      setLoading(true)
      setError('')
      const response = await api.getConfig()
      const configData = response.data as Config

      setConfig(configData)

      // Set risk limits
      if (configData.risk_limits) {
        setMaxDrawdown(configData.risk_limits.max_drawdown_percent)
        setMaxPositionSize(configData.risk_limits.max_position_size_usd)
        setMaxLeverage(configData.risk_limits.max_leverage)
        setDailyLossLimit(configData.risk_limits.daily_loss_limit_usd)
      }

      // Set AI model config
      if (configData.ai_models && configData.ai_models.length > 0) {
        const model = configData.ai_models[0]
        setAiProvider(model.provider)
        setAiModel(model.id)
        setAiApiKey(model.api_key_encrypted || '')
        setAiBaseUrl(model.base_url)
        setAiMaxTokens(model.max_tokens)
        setAiTemperature(model.temperature)
      }

      // Set strategy prompt
      if (configData.default_strategy_prompt) {
        setStrategyPrompt(configData.default_strategy_prompt)
      }
    } catch (err: any) {
      setError(err.message || 'Failed to load configuration')
    } finally {
      setLoading(false)
    }
  }

  const handleSaveRiskLimits = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    setError('')
    setSuccess('')

    try {
      // Validate config exists before spreading
      if (!config) {
        throw new Error('Configuration not loaded')
      }

      const updatedConfig = {
        ...config,
        risk_limits: {
          max_drawdown_percent: maxDrawdown,
          max_position_size_usd: maxPositionSize,
          max_leverage: maxLeverage,
          daily_loss_limit_usd: dailyLossLimit,
        },
      }

      await api.updateConfig(updatedConfig)
      setSuccess('Risk limits updated successfully')

      // Clear previous timeout and set new one
      if (successTimeoutRef.current) {
        clearTimeout(successTimeoutRef.current)
      }
      successTimeoutRef.current = setTimeout(() => setSuccess(''), 3000)

      loadConfig()
    } catch (err: any) {
      setError(err.message || 'Failed to update risk limits')
    } finally {
      setSubmitting(false)
    }
  }

  const handleSaveAIConfig = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    setError('')
    setSuccess('')

    try {
      // Validate config exists before spreading
      if (!config) {
        throw new Error('Configuration not loaded')
      }

      // Safely format model name
      const modelName = aiModel ? aiModel.replace(/-/g, ' ').replace(/\b\w/g, (l) => l.toUpperCase()) : aiModel

      const updatedConfig = {
        ...config,
        ai_models: [
          {
            id: aiModel,
            name: modelName,
            provider: aiProvider,
            api_key_encrypted: aiApiKey,
            base_url: aiBaseUrl,
            max_tokens: aiMaxTokens,
            temperature: aiTemperature,
          },
        ],
      }

      await api.updateConfig(updatedConfig)
      setSuccess('AI configuration updated successfully')

      // Clear previous timeout and set new one
      if (successTimeoutRef.current) {
        clearTimeout(successTimeoutRef.current)
      }
      successTimeoutRef.current = setTimeout(() => setSuccess(''), 3000)

      loadConfig()
    } catch (err: any) {
      setError(err.message || 'Failed to update AI configuration')
    } finally {
      setSubmitting(false)
    }
  }

  const handleSaveStrategyPrompt = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    setError('')
    setSuccess('')

    try {
      // Validate config exists before spreading
      if (!config) {
        throw new Error('Configuration not loaded')
      }

      const updatedConfig = {
        ...config,
        default_strategy_prompt: strategyPrompt,
      }

      await api.updateConfig(updatedConfig)
      setSuccess('Strategy prompt updated successfully')

      // Clear previous timeout and set new one
      if (successTimeoutRef.current) {
        clearTimeout(successTimeoutRef.current)
      }
      successTimeoutRef.current = setTimeout(() => setSuccess(''), 3000)

      loadConfig()
    } catch (err: any) {
      setError(err.message || 'Failed to update strategy prompt')
    } finally {
      setSubmitting(false)
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
                  className="text-gray-400 hover:text-white"
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
                  className="text-primary font-semibold"
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
        <div className="mb-8">
          <h2 className="text-3xl font-bold">Settings</h2>
          <p className="text-gray-400 mt-2">Configure your trading platform</p>
        </div>

        {error && (
          <div className="mb-6 bg-danger/10 border border-danger text-danger px-4 py-3 rounded">
            {error}
          </div>
        )}

        {success && (
          <div className="mb-6 bg-green-500/10 border border-green-500 text-green-500 px-4 py-3 rounded">
            {success}
          </div>
        )}

        {loading ? (
          <div className="text-center py-12">
            <p className="text-gray-400">Loading configuration...</p>
          </div>
        ) : (
          <div className="space-y-6">
            {/* Risk Limits Section */}
            <div className="bg-card border border-border rounded-lg p-6">
              <h3 className="text-xl font-bold mb-4">Risk Limits</h3>
              <p className="text-gray-400 mb-6">
                Configure risk management parameters for all traders
              </p>

              <form onSubmit={handleSaveRiskLimits} className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium mb-2">
                      Max Drawdown (%)
                    </label>
                    <input
                      type="number"
                      step="0.1"
                      min="0"
                      max="100"
                      required
                      value={maxDrawdown}
                      onChange={(e) => setMaxDrawdown(parseFloat(e.target.value))}
                      className="w-full px-3 py-2 border border-border rounded bg-background text-white focus:outline-none focus:ring-2 focus:ring-primary"
                    />
                    <p className="text-xs text-gray-500 mt-1">
                      Maximum allowed portfolio drawdown
                    </p>
                  </div>

                  <div>
                    <label className="block text-sm font-medium mb-2">
                      Max Position Size (USD)
                    </label>
                    <input
                      type="number"
                      step="100"
                      min="0"
                      required
                      value={maxPositionSize}
                      onChange={(e) => setMaxPositionSize(parseFloat(e.target.value))}
                      className="w-full px-3 py-2 border border-border rounded bg-background text-white focus:outline-none focus:ring-2 focus:ring-primary"
                    />
                    <p className="text-xs text-gray-500 mt-1">
                      Maximum position size per trade
                    </p>
                  </div>

                  <div>
                    <label className="block text-sm font-medium mb-2">
                      Max Leverage
                    </label>
                    <input
                      type="number"
                      step="1"
                      min="1"
                      max="125"
                      required
                      value={maxLeverage}
                      onChange={(e) => setMaxLeverage(parseFloat(e.target.value))}
                      className="w-full px-3 py-2 border border-border rounded bg-background text-white focus:outline-none focus:ring-2 focus:ring-primary"
                    />
                    <p className="text-xs text-gray-500 mt-1">
                      Maximum allowed leverage
                    </p>
                  </div>

                  <div>
                    <label className="block text-sm font-medium mb-2">
                      Daily Loss Limit (USD)
                    </label>
                    <input
                      type="number"
                      step="100"
                      min="0"
                      required
                      value={dailyLossLimit}
                      onChange={(e) => setDailyLossLimit(parseFloat(e.target.value))}
                      className="w-full px-3 py-2 border border-border rounded bg-background text-white focus:outline-none focus:ring-2 focus:ring-primary"
                    />
                    <p className="text-xs text-gray-500 mt-1">
                      Maximum daily loss before stopping
                    </p>
                  </div>
                </div>

                <div className="flex justify-end">
                  <button
                    type="submit"
                    disabled={submitting}
                    className="px-6 py-2 bg-primary hover:bg-primary/90 text-white rounded font-medium disabled:opacity-50"
                  >
                    {submitting ? 'Saving...' : 'Save Risk Limits'}
                  </button>
                </div>
              </form>
            </div>

            {/* AI Configuration Section */}
            <div className="bg-card border border-border rounded-lg p-6">
              <h3 className="text-xl font-bold mb-4">AI Configuration</h3>
              <p className="text-gray-400 mb-6">
                Configure the AI model used for trading decisions
              </p>

              {/* Warning if API key is not set */}
              {!aiApiKey && (
                <div className="mb-4 bg-red-500/10 border border-red-500 text-red-500 px-4 py-3 rounded">
                  <strong>⚠️ Required:</strong> AI API Key must be configured before starting traders.
                  <br />
                  Get your API key from{' '}
                  <a
                    href={aiProvider === 'deepseek' ? 'https://platform.deepseek.com' : 'https://platform.openai.com'}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="underline font-semibold"
                  >
                    {aiProvider === 'deepseek' ? 'DeepSeek Platform' : 'OpenAI Platform'}
                  </a>
                </div>
              )}

              <form onSubmit={handleSaveAIConfig} className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium mb-2">
                      AI Provider
                    </label>
                    <select
                      value={aiProvider}
                      onChange={(e) => setAiProvider(e.target.value)}
                      className="w-full px-3 py-2 border border-border rounded bg-background text-white focus:outline-none focus:ring-2 focus:ring-primary"
                    >
                      <option value="deepseek">DeepSeek</option>
                      <option value="openai">OpenAI</option>
                      <option value="anthropic">Anthropic</option>
                      <option value="custom">Custom</option>
                    </select>
                  </div>

                  <div>
                    <label className="block text-sm font-medium mb-2">
                      Model ID
                    </label>
                    <input
                      type="text"
                      required
                      value={aiModel}
                      onChange={(e) => setAiModel(e.target.value)}
                      className="w-full px-3 py-2 border border-border rounded bg-background text-white focus:outline-none focus:ring-2 focus:ring-primary"
                      placeholder="deepseek-chat"
                    />
                  </div>

                  <div className="md:col-span-2">
                    <label className="block text-sm font-medium mb-2">
                      API Key <span className="text-red-500">*</span>
                    </label>
                    <input
                      type="password"
                      required
                      value={aiApiKey}
                      onChange={(e) => setAiApiKey(e.target.value)}
                      className="w-full px-3 py-2 border border-border rounded bg-background text-white focus:outline-none focus:ring-2 focus:ring-primary"
                      placeholder="sk-..."
                    />
                    <p className="text-xs text-gray-500 mt-1">
                      Your {aiProvider} API key. This will be encrypted and stored securely.
                    </p>
                  </div>

                  <div>
                    <label className="block text-sm font-medium mb-2">
                      Base URL
                    </label>
                    <input
                      type="url"
                      required
                      value={aiBaseUrl}
                      onChange={(e) => setAiBaseUrl(e.target.value)}
                      className="w-full px-3 py-2 border border-border rounded bg-background text-white focus:outline-none focus:ring-2 focus:ring-primary"
                      placeholder="https://api.deepseek.com"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium mb-2">
                      Max Tokens
                    </label>
                    <input
                      type="number"
                      step="100"
                      min="100"
                      required
                      value={aiMaxTokens}
                      onChange={(e) => setAiMaxTokens(parseInt(e.target.value))}
                      className="w-full px-3 py-2 border border-border rounded bg-background text-white focus:outline-none focus:ring-2 focus:ring-primary"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium mb-2">
                      Temperature
                    </label>
                    <input
                      type="number"
                      step="0.1"
                      min="0"
                      max="2"
                      required
                      value={aiTemperature}
                      onChange={(e) => setAiTemperature(parseFloat(e.target.value))}
                      className="w-full px-3 py-2 border border-border rounded bg-background text-white focus:outline-none focus:ring-2 focus:ring-primary"
                    />
                    <p className="text-xs text-gray-500 mt-1">
                      Higher values = more creative, lower = more focused
                    </p>
                  </div>
                </div>

                <div className="flex justify-end">
                  <button
                    type="submit"
                    disabled={submitting}
                    className="px-6 py-2 bg-primary hover:bg-primary/90 text-white rounded font-medium disabled:opacity-50"
                  >
                    {submitting ? 'Saving...' : 'Save AI Configuration'}
                  </button>
                </div>
              </form>
            </div>

            {/* Strategy Prompt Section */}
            <div className="bg-card border border-border rounded-lg p-6">
              <h3 className="text-xl font-bold mb-4">Strategy Prompt</h3>
              <p className="text-gray-400 mb-6">
                Customize the system prompt used by the AI for trading decisions
              </p>

              <form onSubmit={handleSaveStrategyPrompt} className="space-y-4">
                <div>
                  <label className="block text-sm font-medium mb-2">
                    Default Strategy Prompt
                  </label>
                  <textarea
                    required
                    value={strategyPrompt}
                    onChange={(e) => setStrategyPrompt(e.target.value)}
                    rows={12}
                    className="w-full px-3 py-2 border border-border rounded bg-background text-white focus:outline-none focus:ring-2 focus:ring-primary font-mono text-sm"
                    placeholder="Enter your strategy prompt..."
                  />
                  <p className="text-xs text-gray-500 mt-1">
                    This prompt guides the AI's trading decisions. Be specific about your strategy.
                  </p>
                </div>

                <div className="flex justify-end">
                  <button
                    type="submit"
                    disabled={submitting}
                    className="px-6 py-2 bg-primary hover:bg-primary/90 text-white rounded font-medium disabled:opacity-50"
                  >
                    {submitting ? 'Saving...' : 'Save Strategy Prompt'}
                  </button>
                </div>
              </form>
            </div>

            {/* Info Section */}
            <div className="bg-yellow-500/10 border border-yellow-500/50 rounded-lg p-6">
              <h3 className="text-lg font-bold text-yellow-500 mb-2">
                Important Notes
              </h3>
              <ul className="list-disc list-inside space-y-2 text-yellow-500/90 text-sm">
                <li>Risk limits apply to all traders on your account</li>
                <li>AI configuration changes will affect new trading decisions</li>
                <li>Strategy prompt changes require trader restart to take effect</li>
                <li>Always test configuration changes with small positions first</li>
              </ul>
            </div>
          </div>
        )}
      </main>
    </div>
  )
}
