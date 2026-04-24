<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import {
  getMe,
  getDashboardTotalRequests,
  getDashboardTokenTotals,
  getDashboardDaily,
  getDashboardDailyTokens,
  getDashboardByProvider,
  getDashboardByModel,
  getDashboardTokensByModel,
  getDashboardToolsUsed,
  getDashboardLastRequest,
} from '@/services/api'
import { sleep } from '@/utils/format'
import StatGrid from '@/views/dashboard/StatGrid.vue'
import RequestsChart from '@/views/dashboard/RequestsChart.vue'
import TokensChart from '@/views/dashboard/TokensChart.vue'
import ProviderBreakdown from '@/views/dashboard/ProviderBreakdown.vue'
import ModelBreakdown from '@/views/dashboard/ModelBreakdown.vue'
import TokensByModel from '@/views/dashboard/TokensByModel.vue'
import ToolsUsed from '@/views/dashboard/ToolsUsed.vue'

interface DailyCount    { date: string; count: number }
interface DailyTokens   { date: string; total: number }
interface ProviderCount { provider: string; count: number }
interface ModelCount    { model: string; count: number }
interface TokenTotals   { totalInput: number; totalOutput: number }
interface LastRequest    { model: string; provider: string; startedAt: string }
interface ModelTokens  { model: string; total: number }
interface ToolCount    { tool: string; count: number }

const userName        = ref('')
const totalRequests   = ref(0)
const tokens          = ref<TokenTotals>({ totalInput: 0, totalOutput: 0 })
const daily           = ref<DailyCount[]>([])
const dailyTokens     = ref<DailyTokens[]>([])
const byProvider      = ref<ProviderCount[]>([])
const byModel         = ref<ModelCount[]>([])
const tokensByModel   = ref<ModelTokens[]>([])
const toolsUsed       = ref<ToolCount[]>([])
const lastRequest     = ref<LastRequest | null>(null)

const loadingUser         = ref(true)
const loadingRequests     = ref(true)
const loadingTokens       = ref(true)
const loadingDaily        = ref(true)
const loadingDailyTokens  = ref(true)
const loadingProviders    = ref(true)
const loadingModels       = ref(true)
const loadingTokensByModel = ref(true)
const loadingTools        = ref(true)
const loadingLastRequest  = ref(true)

const refreshing   = ref(false)
const lastChecked  = ref<Date | null>(null)

let interval: ReturnType<typeof setInterval>

const MIN_SKELETON_MS = 400

function fetchAll() {
  const isRefresh = lastChecked.value !== null
  if (isRefresh) refreshing.value = true
  const delay = isRefresh ? 0 : MIN_SKELETON_MS

  Promise.all([getMe(), sleep(delay)])
    .then(([r]) => { userName.value = r.data.preferredUsername || r.data.username || '' })
    .finally(() => { loadingUser.value = false })

  Promise.all([getDashboardTotalRequests('user'), sleep(delay)])
    .then(([r]) => { totalRequests.value = r.data.totalRequests })
    .finally(() => { loadingRequests.value = false })

  Promise.all([getDashboardTokenTotals('user'), sleep(delay)])
    .then(([r]) => { tokens.value = r.data })
    .finally(() => { loadingTokens.value = false })

  Promise.all([getDashboardDaily('user'), sleep(delay)])
    .then(([r]) => { daily.value = r.data.daily })
    .finally(() => { loadingDaily.value = false })

  Promise.all([getDashboardDailyTokens('user'), sleep(delay)])
    .then(([r]) => { dailyTokens.value = r.data.dailyTokens })
    .finally(() => { loadingDailyTokens.value = false })

  Promise.all([getDashboardByProvider('user'), sleep(delay)])
    .then(([r]) => { byProvider.value = r.data.byProvider })
    .finally(() => { loadingProviders.value = false })

  Promise.all([getDashboardByModel('user'), sleep(delay)])
    .then(([r]) => { byModel.value = r.data.byModel })
    .finally(() => { loadingModels.value = false })

  Promise.all([getDashboardTokensByModel('user'), sleep(delay)])
    .then(([r]) => { tokensByModel.value = r.data.tokensByModel })
    .finally(() => { loadingTokensByModel.value = false })

  Promise.all([getDashboardToolsUsed('user'), sleep(delay)])
    .then(([r]) => { toolsUsed.value = r.data.toolsUsed })
    .finally(() => { loadingTools.value = false })

  Promise.all([getDashboardLastRequest('user'), sleep(delay)])
    .then(([r]) => { lastRequest.value = r.data.lastRequest })
    .finally(() => {
      loadingLastRequest.value = false
      lastChecked.value = new Date()
      refreshing.value = false
    })
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

const chartDays   = computed(() => fillWeek(daily.value,       k => ({ date: k, count: 0 })))
const chartTokens = computed(() => fillWeek(dailyTokens.value, k => ({ date: k, total: 0 })))

function formatTime(d: Date) {
  return d.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}
</script>

<template>
  <div class="dashboard">
    <div class="page-header">
      <div>
        <h1>My Activity</h1>
        <p v-if="!loadingUser && userName" class="subtitle">Welcome back, <strong>{{ userName }}</strong></p>
        <p v-else-if="loadingUser" class="subtitle subtitle-skeleton" />
      </div>
      <div class="header-right">
        <span v-if="lastChecked" class="last-checked">
          <span v-if="refreshing" class="spin-icon" />
          Last updated: {{ formatTime(lastChecked) }}
        </span>
      </div>
    </div>

    <StatGrid
      :total-requests="totalRequests"
      :tokens="tokens"
      :provider-count="byProvider.length"
      :show-active-users="false"
      :last-request="lastRequest"
      :loading-requests="loadingRequests"
      :loading-tokens="loadingTokens"
      :loading-providers="loadingProviders"
      :loading-last-request="loadingLastRequest"
    />

    <div class="charts-row">
      <RequestsChart :days="chartDays" :loading="loadingDaily" />
      <TokensChart :days="chartTokens" :loading="loadingDailyTokens" />
    </div>

    <div class="charts-row">
      <ProviderBreakdown :providers="byProvider" :loading="loadingProviders" />
      <ModelBreakdown :models="byModel" :loading="loadingModels" />
    </div>

    <div class="charts-row">
      <TokensByModel :models="tokensByModel" :loading="loadingTokensByModel" />
      <ToolsUsed :tools="toolsUsed" :loading="loadingTools" />
    </div>
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
.subtitle-skeleton {
  width: 180px; height: 0.9rem; border-radius: 4px; margin-top: 0.4rem;
  background: linear-gradient(90deg, #f1f5f9 25%, #e2e8f0 50%, #f1f5f9 75%);
  background-size: 200% 100%;
  animation: shimmer 1.4s infinite;
}

.header-right {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.4rem;
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
</style>
