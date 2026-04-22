export interface AppConfig {
  apiBaseUrl: string
  bridgeBaseUrl: string
  oidcAuthority: string
  oidcClientId: string
  oidcRedirectUri: string
}

const DEFAULTS: AppConfig = {
  apiBaseUrl:    'http://localhost:8585',
  bridgeBaseUrl: 'http://localhost:8586',
  oidcAuthority: 'http://localhost:8180/realms/ai-bridge',
  oidcClientId:  'ai-bridge-frontend',
  oidcRedirectUri: '',
}

let _config: AppConfig | null = null

export async function loadConfig(): Promise<AppConfig> {
  if (_config) return _config
  try {
    const res = await fetch('/config.json', { cache: 'no-store' })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    _config = { ...DEFAULTS, ...(await res.json()) }
  } catch (err) {
    console.warn('Failed to load /config.json, using defaults', err)
    _config = { ...DEFAULTS }
  }
  return _config!
}

export function getConfig(): AppConfig {
  if (!_config) throw new Error('Config not loaded. Await loadConfig() first.')
  return _config
}
