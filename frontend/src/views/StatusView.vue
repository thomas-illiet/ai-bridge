<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { getStatus, type StatusResponse, type ServiceStatus } from '@/services/api'

const data = ref<StatusResponse | null>(null)
const loading = ref(true)
const refreshing = ref(false)
const error = ref<string | null>(null)
const lastChecked = ref<Date | null>(null)

let interval: ReturnType<typeof setInterval>

async function fetchStatus() {
  if (data.value) refreshing.value = true
  try {
    const res = await getStatus()
    data.value = res.data
    error.value = null
  } catch (e: any) {
    if (e?.response?.data?.services) {
      data.value = e.response.data
      error.value = null
    } else {
      error.value = 'Failed to reach the backend'
    }
  } finally {
    loading.value = false
    refreshing.value = false
    lastChecked.value = new Date()
  }
}

onMounted(() => {
  fetchStatus()
  interval = setInterval(fetchStatus, 30_000)
})

onUnmounted(() => clearInterval(interval))

const SERVICE_LABELS: Record<string, string> = {
  database: 'PostgreSQL',
  keycloak: 'Keycloak',
  openai:   'OpenAI',
  ollama:   'Ollama',
}

function label(s: ServiceStatus) {
  return SERVICE_LABELS[s.name] ?? s.name
}

function statusClass(s: ServiceStatus) {
  return {
    'indicator-up':       s.status === 'up',
    'indicator-down':     s.status === 'down',
    'indicator-disabled': s.status === 'disabled',
  }
}

function statusText(s: ServiceStatus) {
  if (s.status === 'up')   return 'Operational'
  if (s.status === 'down') return 'Unavailable'
  return 'Not configured'
}

function formatTime(d: Date) {
  return d.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}
</script>

<template>
  <div class="status-page">
    <div class="page-header">
      <h1>System Status</h1>
      <span v-if="lastChecked" class="last-checked">
        <span v-if="refreshing" class="spin-icon" />
        Last checked: {{ formatTime(lastChecked) }}
      </span>
    </div>

    <div v-if="loading" class="skeleton-grid">
      <div v-for="i in 5" :key="i" class="skeleton-card">
        <div class="skeleton-line skeleton-title" />
        <div class="skeleton-line skeleton-badge" />
      </div>
    </div>

    <div v-else-if="error" class="state-msg error">{{ error }}</div>

    <template v-else-if="data">
      <div class="overall-banner" :class="data.status === 'healthy' ? 'banner-healthy' : 'banner-degraded'">
        <span class="overall-dot" />
        {{ data.status === 'healthy' ? 'All systems operational' : 'Some services are unavailable' }}
      </div>

      <div class="service-grid">
        <div
          v-for="svc in data.services"
          :key="svc.name"
          class="service-card"
          :class="{ 'card-refreshing': refreshing }"
        >
          <div class="service-header">
            <span class="service-name">{{ label(svc) }}</span>
            <span class="indicator" :class="statusClass(svc)">
              <span v-if="svc.status === 'up'" class="pulse-dot" />
              {{ statusText(svc) }}
            </span>
          </div>
          <p v-if="svc.message" class="service-message">{{ svc.message }}</p>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.status-page {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.page-header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
}

h1 { font-size: 1.75rem; font-weight: 700; }

.last-checked {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.8rem;
  color: #94a3b8;
}

/* spinning refresh icon */
.spin-icon {
  display: inline-block;
  width: 10px;
  height: 10px;
  border: 2px solid #cbd5e1;
  border-top-color: #64748b;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* skeleton loading cards */
.skeleton-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 1rem;
}

.skeleton-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 1rem 1.25rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}

.skeleton-line {
  border-radius: 4px;
  background: linear-gradient(90deg, #f1f5f9 25%, #e2e8f0 50%, #f1f5f9 75%);
  background-size: 200% 100%;
  animation: shimmer 1.4s infinite;
}

.skeleton-title  { width: 90px; height: 14px; }
.skeleton-badge  { width: 70px; height: 20px; border-radius: 999px; }

@keyframes shimmer {
  0%   { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

/* overall banner */
.overall-banner {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem 1.25rem;
  border-radius: 10px;
  font-weight: 600;
  font-size: 0.95rem;
}

.banner-healthy  { background: #f0fdf4; color: #166534; border: 1px solid #bbf7d0; }
.banner-degraded { background: #fff7ed; color: #9a3412; border: 1px solid #fed7aa; }

.overall-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: currentColor;
  flex-shrink: 0;
}

/* service grid */
.service-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 1rem;
}

.service-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 1rem 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  transition: opacity 0.2s;
}

.card-refreshing { opacity: 0.6; }

.service-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.service-name {
  font-weight: 600;
  font-size: 0.95rem;
  color: #1e293b;
}

.indicator {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.2rem 0.6rem;
  border-radius: 999px;
}

.indicator-up       { background: #dcfce7; color: #166534; }
.indicator-down     { background: #fee2e2; color: #991b1b; }
.indicator-disabled { background: #f1f5f9; color: #64748b; }

/* pulsing green dot for operational services */
.pulse-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #16a34a;
  animation: pulse 2s ease-in-out infinite;
  flex-shrink: 0;
}

@keyframes pulse {
  0%, 100% { opacity: 1;   transform: scale(1); }
  50%       { opacity: 0.4; transform: scale(0.85); }
}

.service-message {
  font-size: 0.8rem;
  color: #ef4444;
  margin: 0;
  word-break: break-word;
}

.state-msg { color: #64748b; }
.state-msg.error { color: #ef4444; }
</style>
