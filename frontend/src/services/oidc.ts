import { UserManager, WebStorageStateStore, type User } from 'oidc-client-ts'
import { getConfig } from './config'

let _manager: UserManager | null = null

function getManager(): UserManager {
  if (!_manager) {
    const cfg = getConfig()
    _manager = new UserManager({
      authority: cfg.oidcAuthority,
      client_id: cfg.oidcClientId,
      redirect_uri: cfg.oidcRedirectUri || window.location.origin,
      post_logout_redirect_uri: window.location.origin,
      response_type: 'code',
      scope: 'openid profile email',
      userStore: new WebStorageStateStore({ store: window.sessionStorage }),
    })
  }
  return _manager
}

let _user: User | null = null

export async function initOidc(): Promise<boolean> {
  if (window.location.search.includes('code=') || window.location.search.includes('error=')) {
    try {
      _user = await getManager().signinRedirectCallback()
      window.history.replaceState({}, '', window.location.pathname)
    } catch {
      window.history.replaceState({}, '', window.location.pathname)
    }
  } else {
    _user = await getManager().getUser()
  }

  getManager().events.addUserLoaded((user) => { _user = user })
  getManager().events.addUserUnloaded(() => { _user = null })

  return !!_user && !_user.expired
}

export function login() {
  return getManager().signinRedirect()
}

export function logout() {
  return getManager().signoutRedirect()
}

export async function getValidToken(): Promise<string | undefined> {
  if (!_user || _user.expired) {
    try {
      _user = await getManager().signinSilent()
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
