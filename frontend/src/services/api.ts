import axios from 'axios'
import { getValidToken } from './keycloak'

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
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      import('./keycloak').then(({ login }) => login())
    }
    return Promise.reject(err)
  },
)

export const getMe = () => api.get('/me')
export const getDashboard = () => api.get('/dashboard')

export interface ClientToken {
  id: string
  userId: string
  name: string
  lastUsedAt: string | null
  revokedAt: string | null
  createdAt: string
}

export interface CreateTokenResponse {
  token: ClientToken
  rawToken: string
}

export const listTokens = () => api.get<{ tokens: ClientToken[] }>('/tokens')
export const createToken = (name: string) => api.post<CreateTokenResponse>('/tokens', { name })
export const revokeToken = (id: string) => api.delete(`/tokens/${id}`)

export default api
