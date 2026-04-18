import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { isAuthenticated, getUserInfo, login, logout } from '@/services/keycloak'
import { getMe } from '@/services/api'

interface TokenParsed {
  sub?: string
  preferred_username?: string
  email?: string
  given_name?: string
  family_name?: string
}

export const useAuthStore = defineStore('auth', () => {
  const authenticated = ref(false)
  const tokenParsed   = ref<TokenParsed | null>(null)
  const dbRole        = ref<string>('none') // role from backend DB

  const username = computed(() => tokenParsed.value?.preferred_username ?? '')
  const email    = computed(() => tokenParsed.value?.email ?? '')
  const fullName = computed(
    () => `${tokenParsed.value?.given_name ?? ''} ${tokenParsed.value?.family_name ?? ''}`.trim(),
  )

  function sync() {
    authenticated.value = isAuthenticated()
    tokenParsed.value   = (getUserInfo() as TokenParsed) ?? null
  }

  // Load the DB-managed role from /api/v1/me after authentication.
  async function fetchRole() {
    if (!authenticated.value) { dbRole.value = 'none'; return }
    try {
      const res = await getMe()
      // The backend returns models.User which has Roles: []string{registeredUser.Role}
      dbRole.value = res.data?.roles?.[0] ?? 'none'
    } catch {
      dbRole.value = 'none'
    }
  }

  function hasRole(role: string): boolean {
    return dbRole.value === role
  }

  const isAdmin    = computed(() => dbRole.value === 'admin')
  const isManager  = computed(() => dbRole.value === 'manager')
  const isElevated = computed(() => dbRole.value === 'admin' || dbRole.value === 'manager')

  return {
    authenticated, tokenParsed, dbRole,
    username, email, fullName, isAdmin, isManager, isElevated,
    sync, fetchRole, login, logout, hasRole,
  }
})
