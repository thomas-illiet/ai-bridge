<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useTokenStore } from '@/stores/tokens'
import type { CreateTokenResponse } from '@/services/api'
import { formatDate, tokenStatus } from '@/utils/format'
import PaginationBar from '@/components/PaginationBar.vue'
import CreateTokenModal from '@/views/tokens/CreateTokenModal.vue'

const store = useTokenStore()

const tokPage     = ref(1)
const tokPageSize = ref(10)
watch(tokPageSize, () => { tokPage.value = 1 })
const pagedTokens = computed(() => {
  const start = (tokPage.value - 1) * tokPageSize.value
  return store.tokens.slice(start, start + tokPageSize.value)
})

const showCreateModal      = ref(false)
const createdTokenResult   = ref<CreateTokenResponse | null>(null)
const tokenCopied          = ref(false)
const revokingId           = ref<string | null>(null)

onMounted(() => store.fetchTokens())

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
      <button class="btn btn-primary" @click="showCreateModal = true">New Token</button>
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

    <div v-if="store.loading" class="state-msg">Loading...</div>
    <div v-else-if="store.error" class="state-msg error">{{ store.error }}</div>
    <div v-else-if="store.tokens.length === 0 && !createdTokenResult" class="state-msg">
      No tokens yet. Create one to authenticate API requests programmatically.
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
.state-msg { color: #64748b; font-size: 1rem; }
.state-msg.error { color: #ef4444; }
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
