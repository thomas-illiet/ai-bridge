<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { adminGetHistory, adminGetHistoryDetail, listUsers } from '@/services/api'
import type { InterceptionRow, InterceptionDetail } from '@/services/api'
import { formatDate, fmtNum, interceptionDuration, providerColor } from '@/utils/format'
import PaginationBar from '@/components/PaginationBar.vue'
import HistoryDetailModal from '@/components/HistoryDetailModal.vue'
import { useMinLoad } from '@/composables/useMinLoad'

type SortKey = 'provider' | 'model' | 'startedAt' | 'duration' | 'inputTokens' | 'outputTokens' | 'username'

interface UserOption { id: string; username: string }

const rows       = ref<InterceptionRow[]>([])
const total      = ref(0)
const page       = ref(1)
const pageSize   = ref(10)
const search     = ref('')
const userId     = ref('')
const error      = ref<string | null>(null)
const { loading, withLoad } = useMinLoad(300, true)
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
  error.value = null
  await withLoad(async () => {
    try {
      const res = await adminGetHistory(page.value, pageSize.value, search.value, userId.value, sortBy.value, sortDir.value)
      rows.value  = res.data.interceptions
      total.value = res.data.total
    } catch (e: any) {
      error.value = e?.response?.data?.error ?? 'Failed to load history'
    }
  })
}

async function openDetail(id: string) {
  detail.value = null; detailLoad.value = true
  try {
    const res = await adminGetHistoryDetail(id)
    detail.value = res.data
  } finally { detailLoad.value = false }
}

onMounted(async () => {
  try {
    const res = await listUsers()
    users.value = res.data.users ?? []
  } catch { /* dropdown won't be populated but page still works */ }
  load()
})
</script>

<template>
  <Teleport defer to="#admin-search-portal">
    <div class="portal-controls">
      <select v-model="userId" class="portal-select">
        <option value="">All users</option>
        <option v-for="u in users" :key="u.id" :value="u.id">{{ u.username }}</option>
      </select>
      <div class="portal-search-wrap">
        <svg class="portal-search-icon" xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
        <input v-model="search" type="text" placeholder="Search model or provider…" class="portal-input" />
      </div>
    </div>
  </Teleport>

  <div class="tab-content">
    <div class="card">
    <div class="card-header">
      <h2 class="card-title">Request History <span class="title-count">{{ total }}</span></h2>
    </div>

    <div v-if="!loading && error" class="empty-card empty-card--error">
      <div class="empty-icon empty-icon--error">
        <svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
        </svg>
      </div>
      <p class="empty-title">Failed to load history</p>
      <p class="empty-sub">{{ error }}</p>
    </div>
    <div v-else-if="!loading && rows.length === 0" class="empty-card">
      <div class="empty-icon">
        <svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>
        </svg>
      </div>
      <p class="empty-title">No requests found</p>
      <p class="empty-sub">{{ search || userId ? 'Try adjusting your filters.' : 'API requests will appear here once users start sending them.' }}</p>
    </div>

    <table v-else class="data-table">
      <thead>
        <tr>
          <th class="col-center sortable" :class="{ active: sortBy === 'provider' }" @click="toggleSort('provider')">Provider <span class="sort-icon">{{ sortIcon('provider') }}</span></th>
          <th class="col-center sortable" :class="{ active: sortBy === 'model' }" @click="toggleSort('model')">Model <span class="sort-icon">{{ sortIcon('model') }}</span></th>
          <th class="col-center sortable" :class="{ active: sortBy === 'startedAt' }" @click="toggleSort('startedAt')">Started <span class="sort-icon">{{ sortIcon('startedAt') }}</span></th>
          <th class="col-center sortable" :class="{ active: sortBy === 'duration' }" @click="toggleSort('duration')">Duration <span class="sort-icon">{{ sortIcon('duration') }}</span></th>
          <th class="col-center sortable" :class="{ active: sortBy === 'inputTokens' }" @click="toggleSort('inputTokens')">Input <span class="sort-icon">{{ sortIcon('inputTokens') }}</span></th>
          <th class="col-center sortable" :class="{ active: sortBy === 'outputTokens' }" @click="toggleSort('outputTokens')">Output <span class="sort-icon">{{ sortIcon('outputTokens') }}</span></th>
          <th class="col-center sortable" :class="{ active: sortBy === 'username' }" @click="toggleSort('username')">User <span class="sort-icon">{{ sortIcon('username') }}</span></th>
          <th class="col-center">IP</th>
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
            <td class="col-center"><div class="skeleton-bar skeleton-bar--pill" style="margin:auto" /></td>
            <td class="col-center"><div class="skeleton-bar skeleton-bar--md" style="margin:auto" /></td>
            <td class="col-center"><div class="skeleton-bar skeleton-bar--btn" style="margin:auto" /></td>
          </tr>
        </template>
        <template v-else>
          <tr v-for="row in rows" :key="row.id">
            <td class="col-center">
              <span class="prov-badge" :style="{ background: providerColor(row.providerType) + '22', color: providerColor(row.providerType) }">{{ row.provider }}</span>
            </td>
            <td class="col-center model-cell">{{ row.model }}</td>
            <td class="col-center muted">{{ formatDate(row.startedAt) }}</td>
            <td class="col-center muted">{{ interceptionDuration(row) }}</td>
            <td class="col-center num">{{ fmtNum(row.inputTokens) }}</td>
            <td class="col-center num">{{ fmtNum(row.outputTokens) }}</td>
            <td class="col-center"><span class="user-pill">{{ row.username }}</span></td>
            <td class="col-center ip-cell"><span class="ip-text">{{ row.clientIp || '—' }}</span></td>
            <td class="col-center"><button class="btn-view" @click="openDetail(row.id)"><svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg> View</button></td>
          </tr>
        </template>
      </tbody>
    </table>

    <PaginationBar v-if="!loading" v-model:page="page" v-model:pageSize="pageSize" :total="total" />
    </div>

    <HistoryDetailModal :detail="detail" :loading="detailLoad" @close="detail = null; detailLoad = false" />
  </div>
</template>

<style scoped>
.model-cell { max-width: 200px; }
.ip-cell { white-space: nowrap; }
.ip-text { font-family: monospace; font-size: 0.8rem; color: #475569; }
</style>
