import axios from 'axios'
import { getValidToken } from './oidc'
import { getConfig } from './config'

const api = axios.create({
  headers: { 'Content-Type': 'application/json' },
})

api.interceptors.request.use(async (config) => {
  config.baseURL = `${getConfig().apiBaseUrl}/api/v1`
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

const dashboardParams = (scope: 'user' | 'global') => scope === 'global' ? { params: { scope: 'global' } } : {}
export const getDashboardTotalRequests = (scope: 'user' | 'global' = 'user') => api.get('/dashboard/total-requests', dashboardParams(scope))
export const getDashboardTokenTotals   = (scope: 'user' | 'global' = 'user') => api.get('/dashboard/tokens', dashboardParams(scope))
export const getDashboardDaily         = (scope: 'user' | 'global' = 'user') => api.get('/dashboard/daily', dashboardParams(scope))
export const getDashboardDailyTokens   = (scope: 'user' | 'global' = 'user') => api.get('/dashboard/daily-tokens', dashboardParams(scope))
export const getDashboardByProvider    = (scope: 'user' | 'global' = 'user') => api.get('/dashboard/by-provider', dashboardParams(scope))
export const getDashboardByModel       = (scope: 'user' | 'global' = 'user') => api.get('/dashboard/by-model', dashboardParams(scope))
export const getDashboardTokensByModel = (scope: 'user' | 'global' = 'user') => api.get('/dashboard/tokens-by-model', dashboardParams(scope))
export const getDashboardToolsUsed     = (scope: 'user' | 'global' = 'user') => api.get('/dashboard/tools-used', dashboardParams(scope))
export const getDashboardLastRequest   = (scope: 'user' | 'global' = 'user') => api.get('/dashboard/last-request', dashboardParams(scope))
export const getDashboardActiveUsers   = () => api.get('/dashboard/active-users', { params: { scope: 'global' } })

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

export const getStatus = () => axios.get<StatusResponse>(`${getConfig().apiBaseUrl}/api/status`)

export const listTokens = (includeInactive = false, sortBy = 'created_at', sortDir = 'desc') =>
  api.get<{ tokens: ClientToken[] }>('/tokens', {
    params: {
      ...(includeInactive ? { include_inactive: 'true' } : {}),
      sort_by: sortBy,
      sort_dir: sortDir,
    },
  })
export const createToken = (name: string, durationDays: number, description = '') =>
  api.post<CreateTokenResponse>('/tokens', { name, durationDays, description })
export const patchToken = (id: string, name: string, description: string) =>
  api.patch<{ token: ClientToken }>(`/tokens/${id}`, { name, description })
export const revokeToken = (id: string) => api.delete(`/tokens/${id}`)

export const listUsers = (sortBy = 'created_at', sortDir = 'asc', search = '', includeService = false) =>
  api.get('/admin/users', { params: { sort_by: sortBy, sort_dir: sortDir, ...(search ? { search } : {}), ...(includeService ? { include_service: 'true' } : {}) } })
export const updateUserRole = (id: string, role: string, expiresAt?: string) =>
  api.patch(`/admin/users/${id}`, { role, expiresAt: expiresAt ?? '' })
export const deleteUser = (id: string) => api.delete(`/admin/users/${id}`)
export const getUserTotalRequests = (id: string) => api.get(`/admin/users/${id}/stats/total-requests`)
export const getUserTokenTotals   = (id: string) => api.get(`/admin/users/${id}/stats/tokens`)
export const getUserDailyRequests = (id: string) => api.get(`/admin/users/${id}/stats/daily`)
export const getUserByProvider    = (id: string) => api.get(`/admin/users/${id}/stats/by-provider`)
export const getUserByModel       = (id: string) => api.get(`/admin/users/${id}/stats/by-model`)

export interface InterceptionRow {
  id: string
  initiatorId: string
  username: string
  provider: string
  providerType: string
  model: string
  clientIp: string
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

export const adminListTokens = (page: number, pageSize: number, search: string, includeInactive = false, sortBy = 'created_at', sortDir = 'desc') =>
  api.get<AdminTokensResponse>('/admin/tokens', { params: { page, pageSize, search, sort_by: sortBy, sort_dir: sortDir, ...(includeInactive ? { include_inactive: 'true' } : {}) } })
export const adminRevokeToken   = (id: string) => api.delete(`/admin/tokens/${id}`)
export const adminUnrevokeToken = (id: string) => api.post(`/admin/tokens/${id}/unrevoke`)

export interface ServiceAccount {
  id: string
  username: string
  description: string
  role: 'service'
  createdAt: string
  updatedAt: string
  tokenCount: number
  totalRequests: number
  totalInput: number
  totalOutput: number
}

export interface CreateServiceTokenResponse {
  token: ClientToken
  rawToken: string
}

export const listServiceAccounts = (sortBy = 'created_at', sortDir = 'desc', search = '') =>
  api.get<{ serviceAccounts: ServiceAccount[] }>('/admin/service-accounts', { params: { sort_by: sortBy, sort_dir: sortDir, ...(search ? { search } : {}) } })
export const createServiceAccount = (username: string, description: string) =>
  api.post<ServiceAccount>('/admin/service-accounts', { username, description })
export const deleteServiceAccount = (id: string) =>
  api.delete(`/admin/service-accounts/${id}`)
export const listServiceTokens = (id: string, includeInactive = false, sortBy = 'created_at', sortDir = 'desc') =>
  api.get<{ tokens: ClientToken[] }>(`/admin/service-accounts/${id}/tokens`, {
    params: { sort_by: sortBy, sort_dir: sortDir, ...(includeInactive ? { include_inactive: 'true' } : {}) }
  })
export const createServiceToken = (id: string, name: string, durationDays: number) =>
  api.post<CreateServiceTokenResponse>(`/admin/service-accounts/${id}/tokens`, { name, durationDays })

export interface ProviderInfo {
  name: string
  displayName: string
  type: 'openai' | 'ollama' | 'anthropic'
}

export const getAvailableProviders = () =>
  api.get<{ providers: ProviderInfo[] }>('/providers')

export const getModels = (provider: string) =>
  api.get<{ models: string[] }>('/models', { params: { provider } })

export interface AIProvider {
  id: string
  name: string
  displayName: string
  type: 'openai' | 'ollama' | 'anthropic'
  baseUrl: string
  config: Record<string, unknown>
  enabled: boolean
  apiKeySet: boolean
  createdAt: string
  updatedAt: string
}

export interface CreateProviderBody {
  name: string
  display_name?: string
  type: 'openai' | 'ollama' | 'anthropic'
  base_url: string
  api_key?: string
  config?: Record<string, unknown>
  enabled: boolean
}

export interface UpdateProviderBody {
  name?: string
  display_name?: string
  base_url?: string
  api_key?: string
  config?: Record<string, unknown>
  enabled?: boolean
}

export const listProviders = (search = '') =>
  api.get<{ providers: AIProvider[] }>('/admin/providers', search ? { params: { search } } : {})
export const createProvider = (body: CreateProviderBody) =>
  api.post<{ provider: AIProvider }>('/admin/providers', body)
export const getProvider = (id: string) =>
  api.get<{ provider: AIProvider }>(`/admin/providers/${id}`)
export const updateProvider = (id: string, body: UpdateProviderBody) =>
  api.put<{ provider: AIProvider }>(`/admin/providers/${id}`, body)
export const deleteProvider = (id: string) =>
  api.delete(`/admin/providers/${id}`)
export const reloadProviders = () =>
  api.post('/admin/providers/reload')

export interface MCPServer {
  id: string
  name: string
  displayName: string
  url: string
  headers: Record<string, string>
  allowPattern: string
  denyPattern: string
  enabled: boolean
  createdAt: string
  updatedAt: string
}

export interface CreateMCPServerBody {
  name: string
  display_name?: string
  url: string
  headers?: Record<string, string>
  allow_pattern?: string
  deny_pattern?: string
  enabled: boolean
}

export interface UpdateMCPServerBody {
  display_name?: string
  url?: string
  headers?: Record<string, string>
  allow_pattern?: string
  deny_pattern?: string
  enabled?: boolean
}

export const listMCPServers = (search = '') =>
  api.get<{ mcp_servers: MCPServer[] }>('/admin/mcp-servers', search ? { params: { search } } : {})
export const getMCPServer = (id: string) =>
  api.get<{ mcp_server: MCPServer }>(`/admin/mcp-servers/${id}`)
export const createMCPServer = (body: CreateMCPServerBody) =>
  api.post<{ mcp_server: MCPServer }>('/admin/mcp-servers', body)
export const updateMCPServer = (id: string, body: UpdateMCPServerBody) =>
  api.put<{ mcp_server: MCPServer }>(`/admin/mcp-servers/${id}`, body)
export const deleteMCPServer = (id: string) =>
  api.delete(`/admin/mcp-servers/${id}`)
export const reloadMCP = () =>
  api.post('/admin/mcp-servers/reload')

export interface FirewallRule {
  id: string; cidr: string; description: string
  action: 'allow' | 'deny'; priority: number
  enabled: boolean; createdAt: string
}

export const reloadFirewall = () =>
  api.post('/admin/firewall/reload')

export const listFirewallRules = (sortBy = 'priority', sortDir = 'asc', search = '') =>
  api.get<{ entries: FirewallRule[] }>('/admin/firewall', { params: { sort_by: sortBy, sort_dir: sortDir, ...(search ? { search } : {}) } })
export const addFirewallRule = (cidr: string, description: string, action: 'allow' | 'deny', priority: number) =>
  api.post<FirewallRule>('/admin/firewall', { cidr, description, action, priority })
export const deleteFirewallRule = (id: string) => api.delete(`/admin/firewall/${id}`)
export const toggleFirewallRule = (id: string, enabled: boolean) => api.patch(`/admin/firewall/${id}`, { enabled })
export const moveFirewallRulePriority = (id: string, direction: 'up' | 'down', orderedIds: string[]) =>
  api.post(`/admin/firewall/${id}/move`, { direction, ordered_ids: orderedIds })

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

export const adminListAccessRequests = (status?: string, sortBy = 'created_at', sortDir = 'desc', search = '') =>
  api.get<{ requests: AccessRequest[]; pendingCount: number }>('/admin/access-requests', {
    params: { sort_by: sortBy, sort_dir: sortDir, ...(status ? { status } : {}), ...(search ? { search } : {}) }
  })
export const adminApproveRequest = (id: string, role: string, expiresAt?: string) =>
  api.post<AccessRequest>(`/admin/access-requests/${id}/approve`, { role, expiresAt })
export const adminRejectRequest = (id: string, note: string) =>
  api.post<AccessRequest>(`/admin/access-requests/${id}/reject`, { note })

export default api
