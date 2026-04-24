<script setup lang="ts">
interface TokenTotals   { totalInput: number; totalOutput: number }
interface LastRequest   { model: string; provider: string; startedAt: string }
defineProps<{
  totalRequests: number
  tokens: TokenTotals
  providerCount: number
  activeUsers?: number
  showActiveUsers: boolean
  lastRequest: LastRequest | null
  loadingRequests?: boolean
  loadingTokens?: boolean
  loadingProviders?: boolean
  loadingActiveUsers?: boolean
  loadingLastRequest?: boolean
}>()

function fmtNum(n: number) {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M'
  if (n >= 1_000)     return (n / 1_000).toFixed(1) + 'K'
  return String(n)
}

function relativeDate(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime()
  const mins = Math.floor(diff / 60_000)
  if (mins < 1)  return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24)  return `${hrs}h ago`
  return `${Math.floor(hrs / 24)}d ago`
}
</script>

<template>
  <div class="stat-grid">
    <div class="stat-card">
      <span class="stat-label">Total Requests</span>
      <span v-if="loadingRequests" class="stat-skeleton" />
      <span v-else class="stat-value">{{ fmtNum(totalRequests) }}</span>
    </div>
    <div class="stat-card">
      <span class="stat-label">Input Tokens</span>
      <span v-if="loadingTokens" class="stat-skeleton" />
      <span v-else class="stat-value">{{ fmtNum(tokens.totalInput) }}</span>
    </div>
    <div class="stat-card">
      <span class="stat-label">Output Tokens</span>
      <span v-if="loadingTokens" class="stat-skeleton" />
      <span v-else class="stat-value">{{ fmtNum(tokens.totalOutput) }}</span>
    </div>
    <div class="stat-card">
      <span class="stat-label">Active Providers</span>
      <span v-if="loadingProviders" class="stat-skeleton" />
      <span v-else class="stat-value">{{ providerCount }}</span>
    </div>
    <div class="stat-card" v-if="showActiveUsers">
      <span class="stat-label">Active Users</span>
      <span v-if="loadingActiveUsers" class="stat-skeleton" />
      <span v-else class="stat-value">{{ fmtNum(activeUsers ?? 0) }}</span>
    </div>
    <div class="stat-card stat-card--last" v-if="loadingLastRequest">
      <span class="stat-label">Last Request</span>
      <span class="stat-skeleton" style="margin-top: 0.3rem" />
      <span class="stat-skeleton stat-skeleton--sm" style="margin-top: 0.4rem" />
    </div>
    <div class="stat-card stat-card--last" v-else-if="lastRequest">
      <span class="stat-label">Last Request</span>
      <span class="last-model">{{ lastRequest.model }}</span>
      <div class="last-meta">
        <span class="last-provider">{{ lastRequest.provider }}</span>
        <span class="last-time">{{ relativeDate(lastRequest.startedAt) }}</span>
      </div>
    </div>
    <div class="stat-card stat-card--last" v-else>
      <span class="stat-label">Last Request</span>
      <span class="last-empty">No requests yet</span>
    </div>
  </div>
</template>

<style scoped>
.stat-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: 1rem;
}
.stat-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 1.1rem 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}
.stat-label { font-size: 0.78rem; font-weight: 600; color: #64748b; text-transform: uppercase; letter-spacing: 0.04em; }
.stat-value { font-size: 1.9rem; font-weight: 700; color: #0f172a; line-height: 1; }

.stat-skeleton {
  height: 2rem;
  width: 70%;
  border-radius: 6px;
  background: linear-gradient(90deg, #f1f5f9 25%, #e2e8f0 50%, #f1f5f9 75%);
  background-size: 200% 100%;
  animation: shimmer 1.4s infinite;
}
.stat-skeleton--sm { height: 1rem; width: 50%; }
@keyframes shimmer {
  0%   { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

.stat-card--last { gap: 0.4rem; }
.last-model { font-size: 0.95rem; font-weight: 700; color: #0f172a; font-family: monospace; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.last-meta { display: flex; align-items: center; justify-content: space-between; gap: 0.5rem; margin-top: auto; }
.last-provider { font-size: 0.75rem; font-weight: 600; text-transform: capitalize; background: #f1f5f9; color: #475569; padding: 0.1rem 0.45rem; border-radius: 999px; }
.last-time { font-size: 0.75rem; color: #94a3b8; }
.last-empty { font-size: 0.85rem; color: #94a3b8; }
</style>
