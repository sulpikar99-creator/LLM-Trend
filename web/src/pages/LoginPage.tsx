import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { motion } from 'framer-motion'
import { LogIn, UserPlus, Activity } from 'lucide-react'
import toast from 'react-hot-toast'
import { api } from '@/lib/api'
import { useAuthStore } from '@/stores/authStore'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Card } from '@/components/ui/Card'

const loginSchema = z.object({
  username: z.string().min(3, 'Username must be at least 3 characters'),
  password: z.string().min(6, 'Password must be at least 6 characters'),
})

const registerSchema = z.object({
  username: z.string().min(3, 'Username must be at least 3 characters'),
  email: z.string().email('Invalid email address'),
  password: z.string().min(6, 'Password must be at least 6 characters'),
  beta_code: z.string().min(1, 'Beta code is required'),
})

type LoginFormData = z.infer<typeof loginSchema>
type RegisterFormData = z.infer<typeof registerSchema>

export default function LoginPage() {
  const navigate = useNavigate()
  const { setAuth } = useAuthStore()
  const [isLogin, setIsLogin] = useState(true)
  const [isLoading, setIsLoading] = useState(false)

  const {
    register: registerLogin,
    handleSubmit: handleLoginSubmit,
    formState: { errors: loginErrors },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
  })

  const {
    register: registerSignup,
    handleSubmit: handleRegisterSubmit,
    formState: { errors: registerErrors },
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
  })

  const onLogin = async (data: LoginFormData) => {
    setIsLoading(true)
    try {
      const response = await api.login(data)
      if (response.data && response.data.token) {
        setAuth(response.data.token, {
          id: response.data.user_id,
          username: response.data.username,
          email: '',
          role: response.data.role,
          created_at: '',
        })
        toast.success('Login successful!')
        navigate('/')
      }
    } catch (error: any) {
      toast.error(error.message || 'Login failed')
    } finally {
      setIsLoading(false)
    }
  }

  const onRegister = async (data: RegisterFormData) => {
    setIsLoading(true)
    try {
      const response = await api.register(data)
      if (response.data && response.data.token) {
        setAuth(response.data.token, {
          id: response.data.user_id,
          username: response.data.username,
          email: data.email,
          role: response.data.role,
          created_at: '',
        })
        toast.success('Registration successful!')
        navigate('/')
      }
    } catch (error: any) {
      toast.error(error.message || 'Registration failed')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-background px-4 py-12">
      <div className="absolute inset-0 overflow-hidden">
        <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-primary/10 rounded-full blur-3xl animate-pulse-slow" />
        <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-primary/5 rounded-full blur-3xl animate-pulse-slow" />
      </div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="w-full max-w-md relative z-10"
      >
        {/* Logo */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center p-3 bg-primary/10 rounded-2xl mb-4">
            <Activity className="w-12 h-12 text-primary" />
          </div>
          <h1 className="text-4xl font-bold bg-gradient-to-r from-primary to-primary-400 bg-clip-text text-transparent mb-2">
            LLM Trend
          </h1>
          <p className="text-gray-400">AI-Powered Trading Platform</p>
        </div>

        <Card className="overflow-hidden">
          {/* Toggle Tabs */}
          <div className="flex border-b border-border">
            <button
              onClick={() => setIsLogin(true)}
              className={`flex-1 py-4 font-medium transition-colors relative ${
                isLogin ? 'text-primary' : 'text-gray-400 hover:text-white'
              }`}
            >
              <LogIn className="w-5 h-5 inline mr-2" />
              Login
              {isLogin && (
                <motion.div
                  layoutId="tab-indicator"
                  className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary"
                />
              )}
            </button>
            <button
              onClick={() => setIsLogin(false)}
              className={`flex-1 py-4 font-medium transition-colors relative ${
                !isLogin ? 'text-primary' : 'text-gray-400 hover:text-white'
              }`}
            >
              <UserPlus className="w-5 h-5 inline mr-2" />
              Register
              {!isLogin && (
                <motion.div
                  layoutId="tab-indicator"
                  className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary"
                />
              )}
            </button>
          </div>

          {/* Forms */}
          <div className="p-6">
            {isLogin ? (
              <form onSubmit={handleLoginSubmit(onLogin)} className="space-y-4">
                <Input
                  {...registerLogin('username')}
                  label="Username"
                  placeholder="Enter your username"
                  error={loginErrors.username?.message}
                  disabled={isLoading}
                />
                <Input
                  {...registerLogin('password')}
                  type="password"
                  label="Password"
                  placeholder="Enter your password"
                  error={loginErrors.password?.message}
                  disabled={isLoading}
                />
                <Button
                  type="submit"
                  variant="primary"
                  className="w-full"
                  isLoading={isLoading}
                >
                  Login
                </Button>
              </form>
            ) : (
              <form onSubmit={handleRegisterSubmit(onRegister)} className="space-y-4">
                <Input
                  {...registerSignup('username')}
                  label="Username"
                  placeholder="Choose a username"
                  error={registerErrors.username?.message}
                  disabled={isLoading}
                />
                <Input
                  {...registerSignup('email')}
                  type="email"
                  label="Email"
                  placeholder="your@email.com"
                  error={registerErrors.email?.message}
                  disabled={isLoading}
                />
                <Input
                  {...registerSignup('password')}
                  type="password"
                  label="Password"
                  placeholder="Create a password"
                  error={registerErrors.password?.message}
                  disabled={isLoading}
                />
                <Input
                  {...registerSignup('beta_code')}
                  label="Beta Code"
                  placeholder="Enter beta code"
                  error={registerErrors.beta_code?.message}
                  disabled={isLoading}
                />
                <Button
                  type="submit"
                  variant="primary"
                  className="w-full"
                  isLoading={isLoading}
                >
                  Create Account
                </Button>
              </form>
            )}
          </div>
        </Card>

        <p className="text-center text-sm text-gray-500 mt-6">
          By continuing, you agree to our Terms of Service and Privacy Policy
        </p>
      </motion.div>
    </div>
  )
}
