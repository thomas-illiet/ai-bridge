<script setup lang="ts">
import { computed } from 'vue'

interface ToolCount { tool: string; count: number }
const props = defineProps<{ tools: ToolCount[]; loading?: boolean }>()

const total = computed(() => props.tools.reduce((s, t) => s + t.count, 0))

const PALETTE = ['#6366f1', '#10b981', '#f59e0b', '#ef4444', '#3b82f6', '#06b6d4', '#ec4899', '#f97316']
function color(index: number) { return PALETTE[index % PALETTE.length] }

function fmtNum(n: number) {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M'
  if (n >= 1_000)     return (n / 1_000).toFixed(1) + 'K'
  return String(n)
}

function shortName(tool: string) {
  const parts = tool.split(/[._/-]/)
  return parts[parts.length - 1] || tool
}
</script>

<template>
  <div class="chart-card">
    <div class="chart-header">
      <h2 class="chart-title">Tools Used</h2>
      <span class="top-badge">Top 8</span>
    </div>
    <div v-if="loading" class="list-skeleton"><div v-for="i in 3" :key="i" class="list-skeleton-row" /></div>
    <div v-else-if="tools.length === 0" class="no-data">No tool calls recorded yet.</div>
    <div v-else class="tool-list">
      <div v-for="(t, i) in tools" :key="t.tool" class="tool-row">
        <div class="tool-meta">
          <span class="tool-icon" :style="{ background: color(i) + '22', color: color(i) }">
            <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/>
            </svg>
          </span>
          <span class="tool-name" :title="t.tool">{{ shortName(t.tool) }}</span>
          <span class="tool-count">{{ fmtNum(t.count) }}</span>
          <span class="tool-pct">{{ total > 0 ? Math.round(t.count / total * 100) : 0 }}%</span>
        </div>
        <div class="tool-bar-track">
          <div
            class="tool-bar-fill"
            :style="{ width: (total > 0 ? t.count / total * 100 : 0) + '%', background: color(i) }"
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
.tool-list { display: flex; flex-direction: column; gap: 0.85rem; }
.tool-row  { display: flex; flex-direction: column; gap: 0.3rem; }
.tool-meta { display: flex; align-items: center; gap: 0.5rem; font-size: 0.85rem; }
.tool-icon {
  width: 22px; height: 22px;
  border-radius: 5px;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}
.tool-name  { flex: 1; font-weight: 600; color: #1e293b; font-family: monospace; font-size: 0.82rem; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 160px; }
.tool-count { color: #475569; }
.tool-pct   { color: #94a3b8; font-size: 0.78rem; }
.tool-bar-track { height: 6px; background: #f1f5f9; border-radius: 999px; overflow: hidden; }
.tool-bar-fill  { height: 100%; border-radius: 999px; transition: width 0.4s ease; }

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
