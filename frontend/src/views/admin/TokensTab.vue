<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { adminListTokens, adminRevokeToken, adminUnrevokeToken } from '@/services/api'
import type { AdminTokenRow } from '@/services/api'
import { formatDate, tokenStatus } from '@/utils/format'
import PaginationBar from '@/components/PaginationBar.vue'
import RevokeTokenModal from '@/views/tokens/RevokeTokenModal.vue'
import { useMinLoad } from '@/composables/useMinLoad'

const tokens        = ref<AdminTokenRow[]>([])
const total         = ref(0)
const page          = ref(1)
const pageSize      = ref(10)
const search        = ref('')
const showInactive   = ref(false)
const { loading, withLoad } = useMinLoad(300, true)
const actioningId   = ref<string | null>(null)
const sortBy        = ref('created_at')
const sortDir       = ref<'asc' | 'desc'>('desc')

const confirmToken  = ref<AdminTokenRow | null>(null)
const confirmAction = ref<'revoke' | 'unrevoke'>('revoke')

function toggleSort(col: string) {
  if (sortBy.value === col) { sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc' }
  else { sortBy.value = col; sortDir.value = 'desc' }
  page.value = 1
  load()
}

let searchTimer: ReturnType<typeof setTimeout>
watch(search, () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => { page.value = 1; load() }, 300)
})
watch(page, load)
watch(pageSize, () => { page.value = 1; load() })
watch(showInactive, () => { page.value = 1; load() })

async function load() {
  await withLoad(async () => {
    const res = await adminListTokens(page.value, pageSize.value, search.value, showInactive.value, sortBy.value, sortDir.value)
    tokens.value = res.data.tokens
    total.value  = res.data.total
  })
}

function openConfirm(token: AdminTokenRow, action: 'revoke' | 'unrevoke') {
  confirmToken.value  = token
  confirmAction.value = action
}

function closeConfirm() {
  confirmToken.value = null
}

async function doAction() {
  if (!confirmToken.value) return
  const id = confirmToken.value.id
  actioningId.value = id
  if (confirmAction.value === 'revoke') await adminRevokeToken(id)
  else await adminUnrevokeToken(id)
  actioningId.value = null
  await load()
}

onMounted(load)
</script>

<template>
  <div class="tab-content">
    <div class="card">
    <div class="card-header">
      <h2 class="card-title">Tokens</h2>
      <div class="header-actions">
        <p class="sub">{{ total }} token{{ total !== 1 ? 's' : '' }} total.</p>
        <label class="toggle-label">
          <input type="checkbox" v-model="showInactive" />
          <span class="toggle-switch" />
          Show inactive
        </label>
        <input v-model="search" type="text" placeholder="Search by name or user…" class="search-input" />
      </div>
    </div>

    <div v-if="!loading && tokens.length === 0" class="empty-card">
      <div class="empty-icon">
        <svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4"/>
        </svg>
      </div>
      <p class="empty-title">No tokens found</p>
      <p class="empty-sub">{{ search ? 'Try adjusting your search.' : showInactive ? 'No tokens exist yet.' : 'No active tokens. Inactive tokens (revoked or expired) can be shown with the toggle above.' }}</p>
    </div>

    <table v-else class="data-table">
      <thead>
        <tr>
          <th class="sortable" :class="{ active: sortBy === 'name' }" @click="toggleSort('name')">Name <span class="sort-icon">{{ sortBy === 'name' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
          <th class="col-center sortable" :class="{ active: sortBy === 'username' }" @click="toggleSort('username')">User <span class="sort-icon">{{ sortBy === 'username' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
          <th class="col-center sortable" :class="{ active: sortBy === 'created_at' }" @click="toggleSort('created_at')">Created <span class="sort-icon">{{ sortBy === 'created_at' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
          <th class="col-center sortable" :class="{ active: sortBy === 'expires_at' }" @click="toggleSort('expires_at')">Expires <span class="sort-icon">{{ sortBy === 'expires_at' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
          <th class="col-center sortable" :class="{ active: sortBy === 'last_used_at' }" @click="toggleSort('last_used_at')">Last Used <span class="sort-icon">{{ sortBy === 'last_used_at' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
          <th class="col-center sortable" :class="{ active: sortBy === 'status' }" @click="toggleSort('status')">Status <span class="sort-icon">{{ sortBy === 'status' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
          <th class="col-center">Actions</th>
        </tr>
      </thead>
      <tbody>
        <template v-if="loading">
          <tr v-for="i in 6" :key="i" class="skeleton-row">
            <td><div class="skeleton-bar skeleton-bar--md" /></td>
            <td class="col-center"><div class="skeleton-bar skeleton-bar--pill" style="margin:auto" /></td>
            <td class="col-center"><div class="skeleton-bar skeleton-bar--sm" style="margin:auto" /></td>
            <td class="col-center"><div class="skeleton-bar skeleton-bar--xs" style="margin:auto" /></td>
            <td class="col-center"><div class="skeleton-bar skeleton-bar--sm" style="margin:auto" /></td>
            <td class="col-center"><div class="skeleton-bar skeleton-bar--pill" style="margin:auto" /></td>
            <td class="col-center"><div class="skeleton-bar skeleton-bar--btn" style="margin:auto" /></td>
          </tr>
        </template>
        <template v-else>
          <tr v-for="token in tokens" :key="token.id" :class="{ dimmed: tokenStatus(token) !== 'active' }">
            <td class="bold">{{ token.name }}</td>
            <td class="col-center"><span class="user-pill">{{ token.username }}</span></td>
            <td class="col-center muted">{{ formatDate(token.createdAt) }}</td>
            <td class="col-center muted">{{ token.expiresAt ? formatDate(token.expiresAt) : '—' }}</td>
            <td class="col-center muted">{{ token.lastUsedAt ? formatDate(token.lastUsedAt) : 'Never' }}</td>
            <td class="col-center"><span class="badge" :class="`badge-tok-${tokenStatus(token)}`">{{ tokenStatus(token) }}</span></td>
            <td class="col-center actions">
              <button
                v-if="tokenStatus(token) === 'active'"
                class="btn btn-sm btn-danger"
                :disabled="actioningId === token.id"
                @click="openConfirm(token, 'revoke')"
              >{{ actioningId === token.id ? 'Revoking…' : 'Revoke' }}</button>
              <button
                v-else-if="tokenStatus(token) === 'revoked'"
                class="btn btn-sm btn-secondary"
                :disabled="actioningId === token.id"
                @click="openConfirm(token, 'unrevoke')"
              >{{ actioningId === token.id ? 'Unrevoking…' : 'Unrevoke' }}</button>
              <span v-else class="muted">—</span>
            </td>
          </tr>
        </template>
      </tbody>
    </table>

    <PaginationBar v-if="!loading" v-model:page="page" v-model:pageSize="pageSize" :total="total" />
    </div>

    <RevokeTokenModal
      v-if="confirmToken"
      :token-name="confirmToken.name"
      :action="confirmAction"
      :on-confirm="doAction"
      @done="closeConfirm"
      @close="closeConfirm"
    />
  </div>
</template>

<style scoped>
</style>
