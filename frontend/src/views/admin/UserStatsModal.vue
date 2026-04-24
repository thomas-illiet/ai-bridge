<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import {
  getUserTotalRequests,
  getUserTokenTotals,
  getUserDailyRequests,
  getUserByProvider,
  getUserByModel,
} from '@/services/api'
import { fmtNum } from '@/utils/format'

interface RegisteredUser { id: string; username: string; role: string }
interface UserStats {
  totalRequests: number; totalInput: number; totalOutput: number
  daily: { date: string; count: number }[]
  byProvider: { provider: string; count: number }[]
  byModel:    { model: string; count: number }[]
}

const props = defineProps<{ user: RegisteredUser | null }>()
defineEmits<{ close: [] }>()

const stats = ref<UserStats | null>(null)
const loading = ref(false)
const error   = ref<string | null>(null)

watch(() => props.user, async (u) => {
  if (!u) return
  stats.value = null; error.value = null; loading.value = true
  try {
    const [reqRes, tokRes, dailyRes, provRes, modelRes] = await Promise.all([
      getUserTotalRequests(u.id),
      getUserTokenTotals(u.id),
      getUserDailyRequests(u.id),
      getUserByProvider(u.id),
      getUserByModel(u.id),
    ])
    stats.value = {
      totalRequests: reqRes.data.totalRequests,
      totalInput:    tokRes.data.totalInput,
      totalOutput:   tokRes.data.totalOutput,
      daily:         dailyRes.data.daily,
      byProvider:    provRes.data.byProvider,
      byModel:       modelRes.data.byModel,
    }
  } catch { error.value = 'Failed to load stats' }
  finally  { loading.value = false }
}, { immediate: true })

const chartDays = computed<{ date: string; count: number }[]>(() => {
  const map = new Map((stats.value?.daily ?? []).map(d => [d.date, d.count]))
  const days = []
  for (let i = 6; i >= 0; i--) {
    const d = new Date(); d.setDate(d.getDate() - i)
    const key = d.toISOString().slice(0, 10)
    days.push({ date: key, count: map.get(key) ?? 0 })
  }
  return days
})
const barMax = computed(() => Math.max(1, ...chartDays.value.map(d => d.count)))
function barH(n: number) { return Math.round((n / barMax.value) * 100) }
function shortDate(iso: string) {
  return new Date(iso + 'T00:00:00').toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}
const provTotal = computed(() => (stats.value?.byProvider ?? []).reduce((s, p) => s + p.count, 0))
const PROV_COLORS: Record<string, string> = { openai: '#10b981', ollama: '#6366f1' }
function provColor(n: string) { return PROV_COLORS[n] ?? '#94a3b8' }

function roleBadgeClass(role: string) {
  return { 'badge-admin': role === 'admin', 'badge-user': role === 'user', 'badge-none': role === 'none' }
}
</script>

<template>
  <Teleport to="body">
    <div v-if="user" class="modal-backdrop" @click.self="$emit('close')">
      <div class="modal">
        <div class="modal-header">
          <div>
            <h2>{{ user.username }}</h2>
            <span class="badge" :class="roleBadgeClass(user.role)">{{ user.role }}</span>
          </div>
          <button class="modal-close" @click="$emit('close')">✕</button>
        </div>

        <div v-if="loading" class="state-msg">Loading…</div>
        <div v-else-if="error" class="state-msg error">{{ error }}</div>

        <template v-else-if="stats">
          <div class="stat-grid">
            <div class="stat-card">
              <span class="stat-label">Requests</span>
              <span class="stat-value">{{ fmtNum(stats.totalRequests) }}</span>
            </div>
            <div class="stat-card">
              <span class="stat-label">Input tokens</span>
              <span class="stat-value">{{ fmtNum(stats.totalInput) }}</span>
            </div>
            <div class="stat-card">
              <span class="stat-label">Output tokens</span>
              <span class="stat-value">{{ fmtNum(stats.totalOutput) }}</span>
            </div>
          </div>

          <div v-if="stats.totalRequests === 0" class="no-activity">
            <span class="no-activity-icon">📭</span>
            <p class="no-activity-title">No activity yet</p>
            <p class="no-activity-sub">This user hasn't made any requests through AI Bridge.</p>
          </div>

          <template v-else>
            <div class="charts-row">
              <div class="chart-card">
                <h3 class="chart-title">Requests — Last 7 days</h3>
                <div class="bar-chart">
                  <div v-for="day in chartDays" :key="day.date" class="bar-col">
                    <span class="bar-value">{{ day.count > 0 ? fmtNum(day.count) : '' }}</span>
                    <div class="bar" :style="{ height: barH(day.count) + 'px' }" :class="{ 'bar-zero': day.count === 0 }" />
                    <span class="bar-label">{{ shortDate(day.date) }}</span>
                  </div>
                </div>
              </div>

              <div class="chart-card">
                <h3 class="chart-title">By provider</h3>
                <div v-if="!(stats.byProvider ?? []).length" class="no-data">No data.</div>
                <div v-else class="provider-list">
                  <div v-for="p in stats.byProvider" :key="p.provider" class="provider-row">
                    <div class="provider-meta">
                      <span class="provider-dot" :style="{ background: provColor(p.provider) }" />
                      <span class="provider-name">{{ p.provider }}</span>
                      <span class="provider-count">{{ fmtNum(p.count) }}</span>
                      <span class="provider-pct">{{ provTotal > 0 ? Math.round(p.count / provTotal * 100) : 0 }}%</span>
                    </div>
                    <div class="provider-bar-track">
                      <div class="provider-bar-fill" :style="{ width: (provTotal > 0 ? p.count / provTotal * 100 : 0) + '%', background: provColor(p.provider) }" />
                    </div>
                  </div>
                </div>

                <h3 class="chart-title" style="margin-top:1rem">Top models</h3>
                <div v-if="!(stats.byModel ?? []).length" class="no-data">No data.</div>
                <div v-else class="model-list">
                  <div v-for="m in stats.byModel" :key="m.model" class="model-row">
                    <span class="model-name mono">{{ m.model }}</span>
                    <span class="model-count">{{ fmtNum(m.count) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </template>
        </template>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.modal-backdrop { position: fixed; inset: 0; background: rgba(0,0,0,0.45); display: flex; align-items: center; justify-content: center; z-index: 200; padding: 1rem; }
.modal { background: white; border-radius: 14px; padding: 1.75rem; width: 100%; max-width: 820px; max-height: 90vh; overflow-y: auto; display: flex; flex-direction: column; gap: 1.25rem; box-shadow: 0 20px 60px rgba(0,0,0,0.25); }
.modal-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 1rem; }
.modal-header h2 { font-size: 1.25rem; font-weight: 700; margin: 0 0 0.35rem; }
.modal-close { background: none; border: none; font-size: 1.1rem; cursor: pointer; color: #64748b; padding: 0.2rem 0.4rem; border-radius: 5px; }
.modal-close:hover { background: #f1f5f9; color: #1e293b; }
.badge { padding: 0.18rem 0.55rem; border-radius: 999px; font-size: 0.75rem; font-weight: 600; }
.badge-admin { background: #ede9fe; color: #6d28d9; }
.badge-user  { background: #dcfce7; color: #166534; }
.badge-none  { background: #f1f5f9; color: #64748b; }
.stat-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 0.75rem; }
.stat-card { background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 8px; padding: 0.9rem 1rem; display: flex; flex-direction: column; align-items: center; gap: 0.25rem; text-align: center; }
.stat-label { font-size: 0.72rem; font-weight: 600; color: #64748b; text-transform: uppercase; letter-spacing: 0.04em; }
.stat-value { font-size: 1.6rem; font-weight: 700; color: #0f172a; line-height: 1; }
.charts-row { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
@media (max-width: 600px) { .charts-row { grid-template-columns: 1fr; } }
.chart-card { background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 10px; padding: 1rem; display: flex; flex-direction: column; gap: 0.75rem; }
.chart-title { font-size: 0.82rem; font-weight: 700; color: #475569; margin: 0; text-transform: uppercase; letter-spacing: 0.04em; }
.bar-chart { display: flex; align-items: flex-end; gap: 5px; height: 130px; padding-bottom: 22px; position: relative; }
.bar-col { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: flex-end; gap: 3px; height: 100%; position: relative; }
.bar-value { font-size: 0.6rem; color: #64748b; height: 12px; line-height: 12px; }
.bar { width: 100%; background: #3b82f6; border-radius: 3px 3px 0 0; min-height: 3px; }
.bar.bar-zero { background: #e2e8f0; }
.bar-label { position: absolute; bottom: -18px; font-size: 0.58rem; color: #94a3b8; white-space: nowrap; }
.no-data { font-size: 0.82rem; color: #94a3b8; }
.provider-list { display: flex; flex-direction: column; gap: 0.6rem; }
.provider-row  { display: flex; flex-direction: column; gap: 0.25rem; }
.provider-meta { display: flex; align-items: center; gap: 0.4rem; font-size: 0.82rem; }
.provider-dot  { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.provider-name { flex: 1; font-weight: 600; color: #1e293b; text-transform: capitalize; }
.provider-count { color: #475569; }
.provider-pct   { color: #94a3b8; font-size: 0.75rem; }
.provider-bar-track { height: 5px; background: #e2e8f0; border-radius: 999px; overflow: hidden; }
.provider-bar-fill  { height: 100%; border-radius: 999px; transition: width 0.4s ease; }
.model-list { display: flex; flex-direction: column; gap: 0.35rem; }
.model-row { display: flex; align-items: center; justify-content: space-between; font-size: 0.82rem; padding: 0.25rem 0; border-bottom: 1px solid #e2e8f0; }
.model-row:last-child { border-bottom: none; }
.model-name { color: #334155; font-size: 0.78rem; }
.mono { font-family: monospace; }
.model-count { font-weight: 600; color: #475569; }
.no-activity { display: flex; flex-direction: column; align-items: center; gap: 0.4rem; padding: 2rem 1rem; text-align: center; background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 10px; }
.no-activity-icon { font-size: 2rem; line-height: 1; }
.no-activity-title { font-size: 1rem; font-weight: 600; color: #334155; margin: 0; }
.no-activity-sub { font-size: 0.85rem; color: #94a3b8; margin: 0; }
.state-msg { color: #64748b; font-size: 0.9rem; }
.state-msg.error { color: #ef4444; }
</style>
