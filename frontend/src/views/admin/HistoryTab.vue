<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { adminGetHistory, adminGetHistoryDetail, listUsers } from '@/services/api'
import type { InterceptionRow, InterceptionDetail } from '@/services/api'
import { formatDate, fmtNum, interceptionDuration, providerColor } from '@/utils/format'
import PaginationBar from '@/components/PaginationBar.vue'
import HistoryDetailModal from '@/components/HistoryDetailModal.vue'

type SortKey = 'provider' | 'model' | 'startedAt' | 'duration' | 'inputTokens' | 'outputTokens'

interface UserOption { id: string; username: string }

const rows       = ref<InterceptionRow[]>([])
const total      = ref(0)
const page       = ref(1)
const pageSize   = ref(10)
const search     = ref('')
const userId     = ref('')
const loading    = ref(false)
const users      = ref<UserOption[]>([])
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
  load()
}

function sortIcon(col: SortKey) {
  if (sortBy.value !== col) return '⇅'
  return sortDir.value === 'asc' ? '↑' : '↓'
}

let searchTimer: ReturnType<typeof setTimeout>
watch([search, userId], () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => { page.value = 1; load() }, 300)
})
watch(page, load)
watch(pageSize, () => { page.value = 1; load() })

async function load() {
  loading.value = true
  try {
    const res = await adminGetHistory(page.value, pageSize.value, search.value, userId.value, sortBy.value, sortDir.value)
    rows.value  = res.data.interceptions
    total.value = res.data.total
  } finally { loading.value = false }
}

async function openDetail(id: string) {
  detail.value = null; detailLoad.value = true
  try {
    const res = await adminGetHistoryDetail(id)
    detail.value = res.data
  } finally { detailLoad.value = false }
}

onMounted(async () => {
  const res = await listUsers()
  users.value = res.data.users ?? []
  load()
})
</script>

<template>
  <div class="tab-content">
    <div class="toolbar">
      <p class="sub">{{ total }} request{{ total !== 1 ? 's' : '' }} total.</p>
      <div class="filters">
        <select v-model="userId" class="role-select">
          <option value="">All users</option>
          <option v-for="u in users" :key="u.id" :value="u.id">{{ u.username }}</option>
        </select>
        <input v-model="search" type="text" placeholder="Search model or provider…" class="search-input" />
      </div>
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
      <p class="empty-sub">{{ search || userId ? 'Try adjusting your filters.' : 'API requests will appear here once users start sending them.' }}</p>
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
          <th>User</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="row in rows" :key="row.id">
          <td>
            <span class="prov-badge" :style="{ background: providerColor(row.provider) + '22', color: providerColor(row.provider) }">
              {{ row.provider }}
            </span>
          </td>
          <td class="model-cell">{{ row.model }}</td>
          <td class="muted">{{ formatDate(row.startedAt) }}</td>
          <td class="muted">{{ interceptionDuration(row) }}</td>
          <td class="num">{{ fmtNum(row.inputTokens) }}</td>
          <td class="num">{{ fmtNum(row.outputTokens) }}</td>
          <td><span class="user-pill">{{ row.username }}</span></td>
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
.tab-content { display: flex; flex-direction: column; gap: 1.25rem; }
.toolbar { display: flex; align-items: center; justify-content: space-between; gap: 1rem; flex-wrap: wrap; }
.sub { font-size: 0.85rem; color: #64748b; margin: 0; }
.filters { display: flex; gap: 0.5rem; align-items: center; flex-wrap: wrap; }
.search-input { padding: 0.45rem 0.75rem; border: 1px solid #e2e8f0; border-radius: 6px; font-size: 0.9rem; width: 220px; background: white; outline: none; }
.search-input:focus { border-color: #3b82f6; box-shadow: 0 0 0 2px #bfdbfe; }
.role-select { padding: 0.3rem 0.5rem; border: 1px solid #e2e8f0; border-radius: 6px; font-size: 0.85rem; background: white; cursor: pointer; }
.empty-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.6rem;
  padding: 3rem 2rem;
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
.data-table { width: 100%; border-collapse: collapse; font-size: 0.88rem; background: white; border: 1px solid #e2e8f0; border-radius: 10px; }
.data-table th { text-align: left; padding: 0.55rem 0.9rem; font-size: 0.75rem; font-weight: 600; color: #64748b; text-transform: uppercase; letter-spacing: 0.03em; border-bottom: 1px solid #e2e8f0; background: #f8fafc; white-space: nowrap; }
.data-table th.num { text-align: right; }
.data-table thead tr:first-child th:first-child { border-radius: 10px 0 0 0; }
.data-table thead tr:first-child th:last-child  { border-radius: 0 10px 0 0; }
.data-table tbody tr:last-child td:first-child   { border-radius: 0 0 0 10px; }
.data-table tbody tr:last-child td:last-child    { border-radius: 0 0 10px 0; }
.data-table td { padding: 0.65rem 0.9rem; border-bottom: 1px solid #f1f5f9; }
.data-table tr:last-child td { border-bottom: none; }
.data-table.table-loading { opacity: 0.6; pointer-events: none; }
.sortable { cursor: pointer; user-select: none; }
.sortable:hover { color: #334155; background: #f1f5f9; }
.sortable.active { color: #3b82f6; }
.sort-icon { font-size: 0.7rem; margin-left: 0.25rem; opacity: 0.5; }
.sortable.active .sort-icon { opacity: 1; }
.muted { color: #64748b; }
.num { text-align: right; font-variant-numeric: tabular-nums; color: #334155; font-weight: 500; }
.model-cell { font-family: monospace; font-size: 0.82rem; color: #334155; max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.prov-badge { display: inline-block; padding: 0.15rem 0.55rem; border-radius: 999px; font-size: 0.75rem; font-weight: 600; text-transform: capitalize; }
.user-pill { display: inline-block; background: #f1f5f9; color: #475569; font-size: 0.75rem; font-weight: 600; padding: 0.15rem 0.55rem; border-radius: 999px; }
.btn-view { padding: 0.2rem 0.65rem; border: 1px solid #cbd5e1; border-radius: 6px; background: white; color: #374151; font-size: 0.8rem; font-weight: 500; cursor: pointer; }
.btn-view:hover { background: #f1f5f9; }
</style>
