<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import {
  getDashboardTotalRequests,
  getDashboardTokenTotals,
  getDashboardDaily,
  getDashboardDailyTokens,
  getDashboardByProvider,
  getDashboardByModel,
  getDashboardTokensByModel,
  getDashboardToolsUsed,
  getDashboardLastRequest,
  getDashboardActiveUsers,
  getStatus,
} from '@/services/api'
import type { StatusResponse } from '@/services/api'
import { sleep } from '@/utils/format'
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

const totalRequests   = ref(0)
const activeUsers     = ref<number | undefined>(undefined)
const tokens          = ref<TokenTotals>({ totalInput: 0, totalOutput: 0 })
const daily           = ref<DailyCount[]>([])
const dailyTokens     = ref<DailyTokens[]>([])
const byProvider      = ref<ProviderCount[]>([])
const byModel         = ref<ModelCount[]>([])
const tokensByModel   = ref<ModelTokens[]>([])
const toolsUsed       = ref<ToolCount[]>([])
const lastRequest     = ref<LastRequest | null>(null)
const status          = ref<StatusResponse | null>(null)

const loadingRequests      = ref(true)
const loadingTokens        = ref(true)
const loadingDaily         = ref(true)
const loadingDailyTokens   = ref(true)
const loadingProviders     = ref(true)
const loadingModels        = ref(true)
const loadingTokensByModel = ref(true)
const loadingTools         = ref(true)
const loadingLastRequest   = ref(true)
const loadingActiveUsers   = ref(true)
const loadingStatus        = ref(true)

const refreshing   = ref(false)
const lastChecked  = ref<Date | null>(null)

let interval: ReturnType<typeof setInterval>

const MIN_SKELETON_MS = 400

function fetchAll() {
  const isRefresh = lastChecked.value !== null
  if (isRefresh) refreshing.value = true
  const delay = isRefresh ? 0 : MIN_SKELETON_MS

  Promise.all([getDashboardTotalRequests('global'), sleep(delay)])
    .then(([r]) => { totalRequests.value = r.data.totalRequests })
    .finally(() => { loadingRequests.value = false })

  Promise.all([getDashboardTokenTotals('global'), sleep(delay)])
    .then(([r]) => { tokens.value = r.data })
    .finally(() => { loadingTokens.value = false })

  Promise.all([getDashboardDaily('global'), sleep(delay)])
    .then(([r]) => { daily.value = r.data.daily })
    .finally(() => { loadingDaily.value = false })

  Promise.all([getDashboardDailyTokens('global'), sleep(delay)])
    .then(([r]) => { dailyTokens.value = r.data.dailyTokens })
    .finally(() => { loadingDailyTokens.value = false })

  Promise.all([getDashboardByProvider('global'), sleep(delay)])
    .then(([r]) => { byProvider.value = r.data.byProvider })
    .finally(() => { loadingProviders.value = false })

  Promise.all([getDashboardByModel('global'), sleep(delay)])
    .then(([r]) => { byModel.value = r.data.byModel })
    .finally(() => { loadingModels.value = false })

  Promise.all([getDashboardTokensByModel('global'), sleep(delay)])
    .then(([r]) => { tokensByModel.value = r.data.tokensByModel })
    .finally(() => { loadingTokensByModel.value = false })

  Promise.all([getDashboardToolsUsed('global'), sleep(delay)])
    .then(([r]) => { toolsUsed.value = r.data.toolsUsed })
    .finally(() => { loadingTools.value = false })

  Promise.all([getDashboardLastRequest('global'), sleep(delay)])
    .then(([r]) => { lastRequest.value = r.data.lastRequest })
    .finally(() => { loadingLastRequest.value = false })

  Promise.all([getDashboardActiveUsers(), sleep(delay)])
    .then(([r]) => { activeUsers.value = r.data.activeUsers })
    .catch(() => { activeUsers.value = undefined })
    .finally(() => { loadingActiveUsers.value = false })

  Promise.all([getStatus(), sleep(delay)])
    .then(([r]) => { status.value = r.data })
    .catch(e => {
      if (e?.response?.data?.services) status.value = e.response.data
    })
    .finally(() => {
      loadingStatus.value = false
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
  <div class="overview">
    <div class="overview-toolbar">
      <span v-if="lastChecked" class="last-checked">
        <span v-if="refreshing" class="spin-icon" />
        Last updated: {{ formatTime(lastChecked) }}
      </span>
    </div>

    <StatGrid
      :total-requests="totalRequests"
      :tokens="tokens"
      :provider-count="byProvider.length"
      :active-users="activeUsers"
      :show-active-users="true"
      :last-request="lastRequest"
      :loading-requests="loadingRequests"
      :loading-tokens="loadingTokens"
      :loading-providers="loadingProviders"
      :loading-active-users="loadingActiveUsers"
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

    <ServiceStatus :status="status" :refreshing="refreshing" :loading="loadingStatus" />
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

.charts-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}
@media (max-width: 700px) { .charts-row { grid-template-columns: 1fr; } }
</style>
