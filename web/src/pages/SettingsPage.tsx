import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Save, Bot, Shield, FileText } from 'lucide-react'
import toast from 'react-hot-toast'
import { api } from '@/lib/api'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Select } from '@/components/ui/Select'
import type { Config } from '@/types'

const aiConfigSchema = z.object({
  provider: z.string(),
  model_id: z.string().min(1, 'Model ID is required'),
  api_key: z.string().min(1, 'API Key is required'),
  base_url: z.string().url('Invalid URL'),
  max_tokens: z.number().min(100).max(100000),
  temperature: z.number().min(0).max(2),
})

const riskLimitsSchema = z.object({
  max_drawdown_percent: z.number().min(1).max(100),
  max_position_size_usd: z.number().min(1),
  max_leverage: z.number().min(1).max(125),
  daily_loss_limit_usd: z.number().min(1),
})

type AIConfigFormData = z.infer<typeof aiConfigSchema>
type RiskLimitsFormData = z.infer<typeof riskLimitsSchema>

export default function SettingsPage() {
  const [activeTab, setActiveTab] = useState<'ai' | 'risk' | 'strategy'>('ai')
  const queryClient = useQueryClient()

  const { data: configData, isLoading } = useQuery({
    queryKey: ['config'],
    queryFn: () => api.getConfig(),
  })

  const updateMutation = useMutation({
    mutationFn: (data: Partial<Config>) => api.updateConfig(data),
    onSuccess: () => {
      toast.success('Settings saved successfully')
      queryClient.invalidateQueries({ queryKey: ['config'] })
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to save settings')
    },
  })

  const config = configData?.data as Config

  const {
    register: registerAI,
    handleSubmit: handleAISubmit,
    formState: { errors: aiErrors },
  } = useForm<AIConfigFormData>({
    resolver: zodResolver(aiConfigSchema),
    values: config?.ai_models?.[0]
      ? {
          provider: config.ai_models[0].provider,
          model_id: config.ai_models[0].model_id,
          api_key: config.ai_models[0].api_key || '',
          base_url: config.ai_models[0].base_url,
          max_tokens: config.ai_models[0].max_tokens,
          temperature: config.ai_models[0].temperature,
        }
      : undefined,
  })

  const {
    register: registerRisk,
    handleSubmit: handleRiskSubmit,
    formState: { errors: riskErrors },
  } = useForm<RiskLimitsFormData>({
    resolver: zodResolver(riskLimitsSchema),
    values: config?.risk_limits
      ? {
          max_drawdown_percent: config.risk_limits.max_drawdown_percent,
          max_position_size_usd: config.risk_limits.max_position_size_usd,
          max_leverage: config.risk_limits.max_leverage,
          daily_loss_limit_usd: config.risk_limits.daily_loss_limit_usd,
        }
      : undefined,
  })

  const [strategyPrompt, setStrategyPrompt] = useState(config?.default_strategy_prompt || '')

  const onAISubmit = (data: AIConfigFormData) => {
    updateMutation.mutate({
      ai_models: [
        {
          provider: data.provider,
          model_id: data.model_id,
          api_key: data.api_key,
          base_url: data.base_url,
          max_tokens: data.max_tokens,
          temperature: data.temperature,
        },
      ],
    })
  }

  const onRiskSubmit = (data: RiskLimitsFormData) => {
    updateMutation.mutate({
      risk_limits: data,
    })
  }

  const onStrategySubmit = (e: React.FormEvent) => {
    e.preventDefault()
    updateMutation.mutate({
      default_strategy_prompt: strategyPrompt,
    })
  }

  if (isLoading) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-400">Loading settings...</p>
      </div>
    )
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-4xl font-bold mb-2">Settings</h1>
        <p className="text-gray-400">Configure your AI trading platform</p>
      </div>

      {/* Tabs */}
      <div className="flex space-x-2 border-b border-border">
        {[
          { id: 'ai', label: 'AI Configuration', icon: Bot },
          { id: 'risk', label: 'Risk Limits', icon: Shield },
          { id: 'strategy', label: 'Strategy Prompt', icon: FileText },
        ].map((tab) => {
          const Icon = tab.icon
          return (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id as any)}
              className={`flex items-center space-x-2 px-6 py-3 font-medium border-b-2 transition-colors ${
                activeTab === tab.id
                  ? 'border-primary text-primary'
                  : 'border-transparent text-gray-400 hover:text-white'
              }`}
            >
              <Icon className="w-5 h-5" />
              <span>{tab.label}</span>
            </button>
          )
        })}
      </div>

      {/* AI Configuration Tab */}
      {activeTab === 'ai' && (
        <Card>
          <form onSubmit={handleAISubmit(onAISubmit)} className="space-y-6">
            <div>
              <h3 className="text-xl font-bold mb-4">AI Model Configuration</h3>
              <p className="text-gray-400 text-sm mb-6">
                Configure the AI model used for trading decisions. API keys are encrypted and stored securely.
              </p>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <Select
                {...registerAI('provider')}
                label="AI Provider"
                options={[
                  { value: 'deepseek', label: 'DeepSeek' },
                  { value: 'openai', label: 'OpenAI' },
                  { value: 'claude', label: 'Anthropic Claude' },
                  { value: 'qwen', label: 'Qwen' },
                ]}
                error={aiErrors.provider?.message}
              />

              <Input
                {...registerAI('model_id')}
                label="Model ID"
                placeholder="e.g., deepseek-chat, gpt-4"
                error={aiErrors.model_id?.message}
              />
            </div>

            <Input
              {...registerAI('api_key')}
              type="password"
              label="API Key"
              placeholder="sk-..."
              error={aiErrors.api_key?.message}
            />

            <Input
              {...registerAI('base_url')}
              label="Base URL"
              placeholder="https://api.deepseek.com"
              error={aiErrors.base_url?.message}
            />

            <div className="grid grid-cols-2 gap-4">
              <Input
                {...registerAI('max_tokens', { valueAsNumber: true })}
                type="number"
                label="Max Tokens"
                placeholder="4000"
                error={aiErrors.max_tokens?.message}
              />

              <Input
                {...registerAI('temperature', { valueAsNumber: true })}
                type="number"
                step="0.1"
                label="Temperature"
                placeholder="0.7"
                error={aiErrors.temperature?.message}
              />
            </div>

            <div className="bg-primary/10 border border-primary/50 rounded-lg p-4 text-sm">
              <strong>💡 Tip:</strong> Lower temperature (0.1-0.3) for more conservative trading, higher (0.7-1.0) for more creative strategies.
            </div>

            <Button
              type="submit"
              variant="primary"
              isLoading={updateMutation.isPending}
              className="w-full"
            >
              <Save className="w-4 h-4 mr-2" />
              Save AI Configuration
            </Button>
          </form>
        </Card>
      )}

      {/* Risk Limits Tab */}
      {activeTab === 'risk' && (
        <Card>
          <form onSubmit={handleRiskSubmit(onRiskSubmit)} className="space-y-6">
            <div>
              <h3 className="text-xl font-bold mb-4">Risk Management Limits</h3>
              <p className="text-gray-400 text-sm mb-6">
                Set maximum risk limits for all traders. These limits apply globally to protect your capital.
              </p>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <Input
                {...registerRisk('max_drawdown_percent', { valueAsNumber: true })}
                type="number"
                step="0.1"
                label="Max Drawdown (%)"
                placeholder="20"
                error={riskErrors.max_drawdown_percent?.message}
              />

              <Input
                {...registerRisk('max_position_size_usd', { valueAsNumber: true })}
                type="number"
                label="Max Position Size (USD)"
                placeholder="10000"
                error={riskErrors.max_position_size_usd?.message}
              />

              <Input
                {...registerRisk('max_leverage', { valueAsNumber: true })}
                type="number"
                label="Max Leverage"
                placeholder="10"
                error={riskErrors.max_leverage?.message}
              />

              <Input
                {...registerRisk('daily_loss_limit_usd', { valueAsNumber: true })}
                type="number"
                label="Daily Loss Limit (USD)"
                placeholder="1000"
                error={riskErrors.daily_loss_limit_usd?.message}
              />
            </div>

            <div className="bg-danger/10 border border-danger/50 rounded-lg p-4 text-sm text-danger">
              <strong>⚠️ Warning:</strong> Traders will be automatically stopped if any of these limits are breached.
            </div>

            <Button
              type="submit"
              variant="primary"
              isLoading={updateMutation.isPending}
              className="w-full"
            >
              <Save className="w-4 h-4 mr-2" />
              Save Risk Limits
            </Button>
          </form>
        </Card>
      )}

      {/* Strategy Prompt Tab */}
      {activeTab === 'strategy' && (
        <Card>
          <form onSubmit={onStrategySubmit} className="space-y-6">
            <div>
              <h3 className="text-xl font-bold mb-4">Default Strategy Prompt</h3>
              <p className="text-gray-400 text-sm mb-6">
                This is the default AI trading strategy prompt. You can customize it per-trader when creating a new bot.
              </p>
            </div>

            <div>
              <label className="label">Strategy Prompt</label>
              <textarea
                value={strategyPrompt}
                onChange={(e) => setStrategyPrompt(e.target.value)}
                rows={16}
                className="w-full px-4 py-2.5 bg-background border border-border rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-primary font-mono text-sm"
                placeholder="Enter your default trading strategy..."
              />
              <p className="mt-2 text-xs text-gray-500">
                This prompt guides the AI's trading decisions. Be specific about entry/exit rules, risk management, and indicators.
              </p>
            </div>

            <div className="bg-primary/10 border border-primary/50 rounded-lg p-4 text-sm">
              <strong>💡 Best Practices:</strong>
              <ul className="list-disc list-inside mt-2 space-y-1 text-gray-300">
                <li>Define clear entry and exit rules</li>
                <li>Specify risk management parameters</li>
                <li>Mention technical indicators to use</li>
                <li>Set minimum confidence thresholds</li>
                <li>Include JSON output format requirements</li>
              </ul>
            </div>

            <Button
              type="submit"
              variant="primary"
              isLoading={updateMutation.isPending}
              className="w-full"
            >
              <Save className="w-4 h-4 mr-2" />
              Save Strategy Prompt
            </Button>
          </form>
        </Card>
      )}
    </div>
  )
}
