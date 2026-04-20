<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { getDashboard } from '@/services/api'
import { useAuthStore } from '@/stores/auth'
import axios from 'axios'
import type { StatusResponse } from '@/services/api'
import StatGrid from '@/views/dashboard/StatGrid.vue'
import RequestsChart from '@/views/dashboard/RequestsChart.vue'
import TokensChart from '@/views/dashboard/TokensChart.vue'
import ProviderBreakdown from '@/views/dashboard/ProviderBreakdown.vue'
import ModelBreakdown from '@/views/dashboard/ModelBreakdown.vue'
import ServiceStatus from '@/views/dashboard/ServiceStatus.vue'

interface DailyCount    { date: string; count: number }
interface DailyTokens   { date: string; total: number }
interface ProviderCount { provider: string; count: number }
interface ModelCount    { model: string; count: number }
interface TokenTotals   { totalInput: number; totalOutput: number }
interface LastRequest    { model: string; provider: string; startedAt: string }
interface DashboardData {
  user:          string
  scope:         'user' | 'global'
  totalRequests: number
  activeUsers?:  number
  tokens:        TokenTotals
  daily:         DailyCount[]
  dailyTokens:   DailyTokens[]
  byProvider:    ProviderCount[]
  byModel:       ModelCount[]
  lastRequest:   LastRequest | null
}

const auth       = useAuthStore()
const scope      = ref<'user' | 'global'>('user')
const data       = ref<DashboardData | null>(null)
const status     = ref<StatusResponse | null>(null)
const loading    = ref(true)
const refreshing = ref(false)
const error      = ref<string | null>(null)
const lastChecked = ref<Date | null>(null)

let interval: ReturnType<typeof setInterval>

async function fetchAll() {
  if (data.value) refreshing.value = true
  try {
    const effectiveScope = auth.isElevated ? scope.value : 'user'
    const [dashRes, statusRes] = await Promise.allSettled([
      getDashboard(effectiveScope),
      axios.get<StatusResponse>('/api/status'),
    ])
    if (dashRes.status === 'fulfilled') data.value = dashRes.value.data
    else error.value = 'Failed to load dashboard data'

    if (statusRes.status === 'fulfilled') status.value = statusRes.value.data
    else if (statusRes.status === 'rejected') {
      const e = statusRes.reason as any
      if (e?.response?.data?.services) status.value = e.response.data
    }
    lastChecked.value = new Date()
  } finally {
    loading.value    = false
    refreshing.value = false
  }
}

watch(scope, () => fetchAll())
onMounted(() => { fetchAll(); interval = setInterval(fetchAll, 30_000) })
onUnmounted(() => clearInterval(interval))

function fillWeek<T extends { date: string }>(
  source: T[],
  empty: (date: string) => T
): T[] {
  const map = new Map(source.map(d => [d.date, d]))
  const days: T[] = []
  for (let i = 6; i >= 0; i--) {
    const d = new Date()
    d.setDate(d.getDate() - i)
    const key = d.toISOString().slice(0, 10)
    days.push(map.get(key) ?? empty(key))
  }
  return days
}

const chartDays     = computed(() => fillWeek(data.value?.daily ?? [],       k => ({ date: k, count: 0 })))
const chartTokens   = computed(() => fillWeek(data.value?.dailyTokens ?? [], k => ({ date: k, total: 0 })))

function formatTime(d: Date) {
  return d.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}
</script>

<template>
  <div class="dashboard">
    <div class="page-header">
      <div>
        <h1>Dashboard</h1>
        <p v-if="data" class="subtitle">Welcome back, <strong>{{ data.user }}</strong></p>
      </div>
      <div class="header-right">
        <div v-if="auth.isElevated" class="scope-toggle">
          <button class="scope-btn" :class="{ active: scope === 'user' }" @click="scope = 'user'">
            Mon activity
          </button>
          <button class="scope-btn" :class="{ active: scope === 'global' }" @click="scope = 'global'">
            Global
          </button>
        </div>
        <span v-if="lastChecked" class="last-checked">
          <span v-if="refreshing" class="spin-icon" />
          Last updated: {{ formatTime(lastChecked) }}
        </span>
      </div>
    </div>

    <div v-if="loading" class="skeleton-section">
      <div class="skeleton-grid-5">
        <div v-for="i in 5" :key="i" class="skeleton-card tall" />
      </div>
      <div class="skeleton-grid-2">
        <div class="skeleton-card chart" />
        <div class="skeleton-card chart" />
      </div>
      <div class="skeleton-grid-2">
        <div class="skeleton-card chart" />
        <div class="skeleton-card chart" />
      </div>
    </div>

    <template v-else-if="data">
      <StatGrid
        :total-requests="data.totalRequests"
        :tokens="data.tokens"
        :provider-count="data.byProvider.length"
        :active-users="data.activeUsers"
        :show-active-users="data.scope === 'global'"
        :last-request="data.lastRequest"
      />

      <div class="charts-row">
        <RequestsChart :days="chartDays" />
        <TokensChart :days="chartTokens" />
      </div>

      <div class="charts-row">
        <ProviderBreakdown :providers="data.byProvider" />
        <ModelBreakdown :models="data.byModel" />
      </div>

      <ServiceStatus :status="status" :refreshing="refreshing" />
    </template>

    <div v-else-if="error" class="state-msg error">{{ error }}</div>
  </div>
</template>

<style scoped>
.dashboard { display: flex; flex-direction: column; gap: 1.5rem; }

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 0.5rem;
}
h1 { font-size: 1.75rem; font-weight: 700; margin: 0; }
.subtitle { color: #64748b; font-size: 0.9rem; margin: 0.2rem 0 0; }

.header-right {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.4rem;
}
.scope-toggle {
  display: flex;
  background: #f1f5f9;
  border-radius: 8px;
  padding: 3px;
  gap: 2px;
}
.scope-btn {
  padding: 0.25rem 0.75rem;
  border: none;
  border-radius: 6px;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  background: transparent;
  color: #64748b;
  transition: background 0.15s, color 0.15s;
}
.scope-btn.active {
  background: white;
  color: #0f172a;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}
.last-checked {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.8rem;
  color: #94a3b8;
  padding-top: 0.35rem;
}
.spin-icon {
  display: inline-block;
  width: 10px; height: 10px;
  border: 2px solid #cbd5e1;
  border-top-color: #64748b;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

.skeleton-section { display: flex; flex-direction: column; gap: 1rem; }
.skeleton-grid-5 { display: grid; grid-template-columns: repeat(auto-fill, minmax(150px, 1fr)); gap: 1rem; }
.skeleton-grid-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
.skeleton-card {
  background: linear-gradient(90deg, #f1f5f9 25%, #e2e8f0 50%, #f1f5f9 75%);
  background-size: 200% 100%;
  animation: shimmer 1.4s infinite;
  border-radius: 10px;
}
.skeleton-card.tall  { height: 88px; }
.skeleton-card.chart { height: 220px; }
@keyframes shimmer {
  0%   { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

.charts-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}
@media (max-width: 700px) { .charts-row { grid-template-columns: 1fr; } }

.state-msg { color: #64748b; }
.state-msg.error { color: #ef4444; }
</style>
