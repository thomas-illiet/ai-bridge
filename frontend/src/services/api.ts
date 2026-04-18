import axios from 'axios'
import { getValidToken } from './oidc'

const api = axios.create({
  baseURL: '/api/v1',
  headers: { 'Content-Type': 'application/json' },
})

api.interceptors.request.use(async (config) => {
  const token = await getValidToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (res) => {
    import('@/stores/connectivity').then(({ useConnectivityStore }) => {
      useConnectivityStore().markUp()
    })
    return res
  },
  (err) => {
    if (!err.response) {
      import('@/stores/connectivity').then(({ useConnectivityStore }) => {
        useConnectivityStore().markDown()
      })
    } else if (err.response.status === 401) {
      import('./oidc').then(({ login }) => login())
    }
    return Promise.reject(err)
  },
)

export const getMe = () => api.get('/me')
export const getDashboard = (scope: 'user' | 'global' = 'user') =>
  api.get('/dashboard', scope === 'global' ? { params: { scope: 'global' } } : {})

export interface ClientToken {
  id: string
  userId: string
  name: string
  description: string
  expiresAt: string | null
  lastUsedAt: string | null
  revokedAt: string | null
  createdAt: string
}

export interface AdminTokenRow extends ClientToken {
  username: string
}

export interface AdminTokensResponse {
  tokens: AdminTokenRow[]
  total: number
  page: number
  pageSize: number
}

export interface CreateTokenResponse {
  token: ClientToken
  rawToken: string
}

export interface ServiceStatus {
  name: string
  status: 'up' | 'down' | 'disabled'
  message?: string
  latency_ms?: number
  model_count?: number
}

export interface StatusResponse {
  status: 'healthy' | 'degraded'
  services: ServiceStatus[]
}

export const getStatus = () => axios.get<StatusResponse>('/api/status')

export const listTokens = (includeRevoked = false, sortBy = 'created_at', sortDir = 'desc') =>
  api.get<{ tokens: ClientToken[] }>('/tokens', {
    params: {
      ...(includeRevoked ? { include_revoked: 'true' } : {}),
      sort_by: sortBy,
      sort_dir: sortDir,
    },
  })
export const createToken = (name: string, durationDays: number, description = '') =>
  api.post<CreateTokenResponse>('/tokens', { name, durationDays, description })
export const patchToken = (id: string, name: string, description: string) =>
  api.patch<{ token: ClientToken }>(`/tokens/${id}`, { name, description })
export const revokeToken = (id: string) => api.delete(`/tokens/${id}`)

export const listUsers = (sortBy = 'created_at', sortDir = 'asc') =>
  api.get('/admin/users', { params: { sort_by: sortBy, sort_dir: sortDir } })
export const updateUserRole = (id: string, role: string, expiresAt?: string) =>
  api.patch(`/admin/users/${id}`, { role, expiresAt: expiresAt ?? '' })
export const deleteUser = (id: string) => api.delete(`/admin/users/${id}`)
export const getUserStats = (id: string) => api.get(`/admin/users/${id}/stats`)

export interface InterceptionRow {
  id: string
  initiatorId: string
  username: string
  provider: string
  model: string
  startedAt: string
  endedAt: string | null
  inputTokens: number
  outputTokens: number
}

export interface InterceptionDetail extends InterceptionRow {
  prompts: string[]
}

export interface HistoryResponse {
  interceptions: InterceptionRow[]
  total: number
  page: number
  pageSize: number
}

export interface HistoryStats {
  total: number
  totalInput: number
  totalOutput: number
  topModel: string
}

export const getHistory = (page: number, pageSize: number, search: string, sortBy = 'startedAt', sortDir = 'desc') =>
  api.get<HistoryResponse>('/history', { params: { page, pageSize, search, sortBy, sortDir } })
export const getHistoryStats = () => api.get<HistoryStats>('/history/stats')
export const getHistoryDetail = (id: string) =>
  api.get<InterceptionDetail>(`/history/${id}`)

export const adminGetHistory = (page: number, pageSize: number, search: string, userId: string, sortBy = 'startedAt', sortDir = 'desc') =>
  api.get<HistoryResponse>('/admin/history', { params: { page, pageSize, search, userId, sortBy, sortDir } })
export const adminGetHistoryDetail = (id: string) =>
  api.get<InterceptionDetail>(`/admin/history/${id}`)

export const adminListTokens = (page: number, pageSize: number, search: string, includeRevoked = false, sortBy = 'created_at', sortDir = 'desc') =>
  api.get<AdminTokensResponse>('/admin/tokens', { params: { page, pageSize, search, sort_by: sortBy, sort_dir: sortDir, ...(includeRevoked ? { include_revoked: 'true' } : {}) } })
export const adminRevokeToken   = (id: string) => api.delete(`/admin/tokens/${id}`)
export const adminUnrevokeToken = (id: string) => api.post(`/admin/tokens/${id}/unrevoke`)

export interface ServiceAccount {
  id: string
  username: string
  description: string
  role: 'service'
  createdAt: string
  updatedAt: string
}

export interface CreateServiceTokenResponse {
  token: ClientToken
  rawToken: string
}

export const listServiceAccounts = (sortBy = 'created_at', sortDir = 'desc') =>
  api.get<{ serviceAccounts: ServiceAccount[] }>('/admin/service-accounts', { params: { sort_by: sortBy, sort_dir: sortDir } })
export const createServiceAccount = (username: string, description: string) =>
  api.post<ServiceAccount>('/admin/service-accounts', { username, description })
export const deleteServiceAccount = (id: string) =>
  api.delete(`/admin/service-accounts/${id}`)
export const listServiceTokens = (id: string, includeRevoked = false, sortBy = 'created_at', sortDir = 'desc') =>
  api.get<{ tokens: ClientToken[] }>(`/admin/service-accounts/${id}/tokens`, {
    params: { sort_by: sortBy, sort_dir: sortDir, ...(includeRevoked ? { include_revoked: 'true' } : {}) }
  })
export const createServiceToken = (id: string, name: string, durationDays: number) =>
  api.post<CreateServiceTokenResponse>(`/admin/service-accounts/${id}/tokens`, { name, durationDays })

export const getModels = (provider: 'openai' | 'ollama') =>
  api.get<{ models: string[] }>('/models', { params: { provider } })

export const listWhitelist = (sortBy = 'created_at', sortDir = 'desc') =>
  api.get('/admin/whitelist', { params: { sort_by: sortBy, sort_dir: sortDir } })
export const addWhitelist = (cidr: string, description: string) => api.post('/admin/whitelist', { cidr, description })
export const deleteWhitelist = (id: string) => api.delete(`/admin/whitelist/${id}`)
export const toggleWhitelist = (id: string, enabled: boolean) => api.patch(`/admin/whitelist/${id}`, { enabled })

export interface AccessRequest {
  id: string
  userId: string
  status: 'pending' | 'approved' | 'rejected'
  reason: string
  reviewNote: string
  reviewedBy: string
  reviewedAt: string | null
  createdAt: string
  user?: { id: string; username: string; email: string; role: string }
}

export const createAccessRequest = (reason: string) =>
  api.post<AccessRequest>('/access-requests', { reason })
export const getMyAccessRequest = () =>
  api.get<AccessRequest | null>('/access-requests/me')

export const adminListAccessRequests = (status?: string, sortBy = 'created_at', sortDir = 'desc') =>
  api.get<{ requests: AccessRequest[]; pendingCount: number }>('/admin/access-requests', {
    params: { sort_by: sortBy, sort_dir: sortDir, ...(status ? { status } : {}) }
  })
export const adminApproveRequest = (id: string, role: string, expiresAt?: string) =>
  api.post<AccessRequest>(`/admin/access-requests/${id}/approve`, { role, expiresAt })
export const adminRejectRequest = (id: string, note: string) =>
  api.post<AccessRequest>(`/admin/access-requests/${id}/reject`, { note })

export default api
