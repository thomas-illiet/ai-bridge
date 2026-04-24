<script setup lang="ts">
import { computed } from 'vue'

interface DailyCount { date: string; count: number }
const props = defineProps<{ days: DailyCount[]; loading?: boolean }>()

const barMax = computed(() => Math.max(1, ...props.days.map(d => d.count)))

function barHeight(count: number) { return Math.round((count / barMax.value) * 120) }

function fmtNum(n: number) {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M'
  if (n >= 1_000)     return (n / 1_000).toFixed(1) + 'K'
  return String(n)
}

function shortDate(iso: string) {
  const d = new Date(iso + 'T00:00:00')
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}
</script>

<template>
  <div class="chart-card">
    <h2 class="chart-title">Requests — Last 7 Days</h2>
    <div v-if="loading" class="chart-skeleton" />
    <div v-else class="bar-chart">
      <div v-for="day in days" :key="day.date" class="bar-col">
        <span class="bar-value">{{ day.count > 0 ? fmtNum(day.count) : '' }}</span>
        <div class="bar" :style="{ height: barHeight(day.count) + 'px' }" :class="{ 'bar-zero': day.count === 0 }" />
        <span class="bar-label">{{ shortDate(day.date) }}</span>
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
.chart-title { font-size: 0.9rem; font-weight: 700; color: #1e293b; margin: 0; }
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
.bar-value { font-size: 0.65rem; color: #64748b; height: 14px; line-height: 14px; }
.bar {
  width: 100%;
  background: #3b82f6;
  border-radius: 4px 4px 0 0;
  min-height: 3px;
  transition: height 0.3s ease;
}
.bar.bar-zero { background: #e2e8f0; }
.bar-label { position: absolute; bottom: -20px; font-size: 0.62rem; color: #94a3b8; white-space: nowrap; }

.chart-skeleton {
  height: 131px;
  border-radius: 6px;
  background: linear-gradient(90deg, #f1f5f9 25%, #e2e8f0 50%, #f1f5f9 75%);
  background-size: 200% 100%;
  animation: shimmer 1.4s infinite;
}
@keyframes shimmer {
  0%   { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
</style>
