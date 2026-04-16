import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { isAuthenticated, getUserInfo, login, logout } from '@/services/keycloak'

interface TokenParsed {
  sub?: string
  preferred_username?: string
  email?: string
  given_name?: string
  family_name?: string
  realm_access?: { roles: string[] }
}

export const useAuthStore = defineStore('auth', () => {
  const authenticated = ref(false)
  const tokenParsed = ref<TokenParsed | null>(null)

  const username = computed(() => tokenParsed.value?.preferred_username ?? '')
  const email = computed(() => tokenParsed.value?.email ?? '')
  const fullName = computed(
    () => `${tokenParsed.value?.given_name ?? ''} ${tokenParsed.value?.family_name ?? ''}`.trim(),
  )
  const roles = computed(() => tokenParsed.value?.realm_access?.roles ?? [])

  function sync() {
    authenticated.value = isAuthenticated()
    tokenParsed.value = (getUserInfo() as TokenParsed) ?? null
  }

  function hasRole(role: string): boolean {
    return roles.value.includes(role)
  }

  return { authenticated, tokenParsed, username, email, fullName, roles, sync, login, logout, hasRole }
})
