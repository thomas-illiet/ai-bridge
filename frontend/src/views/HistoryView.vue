<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { getHistory, getHistoryDetail, getHistoryStats } from '@/services/api'
import type { InterceptionRow, InterceptionDetail, HistoryStats } from '@/services/api'
import { formatDate, fmtNum, interceptionDuration, providerColor } from '@/utils/format'
import PaginationBar from '@/components/PaginationBar.vue'
import HistoryDetailModal from '@/components/HistoryDetailModal.vue'
import { useMinLoad } from '@/composables/useMinLoad'

type SortKey = 'provider' | 'model' | 'startedAt' | 'duration' | 'inputTokens' | 'outputTokens'

const rows       = ref<InterceptionRow[]>([])
const total      = ref(0)
const page       = ref(1)
const pageSize   = ref(10)
const search     = ref('')
const { loading, withLoad } = useMinLoad(300, true)
const detail     = ref<InterceptionDetail | null>(null)
const detailLoad = ref(false)
const sortBy     = ref<SortKey>('startedAt')
const sortDir    = ref<'asc' | 'desc'>('desc')
const stats      = ref<HistoryStats | null>(null)

function toggleSort(col: SortKey) {
  if (sortBy.value === col) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortBy.value = col
    sortDir.value = 'desc'
  }
  page.value = 1
  fetchHistory()
}

function sortIcon(col: SortKey) {
  if (sortBy.value !== col) return '⇅'
  return sortDir.value === 'asc' ? '↑' : '↓'
}

let searchTimer: ReturnType<typeof setTimeout>
watch(search, () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => { page.value = 1; fetchHistory() }, 300)
})
watch(page, fetchHistory)
watch(pageSize, () => { page.value = 1; fetchHistory() })

async function fetchHistory() {
  await withLoad(async () => {
    const res = await getHistory(page.value, pageSize.value, search.value, sortBy.value, sortDir.value)
    rows.value  = res.data.interceptions
    total.value = res.data.total
  })
}

async function openDetail(id: string) {
  detail.value     = null
  detailLoad.value = true
  try {
    const res = await getHistoryDetail(id)
    detail.value = res.data
  } finally { detailLoad.value = false }
}

onMounted(() => {
  fetchHistory()
  getHistoryStats().then(r => { stats.value = r.data })
})
</script>

<template>
  <div class="history-page">
    <div class="page-header">
      <div>
        <h1>Request History</h1>
        <p class="subtitle">{{ total }} request{{ total !== 1 ? 's' : '' }} recorded.</p>
      </div>
      <input v-model="search" type="text" placeholder="Search model or provider…" class="search-input" />
    </div>

    <div v-if="stats" class="stat-grid">
      <div class="stat-card">
        <span class="stat-label">Total Requests</span>
        <span class="stat-value">{{ fmtNum(stats.total) }}</span>
      </div>
      <div class="stat-card">
        <span class="stat-label">Input</span>
        <span class="stat-value">{{ fmtNum(stats.totalInput) }}</span>
      </div>
      <div class="stat-card">
        <span class="stat-label">Output</span>
        <span class="stat-value">{{ fmtNum(stats.totalOutput) }}</span>
      </div>
      <div class="stat-card stat-card--model">
        <span class="stat-label">Top Model</span>
        <span v-if="stats.topModel" class="stat-model">{{ stats.topModel }}</span>
        <span v-else class="stat-empty">No requests yet</span>
      </div>
    </div>

    <div v-if="!loading && rows.length === 0" class="empty-card">
      <div class="empty-icon">
        <svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>
        </svg>
      </div>
      <p class="empty-title">No requests found</p>
      <p class="empty-sub">{{ search ? 'Try adjusting your search.' : 'Your API request history will appear here.' }}</p>
    </div>

    <table v-else class="data-table">
      <thead>
        <tr>
          <th class="col-center sortable" :class="{ active: sortBy === 'provider' }" @click="toggleSort('provider')">
            Provider <span class="sort-icon">{{ sortIcon('provider') }}</span>
          </th>
          <th class="col-center sortable" :class="{ active: sortBy === 'model' }" @click="toggleSort('model')">
            Model <span class="sort-icon">{{ sortIcon('model') }}</span>
          </th>
          <th class="col-center sortable" :class="{ active: sortBy === 'startedAt' }" @click="toggleSort('startedAt')">
            Started <span class="sort-icon">{{ sortIcon('startedAt') }}</span>
          </th>
          <th class="col-center sortable" :class="{ active: sortBy === 'duration' }" @click="toggleSort('duration')">
            Duration <span class="sort-icon">{{ sortIcon('duration') }}</span>
          </th>
          <th class="col-center sortable" :class="{ active: sortBy === 'inputTokens' }" @click="toggleSort('inputTokens')">
            Input <span class="sort-icon">{{ sortIcon('inputTokens') }}</span>
          </th>
          <th class="col-center sortable" :class="{ active: sortBy === 'outputTokens' }" @click="toggleSort('outputTokens')">
            Output <span class="sort-icon">{{ sortIcon('outputTokens') }}</span>
          </th>
          <th class="col-center">Actions</th>
        </tr>
      </thead>
      <tbody>
        <template v-if="loading">
          <tr v-for="i in 6" :key="i" class="skeleton-row">
            <td class="col-center"><div class="skeleton-bar skeleton-bar--pill" style="margin:auto" /></td>
            <td class="col-center"><div class="skeleton-bar skeleton-bar--md" style="margin:auto" /></td>
            <td class="col-center"><div class="skeleton-bar skeleton-bar--sm" style="margin:auto" /></td>
            <td class="col-center"><div class="skeleton-bar skeleton-bar--xs" style="margin:auto" /></td>
            <td class="col-center"><div class="skeleton-bar skeleton-bar--xs" style="margin:auto" /></td>
            <td class="col-center"><div class="skeleton-bar skeleton-bar--xs" style="margin:auto" /></td>
            <td class="col-center"><div class="skeleton-bar skeleton-bar--btn" style="margin:auto" /></td>
          </tr>
        </template>
        <template v-else>
          <tr v-for="row in rows" :key="row.id">
            <td class="col-center">
              <span class="provider-badge" :style="{ background: providerColor(row.provider) + '22', color: providerColor(row.provider) }">
                {{ row.provider }}
              </span>
            </td>
            <td class="col-center model-cell">{{ row.model }}</td>
            <td class="col-center muted">{{ formatDate(row.startedAt) }}</td>
            <td class="col-center muted">{{ interceptionDuration(row) }}</td>
            <td class="col-center">{{ fmtNum(row.inputTokens) }}</td>
            <td class="col-center">{{ fmtNum(row.outputTokens) }}</td>
            <td class="col-center">
              <button class="btn-view" @click="openDetail(row.id)">View</button>
            </td>
          </tr>
        </template>
      </tbody>
    </table>

    <PaginationBar v-if="!loading" v-model:page="page" v-model:pageSize="pageSize" :total="total" />

    <HistoryDetailModal :detail="detail" :loading="detailLoad" @close="detail = null; detailLoad = false" />
  </div>
</template>

<style scoped>
.history-page { display: flex; flex-direction: column; gap: 1.5rem; }
.stat-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 1rem; }
.stat-card { background: white; border: 1px solid #e2e8f0; border-radius: 10px; padding: 1.1rem 1.25rem; display: flex; flex-direction: column; gap: 0.3rem; }
.stat-label { font-size: 0.78rem; font-weight: 600; color: #64748b; text-transform: uppercase; letter-spacing: 0.04em; }
.stat-value { font-size: 1.9rem; font-weight: 700; color: #0f172a; line-height: 1; }
.stat-card--model { justify-content: space-between; }
.stat-model { font-size: 0.9rem; font-weight: 700; color: #0f172a; font-family: monospace; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; margin-top: auto; }
.stat-empty { font-size: 0.85rem; color: #94a3b8; margin-top: auto; }
.page-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 1rem; flex-wrap: wrap; }
h1 { font-size: 1.75rem; font-weight: 700; margin: 0; }
.subtitle { font-size: 0.85rem; color: #64748b; margin: 0.2rem 0 0; }
</style>
