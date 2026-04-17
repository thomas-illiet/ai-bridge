<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useTokenStore } from '@/stores/tokens'
import type { CreateTokenResponse } from '@/services/api'
import { formatDate, tokenStatus } from '@/utils/format'
import PaginationBar from '@/components/PaginationBar.vue'
import CreateTokenModal from '@/views/tokens/CreateTokenModal.vue'

const store = useTokenStore()

const showRevoked = ref(false)
const tokPage     = ref(1)
const tokPageSize = ref(10)

watch(showRevoked, (val) => {
  tokPage.value = 1
  store.fetchTokens(val)
})
watch(tokPageSize, () => { tokPage.value = 1 })
const pagedTokens = computed(() => {
  const start = (tokPage.value - 1) * tokPageSize.value
  return store.tokens.slice(start, start + tokPageSize.value)
})

const showCreateModal      = ref(false)
const createdTokenResult   = ref<CreateTokenResponse | null>(null)
const tokenCopied          = ref(false)
const revokingId           = ref<string | null>(null)

onMounted(() => store.fetchTokens(false))

function onTokenCreated(result: CreateTokenResponse) {
  createdTokenResult.value = result
  showCreateModal.value = false
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
  try { await store.deleteToken(id) }
  finally { revokingId.value = null }
}
</script>

<template>
  <div class="tokens-page">
    <div class="page-header">
      <h1>Personal Access Tokens</h1>
      <div class="header-actions">
        <label class="toggle-label">
          <input type="checkbox" v-model="showRevoked" />
          Show revoked
        </label>
        <button class="btn btn-primary" @click="showCreateModal = true">New Token</button>
      </div>
    </div>

    <div v-if="createdTokenResult" class="token-banner">
      <div class="banner-warning">
        Token created — copy it now. It will <strong>not</strong> be shown again.
      </div>
      <div class="token-display">
        <code class="token-value">{{ createdTokenResult.rawToken }}</code>
        <button class="btn btn-sm" @click="copyToken">{{ tokenCopied ? 'Copied!' : 'Copy' }}</button>
      </div>
      <button class="btn btn-outline btn-sm" @click="dismissCreatedToken">I have saved my token</button>
    </div>

    <div v-if="store.loading" class="empty-card">
      <div class="empty-icon">
        <svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/>
        </svg>
      </div>
      <p class="empty-title">Loading tokens…</p>
    </div>
    <div v-else-if="store.error" class="empty-card empty-card--error">
      <div class="empty-icon empty-icon--error">
        <svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
        </svg>
      </div>
      <p class="empty-title">Failed to load tokens</p>
      <p class="empty-sub">{{ store.error }}</p>
    </div>
    <div v-else-if="store.tokens.length === 0 && !createdTokenResult" class="empty-card">
      <div class="empty-icon">
        <svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4"/>
        </svg>
      </div>
      <p class="empty-title">No tokens yet</p>
      <p class="empty-sub">Create a personal access token to authenticate API requests programmatically.</p>
    </div>

    <table v-else-if="store.tokens.length > 0" class="token-table">
      <thead>
        <tr>
          <th>Name</th><th>Created</th><th>Expires</th><th>Last Used</th><th>Status</th><th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="token in pagedTokens" :key="token.id" :class="{ dimmed: tokenStatus(token) !== 'active' }">
          <td>{{ token.name }}</td>
          <td>{{ formatDate(token.createdAt) }}</td>
          <td>{{ token.expiresAt ? formatDate(token.expiresAt) : '—' }}</td>
          <td>{{ token.lastUsedAt ? formatDate(token.lastUsedAt) : 'Never' }}</td>
          <td><span :class="['badge', `badge-${tokenStatus(token)}`]">{{ tokenStatus(token) }}</span></td>
          <td>
            <button
              v-if="tokenStatus(token) === 'active'"
              class="btn btn-sm btn-danger"
              :disabled="revokingId === token.id"
              @click="handleRevoke(token.id)"
            >{{ revokingId === token.id ? 'Revoking…' : 'Revoke' }}</button>
            <span v-else class="muted">—</span>
          </td>
        </tr>
      </tbody>
    </table>

    <PaginationBar
      v-if="store.tokens.length > 0"
      v-model:page="tokPage"
      v-model:pageSize="tokPageSize"
      :total="store.tokens.length"
    />

    <CreateTokenModal v-if="showCreateModal" @close="showCreateModal = false" @created="onTokenCreated" />
  </div>
</template>

<style scoped>
.tokens-page { display: flex; flex-direction: column; gap: 1.5rem; }
.page-header { display: flex; align-items: center; justify-content: space-between; }
h1 { font-size: 1.75rem; font-weight: 700; }
.header-actions { display: flex; align-items: center; gap: 1rem; }
.toggle-label {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.85rem;
  font-weight: 500;
  color: #64748b;
  cursor: pointer;
  user-select: none;
}
.toggle-label input[type="checkbox"] { cursor: pointer; accent-color: #3b82f6; width: 14px; height: 14px; }
.token-banner { background: #fefce8; border: 1px solid #fde047; border-radius: 10px; padding: 1.25rem 1.5rem; display: flex; flex-direction: column; gap: 1rem; }
.banner-warning { font-size: 0.95rem; color: #713f12; }
.token-display { display: flex; align-items: center; gap: 0.75rem; background: #fff; border: 1px solid #e2e8f0; border-radius: 6px; padding: 0.5rem 0.75rem; }
.token-value { flex: 1; font-family: monospace; font-size: 0.8rem; word-break: break-all; color: #1e293b; }
.token-table { width: 100%; border-collapse: collapse; background: white; border: 1px solid #e2e8f0; border-radius: 12px; overflow: hidden; transition: opacity 0.15s; }
.token-table th, .token-table td { padding: 0.75rem 1rem; text-align: left; border-bottom: 1px solid #f1f5f9; font-size: 0.9rem; }
.token-table th { background: #f8fafc; font-weight: 600; color: #64748b; font-size: 0.78rem; text-transform: uppercase; letter-spacing: 0.05em; }
.token-table tr:last-child td { border-bottom: none; }
.token-table tr.dimmed td { opacity: 0.5; }
.badge { display: inline-block; font-size: 0.75rem; font-weight: 600; padding: 0.15rem 0.55rem; border-radius: 4px; text-transform: capitalize; }
.badge-active  { background: #dcfce7; color: #166534; }
.badge-revoked { background: #f1f5f9; color: #64748b; }
.badge-expired { background: #fef3c7; color: #92400e; }
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
.empty-card--error { background: #fff5f5; border-color: #fecaca; }
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
.empty-icon--error { background: #fee2e2; color: #ef4444; }
.empty-title { font-size: 1rem; font-weight: 600; color: #1e293b; margin: 0; }
.empty-sub   { font-size: 0.85rem; color: #94a3b8; margin: 0; max-width: 320px; line-height: 1.5; }
.muted { color: #94a3b8; }
.btn { padding: 0.4rem 1rem; border-radius: 6px; border: none; cursor: pointer; font-size: 0.9rem; font-weight: 500; transition: background 0.15s; }
.btn-sm { padding: 0.25rem 0.65rem; font-size: 0.8rem; }
.btn-primary { background: #3b82f6; color: white; }
.btn-primary:hover:not(:disabled) { background: #2563eb; }
.btn-outline { background: transparent; color: #475569; border: 1px solid #cbd5e1; }
.btn-outline:hover { background: #f1f5f9; }
.btn-danger { background: #fee2e2; color: #dc2626; }
.btn-danger:hover:not(:disabled) { background: #fecaca; }
.btn:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
