<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { adminListTokens, adminRevokeToken } from '@/services/api'
import type { AdminTokenRow } from '@/services/api'
import { formatDate, tokenStatus } from '@/utils/format'
import PaginationBar from '@/components/PaginationBar.vue'

const tokens        = ref<AdminTokenRow[]>([])
const total         = ref(0)
const page          = ref(1)
const pageSize      = ref(10)
const search        = ref('')
const showRevoked   = ref(false)
const loading       = ref(false)
const revokingId    = ref<string | null>(null)

let searchTimer: ReturnType<typeof setTimeout>
watch(search, () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => { page.value = 1; load() }, 300)
})
watch(page, load)
watch(pageSize, () => { page.value = 1; load() })
watch(showRevoked, () => { page.value = 1; load() })

async function load() {
  loading.value = true
  try {
    const res = await adminListTokens(page.value, pageSize.value, search.value, showRevoked.value)
    tokens.value = res.data.tokens
    total.value  = res.data.total
  } finally { loading.value = false }
}

async function revoke(id: string) {
  revokingId.value = id
  try { await adminRevokeToken(id); await load() }
  finally { revokingId.value = null }
}

onMounted(load)
</script>

<template>
  <div class="tab-content">
    <div class="toolbar">
      <p class="sub">{{ total }} token{{ total !== 1 ? 's' : '' }} total.</p>
      <div class="toolbar-right">
        <label class="toggle-label">
          <input type="checkbox" v-model="showRevoked" />
          Show revoked
        </label>
        <input v-model="search" type="text" placeholder="Search by name or user…" class="search-input" />
      </div>
    </div>

    <div v-if="loading && tokens.length === 0" class="empty-card">
      <div class="empty-icon">
        <svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/>
        </svg>
      </div>
      <p class="empty-title">Loading tokens…</p>
    </div>
    <div v-else-if="tokens.length === 0" class="empty-card">
      <div class="empty-icon">
        <svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4"/>
        </svg>
      </div>
      <p class="empty-title">No tokens found</p>
      <p class="empty-sub">{{ search ? 'Try adjusting your search.' : showRevoked ? 'No tokens exist yet.' : 'No active tokens. Revoked tokens can be shown with the toggle above.' }}</p>
    </div>

    <table v-else class="data-table" :class="{ 'table-loading': loading }">
      <thead>
        <tr>
          <th>Name</th><th>User</th><th>Created</th><th>Expires</th><th>Last Used</th><th>Status</th><th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="token in tokens" :key="token.id" :class="{ dimmed: tokenStatus(token) !== 'active' }">
          <td class="bold">{{ token.name }}</td>
          <td><span class="user-pill">{{ token.username }}</span></td>
          <td class="muted">{{ formatDate(token.createdAt) }}</td>
          <td class="muted">{{ token.expiresAt ? formatDate(token.expiresAt) : '—' }}</td>
          <td class="muted">{{ token.lastUsedAt ? formatDate(token.lastUsedAt) : 'Never' }}</td>
          <td><span class="badge" :class="`badge-tok-${tokenStatus(token)}`">{{ tokenStatus(token) }}</span></td>
          <td class="actions">
            <button
              v-if="tokenStatus(token) === 'active'"
              class="btn btn-sm btn-danger"
              :disabled="revokingId === token.id"
              @click="revoke(token.id)"
            >{{ revokingId === token.id ? 'Revoking…' : 'Revoke' }}</button>
            <button v-else class="btn btn-sm btn-danger" :disabled="true">Revoke</button>
          </td>
        </tr>
      </tbody>
    </table>

    <PaginationBar v-model:page="page" v-model:pageSize="pageSize" :total="total" />
  </div>
</template>

<style scoped>
.tab-content { display: flex; flex-direction: column; gap: 1.25rem; }
.toolbar { display: flex; align-items: center; justify-content: space-between; gap: 1rem; flex-wrap: wrap; }
.toolbar-right { display: flex; align-items: center; gap: 0.75rem; }
.sub { font-size: 0.85rem; color: #64748b; margin: 0; }
.toggle-label {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.85rem;
  font-weight: 500;
  color: #64748b;
  cursor: pointer;
  user-select: none;
  white-space: nowrap;
}
.toggle-label input[type="checkbox"] { cursor: pointer; accent-color: #3b82f6; width: 14px; height: 14px; }
.search-input { padding: 0.45rem 0.75rem; border: 1px solid #e2e8f0; border-radius: 6px; font-size: 0.9rem; width: 220px; background: white; outline: none; }
.search-input:focus { border-color: #3b82f6; box-shadow: 0 0 0 2px #bfdbfe; }
.empty-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.6rem;
  padding: 3rem 2rem;
  background: white;
  border: 1px dashed #e2e8f0;
  border-radius: 12px;
  text-align: center;
}
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
.empty-title { font-size: 1rem; font-weight: 600; color: #1e293b; margin: 0; }
.empty-sub   { font-size: 0.85rem; color: #94a3b8; margin: 0; max-width: 320px; line-height: 1.5; }
.data-table { width: 100%; border-collapse: collapse; font-size: 0.88rem; background: white; border: 1px solid #e2e8f0; border-radius: 10px; }
.data-table th { text-align: left; padding: 0.55rem 0.9rem; font-size: 0.75rem; font-weight: 600; color: #64748b; text-transform: uppercase; letter-spacing: 0.03em; border-bottom: 1px solid #e2e8f0; background: #f8fafc; }
.data-table thead tr:first-child th:first-child { border-radius: 10px 0 0 0; }
.data-table thead tr:first-child th:last-child  { border-radius: 0 10px 0 0; }
.data-table tbody tr:last-child td:first-child   { border-radius: 0 0 0 10px; }
.data-table tbody tr:last-child td:last-child    { border-radius: 0 0 10px 0; }
.data-table td { padding: 0.65rem 0.9rem; border-bottom: 1px solid #f1f5f9; }
.data-table tr:last-child td { border-bottom: none; }
.data-table.table-loading { opacity: 0.6; pointer-events: none; }
.data-table tr.dimmed td { opacity: 0.5; }
.bold { font-weight: 600; color: #1e293b; }
.muted { color: #64748b; }
.actions { display: flex; align-items: center; gap: 0.4rem; }
.badge { padding: 0.18rem 0.55rem; border-radius: 999px; font-size: 0.75rem; font-weight: 600; }
.badge-tok-active  { background: #dcfce7; color: #166534; }
.badge-tok-revoked { background: #f1f5f9; color: #64748b; }
.badge-tok-expired { background: #fef3c7; color: #92400e; }
.user-pill { display: inline-block; background: #f1f5f9; color: #475569; font-size: 0.75rem; font-weight: 600; padding: 0.15rem 0.55rem; border-radius: 999px; }
.btn { padding: 0.45rem 1rem; border-radius: 6px; border: none; cursor: pointer; font-size: 0.9rem; font-weight: 500; }
.btn-sm { padding: 0.25rem 0.65rem; font-size: 0.8rem; }
.btn-danger { background: #fee2e2; color: #dc2626; }
.btn-danger:hover:not(:disabled) { background: #fecaca; }
.btn-danger:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
