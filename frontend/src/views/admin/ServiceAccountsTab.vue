<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import {
  listServiceAccounts, createServiceAccount, deleteServiceAccount,
  listServiceTokens, createServiceToken, adminRevokeToken,
} from '@/services/api'
import type { ServiceAccount, ClientToken } from '@/services/api'
import { formatDate, fmtNum, tokenStatus } from '@/utils/format'
import { useMinLoad } from '@/composables/useMinLoad'

const accounts        = ref<ServiceAccount[]>([])
const { loading, withLoad }                   = useMinLoad(300, true)
const { loading: tokensLoading, withLoad: withTokensLoad } = useMinLoad()
const error           = ref<string | null>(null)

const search = ref('')
let searchTimer: ReturnType<typeof setTimeout> | null = null
watch(search, () => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => loadAccounts(), 300)
})

const selectedAccount = ref<ServiceAccount | null>(null)
const tokens          = ref<ClientToken[]>([])
const showInactive     = ref(false)
const revokingId      = ref<string | null>(null)

const sortBy     = ref('created_at')
const sortDir    = ref<'asc' | 'desc'>('desc')
const tokSortBy  = ref('created_at')
const tokSortDir = ref<'asc' | 'desc'>('desc')

function toggleSort(col: string) {
  if (sortBy.value === col) { sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc' }
  else { sortBy.value = col; sortDir.value = 'desc' }
  loadAccounts()
}

function toggleTokSort(col: string) {
  if (tokSortBy.value === col) { tokSortDir.value = tokSortDir.value === 'asc' ? 'desc' : 'asc' }
  else { tokSortBy.value = col; tokSortDir.value = 'desc' }
  loadTokens()
}

const showCreateAccountModal = ref(false)
const showCreateTokenModal   = ref(false)
const newAccountForm  = ref({ username: '', description: '' })
const newTokenForm    = ref({ name: '', durationDays: 90 })
const saving          = ref(false)
const formError       = ref<string | null>(null)

const lastCreatedToken = ref<string | null>(null)
const deleteTarget     = ref<ServiceAccount | null>(null)
const deleting         = ref(false)

async function loadAccounts() {
  error.value = null
  await withLoad(async () => {
    try {
      const res = await listServiceAccounts(sortBy.value, sortDir.value, search.value)
      accounts.value = res.data.serviceAccounts ?? []
    } catch {
      error.value = 'Failed to load service accounts'
    }
  })
}

async function loadTokens() {
  if (!selectedAccount.value) return
  await withTokensLoad(async () => {
    const res = await listServiceTokens(selectedAccount.value!.id, showInactive.value, tokSortBy.value, tokSortDir.value)
    tokens.value = res.data.tokens ?? []
  })
}

async function selectAccount(account: ServiceAccount) {
  if (selectedAccount.value?.id === account.id) {
    selectedAccount.value = null
    tokens.value = []
    return
  }
  selectedAccount.value = account
  tokens.value = []
  showInactive.value = false
  await loadTokens()
}

async function handleCreateAccount() {
  if (!newAccountForm.value.username.trim()) return
  saving.value = true
  formError.value = null
  try {
    await createServiceAccount(newAccountForm.value.username.trim(), newAccountForm.value.description.trim())
    showCreateAccountModal.value = false
    newAccountForm.value = { username: '', description: '' }
    await loadAccounts()
  } catch (e: any) {
    formError.value = e?.response?.data?.error ?? 'Failed to create service account'
  } finally {
    saving.value = false
  }
}

async function handleCreateToken() {
  if (!selectedAccount.value || !newTokenForm.value.name.trim()) return
  saving.value = true
  formError.value = null
  try {
    const res = await createServiceToken(
      selectedAccount.value.id,
      newTokenForm.value.name.trim(),
      newTokenForm.value.durationDays
    )
    showCreateTokenModal.value = false
    newTokenForm.value = { name: '', durationDays: 90 }
    lastCreatedToken.value = res.data.rawToken
    await loadTokens()
  } catch (e: any) {
    formError.value = e?.response?.data?.error ?? 'Failed to create token'
  } finally {
    saving.value = false
  }
}

async function handleRevokeToken(tokenId: string) {
  revokingId.value = tokenId
  try {
    await adminRevokeToken(tokenId)
    await loadTokens()
  } finally {
    revokingId.value = null
  }
}

async function confirmDelete() {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    await deleteServiceAccount(deleteTarget.value.id)
    if (selectedAccount.value?.id === deleteTarget.value.id) {
      selectedAccount.value = null
      tokens.value = []
    }
    deleteTarget.value = null
    await loadAccounts()
  } catch {
    error.value = 'Failed to delete service account'
  } finally {
    deleting.value = false
  }
}

async function toggleShowInactive() {
  showInactive.value = !showInactive.value
  await loadTokens()
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text)
}

function accountInitial(username: string) {
  return username.charAt(0).toUpperCase()
}

onMounted(loadAccounts)
</script>

<template>
  <Teleport defer to="#admin-search-portal">
    <div class="portal-search-wrap">
      <svg class="portal-search-icon" xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
      <input v-model="search" type="text" placeholder="Search accounts…" class="portal-input" />
    </div>
  </Teleport>

  <div class="tab-content">

    <!-- ── Service Accounts card ───────────────────────────────────────────── -->
    <div class="card">
      <div class="card-header">
        <h2 class="card-title">Service Accounts <span class="title-count">{{ accounts.length }}</span></h2>
        <div class="header-actions">
          <button class="btn btn-primary btn-sm" @click="showCreateAccountModal = true; formError = null">
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            New Service Account
          </button>
        </div>
      </div>

      <div v-if="error" class="state-msg error">{{ error }}</div>

      <div v-if="!loading && accounts.length === 0" class="empty-card">
        <div class="empty-icon">
          <svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>
          </svg>
        </div>
        <p class="empty-title">No service accounts</p>
        <p class="empty-sub">Create a service account to generate long-lived tokens for external apps.</p>
      </div>

      <table v-else class="data-table">
        <thead>
          <tr>
            <th class="sortable" :class="{ active: sortBy === 'username' }" @click="toggleSort('username')">Name <span class="sort-icon">{{ sortBy === 'username' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
            <th>Description</th>
            <th class="num sortable" :class="{ active: sortBy === 'token_count' }" @click="toggleSort('token_count')">Tokens <span class="sort-icon">{{ sortBy === 'token_count' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
            <th class="num sortable" :class="{ active: sortBy === 'total_requests' }" @click="toggleSort('total_requests')">Requests <span class="sort-icon">{{ sortBy === 'total_requests' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
            <th class="num sortable" :class="{ active: sortBy === 'total_input' }" @click="toggleSort('total_input')">Input <span class="sort-icon">{{ sortBy === 'total_input' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
            <th class="num sortable" :class="{ active: sortBy === 'total_output' }" @click="toggleSort('total_output')">Output <span class="sort-icon">{{ sortBy === 'total_output' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
            <th class="sortable" :class="{ active: sortBy === 'created_at' }" @click="toggleSort('created_at')">Created <span class="sort-icon">{{ sortBy === 'created_at' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <template v-if="loading">
            <tr v-for="i in 4" :key="i" class="skeleton-row">
              <td><div class="skeleton-bar skeleton-bar--md" /></td>
              <td><div class="skeleton-bar skeleton-bar--lg" /></td>
              <td class="num"><div class="skeleton-bar skeleton-bar--xs" style="margin:auto" /></td>
              <td class="num"><div class="skeleton-bar skeleton-bar--xs" style="margin:auto" /></td>
              <td class="num"><div class="skeleton-bar skeleton-bar--xs" style="margin:auto" /></td>
              <td class="num"><div class="skeleton-bar skeleton-bar--xs" style="margin:auto" /></td>
              <td><div class="skeleton-bar skeleton-bar--sm" /></td>
              <td class="actions"><div class="skeleton-bar skeleton-bar--btn" /></td>
            </tr>
          </template>
          <template v-else>
            <tr
              v-for="a in accounts"
              :key="a.id"
              class="account-row"
              :class="{ 'row-selected': selectedAccount?.id === a.id }"
              @click="selectAccount(a)"
            >
              <td>
                <div class="account-name-cell">
                  <span class="account-initial">{{ accountInitial(a.username) }}</span>
                  <span class="bold">{{ a.username }}</span>
                </div>
              </td>
              <td class="muted">{{ a.description || '—' }}</td>
              <td class="num">{{ fmtNum(a.tokenCount) }}</td>
              <td class="num">{{ fmtNum(a.totalRequests) }}</td>
              <td class="num">{{ fmtNum(a.totalInput) }}</td>
              <td class="num">{{ fmtNum(a.totalOutput) }}</td>
              <td class="muted">{{ formatDate(a.createdAt) }}</td>
              <td class="actions" @click.stop>
                <button class="btn btn-sm btn-danger" @click="deleteTarget = a">
                  <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"/></svg>
                  Delete
                </button>
                <span v-if="selectedAccount?.id === a.id" class="row-chevron">▶</span>
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </div>

    <!-- ── Tokens card (separate) ──────────────────────────────────────────── -->
    <Transition name="token-card">
      <div v-if="selectedAccount" class="card">
        <div class="card-header">
          <div class="token-card-heading">
            <div class="account-avatar">{{ accountInitial(selectedAccount.username) }}</div>
            <div>
              <h2 class="card-title">Tokens <span class="title-count">{{ tokens.length }}</span></h2>
              <p class="card-subtitle">for <strong>{{ selectedAccount.username }}</strong></p>
            </div>
          </div>
          <div class="header-actions">
            <label class="toggle-label">
              <input type="checkbox" :checked="showInactive" @change="toggleShowInactive" />
              <span class="toggle-switch" />
              Show inactive
            </label>
            <button class="btn btn-sm btn-primary" @click="showCreateTokenModal = true; formError = null">
              <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
              New Token
            </button>
            <button class="btn btn-sm btn-outline" @click="selectedAccount = null; tokens = []">
              <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="19" y1="12" x2="5" y2="12"/><polyline points="12 19 5 12 12 5"/></svg>
              Close
            </button>
          </div>
        </div>

        <div v-if="!tokensLoading && tokens.length === 0" class="empty-card">
          <div class="empty-icon">
            <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>
            </svg>
          </div>
          <p class="empty-title">No tokens yet</p>
          <p class="empty-sub">Generate a token so <strong>{{ selectedAccount.username }}</strong> can authenticate with the API.</p>
          <button class="btn btn-primary btn-sm" style="margin-top:0.25rem" @click="showCreateTokenModal = true; formError = null">
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            New Token
          </button>
        </div>

        <table v-else class="data-table">
          <thead>
            <tr>
              <th class="sortable" :class="{ active: tokSortBy === 'name' }" @click="toggleTokSort('name')">Name <span class="sort-icon">{{ tokSortBy === 'name' ? (tokSortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
              <th class="sortable" :class="{ active: tokSortBy === 'created_at' }" @click="toggleTokSort('created_at')">Created <span class="sort-icon">{{ tokSortBy === 'created_at' ? (tokSortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
              <th class="sortable" :class="{ active: tokSortBy === 'expires_at' }" @click="toggleTokSort('expires_at')">Expires <span class="sort-icon">{{ tokSortBy === 'expires_at' ? (tokSortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
              <th class="sortable" :class="{ active: tokSortBy === 'last_used_at' }" @click="toggleTokSort('last_used_at')">Last Used <span class="sort-icon">{{ tokSortBy === 'last_used_at' ? (tokSortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <template v-if="tokensLoading">
              <tr v-for="i in 3" :key="i" class="skeleton-row">
                <td><div class="skeleton-bar skeleton-bar--md" /></td>
                <td><div class="skeleton-bar skeleton-bar--sm" /></td>
                <td><div class="skeleton-bar skeleton-bar--xs" /></td>
                <td><div class="skeleton-bar skeleton-bar--sm" /></td>
                <td><div class="skeleton-bar skeleton-bar--pill" /></td>
                <td><div class="skeleton-bar skeleton-bar--btn" /></td>
              </tr>
            </template>
            <template v-else>
              <tr v-for="t in tokens" :key="t.id" :class="{ dimmed: tokenStatus(t) !== 'active' }">
                <td class="bold">{{ t.name }}</td>
                <td class="muted">{{ formatDate(t.createdAt) }}</td>
                <td class="muted">{{ t.expiresAt ? formatDate(t.expiresAt) : '—' }}</td>
                <td class="muted">{{ t.lastUsedAt ? formatDate(t.lastUsedAt) : 'Never' }}</td>
                <td><span class="badge" :class="`badge-tok-${tokenStatus(t)}`">{{ tokenStatus(t) }}</span></td>
                <td class="actions">
                  <button
                    v-if="tokenStatus(t) === 'active'"
                    class="btn btn-sm btn-danger"
                    :disabled="revokingId === t.id"
                    @click="handleRevokeToken(t.id)"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="4.93" y1="4.93" x2="19.07" y2="19.07"/></svg>
                    {{ revokingId === t.id ? 'Revoking…' : 'Revoke' }}
                  </button>
                  <span v-else class="muted" style="font-size:0.8rem">—</span>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>
    </Transition>

  </div>

  <!-- ── Create Service Account Modal ───────────────────────────────────────── -->
  <Teleport to="body">
    <div v-if="showCreateAccountModal" class="modal-backdrop" @click.self="showCreateAccountModal = false">
      <div class="modal">
        <h3>New Service Account</h3>
        <div class="form-group">
          <label>Name <span class="required">*</span></label>
          <input
            v-model="newAccountForm.username"
            type="text"
            class="text-input"
            placeholder="e.g. ci-pipeline"
            maxlength="100"
            @keydown.enter="handleCreateAccount"
          />
        </div>
        <div class="form-group">
          <label>Description <span class="optional">(optional)</span></label>
          <input
            v-model="newAccountForm.description"
            type="text"
            class="text-input"
            placeholder="e.g. Used by the CI pipeline"
            maxlength="255"
          />
        </div>
        <p v-if="formError" class="form-error">{{ formError }}</p>
        <div class="modal-actions">
          <button class="btn-cancel" @click="showCreateAccountModal = false">
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            Cancel
          </button>
          <button
            class="btn-primary"
            :disabled="saving || !newAccountForm.username.trim()"
            @click="handleCreateAccount"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            {{ saving ? 'Creating…' : 'Create' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>

  <!-- ── Create Token Modal ─────────────────────────────────────────────────── -->
  <Teleport to="body">
    <div v-if="showCreateTokenModal" class="modal-backdrop" @click.self="showCreateTokenModal = false">
      <div class="modal">
        <h3>New Token</h3>
        <p class="modal-sub">Account: <strong>{{ selectedAccount?.username }}</strong></p>
        <div class="form-group">
          <label>Token Name <span class="required">*</span></label>
          <input
            v-model="newTokenForm.name"
            type="text"
            class="text-input"
            placeholder="e.g. prod-key"
            maxlength="100"
          />
        </div>
        <div class="form-group">
          <label>Duration (days) <span class="hint">max 365</span></label>
          <input
            v-model.number="newTokenForm.durationDays"
            type="number"
            class="text-input"
            min="1"
            max="365"
          />
        </div>
        <p v-if="formError" class="form-error">{{ formError }}</p>
        <div class="modal-actions">
          <button class="btn-cancel" @click="showCreateTokenModal = false">
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            Cancel
          </button>
          <button
            class="btn-primary"
            :disabled="saving || !newTokenForm.name.trim() || newTokenForm.durationDays < 1 || newTokenForm.durationDays > 365"
            @click="handleCreateToken"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            {{ saving ? 'Creating…' : 'Create Token' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>

  <!-- ── Raw Token Display Modal (one-time) ────────────────────────────────── -->
  <Teleport to="body">
    <div v-if="lastCreatedToken" class="modal-backdrop">
      <div class="modal modal-wide">
        <h3>Token Created</h3>
        <div class="token-warning">
          This token will not be shown again. Copy it now and store it securely.
        </div>
        <div class="token-display">
          <code class="token-code">{{ lastCreatedToken }}</code>
          <button class="btn btn-sm btn-copy" @click="copyToClipboard(lastCreatedToken!)">
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
            Copy
          </button>
        </div>
        <div class="modal-actions">
          <button class="btn-primary" @click="lastCreatedToken = null">
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
            Done, I've copied it
          </button>
        </div>
      </div>
    </div>
  </Teleport>

  <!-- ── Delete Confirm Modal ───────────────────────────────────────────────── -->
  <Teleport to="body">
    <div v-if="deleteTarget" class="modal-backdrop" @click.self="deleteTarget = null">
      <div class="modal">
        <h3>Delete Service Account</h3>
        <p class="modal-sub">
          Delete <strong>{{ deleteTarget.username }}</strong>? This will also revoke all its tokens. This action cannot be undone.
        </p>
        <div class="modal-actions">
          <button class="btn-cancel" @click="deleteTarget = null">
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            Cancel
          </button>
          <button class="btn-danger-solid" :disabled="deleting" @click="confirmDelete">
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"/></svg>
            {{ deleting ? 'Deleting…' : 'Delete' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.hint { font-weight: 400; font-size: 0.78rem; color: #94a3b8; }

/* Clickable account rows */
.account-row { cursor: pointer; transition: background 0.1s; }
.account-row:hover:not(.row-selected) td { background: #f8fafc; }

.account-name-cell { display: flex; align-items: center; gap: 0.6rem; }
.account-initial {
  display: inline-flex; align-items: center; justify-content: center;
  width: 28px; height: 28px; border-radius: 50%;
  background: #dbeafe; color: #1d4ed8;
  font-size: 0.75rem; font-weight: 700; flex-shrink: 0;
}
.row-chevron { color: #3b82f6; font-size: 0.65rem; }

/* Token card heading */
.token-card-heading { display: flex; align-items: center; gap: 0.9rem; }
.account-avatar {
  display: flex; align-items: center; justify-content: center;
  width: 40px; height: 40px; border-radius: 50%;
  background: #dbeafe; color: #1d4ed8;
  font-size: 1rem; font-weight: 700; flex-shrink: 0;
}
.card-subtitle { font-size: 0.82rem; color: #64748b; margin: 0.1rem 0 0; }

/* Token display modal */
.token-warning {
  background: #fef3c7; border: 1px solid #fde68a; border-radius: 8px;
  padding: 0.75rem 1rem; font-size: 0.85rem; color: #92400e; font-weight: 500;
}
.token-display {
  display: flex; align-items: flex-start; gap: 0.6rem;
  background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 8px; padding: 0.75rem;
}
.token-code { flex: 1; font-family: monospace; font-size: 0.8rem; word-break: break-all; color: #1e293b; }

/* Token card slide-in transition */
.token-card-enter-active { transition: opacity 0.2s ease, transform 0.2s ease; }
.token-card-leave-active { transition: opacity 0.15s ease, transform 0.15s ease; }
.token-card-enter-from  { opacity: 0; transform: translateY(-6px); }
.token-card-leave-to    { opacity: 0; transform: translateY(-4px); }
</style>
