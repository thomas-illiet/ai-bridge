<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { getStatus, type StatusResponse, type ServiceStatus } from '@/services/api'

const data = ref<StatusResponse | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)
const lastChecked = ref<Date | null>(null)

let interval: ReturnType<typeof setInterval>

async function fetchStatus() {
  try {
    const res = await getStatus()
    data.value = res.data
    error.value = null
  } catch {
    error.value = 'Failed to reach the backend'
  } finally {
    loading.value = false
    lastChecked.value = new Date()
  }
}

onMounted(() => {
  fetchStatus()
  interval = setInterval(fetchStatus, 30_000)
})

onUnmounted(() => clearInterval(interval))

const SERVICE_LABELS: Record<string, string> = {
  database:  'PostgreSQL',
  keycloak:  'Keycloak',
  anthropic: 'Anthropic',
  openai:    'OpenAI',
  ollama:    'Ollama',
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
  if (s.status === 'up')       return 'Operational'
  if (s.status === 'down')     return 'Unavailable'
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
        Last checked: {{ formatTime(lastChecked) }}
      </span>
    </div>

    <div v-if="loading" class="state-msg">Checking services…</div>
    <div v-else-if="error" class="state-msg error">{{ error }}</div>

    <template v-else-if="data">
      <div class="overall-banner" :class="data.status === 'healthy' ? 'banner-healthy' : 'banner-degraded'">
        <span class="overall-dot" />
        {{ data.status === 'healthy' ? 'All systems operational' : 'Some services are unavailable' }}
      </div>

      <div class="service-grid">
        <div v-for="svc in data.services" :key="svc.name" class="service-card">
          <div class="service-header">
            <span class="service-name">{{ label(svc) }}</span>
            <span class="indicator" :class="statusClass(svc)">{{ statusText(svc) }}</span>
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

h1 {
  font-size: 1.75rem;
  font-weight: 700;
}

.last-checked {
  font-size: 0.8rem;
  color: #94a3b8;
}

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
}

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
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.2rem 0.6rem;
  border-radius: 999px;
}

.indicator-up       { background: #dcfce7; color: #166534; }
.indicator-down     { background: #fee2e2; color: #991b1b; }
.indicator-disabled { background: #f1f5f9; color: #64748b; }

.service-message {
  font-size: 0.8rem;
  color: #ef4444;
  margin: 0;
  word-break: break-word;
}

.state-msg { color: #64748b; }
.state-msg.error { color: #ef4444; }
</style>
