import type {
  ApiResponse,
  LoginRequest,
  RegisterRequest,
  AuthResponse,
  Trader,
  CreateTraderRequest,
  Config,
} from '@/types'

const API_BASE_URL = '/api'

class ApiClient {
  private getHeaders(): HeadersInit {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    }

    const token = localStorage.getItem('auth-storage')
    if (token) {
      try {
        const parsed = JSON.parse(token)
        if (parsed.state?.token) {
          headers['Authorization'] = `Bearer ${parsed.state.token}`
        }
      } catch (e) {
        console.error('Failed to parse auth token:', e)
      }
    }

    return headers
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    try {
      const response = await fetch(`${API_BASE_URL}${endpoint}`, {
        ...options,
        headers: {
          ...this.getHeaders(),
          ...options.headers,
        },
      })

      if (!response.ok) {
        let errorMessage = 'Request failed'
        try {
          const errorData = await response.json()
          errorMessage = errorData.error || errorData.message || errorMessage
        } catch {
          const errorText = await response.text()
          errorMessage = errorText || `HTTP ${response.status}: ${response.statusText}`
        }
        throw new Error(errorMessage)
      }

      const data = await response.json()
      return data
    } catch (error) {
      if (error instanceof Error) {
        throw error
      }
      throw new Error('An unknown error occurred')
    }
  }

  // Auth endpoints
  async login(data: LoginRequest): Promise<AuthResponse> {
    return this.request<AuthResponse['data']>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async register(data: RegisterRequest): Promise<AuthResponse> {
    return this.request<AuthResponse['data']>('/auth/register', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  // Trader endpoints
  async listTraders(): Promise<ApiResponse<Trader[]>> {
    return this.request<Trader[]>('/traders')
  }

  async getTrader(id: string): Promise<ApiResponse<Trader>> {
    return this.request<Trader>(`/traders/${id}`)
  }

  async createTrader(data: CreateTraderRequest): Promise<ApiResponse<Trader>> {
    return this.request<Trader>('/traders', {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  async startTrader(id: string): Promise<ApiResponse<{ message: string }>> {
    return this.request<{ message: string }>(`/traders/${id}/start`, {
      method: 'POST',
    })
  }

  async stopTrader(id: string): Promise<ApiResponse<{ message: string }>> {
    return this.request<{ message: string }>(`/traders/${id}/stop`, {
      method: 'POST',
    })
  }

  async deleteTrader(id: string): Promise<ApiResponse<{ message: string }>> {
    return this.request<{ message: string }>(`/traders/${id}`, {
      method: 'DELETE',
    })
  }

  // Analytics endpoints
  async getPerformance(traderId?: string): Promise<ApiResponse<any>> {
    const query = traderId ? `?trader_id=${traderId}` : ''
    return this.request<any>(`/analytics/performance${query}`)
  }

  async getDrawdown(traderId?: string): Promise<ApiResponse<any>> {
    const query = traderId ? `?trader_id=${traderId}` : ''
    return this.request<any>(`/analytics/drawdown${query}`)
  }

  async getMonteCarlo(traderId?: string): Promise<ApiResponse<any>> {
    const query = traderId ? `?trader_id=${traderId}` : ''
    return this.request<any>(`/analytics/montecarlo${query}`)
  }

  async getCorrelation(): Promise<ApiResponse<any>> {
    return this.request<any>('/analytics/correlation')
  }

  // Config endpoints
  async getConfig(): Promise<ApiResponse<Config>> {
    return this.request<Config>('/config')
  }

  async updateConfig(data: Partial<Config>): Promise<ApiResponse<Config>> {
    return this.request<Config>('/config', {
      method: 'PUT',
      body: JSON.stringify(data),
    })
  }

  // Health check
  async health(): Promise<ApiResponse<{ status: string }>> {
    return this.request<{ status: string }>('/health')
  }
}

export const api = new ApiClient()
