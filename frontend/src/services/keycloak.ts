import Keycloak from 'keycloak-js'

const keycloak = new Keycloak({
  url: import.meta.env.VITE_KEYCLOAK_URL ?? 'http://localhost:8180',
  realm: import.meta.env.VITE_KEYCLOAK_REALM ?? 'ai-bridge',
  clientId: import.meta.env.VITE_KEYCLOAK_CLIENT_ID ?? 'ai-bridge-frontend',
})

export async function initKeycloak(): Promise<boolean> {
  keycloak.onTokenExpired = () => {
    keycloak.updateToken(30).catch(() => keycloak.login())
  }

  keycloak.onAuthRefreshError = () => {
    keycloak.login()
  }

  return keycloak.init({
    onLoad: 'check-sso',
    silentCheckSsoRedirectUri: `${window.location.origin}/silent-check-sso.html`,
    pkceMethod: 'S256',
  })
}

export function login() {
  return keycloak.login()
}

export function logout() {
  return keycloak.logout({ redirectUri: window.location.origin })
}

export function getToken(): string | undefined {
  return keycloak.token
}

export async function getValidToken(): Promise<string | undefined> {
  try {
    if (keycloak.isTokenExpired(30)) {
      await keycloak.updateToken(30)
    }
    return keycloak.token
  } catch {
    keycloak.login()
    return undefined
  }
}

export function isAuthenticated(): boolean {
  return !!keycloak.authenticated
}

export function getUserInfo() {
  return keycloak.tokenParsed
}

export default keycloak
