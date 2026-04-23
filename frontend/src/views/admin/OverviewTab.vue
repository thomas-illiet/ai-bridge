<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { getDashboard, getStatus } from '@/services/api'
import type { StatusResponse } from '@/services/api'
import StatGrid from '@/views/dashboard/StatGrid.vue'
import RequestsChart from '@/views/dashboard/RequestsChart.vue'
import TokensChart from '@/views/dashboard/TokensChart.vue'
import ProviderBreakdown from '@/views/dashboard/ProviderBreakdown.vue'
import ModelBreakdown from '@/views/dashboard/ModelBreakdown.vue'
import TokensByModel from '@/views/dashboard/TokensByModel.vue'
import ToolsUsed from '@/views/dashboard/ToolsUsed.vue'
import ServiceStatus from '@/views/dashboard/ServiceStatus.vue'

interface DailyCount    { date: string; count: number }
interface DailyTokens   { date: string; total: number }
interface ProviderCount { provider: string; count: number }
interface ModelCount    { model: string; count: number }
interface ModelTokens   { model: string; total: number }
interface ToolCount     { tool: string; count: number }
interface TokenTotals   { totalInput: number; totalOutput: number }
interface LastRequest    { model: string; provider: string; startedAt: string }
interface DashboardData {
  user:           string
  scope:          'user' | 'global'
  totalRequests:  number
  activeUsers?:   number
  tokens:         TokenTotals
  daily:          DailyCount[]
  dailyTokens:    DailyTokens[]
  byProvider:     ProviderCount[]
  byModel:        ModelCount[]
  tokensByModel:  ModelTokens[]
  toolsUsed:      ToolCount[]
  lastRequest:    LastRequest | null
}

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
    const [dashRes, statusRes] = await Promise.allSettled([
      getDashboard('global'),
      getStatus(),
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

const chartDays   = computed(() => fillWeek(data.value?.daily ?? [],       k => ({ date: k, count: 0 })))
const chartTokens = computed(() => fillWeek(data.value?.dailyTokens ?? [], k => ({ date: k, total: 0 })))

function formatTime(d: Date) {
  return d.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}
</script>

<template>
  <div class="overview">
    <div class="overview-toolbar">
      <span v-if="lastChecked" class="last-checked">
        <span v-if="refreshing" class="spin-icon" />
        Last updated: {{ formatTime(lastChecked) }}
      </span>
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
        :show-active-users="true"
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

      <div class="charts-row">
        <TokensByModel :models="data.tokensByModel" />
        <ToolsUsed :tools="data.toolsUsed" />
      </div>

      <ServiceStatus :status="status" :refreshing="refreshing" />
    </template>

    <div v-else-if="error" class="state-msg error">{{ error }}</div>
  </div>
</template>

<style scoped>
.overview { display: flex; flex-direction: column; gap: 1.5rem; }

.overview-toolbar {
  display: flex;
  justify-content: flex-end;
}

.last-checked {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.8rem;
  color: #94a3b8;
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
