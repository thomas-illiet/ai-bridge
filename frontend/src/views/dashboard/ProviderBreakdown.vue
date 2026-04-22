<script setup lang="ts">
import { computed } from 'vue'

interface ProviderCount { provider: string; count: number }
const props = defineProps<{ providers: ProviderCount[] }>()

const total = computed(() => props.providers.reduce((s, p) => s + p.count, 0))

const COLORS: Record<string, string> = { openai: '#10b981', ollama: '#6366f1', anthropic: '#f59e0b' }
function color(name: string) { return COLORS[name] ?? '#94a3b8' }

function fmtNum(n: number) {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M'
  if (n >= 1_000)     return (n / 1_000).toFixed(1) + 'K'
  return String(n)
}
</script>

<template>
  <div class="chart-card">
    <div class="chart-header">
      <h2 class="chart-title">Requests by Provider</h2>
      <span class="top-badge">Top 5</span>
    </div>
    <div v-if="providers.length === 0" class="no-data">No requests recorded yet.</div>
    <div v-else class="provider-list">
      <div v-for="p in providers" :key="p.provider" class="provider-row">
        <div class="provider-meta">
          <span class="provider-dot" :style="{ background: color(p.provider) }" />
          <span class="provider-name">{{ p.provider }}</span>
          <span class="provider-count">{{ fmtNum(p.count) }}</span>
          <span class="provider-pct">{{ total > 0 ? Math.round(p.count / total * 100) : 0 }}%</span>
        </div>
        <div class="provider-bar-track">
          <div
            class="provider-bar-fill"
            :style="{ width: (total > 0 ? p.count / total * 100 : 0) + '%', background: color(p.provider) }"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.chart-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.chart-header { display: flex; align-items: center; justify-content: space-between; }
.chart-title { font-size: 0.9rem; font-weight: 700; color: #1e293b; margin: 0; }
.top-badge { font-size: 0.7rem; font-weight: 700; background: #f1f5f9; color: #64748b; padding: 0.15rem 0.5rem; border-radius: 999px; }
.no-data { font-size: 0.85rem; color: #94a3b8; }
.provider-list { display: flex; flex-direction: column; gap: 0.85rem; }
.provider-row  { display: flex; flex-direction: column; gap: 0.3rem; }
.provider-meta { display: flex; align-items: center; gap: 0.5rem; font-size: 0.85rem; }
.provider-dot  { width: 9px; height: 9px; border-radius: 50%; flex-shrink: 0; }
.provider-name { flex: 1; font-weight: 600; color: #1e293b; text-transform: capitalize; }
.provider-count { color: #475569; }
.provider-pct   { color: #94a3b8; font-size: 0.78rem; }
.provider-bar-track { height: 6px; background: #f1f5f9; border-radius: 999px; overflow: hidden; }
.provider-bar-fill  { height: 100%; border-radius: 999px; transition: width 0.4s ease; }
</style>
