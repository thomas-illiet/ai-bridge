<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { listWhitelist, addWhitelist, deleteWhitelist, toggleWhitelist } from '@/services/api'
import { formatDate } from '@/utils/format'
import PaginationBar from '@/components/PaginationBar.vue'

interface WhitelistEntry {
  id: string; cidr: string; description: string
  enabled: boolean; createdBy: string; createdAt: string
}

const entries      = ref<WhitelistEntry[]>([])
const loading      = ref(true)
const error        = ref<string | null>(null)
const saving       = ref(false)
const newCIDR      = ref('')
const newDesc      = ref('')
const formError    = ref<string | null>(null)
const confirmDelId = ref<string | null>(null)

const isEmpty   = computed(() => entries.value.length === 0)
const noEnabled = computed(() => entries.value.filter(e => e.enabled).length === 0)

const page     = ref(1)
const pageSize = ref(10)
watch(pageSize, () => { page.value = 1 })
const pagedEntries = computed(() => {
  const s = (page.value - 1) * pageSize.value
  return entries.value.slice(s, s + pageSize.value)
})

async function loadWL() {
  try { const res = await listWhitelist(); entries.value = res.data.entries ?? [] }
  catch { error.value = 'Failed to load whitelist' }
  finally { loading.value = false }
}

async function addEntry() {
  if (!newCIDR.value.trim()) { formError.value = 'IP or CIDR is required'; return }
  formError.value = null; saving.value = true
  try {
    await addWhitelist(newCIDR.value.trim(), newDesc.value.trim())
    newCIDR.value = ''; newDesc.value = ''
    await loadWL()
  } catch (e: any) { formError.value = e?.response?.data?.error ?? 'Failed to add entry' }
  finally { saving.value = false }
}

async function toggle(entry: WhitelistEntry) {
  try { await toggleWhitelist(entry.id, !entry.enabled); entry.enabled = !entry.enabled }
  catch { error.value = 'Failed to toggle entry' }
}

async function remove(id: string) {
  try { await deleteWhitelist(id); entries.value = entries.value.filter(e => e.id !== id) }
  catch { error.value = 'Failed to delete entry' }
  finally { confirmDelId.value = null }
}

onMounted(loadWL)
</script>

<template>
  <div class="tab-content">
    <div class="mode-banner" :class="noEnabled ? 'banner-open' : 'banner-restricted'">
      <span class="banner-dot" />
      <span v-if="noEnabled"><strong>Open mode</strong> — no enabled entries, all authenticated users can access the proxy.</span>
      <span v-else><strong>Restricted mode</strong> — only whitelisted IPs can access the proxy.</span>
    </div>

    <div class="card">
      <h2 class="card-title">Add IP / CIDR</h2>
      <div class="add-form">
        <div class="field">
          <label>IP address or CIDR</label>
          <input v-model="newCIDR" placeholder="192.168.1.10  or  10.0.0.0/8" @keyup.enter="addEntry" />
        </div>
        <div class="field">
          <label>Description <span class="optional">(optional)</span></label>
          <input v-model="newDesc" placeholder="e.g. Office network" @keyup.enter="addEntry" />
        </div>
        <button class="btn btn-primary" :disabled="saving" @click="addEntry">{{ saving ? 'Adding…' : 'Add' }}</button>
      </div>
      <p v-if="formError" class="form-error">{{ formError }}</p>
    </div>

    <div class="card">
      <h2 class="card-title">Whitelist entries</h2>
      <div v-if="loading" class="state-msg">Loading…</div>
      <div v-else-if="error" class="state-msg error">{{ error }}</div>
      <div v-else-if="isEmpty" class="state-msg muted">No entries yet.</div>
      <table v-else class="data-table">
        <thead><tr><th>IP / CIDR</th><th>Description</th><th>Status</th><th>Added by</th><th>Date</th><th>Actions</th></tr></thead>
        <tbody>
          <tr v-for="entry in pagedEntries" :key="entry.id">
            <td class="mono">{{ entry.cidr }}</td>
            <td class="muted">{{ entry.description || '—' }}</td>
            <td>
              <button class="status-toggle" :class="entry.enabled ? 'toggle-on' : 'toggle-off'" @click="toggle(entry)">
                {{ entry.enabled ? 'Enabled' : 'Disabled' }}
              </button>
            </td>
            <td class="muted">{{ entry.createdBy }}</td>
            <td class="muted">{{ formatDate(entry.createdAt) }}</td>
            <td class="actions">
              <template v-if="confirmDelId === entry.id">
                <span class="confirm-text">Delete?</span>
                <button class="btn btn-sm btn-danger" @click="remove(entry.id)">Yes</button>
                <button class="btn-link" @click="confirmDelId = null">No</button>
              </template>
              <button v-else class="btn btn-sm btn-danger" @click="confirmDelId = entry.id">Delete</button>
            </td>
          </tr>
        </tbody>
      </table>
      <PaginationBar v-if="entries.length > 0" v-model:page="page" v-model:pageSize="pageSize" :total="entries.length" />
    </div>
  </div>
</template>

<style scoped>
.tab-content { display: flex; flex-direction: column; gap: 1.25rem; }
.state-msg { color: #64748b; font-size: 0.9rem; }
.state-msg.error { color: #ef4444; }
.state-msg.muted { color: #94a3b8; }
.mode-banner { display: flex; align-items: center; gap: 0.75rem; padding: 0.9rem 1.25rem; border-radius: 10px; font-size: 0.9rem; }
.banner-open       { background: #f0fdf4; color: #166534; border: 1px solid #bbf7d0; }
.banner-restricted { background: #fffbeb; color: #92400e; border: 1px solid #fde68a; }
.banner-dot { width: 9px; height: 9px; border-radius: 50%; background: currentColor; flex-shrink: 0; }
.card { background: white; border: 1px solid #e2e8f0; border-radius: 10px; padding: 1.25rem; display: flex; flex-direction: column; gap: 1rem; }
.card-title { font-size: 1rem; font-weight: 700; margin: 0; }
.add-form { display: flex; align-items: flex-end; gap: 0.75rem; flex-wrap: wrap; }
.field { display: flex; flex-direction: column; gap: 0.3rem; flex: 1; min-width: 160px; }
.field label { font-size: 0.78rem; font-weight: 600; color: #64748b; }
.optional { font-weight: 400; color: #94a3b8; }
.field input { padding: 0.45rem 0.75rem; border: 1px solid #e2e8f0; border-radius: 6px; font-size: 0.9rem; outline: none; }
.field input:focus { border-color: #3b82f6; }
.form-error { color: #ef4444; font-size: 0.83rem; margin: 0; }
.btn { padding: 0.45rem 1rem; border-radius: 6px; border: none; cursor: pointer; font-size: 0.9rem; font-weight: 500; }
.btn-sm { padding: 0.25rem 0.65rem; font-size: 0.8rem; }
.btn-primary { background: #3b82f6; color: white; }
.btn-primary:hover:not(:disabled) { background: #2563eb; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-danger { background: #fee2e2; color: #dc2626; }
.btn-danger:hover:not(:disabled) { background: #fecaca; }
.btn-link { background: none; border: none; cursor: pointer; font-size: 0.83rem; padding: 0 0.25rem; color: #64748b; }
.btn-link:hover { color: #1e293b; }
.data-table { width: 100%; border-collapse: collapse; font-size: 0.88rem; background: white; border: 1px solid #e2e8f0; border-radius: 10px; }
.data-table th { text-align: left; padding: 0.55rem 0.9rem; font-size: 0.75rem; font-weight: 600; color: #64748b; text-transform: uppercase; letter-spacing: 0.03em; border-bottom: 1px solid #e2e8f0; background: #f8fafc; }
.data-table thead tr:first-child th:first-child { border-radius: 10px 0 0 0; }
.data-table thead tr:first-child th:last-child  { border-radius: 0 10px 0 0; }
.data-table tbody tr:last-child td:first-child   { border-radius: 0 0 0 10px; }
.data-table tbody tr:last-child td:last-child    { border-radius: 0 0 10px 0; }
.data-table td { padding: 0.65rem 0.9rem; border-bottom: 1px solid #f1f5f9; }
.data-table tr:last-child td { border-bottom: none; }
.mono { font-family: monospace; }
.muted { color: #64748b; }
.actions { display: flex; align-items: center; gap: 0.4rem; white-space: nowrap; }
.confirm-text { font-size: 0.83rem; color: #ef4444; font-weight: 600; }
.status-toggle { padding: 0.15rem 0.55rem; border-radius: 999px; border: none; cursor: pointer; font-size: 0.75rem; font-weight: 600; }
.toggle-on  { background: #dcfce7; color: #166534; }
.toggle-off { background: #f1f5f9; color: #64748b; }
.status-toggle:hover { opacity: 0.75; }
</style>
