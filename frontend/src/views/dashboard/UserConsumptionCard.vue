<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { listUsers } from '@/services/api'
import { fmtNum } from '@/utils/format'

interface UserRow {
  id: string
  username: string
  role: string
  totalRequests: number
  totalInput: number
  totalOutput: number
}

type SortKey = 'totalRequests' | 'totalInput' | 'totalOutput'

const LIMIT_OPTIONS = [10, 15, 25, 50] as const
type LimitOption = typeof LIMIT_OPTIONS[number]

const METRICS: { key: SortKey; label: string; color: string }[] = [
  { key: 'totalRequests', label: 'Requests', color: '#3b82f6' },
  { key: 'totalInput',    label: 'Input',    color: '#8b5cf6' },
  { key: 'totalOutput',   label: 'Output',   color: '#10b981' },
]

const users   = ref<UserRow[]>([])
const loading = ref(true)
const error   = ref<string | null>(null)
const sortBy  = ref<SortKey>('totalRequests')
const limit   = ref<LimitOption>(15)

onMounted(async () => {
  try {
    const res = await listUsers('total_requests', 'desc')
    users.value = (res.data.users ?? []).map((u: any) => ({
      id:            String(u.id ?? ''),
      username:      String(u.username ?? ''),
      role:          String(u.role ?? ''),
      totalRequests: Number(u.totalRequests) || 0,
      totalInput:    Number(u.totalInput)    || 0,
      totalOutput:   Number(u.totalOutput)   || 0,
    }))
  } catch (e: any) {
    error.value = e?.response?.data?.error ?? 'Failed to load user data'
  } finally {
    loading.value = false
  }
})

const activeMetric = computed(() => METRICS.find(m => m.key === sortBy.value)!)

const sortedRows = computed(() => {
  const sorted = [...users.value].sort((a, b) => b[sortBy.value] - a[sortBy.value])
  const top    = sorted.slice(0, limit.value)
  const maxVal = top.reduce((m, r) => (r[sortBy.value] > m ? r[sortBy.value] : m), 0)

  return top.map(r => ({
    ...r,
    barHeight: maxVal > 0 ? Math.max(3, Math.round((r[sortBy.value] / maxVal) * 120)) : 3,
    value:     r[sortBy.value],
  }))
})
</script>

<template>
  <div class="chart-card">
    <!-- header -->
    <div class="chart-header">
      <div class="title-group">
        <h2 class="chart-title">User Consumption</h2>
        <span v-if="!loading" class="count-badge">
          {{ Math.min(users.length, limit) }} / {{ users.length }}
        </span>
      </div>
      <div class="header-right">
        <div class="metric-tabs">
          <button
            v-for="m in METRICS"
            :key="m.key"
            class="metric-tab"
            :class="{ active: sortBy === m.key }"
            :style="sortBy === m.key ? { '--c': m.color } : {}"
            @click="sortBy = m.key"
          >{{ m.label }}</button>
        </div>
        <select v-model="limit" class="limit-select">
          <option v-for="n in LIMIT_OPTIONS" :key="n" :value="n">Top {{ n }}</option>
        </select>
      </div>
    </div>

    <!-- error -->
    <p v-if="error" class="state-msg err">{{ error }}</p>

    <!-- empty -->
    <p v-else-if="!loading && users.length === 0" class="state-msg">No user data available.</p>

    <!-- chart -->
    <div v-else class="bar-chart">
      <!-- skeleton -->
      <template v-if="loading">
        <div v-for="i in 8" :key="i" class="bar-col">
          <div class="bar-value" />
          <div
            class="bar bar-zero"
            :style="{ height: (30 + (i % 4) * 22) + 'px' }"
          />
          <span class="bar-label sk-label" />
        </div>
      </template>

      <!-- data bars -->
      <template v-else>
        <div v-for="u in sortedRows" :key="u.id" class="bar-col">
          <span class="bar-value">{{ u.value > 0 ? fmtNum(u.value) : '' }}</span>
          <div
            class="bar"
            :class="{ 'bar-zero': u.value === 0 }"
            :style="{
              height:     u.barHeight + 'px',
              background: u.value > 0 ? activeMetric.color : undefined,
            }"
          />
          <span class="bar-label" :title="u.username">
            {{ u.username.length > 9 ? u.username.slice(0, 8) + '…' : u.username }}
          </span>
        </div>
      </template>
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

/* ── header ── */
.chart-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.title-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.chart-title {
  font-size: 0.9rem;
  font-weight: 700;
  color: #1e293b;
  margin: 0;
}

.count-badge {
  font-size: 0.7rem;
  font-weight: 700;
  background: #f1f5f9;
  color: #64748b;
  padding: 0.15rem 0.5rem;
  border-radius: 999px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.metric-tabs {
  display: flex;
  background: #f1f5f9;
  border-radius: 7px;
  padding: 3px;
  gap: 2px;
}

.metric-tab {
  padding: 0.22rem 0.65rem;
  border: none;
  border-radius: 5px;
  font-size: 0.75rem;
  font-weight: 600;
  cursor: pointer;
  background: transparent;
  color: #64748b;
  transition: background 0.15s, color 0.15s, box-shadow 0.15s;
}

.metric-tab.active {
  background: white;
  color: var(--c, #3b82f6);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.limit-select {
  padding: 0.24rem 0.55rem;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 600;
  color: #475569;
  background: white;
  cursor: pointer;
  outline: none;
}
.limit-select:focus { border-color: #3b82f6; box-shadow: 0 0 0 2px #bfdbfe; }

/* ── chart — identical pattern to TokensChart / RequestsChart ── */
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
  min-width: 0;
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
  transition: height 0.3s ease, background 0.3s ease;
}

.bar.bar-zero { background: #e2e8f0; }

.bar-label {
  position: absolute;
  bottom: -20px;
  font-size: 0.62rem;
  color: #94a3b8;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100%;
}

/* skeleton label */
.sk-label {
  display: block;
  width: 80%;
  height: 8px;
  background: linear-gradient(90deg, #f1f5f9 25%, #e2e8f0 50%, #f1f5f9 75%);
  background-size: 200% 100%;
  animation: shimmer 1.4s infinite;
  border-radius: 4px;
}

@keyframes shimmer {
  0%   { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

/* ── state ── */
.state-msg { font-size: 0.85rem; color: #94a3b8; margin: 0; }
.err { color: #ef4444; }
</style>
