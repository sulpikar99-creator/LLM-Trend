import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { motion } from 'framer-motion'
import { Plus, Play, Square, Trash2 } from 'lucide-react'
import toast from 'react-hot-toast'
import { api } from '@/lib/api'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Modal } from '@/components/ui/Modal'
import { Input } from '@/components/ui/Input'
import { Select } from '@/components/ui/Select'
import { Badge } from '@/components/ui/Badge'
import type { Trader, CreateTraderRequest } from '@/types'

const createTraderSchema = z.object({
  name: z.string().min(3, 'Name must be at least 3 characters'),
  exchange_type: z.string(),
  symbol: z.string().min(1, 'Symbol is required'),
  interval: z.string(),
  api_key: z.string().min(1, 'API Key is required'),
  api_secret: z.string().min(1, 'API Secret is required'),
  testnet: z.boolean(),
  strategy_prompt: z.string().min(10, 'Strategy prompt is required'),
})

type CreateTraderFormData = z.infer<typeof createTraderSchema>

export default function TradersPage() {
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)
  const queryClient = useQueryClient()

  const { data: tradersData, isLoading } = useQuery({
    queryKey: ['traders'],
    queryFn: () => api.listTraders(),
  })

  const createMutation = useMutation({
    mutationFn: (data: CreateTraderRequest) => api.createTrader(data),
    onSuccess: () => {
      toast.success('Trader created successfully')
      queryClient.invalidateQueries({ queryKey: ['traders'] })
      setIsCreateModalOpen(false)
      reset()
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to create trader')
    },
  })

  const startMutation = useMutation({
    mutationFn: (id: string) => api.startTrader(id),
    onSuccess: () => {
      toast.success('Trader started')
      queryClient.invalidateQueries({ queryKey: ['traders'] })
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to start trader')
    },
  })

  const stopMutation = useMutation({
    mutationFn: (id: string) => api.stopTrader(id),
    onSuccess: () => {
      toast.success('Trader stopped')
      queryClient.invalidateQueries({ queryKey: ['traders'] })
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to stop trader')
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.deleteTrader(id),
    onSuccess: () => {
      toast.success('Trader deleted')
      queryClient.invalidateQueries({ queryKey: ['traders'] })
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to delete trader')
    },
  })

  const {
    register,
    handleSubmit,
    watch,
    reset,
    formState: { errors },
  } = useForm<CreateTraderFormData>({
    resolver: zodResolver(createTraderSchema),
    defaultValues: {
      exchange_type: 'binance_futures',
      interval: '15m',
      testnet: true,
      strategy_prompt: `You are a cryptocurrency trading expert. Analyze market data and make informed decisions.

Trading Rules:
- Only trade when high confidence based on technical indicators
- Use RSI, MACD, Moving Averages for confirmation
- Minimum risk/reward ratio of 1:2
- Set stop loss 2% from entry
- Set take profit 4% from entry
- Never risk more than 2% per trade

Output JSON format:
{
  "action": "buy" | "sell" | "hold" | "close",
  "confidence": 0-100,
  "reason": "explanation",
  "entry_price": number,
  "stop_loss": number,
  "take_profit": number,
  "position_size_pct": 1-10
}`,
    },
  })

  const testnet = watch('testnet')

  const onSubmit = (data: CreateTraderFormData) => {
    createMutation.mutate(data)
  }

  const handleStart = (id: string) => {
    startMutation.mutate(id)
  }

  const handleStop = (id: string) => {
    stopMutation.mutate(id)
  }

  const handleDelete = (id: string) => {
    if (confirm('Are you sure you want to delete this trader?')) {
      deleteMutation.mutate(id)
    }
  }

  const traders = (tradersData?.data || []) as Trader[]

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-4xl font-bold mb-2">AI Traders</h1>
          <p className="text-gray-400">Manage your automated trading bots</p>
        </div>
        <Button
          onClick={() => setIsCreateModalOpen(true)}
          className="flex items-center space-x-2"
        >
          <Plus className="w-5 h-5" />
          <span>Create Trader</span>
        </Button>
      </div>

      {/* Traders Grid */}
      {isLoading ? (
        <div className="text-center py-12">
          <p className="text-gray-400">Loading traders...</p>
        </div>
      ) : traders.length === 0 ? (
        <Card>
          <div className="text-center py-12">
            <p className="text-gray-400 mb-4">No traders yet. Create your first AI trading bot!</p>
            <Button onClick={() => setIsCreateModalOpen(true)}>
              <Plus className="w-5 h-5 mr-2" />
              Create Trader
            </Button>
          </div>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {traders.map((trader, index) => (
            <motion.div
              key={trader.id}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.1 }}
            >
              <Card hover>
                <div className="flex justify-between items-start mb-4">
                  <div>
                    <h3 className="text-xl font-bold mb-1">{trader.name}</h3>
                    <p className="text-sm text-gray-400">{trader.exchange_type}</p>
                    <p className="text-sm text-gray-500 mt-1">
                      {trader.symbol} • {trader.interval}
                    </p>
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

                {trader.exchange_config?.testnet && (
                  <Badge variant="warning" size="sm" className="mb-4">
                    Testnet Mode
                  </Badge>
                )}

                <div className="flex space-x-2">
                  {trader.status === 'running' ? (
                    <Button
                      variant="warning"
                      size="sm"
                      onClick={() => handleStop(trader.id)}
                      disabled={stopMutation.isPending}
                      className="flex-1"
                    >
                      <Square className="w-4 h-4 mr-1" />
                      Stop
                    </Button>
                  ) : (
                    <Button
                      variant="success"
                      size="sm"
                      onClick={() => handleStart(trader.id)}
                      disabled={startMutation.isPending}
                      className="flex-1"
                    >
                      <Play className="w-4 h-4 mr-1" />
                      Start
                    </Button>
                  )}
                  <Button
                    variant="danger"
                    size="sm"
                    onClick={() => handleDelete(trader.id)}
                    disabled={deleteMutation.isPending}
                  >
                    <Trash2 className="w-4 h-4" />
                  </Button>
                </div>
              </Card>
            </motion.div>
          ))}
        </div>
      )}

      {/* Create Trader Modal */}
      <Modal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        title="Create New AI Trader"
        size="xl"
      >
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          {/* Basic Info */}
          <div className="space-y-4">
            <h3 className="text-lg font-semibold text-primary">Basic Information</h3>

            <Input
              {...register('name')}
              label="Trader Name"
              placeholder="e.g., Bitcoin Scalper"
              error={errors.name?.message}
            />

            <div className="grid grid-cols-2 gap-4">
              <Select
                {...register('exchange_type')}
                label="Exchange"
                options={[{ value: 'binance_futures', label: 'Binance Futures' }]}
                error={errors.exchange_type?.message}
              />

              <Input
                {...register('symbol')}
                label="Symbol"
                placeholder="BTCUSDT"
                error={errors.symbol?.message}
              />
            </div>

            <Select
              {...register('interval')}
              label="Timeframe"
              options={[
                { value: '1m', label: '1 Minute' },
                { value: '5m', label: '5 Minutes' },
                { value: '15m', label: '15 Minutes (Recommended)' },
                { value: '30m', label: '30 Minutes' },
                { value: '1h', label: '1 Hour' },
                { value: '4h', label: '4 Hours' },
                { value: '1d', label: '1 Day' },
              ]}
              error={errors.interval?.message}
            />
          </div>

          {/* Exchange Credentials */}
          <div className="space-y-4 border-t border-border pt-6">
            <h3 className="text-lg font-semibold text-primary">Exchange Credentials</h3>

            <Input
              {...register('api_key')}
              label="API Key"
              placeholder="Your Binance API Key"
              error={errors.api_key?.message}
            />

            <Input
              {...register('api_secret')}
              type="password"
              label="API Secret"
              placeholder="Your Binance API Secret"
              error={errors.api_secret?.message}
            />

            <div className="flex items-center space-x-2">
              <input
                {...register('testnet')}
                type="checkbox"
                className="w-4 h-4 rounded bg-background border-border"
              />
              <label className="text-sm font-medium">
                Use Testnet Mode (Recommended for testing)
              </label>
            </div>

            {testnet ? (
              <div className="bg-success/10 border border-success/50 rounded-lg p-4 text-sm text-success">
                <strong>✅ Testnet Mode:</strong> No real money will be used. Get free API keys from{' '}
                <a
                  href="https://testnet.binancefuture.com"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="underline font-semibold"
                >
                  testnet.binancefuture.com
                </a>
              </div>
            ) : (
              <div className="bg-danger/10 border border-danger rounded-lg p-4 text-sm text-danger">
                <strong>⚠️ REAL TRADING MODE:</strong> Real money will be traded! Make sure you understand the risks.
              </div>
            )}
          </div>

          {/* Strategy Prompt */}
          <div className="space-y-4 border-t border-border pt-6">
            <h3 className="text-lg font-semibold text-primary">AI Trading Strategy</h3>

            <div>
              <label className="label">Strategy Prompt</label>
              <textarea
                {...register('strategy_prompt')}
                rows={10}
                className="w-full px-4 py-2.5 bg-background border border-border rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-primary font-mono text-xs"
                placeholder="Enter AI trading strategy..."
              />
              {errors.strategy_prompt && (
                <p className="mt-1 text-sm text-danger">{errors.strategy_prompt.message}</p>
              )}
              <p className="mt-1 text-xs text-gray-500">
                This prompt guides the AI's trading decisions. Be specific about rules and risk management.
              </p>
            </div>
          </div>

          {/* Actions */}
          <div className="flex space-x-3 pt-4 border-t border-border">
            <Button
              type="button"
              variant="secondary"
              onClick={() => setIsCreateModalOpen(false)}
              className="flex-1"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              variant="primary"
              isLoading={createMutation.isPending}
              className="flex-1"
            >
              Create Trader
            </Button>
          </div>
        </form>
      </Modal>
    </div>
  )
}
