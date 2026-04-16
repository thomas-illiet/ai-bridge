<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useTokenStore } from '@/stores/tokens'
import type { CreateTokenResponse } from '@/services/api'

const store = useTokenStore()

const showCreateModal = ref(false)
const newTokenName = ref('')
const creating = ref(false)
const createError = ref<string | null>(null)

const createdTokenResult = ref<CreateTokenResponse | null>(null)
const tokenCopied = ref(false)

const revokingId = ref<string | null>(null)

onMounted(() => store.fetchTokens())

async function handleCreate() {
  creating.value = true
  createError.value = null
  try {
    createdTokenResult.value = await store.generateToken(newTokenName.value.trim())
    newTokenName.value = ''
    showCreateModal.value = false
  } catch {
    createError.value = 'Failed to create token'
  } finally {
    creating.value = false
  }
}

function dismissCreatedToken() {
  createdTokenResult.value = null
  tokenCopied.value = false
  store.fetchTokens()
}

async function copyToken() {
  if (createdTokenResult.value) {
    await navigator.clipboard.writeText(createdTokenResult.value.rawToken)
    tokenCopied.value = true
  }
}

async function handleRevoke(id: string) {
  revokingId.value = id
  try {
    await store.deleteToken(id)
  } finally {
    revokingId.value = null
  }
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}
</script>

<template>
  <div class="tokens-page">
    <div class="page-header">
      <h1>Personal Access Tokens</h1>
      <button class="btn btn-primary" @click="showCreateModal = true">New Token</button>
    </div>

    <div v-if="createdTokenResult" class="token-banner">
      <div class="banner-warning">
        Token created — copy it now. It will <strong>not</strong> be shown again.
      </div>
      <div class="token-display">
        <code class="token-value">{{ createdTokenResult.rawToken }}</code>
        <button class="btn btn-sm" @click="copyToken">
          {{ tokenCopied ? 'Copied!' : 'Copy' }}
        </button>
      </div>
      <button class="btn btn-outline btn-sm" @click="dismissCreatedToken">
        I have saved my token
      </button>
    </div>

    <div v-if="store.loading" class="state-msg">Loading...</div>
    <div v-else-if="store.error" class="state-msg error">{{ store.error }}</div>
    <div v-else-if="store.tokens.length === 0 && !createdTokenResult" class="state-msg">
      No tokens yet. Create one to authenticate API requests programmatically.
    </div>

    <table v-else-if="store.tokens.length > 0" class="token-table">
      <thead>
        <tr>
          <th>Name</th>
          <th>Created</th>
          <th>Last Used</th>
          <th>Status</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="token in store.tokens" :key="token.id" :class="{ revoked: token.revokedAt }">
          <td>{{ token.name }}</td>
          <td>{{ formatDate(token.createdAt) }}</td>
          <td>{{ token.lastUsedAt ? formatDate(token.lastUsedAt) : 'Never' }}</td>
          <td>
            <span :class="['badge', token.revokedAt ? 'badge-revoked' : 'badge-active']">
              {{ token.revokedAt ? 'Revoked' : 'Active' }}
            </span>
          </td>
          <td>
            <button
              v-if="!token.revokedAt"
              class="btn btn-sm btn-danger"
              :disabled="revokingId === token.id"
              @click="handleRevoke(token.id)"
            >
              {{ revokingId === token.id ? 'Revoking...' : 'Revoke' }}
            </button>
            <span v-else class="muted">—</span>
          </td>
        </tr>
      </tbody>
    </table>

    <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
      <div class="modal">
        <h2>Create Token</h2>
        <form @submit.prevent="handleCreate">
          <label for="token-name">Token name</label>
          <input
            id="token-name"
            v-model="newTokenName"
            type="text"
            placeholder="e.g. CI Pipeline, Local Dev"
            maxlength="100"
            autofocus
          />
          <div v-if="createError" class="error-msg">{{ createError }}</div>
          <div class="modal-actions">
            <button type="button" class="btn btn-outline" @click="showCreateModal = false">
              Cancel
            </button>
            <button
              type="submit"
              class="btn btn-primary"
              :disabled="creating || !newTokenName.trim()"
            >
              {{ creating ? 'Creating...' : 'Create Token' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
.tokens-page {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

h1 {
  font-size: 1.75rem;
  font-weight: 700;
}

.token-banner {
  background: #fefce8;
  border: 1px solid #fde047;
  border-radius: 10px;
  padding: 1.25rem 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.banner-warning {
  font-size: 0.95rem;
  color: #713f12;
}

.token-display {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  padding: 0.5rem 0.75rem;
}

.token-value {
  flex: 1;
  font-family: monospace;
  font-size: 0.8rem;
  word-break: break-all;
  color: #1e293b;
}

.token-table {
  width: 100%;
  border-collapse: collapse;
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  overflow: hidden;
}

.token-table th,
.token-table td {
  padding: 0.75rem 1rem;
  text-align: left;
  border-bottom: 1px solid #f1f5f9;
  font-size: 0.9rem;
}

.token-table th {
  background: #f8fafc;
  font-weight: 600;
  color: #64748b;
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.token-table tr:last-child td {
  border-bottom: none;
}

.token-table tr.revoked td {
  opacity: 0.5;
}

.badge {
  display: inline-block;
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.15rem 0.55rem;
  border-radius: 4px;
}

.badge-active {
  background: #dcfce7;
  color: #166534;
}

.badge-revoked {
  background: #f1f5f9;
  color: #64748b;
}

.state-msg {
  color: #64748b;
  font-size: 1rem;
}

.state-msg.error {
  color: #ef4444;
}

.muted {
  color: #94a3b8;
}

.btn {
  padding: 0.4rem 1rem;
  border-radius: 6px;
  border: none;
  cursor: pointer;
  font-size: 0.9rem;
  font-weight: 500;
  transition: background 0.15s;
}

.btn-sm {
  padding: 0.25rem 0.65rem;
  font-size: 0.8rem;
}

.btn-primary {
  background: #3b82f6;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #2563eb;
}

.btn-outline {
  background: transparent;
  color: #475569;
  border: 1px solid #cbd5e1;
}

.btn-outline:hover {
  background: #f1f5f9;
}

.btn-danger {
  background: #fee2e2;
  color: #dc2626;
}

.btn-danger:hover:not(:disabled) {
  background: #fecaca;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

.modal {
  background: white;
  border-radius: 12px;
  padding: 2rem;
  width: 100%;
  max-width: 440px;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.modal h2 {
  font-size: 1.25rem;
  font-weight: 700;
}

.modal label {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  margin-bottom: 0.4rem;
  color: #374151;
}

.modal input {
  width: 100%;
  padding: 0.5rem 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.95rem;
  box-sizing: border-box;
}

.modal input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px #bfdbfe;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}

.error-msg {
  color: #ef4444;
  font-size: 0.875rem;
}
</style>
