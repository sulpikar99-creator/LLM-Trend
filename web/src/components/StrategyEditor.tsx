import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { motion } from 'framer-motion'
import { Brain, Save, RotateCcw, Sparkles, AlertCircle, CheckCircle } from 'lucide-react'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'

interface StrategyEditorProps {
  traderId: string
  currentStrategy: string
  traderName: string
}

const STRATEGY_TEMPLATES = [
  {
    name: 'Conservative Trend Following',
    description: 'Low risk, follows strong trends with strict risk management',
    prompt: `You are a conservative trend-following trading AI. Your strategy:

1. ONLY trade when strong trend is confirmed (ADX > 25, clear MA alignment)
2. NEVER risk more than 1% of capital per trade
3. Use wide stop losses (2-3% from entry) to avoid noise
4. Take profit at 1:2 or 1:3 risk-reward ratio
5. Avoid trading during high volatility or unclear market conditions
6. Wait for pullbacks to enter in the direction of the trend

Be patient and prioritize capital preservation over frequent trading.`,
  },
  {
    name: 'Aggressive Scalping',
    description: 'High frequency, quick profits on small movements',
    prompt: `You are an aggressive scalping trading AI. Your strategy:

1. Take trades on 5-15 minute timeframes
2. Target 0.3-0.5% profit per trade with tight stop losses
3. Use 10-25x leverage for amplified returns
4. Trade based on momentum indicators (RSI, MACD crossovers)
5. Exit quickly - don't hold positions longer than 30 minutes
6. Trade multiple times per day when opportunities arise

Be aggressive but disciplined with your stop losses.`,
  },
  {
    name: 'Breakout Hunter',
    description: 'Captures major moves from consolidation breakouts',
    prompt: `You are a breakout trading AI. Your strategy:

1. Identify consolidation patterns (triangles, flags, ranges)
2. Wait for volume-confirmed breakouts above/below key levels
3. Enter immediately on breakout with 2-3x leverage
4. Stop loss just below/above the breakout level (1-2%)
5. Target previous swing high/low or measured move
6. Only trade clean, obvious breakout setups

Be selective - quality over quantity for breakout trades.`,
  },
  {
    name: 'Mean Reversion',
    description: 'Profits from price returning to average after extremes',
    prompt: `You are a mean reversion trading AI. Your strategy:

1. Trade when price deviates significantly from moving averages (> 2 standard deviations)
2. Enter counter-trend positions expecting reversion to mean
3. Use Bollinger Bands and RSI to identify overbought/oversold conditions
4. Tight stop losses (1-1.5%) as reversals can continue
5. Take profit when price returns to moving average
6. Avoid mean reversion in strong trending markets

Be quick to exit if reversal doesn't happen.`,
  },
  {
    name: 'News/Event Driven',
    description: 'Reacts to major news and market events',
    prompt: `You are an event-driven trading AI. Your strategy:

1. Monitor for major economic announcements and crypto news
2. Analyze sentiment and likely market reaction
3. Take directional positions ahead of anticipated volatility
4. Use tighter stops due to unpredictable event outcomes
5. Scale out of positions as volatility subsides
6. Avoid trading during extreme uncertainty

Be decisive but respect that events can be unpredictable.`,
  },
]

export function StrategyEditor({ traderId, currentStrategy, traderName }: StrategyEditorProps) {
  const [strategy, setStrategy] = useState(currentStrategy || '')
  const [selectedTemplate, setSelectedTemplate] = useState<number | null>(null)
  const [hasChanges, setHasChanges] = useState(false)

  const queryClient = useQueryClient()

  const updateStrategyMutation = useMutation({
    mutationFn: async (newStrategy: string) => {
      const token = localStorage.getItem('auth-storage')
      let authToken = ''
      if (token) {
        try {
          const parsed = JSON.parse(token)
          authToken = parsed.state?.token || ''
        } catch (e) {
          console.error('Failed to parse auth token:', e)
        }
      }

      const response = await fetch(`/api/traders/${traderId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${authToken}`,
        },
        body: JSON.stringify({ strategy_prompt: newStrategy }),
      })

      if (!response.ok) {
        throw new Error('Failed to update strategy')
      }

      return response.json()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['trader', traderId] })
      setHasChanges(false)
    },
  })

  const handleTemplateSelect = (index: number) => {
    setSelectedTemplate(index)
    setStrategy(STRATEGY_TEMPLATES[index].prompt)
    setHasChanges(true)
  }

  const handleStrategyChange = (value: string) => {
    setStrategy(value)
    setHasChanges(true)
    setSelectedTemplate(null)
  }

  const handleSave = () => {
    updateStrategyMutation.mutate(strategy)
  }

  const handleReset = () => {
    setStrategy(currentStrategy || '')
    setHasChanges(false)
    setSelectedTemplate(null)
  }

  return (
    <div className="space-y-6">
      <Card>
        <div className="space-y-4">
          {/* Header */}
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <div className="p-2 bg-primary/10 rounded-lg">
                <Brain className="w-6 h-6 text-primary" />
              </div>
              <div>
                <h3 className="text-lg font-bold">AI Strategy Editor</h3>
                <p className="text-sm text-gray-400">{traderName}</p>
              </div>
            </div>
            {hasChanges && (
              <Badge variant="warning">
                <Sparkles className="w-3 h-3 mr-1" />
                Unsaved Changes
              </Badge>
            )}
          </div>

          {/* Instructions */}
          <div className="p-4 bg-primary/5 border border-primary/20 rounded-lg">
            <div className="flex items-start space-x-3">
              <AlertCircle className="w-5 h-5 text-primary mt-0.5" />
              <div className="text-sm">
                <p className="font-semibold mb-1">How to customize your AI strategy:</p>
                <ul className="space-y-1 text-gray-400">
                  <li>• Choose a template below or write your own strategy</li>
                  <li>• Be specific about entry/exit rules, risk management, and market conditions</li>
                  <li>• The AI will follow these instructions when analyzing the market</li>
                  <li>• More detailed strategies lead to better AI decision-making</li>
                </ul>
              </div>
            </div>
          </div>

          {/* Strategy Templates */}
          <div>
            <h4 className="font-semibold mb-3">Strategy Templates</h4>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              {STRATEGY_TEMPLATES.map((template, index) => (
                <motion.div
                  key={index}
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                >
                  <button
                    onClick={() => handleTemplateSelect(index)}
                    className={`w-full p-4 text-left rounded-lg border-2 transition-all ${
                      selectedTemplate === index
                        ? 'border-primary bg-primary/10'
                        : 'border-border hover:border-primary/50'
                    }`}
                  >
                    <div className="flex items-start justify-between mb-2">
                      <h5 className="font-semibold">{template.name}</h5>
                      {selectedTemplate === index && (
                        <CheckCircle className="w-5 h-5 text-primary" />
                      )}
                    </div>
                    <p className="text-sm text-gray-400">{template.description}</p>
                  </button>
                </motion.div>
              ))}
            </div>
          </div>

          {/* Strategy Prompt Editor */}
          <div>
            <label className="block font-semibold mb-2">
              Custom Strategy Prompt
              <span className="text-sm text-gray-400 font-normal ml-2">
                (AI will follow these exact instructions)
              </span>
            </label>
            <textarea
              value={strategy}
              onChange={(e) => handleStrategyChange(e.target.value)}
              className="w-full h-64 p-4 bg-background border border-border rounded-lg focus:border-primary focus:outline-none font-mono text-sm"
              placeholder="Describe your trading strategy in detail..."
            />
            <p className="text-xs text-gray-500 mt-2">
              {strategy.length} characters • Be as specific as possible for best results
            </p>
          </div>

          {/* Action Buttons */}
          <div className="flex space-x-3">
            <Button
              variant="primary"
              onClick={handleSave}
              isLoading={updateStrategyMutation.isPending}
              disabled={!hasChanges || !strategy.trim()}
              className="flex-1"
            >
              <Save className="w-4 h-4 mr-2" />
              Save Strategy
            </Button>
            <Button
              variant="secondary"
              onClick={handleReset}
              disabled={!hasChanges}
            >
              <RotateCcw className="w-4 h-4 mr-2" />
              Reset
            </Button>
          </div>

          {/* Success/Error Messages */}
          {updateStrategyMutation.isSuccess && (
            <div className="flex items-center space-x-2 p-3 bg-success/10 border border-success/20 rounded-lg">
              <CheckCircle className="w-4 h-4 text-success" />
              <span className="text-sm text-success">
                Strategy updated successfully! AI will use new strategy on next trading cycle.
              </span>
            </div>
          )}

          {updateStrategyMutation.isError && (
            <div className="flex items-center space-x-2 p-3 bg-danger/10 border border-danger/20 rounded-lg">
              <AlertCircle className="w-4 h-4 text-danger" />
              <span className="text-sm text-danger">
                Failed to update strategy: {updateStrategyMutation.error?.message}
              </span>
            </div>
          )}
        </div>
      </Card>

      {/* Tips Card */}
      <Card>
        <h4 className="font-semibold mb-3">Strategy Writing Tips</h4>
        <div className="space-y-2 text-sm text-gray-400">
          <div className="flex items-start space-x-2">
            <span className="text-primary">✓</span>
            <p><strong>Be Specific:</strong> "Enter when RSI below 30" is better than "buy when oversold"</p>
          </div>
          <div className="flex items-start space-x-2">
            <span className="text-primary">✓</span>
            <p><strong>Include Risk Rules:</strong> Always specify stop loss, take profit, and position sizing</p>
          </div>
          <div className="flex items-start space-x-2">
            <span className="text-primary">✓</span>
            <p><strong>Define Market Conditions:</strong> Explain when NOT to trade (high volatility, news events, etc.)</p>
          </div>
          <div className="flex items-start space-x-2">
            <span className="text-primary">✓</span>
            <p><strong>Use Numbers:</strong> "Stop loss at 2%" is clearer than "tight stop loss"</p>
          </div>
          <div className="flex items-start space-x-2">
            <span className="text-primary">✓</span>
            <p><strong>Prioritize Actions:</strong> Number your rules (1., 2., 3.) for clarity</p>
          </div>
        </div>
      </Card>
    </div>
  )
}
