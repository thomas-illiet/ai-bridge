import { defineStore } from 'pinia'
import { ref } from 'vue'
import { listTokens, createToken, revokeToken, type ClientToken, type CreateTokenResponse } from '@/services/api'

export const useTokenStore = defineStore('tokens', () => {
  const tokens = ref<ClientToken[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchTokens(): Promise<void> {
    loading.value = true
    error.value = null
    try {
      const res = await listTokens()
      tokens.value = res.data.tokens
    } catch {
      error.value = 'Failed to load tokens'
    } finally {
      loading.value = false
    }
  }

  async function generateToken(name: string): Promise<CreateTokenResponse> {
    const res = await createToken(name)
    return res.data
  }

  async function deleteToken(id: string): Promise<void> {
    await revokeToken(id)
    const token = tokens.value.find((t) => t.id === id)
    if (token) {
      token.revokedAt = new Date().toISOString()
    }
  }

  return { tokens, loading, error, fetchTokens, generateToken, deleteToken }
})
