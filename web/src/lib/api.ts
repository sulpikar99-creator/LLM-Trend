const API_BASE_URL = '/api'

interface ApiResponse<T> {
  success: boolean
  data: T
  error?: string
}

async function fetchApi<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<ApiResponse<T>> {
  const token = localStorage.getItem('token')

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  }

  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers,
  })

  const data = await response.json()

  if (!response.ok) {
    throw new Error(data.error || 'Request failed')
  }

  return data
}

export const api = {
  // Auth endpoints
  login: (username: string, password: string) =>
    fetchApi<{ token: string; user_id: string; username: string; role: string }>(
      '/auth/login',
      {
        method: 'POST',
        body: JSON.stringify({ username, password }),
      }
    ),

  register: (username: string, email: string, password: string, betaCode?: string) =>
    fetchApi<{ token: string; user_id: string; username: string; role: string }>(
      '/auth/register',
      {
        method: 'POST',
        body: JSON.stringify({
          username,
          email,
          password,
          ...(betaCode && { beta_code: betaCode })
        }),
      }
    ),

  // User endpoints
  getProfile: () => fetchApi<any>('/user/profile'),

  // Trader endpoints
  listTraders: () => fetchApi<any>('/traders'),
  getTrader: (id: string) => fetchApi<any>(`/traders/${id}`),
  createTrader: (data: any) => fetchApi<any>('/traders', {
    method: 'POST',
    body: JSON.stringify(data),
  }),
  updateTrader: (id: string, data: any) => fetchApi<any>(`/traders/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  }),
  deleteTrader: (id: string) => fetchApi<any>(`/traders/${id}`, {
    method: 'DELETE',
  }),
  startTrader: (
    id: string,
    data?: { symbol?: string; interval?: string }
  ) =>
    fetchApi<any>(`/traders/${id}/start`, {
      method: 'POST',
      body: JSON.stringify(data ?? {}),
    }),
  stopTrader: (id: string) => fetchApi<any>(`/traders/${id}/stop`, {
    method: 'POST',
  }),

  // Analytics endpoints
  getDrawdown: () => fetchApi<any>('/analytics/drawdown'),
  getMonteCarlo: () => fetchApi<any>('/analytics/montecarlo'),
  getCorrelation: () => fetchApi<any>('/analytics/correlation'),
  getPerformance: () => fetchApi<any>('/analytics/performance'),

  // Config endpoints
  getConfig: () => fetchApi<any>('/config'),
  updateConfig: (data: any) => fetchApi<any>('/config', {
    method: 'PUT',
    body: JSON.stringify(data),
  }),
}
