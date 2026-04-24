<script setup lang="ts">
import { computed } from 'vue'

interface ModelTokens { model: string; total: number }
const props = defineProps<{ models: ModelTokens[]; loading?: boolean }>()

const grandTotal = computed(() => props.models.reduce((s, m) => s + m.total, 0))

const PALETTE = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4', '#ec4899', '#f97316']
function color(index: number) { return PALETTE[index % PALETTE.length] }

function fmtNum(n: number) {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M'
  if (n >= 1_000)     return (n / 1_000).toFixed(1) + 'K'
  return String(n)
}
</script>

<template>
  <div class="chart-card">
    <div class="chart-header">
      <h2 class="chart-title">Tokens by Model</h2>
      <span class="top-badge">Top 8</span>
    </div>
    <div v-if="loading" class="list-skeleton"><div v-for="i in 3" :key="i" class="list-skeleton-row" /></div>
    <div v-else-if="models.length === 0" class="no-data">No token usage recorded yet.</div>
    <div v-else class="model-list">
      <div v-for="(m, i) in models" :key="m.model" class="model-row">
        <div class="model-meta">
          <span class="model-dot" :style="{ background: color(i) }" />
          <span class="model-name">{{ m.model }}</span>
          <span class="model-count">{{ fmtNum(m.total) }}</span>
          <span class="model-pct">{{ grandTotal > 0 ? Math.round(m.total / grandTotal * 100) : 0 }}%</span>
        </div>
        <div class="model-bar-track">
          <div
            class="model-bar-fill"
            :style="{ width: (grandTotal > 0 ? m.total / grandTotal * 100 : 0) + '%', background: color(i) }"
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
.model-list { display: flex; flex-direction: column; gap: 0.85rem; }
.model-row  { display: flex; flex-direction: column; gap: 0.3rem; }
.model-meta { display: flex; align-items: center; gap: 0.5rem; font-size: 0.85rem; }
.model-dot  { width: 9px; height: 9px; border-radius: 50%; flex-shrink: 0; }
.model-name { flex: 1; font-weight: 600; color: #1e293b; font-family: monospace; font-size: 0.82rem; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 160px; }
.model-count { color: #475569; }
.model-pct   { color: #94a3b8; font-size: 0.78rem; }
.model-bar-track { height: 6px; background: #f1f5f9; border-radius: 999px; overflow: hidden; }
.model-bar-fill  { height: 100%; border-radius: 999px; transition: width 0.4s ease; }

.list-skeleton { display: flex; flex-direction: column; gap: 0.85rem; }
.list-skeleton-row {
  height: 36px; border-radius: 6px;
  background: linear-gradient(90deg, #f1f5f9 25%, #e2e8f0 50%, #f1f5f9 75%);
  background-size: 200% 100%;
  animation: shimmer 1.4s infinite;
}
.list-skeleton-row:nth-child(2) { animation-delay: 0.1s; }
.list-skeleton-row:nth-child(3) { animation-delay: 0.2s; }
@keyframes shimmer {
  0%   { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
</style>
