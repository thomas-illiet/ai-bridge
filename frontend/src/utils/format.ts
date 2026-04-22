export function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' })
}

export function fmtNum(n: number): string {
  if (n >= 1_000_000_000_000) return (n / 1_000_000_000_000).toFixed(1) + 'T'
  if (n >= 1_000_000_000)     return (n / 1_000_000_000).toFixed(1) + 'B'
  if (n >= 1_000_000)         return (n / 1_000_000).toFixed(1) + 'M'
  if (n >= 1_000)             return (n / 1_000).toFixed(1) + 'K'
  return String(n)
}

export function tokenStatus(token: { revokedAt: string | null; expiresAt: string | null }): 'active' | 'revoked' | 'expired' {
  if (token.revokedAt) return 'revoked'
  if (token.expiresAt && new Date(token.expiresAt) < new Date()) return 'expired'
  return 'active'
}

export function isExpiringSoon(token: { revokedAt: string | null; expiresAt: string | null }, days = 3): boolean {
  if (token.revokedAt || !token.expiresAt) return false
  const expiry = new Date(token.expiresAt)
  const now = new Date()
  if (expiry <= now) return false
  return expiry.getTime() - now.getTime() < days * 24 * 60 * 60 * 1000
}

export function interceptionDuration(row: { startedAt: string; endedAt: string | null }): string {
  if (!row.endedAt) return '—'
  const ms = new Date(row.endedAt).getTime() - new Date(row.startedAt).getTime()
  return ms < 1000 ? ms + 'ms' : (ms / 1000).toFixed(1) + 's'
}

export const PROVIDER_COLORS: Record<string, string> = { openai: '#10b981', ollama: '#6366f1' }
export function providerColor(p: string): string { return PROVIDER_COLORS[p] ?? '#94a3b8' }
