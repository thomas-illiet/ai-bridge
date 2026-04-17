<script setup lang="ts">
import type { StatusResponse, ServiceStatus } from '@/services/api'

defineProps<{ status: StatusResponse | null; refreshing: boolean }>()

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
</script>

<template>
  <div>
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
  gap: 0.35rem;
  transition: opacity 0.2s;
}
.card-refreshing { opacity: 0.6; }
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
</style>
