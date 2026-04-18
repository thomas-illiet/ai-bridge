<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useTokenStore } from '@/stores/tokens'
import type { ClientToken, CreateTokenResponse } from '@/services/api'
import { formatDate, tokenStatus, isExpiringSoon } from '@/utils/format'
import PaginationBar from '@/components/PaginationBar.vue'
import CreateTokenModal from '@/views/tokens/CreateTokenModal.vue'
import EditTokenModal from '@/views/tokens/EditTokenModal.vue'
import RevokeTokenModal from '@/views/tokens/RevokeTokenModal.vue'
import TokenCreatedModal from '@/views/tokens/TokenCreatedModal.vue'

const store = useTokenStore()

const showRevoked = ref(false)
const tokPage     = ref(1)
const tokPageSize = ref(10)

watch(showRevoked, (val) => {
  tokPage.value = 1
  store.fetchTokens(val, sortBy.value, sortDir.value)
})
watch(tokPageSize, () => { tokPage.value = 1 })
const pagedTokens = computed(() => {
  const start = (tokPage.value - 1) * tokPageSize.value
  return store.tokens.slice(start, start + tokPageSize.value)
})

const expiringTokens = computed(() =>
  store.tokens.filter((t) => isExpiringSoon(t))
)

const kpis = computed(() => {
  const all = store.tokens
  const active = all.filter((t) => tokenStatus(t) === 'active')
  const neverUsed = active.filter((t) => !t.lastUsedAt).length
  const lastUsedAt = all
    .map((t) => t.lastUsedAt)
    .filter(Boolean)
    .map((d) => new Date(d!).getTime())
    .reduce((max, v) => (v > max ? v : max), 0)
  return {
    active: active.length,
    expiringSoon: expiringTokens.value.length,
    neverUsed,
    lastUsedAt: lastUsedAt ? relativeDate(lastUsedAt) : null,
  }
})

function relativeDate(ts: number): string {
  const diff = Date.now() - ts
  const mins = Math.floor(diff / 60_000)
  if (mins < 1)  return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24)  return `${hrs}h ago`
  return `${Math.floor(hrs / 24)}d ago`
}

const sortBy  = ref('created_at')
const sortDir = ref<'asc' | 'desc'>('desc')

function toggleSort(col: string) {
  if (sortBy.value === col) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortBy.value = col
    sortDir.value = 'desc'
  }
  tokPage.value = 1
  store.fetchTokens(showRevoked.value, sortBy.value, sortDir.value)
}

const showCreateModal    = ref(false)
const editingToken       = ref<ClientToken | null>(null)
const revokingToken      = ref<ClientToken | null>(null)
const createdTokenResult = ref<CreateTokenResponse | null>(null)

const openMenuId = ref<string | null>(null)
function toggleMenu(id: string) { openMenuId.value = openMenuId.value === id ? null : id }
function closeMenus() { openMenuId.value = null }
function onDocClick() { closeMenus() }

onMounted(() => { document.addEventListener('click', onDocClick); store.fetchTokens(false, sortBy.value, sortDir.value) })
onBeforeUnmount(() => document.removeEventListener('click', onDocClick))

function onTokenCreated(result: CreateTokenResponse) {
  createdTokenResult.value = result
  showCreateModal.value = false
}

function dismissCreatedToken() {
  createdTokenResult.value = null
  store.fetchTokens()
}

function daysUntilExpiry(token: ClientToken): number {
  if (!token.expiresAt) return Infinity
  return Math.ceil((new Date(token.expiresAt).getTime() - Date.now()) / 86400000)
}
</script>

<template>
  <div class="tokens-page">
    <div class="page-header">
      <h1>Personal Access Tokens</h1>
      <div class="header-actions">
        <label class="toggle-label">
          <input type="checkbox" v-model="showRevoked" />
          <span class="toggle-switch" />
          Show revoked
        </label>
        <button class="btn btn-primary" @click="showCreateModal = true">New Token</button>
      </div>
    </div>

    <div v-if="!store.loading && !store.error" class="stat-grid">
      <div class="stat-card">
        <span class="stat-label">Active</span>
        <span class="stat-value">{{ kpis.active }}</span>
      </div>
      <div class="stat-card" :class="{ 'stat-card--warn': kpis.expiringSoon > 0 }">
        <span class="stat-label">Expiring soon</span>
        <span class="stat-value">{{ kpis.expiringSoon }}</span>
      </div>
      <div class="stat-card">
        <span class="stat-label">Never used</span>
        <span class="stat-value">{{ kpis.neverUsed }}</span>
      </div>
      <div class="stat-card stat-card--last">
        <span class="stat-label">Last used</span>
        <span v-if="kpis.lastUsedAt" class="stat-value stat-value--sm">{{ kpis.lastUsedAt }}</span>
        <span v-else class="stat-empty">No activity yet</span>
      </div>
    </div>

    <div v-if="expiringTokens.length > 0" class="expiry-banner">
      <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/>
      </svg>
      <div>
        <span v-if="expiringTokens.length === 1">
          Token <strong>{{ expiringTokens[0].name }}</strong> expires in {{ daysUntilExpiry(expiringTokens[0]) }} day{{ daysUntilExpiry(expiringTokens[0]) === 1 ? '' : 's' }} — revoke it and create a new one to avoid service interruption.
        </span>
        <span v-else>
          {{ expiringTokens.length }} tokens expire within 3 days:
          <strong>{{ expiringTokens.map(t => t.name).join(', ') }}</strong>
        </span>
      </div>
    </div>

    <div v-if="!store.loading && store.error" class="empty-card empty-card--error">
      <div class="empty-icon empty-icon--error">
        <svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
        </svg>
      </div>
      <p class="empty-title">Failed to load tokens</p>
      <p class="empty-sub">{{ store.error }}</p>
    </div>
    <div v-else-if="!store.loading && store.tokens.length === 0 && !createdTokenResult" class="empty-card">
      <div class="empty-icon">
        <svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4"/>
        </svg>
      </div>
      <p class="empty-title">No tokens yet</p>
      <p class="empty-sub">Create a personal access token to authenticate API requests programmatically.</p>
    </div>

    <table v-else class="data-table">
      <thead>
        <tr>
          <th class="sortable" :class="{ active: sortBy === 'name' }" @click="toggleSort('name')">Name <span class="sort-icon">{{ sortBy === 'name' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
          <th class="col-center sortable" :class="{ active: sortBy === 'created_at' }" @click="toggleSort('created_at')">Created <span class="sort-icon">{{ sortBy === 'created_at' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
          <th class="col-center sortable" :class="{ active: sortBy === 'expires_at' }" @click="toggleSort('expires_at')">Expires <span class="sort-icon">{{ sortBy === 'expires_at' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
          <th class="col-center sortable" :class="{ active: sortBy === 'last_used_at' }" @click="toggleSort('last_used_at')">Last Used <span class="sort-icon">{{ sortBy === 'last_used_at' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
          <th class="col-center sortable" :class="{ active: sortBy === 'status' }" @click="toggleSort('status')">Status <span class="sort-icon">{{ sortBy === 'status' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
          <th class="col-center">Actions</th>
        </tr>
      </thead>
      <tbody>
        <template v-if="store.loading">
          <tr v-for="i in 5" :key="i" class="skeleton-row">
            <td><div class="skeleton-bar skeleton-bar--lg" /></td>
            <td class="col-center"><div class="skeleton-bar skeleton-bar--sm" style="margin:auto" /></td>
            <td class="col-center"><div class="skeleton-bar skeleton-bar--xs" style="margin:auto" /></td>
            <td class="col-center"><div class="skeleton-bar skeleton-bar--sm" style="margin:auto" /></td>
            <td class="col-center"><div class="skeleton-bar skeleton-bar--pill" style="margin:auto" /></td>
            <td class="col-center"><div class="skeleton-bar skeleton-bar--btn" style="margin:auto" /></td>
          </tr>
        </template>
        <template v-else>
          <tr v-for="token in pagedTokens" :key="token.id" :class="{ dimmed: tokenStatus(token) !== 'active' }">
            <td>
              <div class="token-name">{{ token.name }}</div>
              <div v-if="token.description" class="token-desc">{{ token.description }}</div>
            </td>
            <td class="col-center">{{ formatDate(token.createdAt) }}</td>
            <td class="col-center">{{ token.expiresAt ? formatDate(token.expiresAt) : '—' }}</td>
            <td class="col-center">{{ token.lastUsedAt ? formatDate(token.lastUsedAt) : 'Never' }}</td>
            <td class="col-center">
              <div class="badge-group">
                <span v-if="!isExpiringSoon(token)" :class="['badge', `badge-${tokenStatus(token)}`]">{{ tokenStatus(token) }}</span>
                <span v-else class="badge badge-warning">Expires soon</span>
              </div>
            </td>
            <td class="col-center">
              <div v-if="tokenStatus(token) === 'active'" class="action-menu">
                <button class="btn-action-trigger" @click.stop="toggleMenu(token.id)">
                  Actions <span class="chevron-down">▾</span>
                </button>
                <div v-if="openMenuId === token.id" class="action-dropdown">
                  <button class="action-item" @click="editingToken = token; closeMenus()">Edit</button>
                  <div class="action-divider" />
                  <button class="action-item danger" @click="revokingToken = token; closeMenus()">Revoke</button>
                </div>
              </div>
              <span v-else class="muted">—</span>
            </td>
          </tr>
        </template>
      </tbody>
    </table>

    <PaginationBar
      v-if="store.tokens.length > 0"
      v-model:page="tokPage"
      v-model:pageSize="tokPageSize"
      :total="store.tokens.length"
    />

    <CreateTokenModal v-if="showCreateModal" @close="showCreateModal = false" @created="onTokenCreated" />
    <TokenCreatedModal v-if="createdTokenResult" :result="createdTokenResult" @close="dismissCreatedToken" />
    <EditTokenModal v-if="editingToken" :token="editingToken" @close="editingToken = null" />
    <RevokeTokenModal
      v-if="revokingToken"
      :token-name="revokingToken.name"
      :on-confirm="() => store.deleteToken(revokingToken!.id)"
      @done="revokingToken = null"
      @close="revokingToken = null"
    />
  </div>
</template>

<style scoped>
.tokens-page { display: flex; flex-direction: column; gap: 1.5rem; }
.page-header { display: flex; align-items: center; justify-content: space-between; }
h1 { font-size: 1.75rem; font-weight: 700; }
.header-actions { display: flex; align-items: center; gap: 1rem; }
.stat-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 1rem; }
.stat-card { background: white; border: 1px solid #e2e8f0; border-radius: 10px; padding: 1.1rem 1.25rem; display: flex; flex-direction: column; gap: 0.3rem; }
.stat-card--warn { border-color: #fed7aa; background: #fff7ed; }
.stat-card--warn .stat-value { color: #ea580c; }
.stat-label { font-size: 0.78rem; font-weight: 600; color: #64748b; text-transform: uppercase; letter-spacing: 0.04em; }
.stat-value { font-size: 1.9rem; font-weight: 700; color: #0f172a; line-height: 1; }
.stat-value--sm { font-size: 1.25rem; }
.stat-empty { font-size: 0.85rem; color: #94a3b8; }
.expiry-banner {
  display: flex; align-items: flex-start; gap: 0.75rem;
  background: #fff7ed; border: 1px solid #fed7aa; border-radius: 10px;
  padding: 0.9rem 1.25rem; font-size: 0.875rem; color: #9a3412;
}
.expiry-banner svg { flex-shrink: 0; margin-top: 1px; color: #ea580c; }
.token-name { font-weight: 500; }
.token-desc { font-size: 0.78rem; color: #94a3b8; margin-top: 0.15rem; }
.badge-group { display: block; }
/* badge border-radius override: square corners for status badges in this table */
.badge { border-radius: 4px; }
</style>
