<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  listServiceAccounts, createServiceAccount, deleteServiceAccount,
  listServiceTokens, createServiceToken, adminRevokeToken,
} from '@/services/api'
import type { ServiceAccount, ClientToken } from '@/services/api'
import { formatDate, tokenStatus } from '@/utils/format'

const accounts        = ref<ServiceAccount[]>([])
const loading         = ref(true)
const error           = ref<string | null>(null)

const selectedAccount = ref<ServiceAccount | null>(null)
const tokens          = ref<ClientToken[]>([])
const tokensLoading   = ref(false)
const showRevoked     = ref(false)
const revokingId      = ref<string | null>(null)

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
  loading.value = true
  error.value = null
  try {
    const res = await listServiceAccounts()
    accounts.value = res.data.serviceAccounts ?? []
  } catch {
    error.value = 'Failed to load service accounts'
  } finally {
    loading.value = false
  }
}

async function loadTokens() {
  if (!selectedAccount.value) return
  tokensLoading.value = true
  try {
    const res = await listServiceTokens(selectedAccount.value.id, showRevoked.value)
    tokens.value = res.data.tokens ?? []
  } finally {
    tokensLoading.value = false
  }
}

async function selectAccount(account: ServiceAccount) {
  selectedAccount.value = account
  tokens.value = []
  showRevoked.value = false
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

async function toggleShowRevoked() {
  showRevoked.value = !showRevoked.value
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

    <!-- Loading -->
    <div v-if="loading" class="empty-card">
      <p class="empty-title">Loading…</p>
    </div>

    <!-- Empty -->
    <div v-else-if="accounts.length === 0 && !loading" class="empty-card">
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
          <th>Name</th>
          <th>Description</th>
          <th>Created</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
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
            <input type="checkbox" :checked="showRevoked" @change="toggleShowRevoked" />
            Show revoked
          </label>
          <button class="btn btn-sm btn-primary" @click="showCreateTokenModal = true; formError = null">
            + New Token
          </button>
          <button class="btn btn-sm btn-outline" @click="selectedAccount = null">Close</button>
        </div>
      </div>

      <div v-if="tokensLoading" class="panel-loading">Loading tokens…</div>
      <div v-else-if="tokens.length === 0" class="panel-empty">No tokens yet.</div>
      <table v-else class="data-table">
        <thead>
          <tr>
            <th>Name</th><th>Created</th><th>Expires</th><th>Last Used</th><th>Status</th><th>Actions</th>
          </tr>
        </thead>
        <tbody>
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
.tab-content { display: flex; flex-direction: column; gap: 1.25rem; }
.toolbar { display: flex; align-items: center; justify-content: space-between; gap: 1rem; flex-wrap: wrap; }
.sub { font-size: 0.85rem; color: #64748b; margin: 0; }
.state-msg { font-size: 0.9rem; }
.state-msg.error { color: #ef4444; }

/* Empty state */
.empty-card { display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 0.6rem; padding: 3rem 2rem; background: white; border: 1px dashed #e2e8f0; border-radius: 12px; text-align: center; }
.empty-icon { display: flex; align-items: center; justify-content: center; width: 64px; height: 64px; border-radius: 16px; background: #f1f5f9; color: #94a3b8; margin-bottom: 0.25rem; }
.empty-title { font-size: 1rem; font-weight: 600; color: #1e293b; margin: 0; }
.empty-sub { font-size: 0.85rem; color: #94a3b8; margin: 0; max-width: 320px; line-height: 1.5; }

/* Table */
.data-table { width: 100%; border-collapse: collapse; font-size: 0.88rem; background: white; border: 1px solid #e2e8f0; border-radius: 10px; }
.data-table th { text-align: left; padding: 0.55rem 0.9rem; font-size: 0.75rem; font-weight: 600; color: #64748b; text-transform: uppercase; letter-spacing: 0.03em; border-bottom: 1px solid #e2e8f0; background: #f8fafc; }
.data-table thead tr:first-child th:first-child { border-radius: 10px 0 0 0; }
.data-table thead tr:first-child th:last-child  { border-radius: 0 10px 0 0; }
.data-table tbody tr:last-child td:first-child   { border-radius: 0 0 0 10px; }
.data-table tbody tr:last-child td:last-child    { border-radius: 0 0 10px 0; }
.data-table td { padding: 0.65rem 0.9rem; border-bottom: 1px solid #f1f5f9; }
.data-table tr:last-child td { border-bottom: none; }
.data-table tr.row-selected { background: #f0f9ff; }
.data-table tr.dimmed td { opacity: 0.5; }
.bold { font-weight: 600; color: #1e293b; }
.muted { color: #64748b; }
.actions { display: flex; align-items: center; gap: 0.4rem; white-space: nowrap; }

/* Badges */
.badge { padding: 0.18rem 0.55rem; border-radius: 999px; font-size: 0.75rem; font-weight: 600; }
.badge-tok-active  { background: #dcfce7; color: #166534; }
.badge-tok-revoked { background: #f1f5f9; color: #64748b; }
.badge-tok-expired { background: #fef3c7; color: #92400e; }

/* Buttons */
.btn { padding: 0.45rem 1rem; border-radius: 6px; border: none; cursor: pointer; font-size: 0.9rem; font-weight: 500; transition: background 0.12s; }
.btn-sm { padding: 0.25rem 0.65rem; font-size: 0.8rem; }
.btn-primary { background: #3b82f6; color: white; }
.btn-primary:hover:not(:disabled) { background: #2563eb; }
.btn-primary:disabled { opacity: 0.55; cursor: not-allowed; }
.btn-danger { background: #fee2e2; color: #dc2626; }
.btn-danger:hover:not(:disabled) { background: #fecaca; }
.btn-danger:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-outline { background: white; color: #374151; border: 1px solid #e2e8f0; }
.btn-outline:hover { background: #f8fafc; }
.btn-copy { background: #f1f5f9; color: #374151; }
.btn-copy:hover { background: #e2e8f0; }
.btn-danger-solid { padding: 0.4rem 1rem; border: none; border-radius: 6px; background: #dc2626; color: white; font-size: 0.88rem; font-weight: 600; cursor: pointer; }
.btn-danger-solid:hover:not(:disabled) { background: #b91c1c; }
.btn-danger-solid:disabled { opacity: 0.55; cursor: not-allowed; }

/* Token sub-panel */
.token-panel { border: 1px solid #e2e8f0; border-radius: 10px; overflow: hidden; }
.panel-header { display: flex; align-items: center; justify-content: space-between; gap: 1rem; flex-wrap: wrap; padding: 0.75rem 1rem; background: #f8fafc; border-bottom: 1px solid #e2e8f0; }
.panel-title { display: flex; align-items: center; gap: 0.4rem; font-size: 0.9rem; }
.panel-label { color: #64748b; }
.panel-actions { display: flex; align-items: center; gap: 0.6rem; flex-wrap: wrap; }
.panel-loading { padding: 1rem; color: #64748b; font-size: 0.88rem; }
.panel-empty { padding: 1rem; color: #94a3b8; font-size: 0.88rem; font-style: italic; }
.toggle-label { display: flex; align-items: center; gap: 0.4rem; font-size: 0.85rem; font-weight: 500; color: #64748b; cursor: pointer; user-select: none; white-space: nowrap; }
.toggle-label input[type="checkbox"] { cursor: pointer; accent-color: #3b82f6; width: 14px; height: 14px; }

/* Modal */
.modal-backdrop { position: fixed; inset: 0; background: rgba(0,0,0,0.45); display: flex; align-items: center; justify-content: center; z-index: 200; }
.modal { background: white; border-radius: 12px; padding: 1.75rem; width: 100%; max-width: 400px; display: flex; flex-direction: column; gap: 1rem; }
.modal-wide { max-width: 560px; }
.modal h3 { font-size: 1.1rem; font-weight: 700; margin: 0; }
.modal-sub { font-size: 0.9rem; color: #475569; margin: 0; }
.form-group { display: flex; flex-direction: column; gap: 0.35rem; }
.form-group label { font-size: 0.85rem; font-weight: 600; color: #374151; }
.required { color: #ef4444; }
.optional { font-weight: 400; color: #94a3b8; }
.hint { font-weight: 400; font-size: 0.78rem; color: #94a3b8; }
.text-input { padding: 0.45rem 0.6rem; border: 1px solid #d1d5db; border-radius: 6px; font-size: 0.9rem; background: white; outline: none; width: 100%; box-sizing: border-box; }
.text-input:focus { border-color: #3b82f6; box-shadow: 0 0 0 2px #bfdbfe; }
.form-error { color: #ef4444; font-size: 0.85rem; margin: 0; }
.modal-actions { display: flex; justify-content: flex-end; gap: 0.6rem; }
.btn-cancel { padding: 0.4rem 1rem; border: 1px solid #e2e8f0; border-radius: 6px; background: white; color: #374151; font-size: 0.88rem; cursor: pointer; }
.btn-cancel:hover { background: #f8fafc; }

/* Raw token display */
.token-warning { background: #fef3c7; border: 1px solid #fde68a; border-radius: 8px; padding: 0.75rem 1rem; font-size: 0.85rem; color: #92400e; font-weight: 500; }
.token-display { display: flex; align-items: flex-start; gap: 0.6rem; background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 8px; padding: 0.75rem; }
.token-code { flex: 1; font-family: monospace; font-size: 0.8rem; word-break: break-all; color: #1e293b; }
</style>
