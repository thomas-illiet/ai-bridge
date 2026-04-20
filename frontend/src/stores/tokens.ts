import { defineStore } from 'pinia'
import { ref } from 'vue'
import { listTokens, createToken, patchToken, revokeToken, type ClientToken, type CreateTokenResponse } from '@/services/api'

export const useTokenStore = defineStore('tokens', () => {
  const tokens = ref<ClientToken[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchTokens(includeInactive = false, sortBy = 'created_at', sortDir = 'desc'): Promise<void> {
    loading.value = true
    error.value = null
    try {
      await Promise.all([
        listTokens(includeInactive, sortBy, sortDir).then(res => { tokens.value = res.data.tokens }).catch(() => { error.value = 'Failed to load tokens' }),
        new Promise<void>(r => setTimeout(r, 300)),
      ])
    } finally {
      loading.value = false
    }
  }

  async function generateToken(name: string, durationDays: number, description = ''): Promise<CreateTokenResponse> {
    const res = await createToken(name, durationDays, description)
    return res.data
  }

  async function updateToken(id: string, name: string, description: string): Promise<void> {
    const res = await patchToken(id, name, description)
    const idx = tokens.value.findIndex((t) => t.id === id)
    if (idx !== -1) tokens.value[idx] = res.data.token
  }

  async function deleteToken(id: string): Promise<void> {
    await revokeToken(id)
    const token = tokens.value.find((t) => t.id === id)
    if (token) {
      token.revokedAt = new Date().toISOString()
    }
  }

  return { tokens, loading, error, fetchTokens, generateToken, updateToken, deleteToken }
})
