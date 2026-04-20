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
  await withLoad(async () => {
    const res = await adminGetHistory(page.value, pageSize.value, search.value, userId.value, sortBy.value, sortDir.value)
    rows.value  = res.data.interceptions
    total.value = res.data.total
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

    <div v-if="!loading && rows.length === 0" class="empty-card">
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
            <td class="col-center"><button class="btn-view" @click="openDetail(row.id)">View</button></td>
          </tr>
        </template>
      </tbody>
    </table>

    <PaginationBar v-if="!loading" v-model:page="page" v-model:pageSize="pageSize" :total="total" />

    <HistoryDetailModal :detail="detail" :loading="detailLoad" @close="detail = null; detailLoad = false" />
  </div>
</template>

<style scoped>
/* model-cell max-width override for narrower admin layout */
.model-cell { max-width: 200px; }
</style>
