import { UserManager, WebStorageStateStore, type User } from 'oidc-client-ts'

const manager = new UserManager({
  authority: import.meta.env.VITE_OIDC_AUTHORITY ?? 'http://localhost:8180/realms/ai-bridge',
  client_id: import.meta.env.VITE_OIDC_CLIENT_ID ?? 'ai-bridge-frontend',
  redirect_uri: import.meta.env.VITE_OIDC_REDIRECT_URI ?? window.location.origin,
  post_logout_redirect_uri: window.location.origin,
  response_type: 'code',
  scope: 'openid profile email',
  userStore: new WebStorageStateStore({ store: window.sessionStorage }),
})

let _user: User | null = null

export async function initOidc(): Promise<boolean> {
  if (window.location.search.includes('code=') || window.location.search.includes('error=')) {
    try {
      _user = await manager.signinRedirectCallback()
      window.history.replaceState({}, '', window.location.pathname)
    } catch {
      window.history.replaceState({}, '', window.location.pathname)
    }
  } else {
    _user = await manager.getUser()
  }

  manager.events.addUserLoaded((user) => { _user = user })
  manager.events.addUserUnloaded(() => { _user = null })

  return !!_user && !_user.expired
}

export function login() {
  return manager.signinRedirect()
}

export function logout() {
  return manager.signoutRedirect()
}

export async function getValidToken(): Promise<string | undefined> {
  if (!_user || _user.expired) {
    try {
      _user = await manager.signinSilent()
    } catch {
      login()
      return undefined
    }
  }
  return _user?.access_token
}

export function isAuthenticated(): boolean {
  return !!_user && !_user.expired
}

export function getUserInfo() {
  return _user?.profile ?? null
}
