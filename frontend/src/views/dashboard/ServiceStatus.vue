<script setup lang="ts">
import type { StatusResponse, ServiceStatus } from '@/services/api'

defineProps<{ status: StatusResponse | null; refreshing: boolean; loading?: boolean }>()

const SERVICE_LABELS: Record<string, string> = {
  database: 'PostgreSQL',
  oidc:     'OIDC Provider',
  openai:   'OpenAI',
  ollama:   'Ollama',
}

const SERVICE_DESCRIPTIONS: Record<string, string> = {
  database: 'Primary data store',
  oidc:     'Authentication provider',
  openai:   'Cloud AI inference',
  ollama:   'Local AI inference',
}

function svcLabel(s: ServiceStatus) { return SERVICE_LABELS[s.name] ?? s.name }
function svcDesc(s: ServiceStatus)  { return SERVICE_DESCRIPTIONS[s.name] ?? '' }
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
function svcLatency(s: ServiceStatus) {
  if (s.latency_ms == null || s.status !== 'up') return null
  return s.latency_ms < 1 ? '<1 ms' : `${s.latency_ms} ms`
}
function svcModelCount(s: ServiceStatus) {
  if (s.model_count == null || s.status !== 'up') return null
  return `${s.model_count} model${s.model_count !== 1 ? 's' : ''}`
}
</script>

<template>
  <div>
    <div class="section-header">
      <h2>Service Status</h2>
      <span v-if="status" class="overall-pill" :class="status.status === 'healthy' ? 'pill-healthy' : 'pill-degraded'">
        {{ status.status === 'healthy' ? 'All operational' : 'Degraded' }}
      </span>
    </div>

    <div v-if="loading" class="service-grid">
      <div v-for="i in 4" :key="i" class="service-skeleton" />
    </div>
    <div v-else-if="status" class="service-grid">
      <div
        v-for="svc in status.services"
        :key="svc.name"
        class="service-card"
        :class="{ 'card-refreshing': refreshing, 'card-down': svc.status === 'down' }"
      >
        <div class="service-header">
          <span class="service-name">{{ svcLabel(svc) }}</span>
          <span class="indicator" :class="svcClass(svc)">
            <span v-if="svc.status === 'up'" class="pulse-dot" />
            {{ svcText(svc) }}
          </span>
        </div>
        <p class="service-desc">{{ svcDesc(svc) }}</p>
        <div class="service-meta">
          <span v-if="svcLatency(svc)" class="meta-tag">
            <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
            {{ svcLatency(svc) }}
          </span>
          <span v-if="svcModelCount(svc)" class="meta-tag">
            <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>
            {{ svcModelCount(svc) }}
          </span>
          <span v-if="svc.message" class="meta-error">{{ svc.message }}</span>
        </div>
      </div>
    </div>
    <p v-else class="muted">Unable to reach status endpoint.</p>
  </div>
</template>

<style scoped>
.section-header { display: flex; align-items: center; gap: 0.75rem; }
.section-header h2 { font-size: 1.1rem; font-weight: 700; margin: 0; }

.overall-pill { font-size: 0.75rem; font-weight: 600; padding: 0.2rem 0.65rem; border-radius: 999px; }
.pill-healthy  { background: #dcfce7; color: #166534; }
.pill-degraded { background: #fee2e2; color: #991b1b; }

.service-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 1rem;
  margin-top: 1rem;
}
.service-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 0.9rem 1.1rem;
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  transition: opacity 0.2s;
}
.card-refreshing { opacity: 0.6; }
.card-down { border-color: #fecaca; background: #fff8f8; }

.service-header { display: flex; align-items: center; justify-content: space-between; }
.service-name { font-weight: 600; font-size: 0.9rem; color: #1e293b; }

.indicator {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.72rem;
  font-weight: 600;
  padding: 0.18rem 0.55rem;
  border-radius: 999px;
  white-space: nowrap;
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

.service-desc { font-size: 0.78rem; color: #94a3b8; margin: 0; }

.service-meta { display: flex; align-items: center; flex-wrap: wrap; gap: 0.4rem; margin-top: 0.15rem; }

.meta-tag {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.72rem;
  color: #64748b;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 4px;
  padding: 0.1rem 0.4rem;
}

.meta-error { font-size: 0.72rem; color: #ef4444; word-break: break-word; }

.muted { color: #94a3b8; font-size: 0.85rem; }

.service-skeleton {
  height: 88px; border-radius: 10px;
  background: linear-gradient(90deg, #f1f5f9 25%, #e2e8f0 50%, #f1f5f9 75%);
  background-size: 200% 100%;
  animation: shimmer 1.4s infinite;
}
.service-skeleton:nth-child(2) { animation-delay: 0.1s; }
.service-skeleton:nth-child(3) { animation-delay: 0.2s; }
.service-skeleton:nth-child(4) { animation-delay: 0.3s; }
@keyframes shimmer {
  0%   { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
</style>
