<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  listServiceAccounts, createServiceAccount, deleteServiceAccount,
  listServiceTokens, createServiceToken, adminRevokeToken,
} from '@/services/api'
import type { ServiceAccount, ClientToken } from '@/services/api'
import { formatDate, tokenStatus } from '@/utils/format'
import { useMinLoad } from '@/composables/useMinLoad'

const accounts        = ref<ServiceAccount[]>([])
const { loading, withLoad }                   = useMinLoad(300, true)
const { loading: tokensLoading, withLoad: withTokensLoad } = useMinLoad()
const error           = ref<string | null>(null)

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
      const res = await listServiceAccounts(sortBy.value, sortDir.value)
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

onMounted(loadAccounts)
</script>

<template>
  <div class="tab-content">
    <!-- Header -->
    <div class="toolbar">
      <p class="sub">{{ accounts.length }} service account{{ accounts.length !== 1 ? 's' : '' }}</p>
      <button class="btn btn-primary" @click="showCreateAccountModal = true; formError = null">
        + New Service Account
      </button>
    </div>

    <!-- Error -->
    <div v-if="error" class="state-msg error">{{ error }}</div>

    <!-- Empty -->
    <div v-if="!loading && accounts.length === 0" class="empty-card">
      <div class="empty-icon">
        <svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>
        </svg>
      </div>
      <p class="empty-title">No service accounts</p>
      <p class="empty-sub">Create a service account to generate long-lived tokens for external apps.</p>
    </div>

    <!-- Accounts table -->
    <table v-else class="data-table">
      <thead>
        <tr>
          <th class="sortable" :class="{ active: sortBy === 'username' }" @click="toggleSort('username')">Name <span class="sort-icon">{{ sortBy === 'username' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
          <th>Description</th>
          <th class="sortable" :class="{ active: sortBy === 'created_at' }" @click="toggleSort('created_at')">Created <span class="sort-icon">{{ sortBy === 'created_at' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <template v-if="loading">
          <tr v-for="i in 4" :key="i" class="skeleton-row">
            <td><div class="skeleton-bar skeleton-bar--md" /></td>
            <td><div class="skeleton-bar skeleton-bar--lg" /></td>
            <td><div class="skeleton-bar skeleton-bar--sm" /></td>
            <td class="actions"><div class="skeleton-bar skeleton-bar--btn" /></td>
          </tr>
        </template>
        <template v-else>
          <tr v-for="a in accounts" :key="a.id" :class="{ 'row-selected': selectedAccount?.id === a.id }">
            <td class="bold">{{ a.username }}</td>
            <td class="muted">{{ a.description || '—' }}</td>
            <td class="muted">{{ formatDate(a.createdAt) }}</td>
            <td class="actions">
              <button class="btn btn-sm" @click="selectAccount(a)">
                {{ selectedAccount?.id === a.id ? 'Tokens ▾' : 'Manage Tokens' }}
              </button>
              <button class="btn btn-sm btn-danger" @click="deleteTarget = a">Delete</button>
            </td>
          </tr>
        </template>
      </tbody>
    </table>

    <!-- Token sub-panel -->
    <div v-if="selectedAccount" class="token-panel">
      <div class="panel-header">
        <div class="panel-title">
          <span class="panel-label">Tokens for</span>
          <strong>{{ selectedAccount.username }}</strong>
        </div>
        <div class="panel-actions">
          <label class="toggle-label">
            <input type="checkbox" :checked="showInactive" @change="toggleShowInactive" />
            <span class="toggle-switch" />
            Show inactive
          </label>
          <button class="btn btn-sm btn-primary" @click="showCreateTokenModal = true; formError = null">
            + New Token
          </button>
          <button class="btn btn-sm btn-outline" @click="selectedAccount = null">Close</button>
        </div>
      </div>

      <div v-if="!tokensLoading && tokens.length === 0" class="panel-empty">No tokens yet.</div>
      <table v-else class="data-table">
        <thead>
          <tr>
            <th class="sortable" :class="{ active: tokSortBy === 'name' }" @click="toggleTokSort('name')">Name <span class="sort-icon">{{ tokSortBy === 'name' ? (tokSortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
            <th class="sortable" :class="{ active: tokSortBy === 'created_at' }" @click="toggleTokSort('created_at')">Created <span class="sort-icon">{{ tokSortBy === 'created_at' ? (tokSortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
            <th class="sortable" :class="{ active: tokSortBy === 'expires_at' }" @click="toggleTokSort('expires_at')">Expires <span class="sort-icon">{{ tokSortBy === 'expires_at' ? (tokSortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
            <th class="sortable" :class="{ active: tokSortBy === 'last_used_at' }" @click="toggleTokSort('last_used_at')">Last Used <span class="sort-icon">{{ tokSortBy === 'last_used_at' ? (tokSortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
            <th>Status</th><th>Actions</th>
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
                >{{ revokingId === t.id ? 'Revoking…' : 'Revoke' }}</button>
                <span v-else class="muted" style="font-size:0.8rem">—</span>
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </div>

    <!-- Create Service Account Modal -->
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
            <button class="btn-cancel" @click="showCreateAccountModal = false">Cancel</button>
            <button
              class="btn-primary"
              :disabled="saving || !newAccountForm.username.trim()"
              @click="handleCreateAccount"
            >{{ saving ? 'Creating…' : 'Create' }}</button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Create Token Modal -->
    <Teleport to="body">
      <div v-if="showCreateTokenModal" class="modal-backdrop" @click.self="showCreateTokenModal = false">
        <div class="modal">
          <h3>New Service Token</h3>
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
            <button class="btn-cancel" @click="showCreateTokenModal = false">Cancel</button>
            <button
              class="btn-primary"
              :disabled="saving || !newTokenForm.name.trim() || newTokenForm.durationDays < 1 || newTokenForm.durationDays > 365"
              @click="handleCreateToken"
            >{{ saving ? 'Creating…' : 'Create Token' }}</button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Raw Token Display Modal (one-time) -->
    <Teleport to="body">
      <div v-if="lastCreatedToken" class="modal-backdrop">
        <div class="modal modal-wide">
          <h3>Token Created</h3>
          <div class="token-warning">
            This token will not be shown again. Copy it now and store it securely.
          </div>
          <div class="token-display">
            <code class="token-code">{{ lastCreatedToken }}</code>
            <button class="btn btn-sm btn-copy" @click="copyToClipboard(lastCreatedToken!)">Copy</button>
          </div>
          <div class="modal-actions">
            <button class="btn-primary" @click="lastCreatedToken = null">Done, I've copied it</button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Delete Confirm Modal -->
    <Teleport to="body">
      <div v-if="deleteTarget" class="modal-backdrop" @click.self="deleteTarget = null">
        <div class="modal">
          <h3>Delete Service Account</h3>
          <p class="modal-sub">
            Delete <strong>{{ deleteTarget.username }}</strong>? This will also revoke all its tokens. This action cannot be undone.
          </p>
          <div class="modal-actions">
            <button class="btn-cancel" @click="deleteTarget = null">Cancel</button>
            <button class="btn-danger-solid" :disabled="deleting" @click="confirmDelete">
              {{ deleting ? 'Deleting…' : 'Delete' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.hint { font-weight: 400; font-size: 0.78rem; color: #94a3b8; }
.token-panel { border: 1px solid #e2e8f0; border-radius: 10px; overflow: hidden; }
.panel-header { display: flex; align-items: center; justify-content: space-between; gap: 1rem; flex-wrap: wrap; padding: 0.75rem 1rem; background: #f8fafc; border-bottom: 1px solid #e2e8f0; }
.panel-title { display: flex; align-items: center; gap: 0.4rem; font-size: 0.9rem; }
.panel-label { color: #64748b; }
.panel-actions { display: flex; align-items: center; gap: 0.6rem; flex-wrap: wrap; }
.panel-loading { padding: 1rem; color: #64748b; font-size: 0.88rem; }
.panel-empty { padding: 1rem; color: #94a3b8; font-size: 0.88rem; font-style: italic; }
.token-warning { background: #fef3c7; border: 1px solid #fde68a; border-radius: 8px; padding: 0.75rem 1rem; font-size: 0.85rem; color: #92400e; font-weight: 500; }
.token-display { display: flex; align-items: flex-start; gap: 0.6rem; background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 8px; padding: 0.75rem; }
.token-code { flex: 1; font-family: monospace; font-size: 0.8rem; word-break: break-all; color: #1e293b; }
</style>
