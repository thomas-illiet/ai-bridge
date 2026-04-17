<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { getDashboard } from '@/services/api'
import axios from 'axios'
import type { StatusResponse, ServiceStatus } from '@/services/api'

// ── types ──────────────────────────────────────────────────────────────────
interface DailyCount   { date: string; count: number }
interface ProviderCount { provider: string; count: number }
interface TokenTotals  { totalInput: number; totalOutput: number }

interface DashboardData {
  user:          string
  totalRequests: number
  tokens:        TokenTotals
  daily:         DailyCount[]
  byProvider:    ProviderCount[]
}

// ── state ──────────────────────────────────────────────────────────────────
const data      = ref<DashboardData | null>(null)
const status    = ref<StatusResponse | null>(null)
const loading   = ref(true)
const refreshing = ref(false)
const error     = ref<string | null>(null)
const lastChecked = ref<Date | null>(null)

let interval: ReturnType<typeof setInterval>

// ── data fetching ──────────────────────────────────────────────────────────
async function fetchAll() {
  if (data.value) refreshing.value = true
  try {
    const [dashRes, statusRes] = await Promise.allSettled([
      getDashboard(),
      axios.get<StatusResponse>('/api/status'),
    ])
    if (dashRes.status === 'fulfilled') data.value  = dashRes.value.data
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

onMounted(() => {
  fetchAll()
  interval = setInterval(fetchAll, 30_000)
})
onUnmounted(() => clearInterval(interval))

// ── status helpers ─────────────────────────────────────────────────────────
const SERVICE_LABELS: Record<string, string> = {
  database: 'PostgreSQL',
  keycloak: 'Keycloak',
  openai:   'OpenAI',
  ollama:   'Ollama',
}
function svcLabel(s: ServiceStatus) { return SERVICE_LABELS[s.name] ?? s.name }
function svcClass(s: ServiceStatus) {
  return {
    'indicator-up':       s.status === 'up',
    'indicator-down':     s.status === 'down',
    'indicator-disabled': s.status === 'disabled',
  }
}
function svcText(s: ServiceStatus) {
  if (s.status === 'up')   return 'Operational'
  if (s.status === 'down') return 'Unavailable'
  return 'Not configured'
}
function formatTime(d: Date) {
  return d.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

// ── chart helpers ──────────────────────────────────────────────────────────
function fmtNum(n: number) {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M'
  if (n >= 1_000)     return (n / 1_000).toFixed(1) + 'K'
  return String(n)
}

// Fill gaps so we always show 7 days
const chartDays = computed<DailyCount[]>(() => {
  const map = new Map((data.value?.daily ?? []).map(d => [d.date, d.count]))
  const days: DailyCount[] = []
  for (let i = 6; i >= 0; i--) {
    const d = new Date()
    d.setDate(d.getDate() - i)
    const key = d.toISOString().slice(0, 10)
    days.push({ date: key, count: map.get(key) ?? 0 })
  }
  return days
})

const barMax = computed(() => Math.max(1, ...chartDays.value.map(d => d.count)))

function barHeight(count: number) {
  return Math.round((count / barMax.value) * 120)
}

function shortDate(iso: string) {
  const d = new Date(iso + 'T00:00:00')
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}

const providerTotal = computed(() =>
  (data.value?.byProvider ?? []).reduce((s, p) => s + p.count, 0)
)

const PROVIDER_COLORS: Record<string, string> = {
  openai: '#10b981',
  ollama: '#6366f1',
}
function providerColor(name: string) {
  return PROVIDER_COLORS[name] ?? '#94a3b8'
}
</script>

<template>
  <div class="dashboard">
    <!-- header -->
    <div class="page-header">
      <div>
        <h1>Dashboard</h1>
        <p v-if="data" class="subtitle">Welcome back, <strong>{{ data.user }}</strong></p>
      </div>
      <span v-if="lastChecked" class="last-checked">
        <span v-if="refreshing" class="spin-icon" />
        Last updated: {{ formatTime(lastChecked) }}
      </span>
    </div>

    <div v-if="loading" class="skeleton-section">
      <div class="skeleton-grid-4">
        <div v-for="i in 4" :key="i" class="skeleton-card tall" />
      </div>
      <div class="skeleton-grid-2">
        <div class="skeleton-card chart" />
        <div class="skeleton-card chart" />
      </div>
    </div>

    <template v-else-if="data">
      <!-- ── stat cards ── -->
      <div class="stat-grid">
        <div class="stat-card">
          <span class="stat-label">Total Requests</span>
          <span class="stat-value">{{ fmtNum(data.totalRequests) }}</span>
        </div>
        <div class="stat-card">
          <span class="stat-label">Input Tokens</span>
          <span class="stat-value">{{ fmtNum(data.tokens.totalInput) }}</span>
        </div>
        <div class="stat-card">
          <span class="stat-label">Output Tokens</span>
          <span class="stat-value">{{ fmtNum(data.tokens.totalOutput) }}</span>
        </div>
        <div class="stat-card">
          <span class="stat-label">Active Providers</span>
          <span class="stat-value">{{ data.byProvider.length }}</span>
        </div>
      </div>

      <!-- ── charts row ── -->
      <div class="charts-row">
        <!-- bar chart: daily requests -->
        <div class="chart-card">
          <h2 class="chart-title">Requests — Last 7 Days</h2>
          <div class="bar-chart">
            <div
              v-for="day in chartDays"
              :key="day.date"
              class="bar-col"
            >
              <span class="bar-value">{{ day.count > 0 ? fmtNum(day.count) : '' }}</span>
              <div
                class="bar"
                :style="{ height: barHeight(day.count) + 'px' }"
                :class="{ 'bar-zero': day.count === 0 }"
              />
              <span class="bar-label">{{ shortDate(day.date) }}</span>
            </div>
          </div>
        </div>

        <!-- provider breakdown -->
        <div class="chart-card">
          <h2 class="chart-title">Requests by Provider</h2>
          <div v-if="data.byProvider.length === 0" class="no-data">No requests recorded yet.</div>
          <div v-else class="provider-list">
            <div
              v-for="p in data.byProvider"
              :key="p.provider"
              class="provider-row"
            >
              <div class="provider-meta">
                <span class="provider-dot" :style="{ background: providerColor(p.provider) }" />
                <span class="provider-name">{{ p.provider }}</span>
                <span class="provider-count">{{ fmtNum(p.count) }}</span>
                <span class="provider-pct">
                  {{ providerTotal > 0 ? Math.round(p.count / providerTotal * 100) : 0 }}%
                </span>
              </div>
              <div class="provider-bar-track">
                <div
                  class="provider-bar-fill"
                  :style="{
                    width: (providerTotal > 0 ? p.count / providerTotal * 100 : 0) + '%',
                    background: providerColor(p.provider),
                  }"
                />
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- ── service status ── -->
      <div class="section-header">
        <h2>Service Status</h2>
        <span v-if="status" class="overall-pill" :class="status.status === 'healthy' ? 'pill-healthy' : 'pill-degraded'">
          {{ status.status === 'healthy' ? 'All operational' : 'Degraded' }}
        </span>
      </div>

      <div v-if="status" class="service-grid">
        <div
          v-for="svc in status.services"
          :key="svc.name"
          class="service-card"
          :class="{ 'card-refreshing': refreshing }"
        >
          <div class="service-header">
            <span class="service-name">{{ svcLabel(svc) }}</span>
            <span class="indicator" :class="svcClass(svc)">
              <span v-if="svc.status === 'up'" class="pulse-dot" />
              {{ svcText(svc) }}
            </span>
          </div>
          <p v-if="svc.message" class="service-message">{{ svc.message }}</p>
        </div>
      </div>
      <p v-else class="muted">Unable to reach status endpoint.</p>
    </template>

    <div v-else-if="error" class="state-msg error">{{ error }}</div>
  </div>
</template>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

/* header */
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 0.5rem;
}
h1 { font-size: 1.75rem; font-weight: 700; margin: 0; }
.subtitle { color: #64748b; font-size: 0.9rem; margin: 0.2rem 0 0; }

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

/* skeleton */
.skeleton-grid-4 {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: 1rem;
}
.skeleton-grid-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}
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

/* stat cards */
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

/* charts row */
.charts-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}
@media (max-width: 700px) { .charts-row { grid-template-columns: 1fr; } }

.chart-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.chart-title { font-size: 0.9rem; font-weight: 700; color: #1e293b; margin: 0; }

/* bar chart */
.bar-chart {
  display: flex;
  align-items: flex-end;
  gap: 6px;
  height: 155px;
  padding-bottom: 24px;
  position: relative;
}
.bar-col {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-end;
  gap: 4px;
  height: 100%;
  position: relative;
}
.bar-value {
  font-size: 0.65rem;
  color: #64748b;
  height: 14px;
  line-height: 14px;
}
.bar {
  width: 100%;
  background: #3b82f6;
  border-radius: 4px 4px 0 0;
  min-height: 3px;
  transition: height 0.3s ease;
}
.bar.bar-zero { background: #e2e8f0; min-height: 3px; }
.bar-label {
  position: absolute;
  bottom: -20px;
  font-size: 0.62rem;
  color: #94a3b8;
  white-space: nowrap;
}

/* provider breakdown */
.no-data { font-size: 0.85rem; color: #94a3b8; }
.provider-list { display: flex; flex-direction: column; gap: 0.85rem; }
.provider-row  { display: flex; flex-direction: column; gap: 0.3rem; }
.provider-meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.85rem;
}
.provider-dot  { width: 9px; height: 9px; border-radius: 50%; flex-shrink: 0; }
.provider-name { flex: 1; font-weight: 600; color: #1e293b; text-transform: capitalize; }
.provider-count { color: #475569; }
.provider-pct   { color: #94a3b8; font-size: 0.78rem; }
.provider-bar-track {
  height: 6px;
  background: #f1f5f9;
  border-radius: 999px;
  overflow: hidden;
}
.provider-bar-fill {
  height: 100%;
  border-radius: 999px;
  transition: width 0.4s ease;
}

/* section header */
.section-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}
.section-header h2 { font-size: 1.1rem; font-weight: 700; margin: 0; }

.overall-pill {
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.2rem 0.65rem;
  border-radius: 999px;
}
.pill-healthy  { background: #dcfce7; color: #166534; }
.pill-degraded { background: #fee2e2; color: #991b1b; }

/* service grid */
.service-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 1rem;
}
.service-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 0.9rem 1.1rem;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  transition: opacity 0.2s;
}
.card-refreshing { opacity: 0.6; }
.service-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.service-name { font-weight: 600; font-size: 0.9rem; color: #1e293b; }
.indicator {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.72rem;
  font-weight: 600;
  padding: 0.18rem 0.55rem;
  border-radius: 999px;
}
.indicator-up       { background: #dcfce7; color: #166534; }
.indicator-down     { background: #fee2e2; color: #991b1b; }
.indicator-disabled { background: #f1f5f9; color: #64748b; }
.pulse-dot {
  width: 7px; height: 7px;
  border-radius: 50%;
  background: #16a34a;
  animation: pulse 2s ease-in-out infinite;
  flex-shrink: 0;
}
@keyframes pulse {
  0%, 100% { opacity: 1;   transform: scale(1); }
  50%       { opacity: 0.4; transform: scale(0.85); }
}
.service-message { font-size: 0.78rem; color: #ef4444; margin: 0; word-break: break-word; }

.muted { color: #94a3b8; font-size: 0.85rem; }
.state-msg { color: #64748b; }
.state-msg.error { color: #ef4444; }
</style>
