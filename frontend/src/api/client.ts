const BASE_URL = '/api/v1'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const token = localStorage.getItem('token')
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options?.headers as Record<string, string>),
  }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const res = await fetch(`${BASE_URL}${path}`, { ...options, headers })

  if (res.status === 401) {
    localStorage.removeItem('token')
    window.location.href = '/login'
    throw new Error('Unauthorized')
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error((body as { error?: string }).error || `HTTP ${res.status}`)
  }

  return res.json()
}

export const api = {
  // Auth
  login: (username: string, password: string) =>
    request<{ token: string }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    }),

  // Health
  health: () => request<{ status: string }>('/health'),

  // Metrics
  queryMetrics: (name: string, host: string, limit = 100) =>
    request<{ data: unknown[]; count: number }>(
      `/metrics/query?name=${name}&host=${host}&limit=${limit}`
    ),

  // Hosts
  getHosts: () => request<{ data: unknown[]; count: number }>('/hosts'),

  // Rules
  getRules: (page = 1, pageSize = 20) =>
    request<{ data: unknown[]; total: number }>(`/rules?page=${page}&page_size=${pageSize}`),
  createRule: (rule: unknown) =>
    request<unknown>('/rules', { method: 'POST', body: JSON.stringify(rule) }),
  updateRule: (id: string, rule: unknown) =>
    request<unknown>(`/rules/${id}`, { method: 'PUT', body: JSON.stringify(rule) }),
  deleteRule: (id: string) =>
    request<unknown>(`/rules/${id}`, { method: 'DELETE' }),

  // Alerts
  getAlerts: (page = 1, pageSize = 20, status = '', severity = '') => {
    const params = new URLSearchParams({ page: String(page), page_size: String(pageSize) })
    if (status) params.set('status', status)
    if (severity) params.set('severity', severity)
    return request<{ data: unknown[]; total: number }>(`/alerts?${params}`)
  },
  getAlertStats: () => request<{ firing: number; resolved: number; total: number }>('/alerts/stats'),
}
