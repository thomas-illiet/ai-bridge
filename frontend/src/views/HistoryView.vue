<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { getHistory, getHistoryDetail } from '@/services/api'
import type { InterceptionRow, InterceptionDetail } from '@/services/api'
import { formatDate, fmtNum, interceptionDuration, providerColor } from '@/utils/format'
import PaginationBar from '@/components/PaginationBar.vue'
import HistoryDetailModal from '@/components/HistoryDetailModal.vue'

type SortKey = 'provider' | 'model' | 'startedAt' | 'duration' | 'inputTokens' | 'outputTokens'

const rows       = ref<InterceptionRow[]>([])
const total      = ref(0)
const page       = ref(1)
const pageSize   = ref(10)
const search     = ref('')
const loading    = ref(false)
const detail     = ref<InterceptionDetail | null>(null)
const detailLoad = ref(false)
const sortBy     = ref<SortKey>('startedAt')
const sortDir    = ref<'asc' | 'desc'>('desc')

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
  loading.value = true
  try {
    const res = await getHistory(page.value, pageSize.value, search.value, sortBy.value, sortDir.value)
    rows.value  = res.data.interceptions
    total.value = res.data.total
  } finally { loading.value = false }
}

async function openDetail(id: string) {
  detail.value     = null
  detailLoad.value = true
  try {
    const res = await getHistoryDetail(id)
    detail.value = res.data
  } finally { detailLoad.value = false }
}

onMounted(fetchHistory)
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

    <div v-if="loading && rows.length === 0" class="empty-card">
      <div class="empty-icon">
        <svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/>
        </svg>
      </div>
      <p class="empty-title">Loading history…</p>
    </div>
    <div v-else-if="rows.length === 0" class="empty-card">
      <div class="empty-icon">
        <svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>
        </svg>
      </div>
      <p class="empty-title">No requests found</p>
      <p class="empty-sub">{{ search ? 'Try adjusting your search.' : 'Your API request history will appear here.' }}</p>
    </div>

    <table v-else class="data-table" :class="{ 'table-loading': loading }">
      <thead>
        <tr>
          <th class="sortable" :class="{ active: sortBy === 'provider' }" @click="toggleSort('provider')">
            Provider <span class="sort-icon">{{ sortIcon('provider') }}</span>
          </th>
          <th class="sortable" :class="{ active: sortBy === 'model' }" @click="toggleSort('model')">
            Model <span class="sort-icon">{{ sortIcon('model') }}</span>
          </th>
          <th class="sortable" :class="{ active: sortBy === 'startedAt' }" @click="toggleSort('startedAt')">
            Started <span class="sort-icon">{{ sortIcon('startedAt') }}</span>
          </th>
          <th class="sortable" :class="{ active: sortBy === 'duration' }" @click="toggleSort('duration')">
            Duration <span class="sort-icon">{{ sortIcon('duration') }}</span>
          </th>
          <th class="num sortable" :class="{ active: sortBy === 'inputTokens' }" @click="toggleSort('inputTokens')">
            Input <span class="sort-icon">{{ sortIcon('inputTokens') }}</span>
          </th>
          <th class="num sortable" :class="{ active: sortBy === 'outputTokens' }" @click="toggleSort('outputTokens')">
            Output <span class="sort-icon">{{ sortIcon('outputTokens') }}</span>
          </th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="row in rows" :key="row.id">
          <td>
            <span class="provider-badge" :style="{ background: providerColor(row.provider) + '22', color: providerColor(row.provider) }">
              {{ row.provider }}
            </span>
          </td>
          <td class="model-cell">{{ row.model }}</td>
          <td class="muted">{{ formatDate(row.startedAt) }}</td>
          <td class="muted">{{ interceptionDuration(row) }}</td>
          <td class="num">{{ fmtNum(row.inputTokens) }}</td>
          <td class="num">{{ fmtNum(row.outputTokens) }}</td>
          <td>
            <button class="btn-view" @click="openDetail(row.id)">View</button>
          </td>
        </tr>
      </tbody>
    </table>

    <PaginationBar v-model:page="page" v-model:pageSize="pageSize" :total="total" />

    <HistoryDetailModal :detail="detail" :loading="detailLoad" @close="detail = null; detailLoad = false" />
  </div>
</template>

<style scoped>
.history-page { display: flex; flex-direction: column; gap: 1.5rem; }
.page-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 1rem; flex-wrap: wrap; }
h1 { font-size: 1.75rem; font-weight: 700; margin: 0; }
.subtitle { font-size: 0.85rem; color: #64748b; margin: 0.2rem 0 0; }
.search-input { padding: 0.45rem 0.75rem; border: 1px solid #d1d5db; border-radius: 6px; font-size: 0.9rem; width: 220px; background: white; outline: none; }
.search-input:focus { border-color: #3b82f6; box-shadow: 0 0 0 2px #bfdbfe; }
.data-table { width: 100%; border-collapse: collapse; background: white; border: 1px solid #e2e8f0; border-radius: 12px; overflow: hidden; transition: opacity 0.15s; }
.data-table.table-loading { opacity: 0.6; pointer-events: none; }
.data-table th, .data-table td { padding: 0.7rem 1rem; text-align: left; border-bottom: 1px solid #f1f5f9; font-size: 0.88rem; }
.data-table th { background: #f8fafc; font-weight: 600; color: #64748b; font-size: 0.75rem; text-transform: uppercase; letter-spacing: 0.04em; white-space: nowrap; }
.data-table th.num { text-align: right; }
.data-table tr:last-child td { border-bottom: none; }
.sortable { cursor: pointer; user-select: none; }
.sortable:hover { color: #334155; background: #f1f5f9; }
.sortable.active { color: #3b82f6; }
.sort-icon { font-size: 0.7rem; margin-left: 0.25rem; opacity: 0.5; }
.sortable.active .sort-icon { opacity: 1; }
.num { text-align: right; font-variant-numeric: tabular-nums; color: #334155; font-weight: 500; }
.muted { color: #64748b; }
.model-cell { font-family: monospace; font-size: 0.82rem; color: #334155; max-width: 240px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.provider-badge { display: inline-block; padding: 0.15rem 0.55rem; border-radius: 999px; font-size: 0.75rem; font-weight: 600; text-transform: capitalize; }
.btn-view { padding: 0.2rem 0.65rem; border: 1px solid #cbd5e1; border-radius: 6px; background: white; color: #374151; font-size: 0.8rem; font-weight: 500; cursor: pointer; transition: background 0.12s; }
.btn-view:hover { background: #f1f5f9; }
.empty-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.6rem;
  padding: 3.5rem 2rem;
  background: white;
  border: 1px dashed #e2e8f0;
  border-radius: 12px;
  text-align: center;
}
.empty-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  border-radius: 16px;
  background: #f1f5f9;
  color: #94a3b8;
  margin-bottom: 0.25rem;
}
.empty-title { font-size: 1rem; font-weight: 600; color: #1e293b; margin: 0; }
.empty-sub   { font-size: 0.85rem; color: #94a3b8; margin: 0; max-width: 320px; line-height: 1.5; }
</style>
