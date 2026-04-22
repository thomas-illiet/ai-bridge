<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { listWhitelist, addWhitelist, deleteWhitelist, toggleWhitelist } from '@/services/api'
import { formatDate } from '@/utils/format'
import PaginationBar from '@/components/PaginationBar.vue'
import { useMinLoad } from '@/composables/useMinLoad'

interface WhitelistEntry {
  id: string; cidr: string; description: string
  enabled: boolean; createdBy: string; createdAt: string
}

const entries      = ref<WhitelistEntry[]>([])
const { loading, withLoad } = useMinLoad(300, true)
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

const sortBy  = ref('created_at')
const sortDir = ref<'asc' | 'desc'>('desc')

function toggleSort(col: string) {
  if (sortBy.value === col) { sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc' }
  else { sortBy.value = col; sortDir.value = 'desc' }
  page.value = 1
  loadWL()
}

async function loadWL() {
  error.value = null
  await withLoad(async () => {
    try { const res = await listWhitelist(sortBy.value, sortDir.value); entries.value = res.data.entries ?? [] }
    catch { error.value = 'Failed to load whitelist' }
  })
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
        <div class="wl-field">
          <label>IP address or CIDR</label>
          <input v-model="newCIDR" placeholder="192.168.1.10  or  10.0.0.0/8" @keyup.enter="addEntry" />
        </div>
        <div class="wl-field">
          <label>Description <span class="optional">(optional)</span></label>
          <input v-model="newDesc" placeholder="e.g. Office network" @keyup.enter="addEntry" />
        </div>
        <button class="btn btn-primary" :disabled="saving" @click="addEntry">{{ saving ? 'Adding…' : 'Add' }}</button>
      </div>
      <p v-if="formError" class="form-error">{{ formError }}</p>
    </div>

    <div class="card">
      <h2 class="card-title">Whitelist entries</h2>
      <div v-if="!loading && error" class="empty-card empty-card--error">
        <div class="empty-icon empty-icon--error">
          <svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
          </svg>
        </div>
        <p class="empty-title">Failed to load</p>
        <p class="empty-sub">{{ error }}</p>
      </div>
      <div v-else-if="!loading && isEmpty" class="empty-card">
        <div class="empty-icon">
          <svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
            <path d="M7 11V7a5 5 0 0 1 10 0v4"/>
          </svg>
        </div>
        <p class="empty-title">No IP rules yet</p>
        <p class="empty-sub">All authenticated users can currently access the proxy. Add an IP or CIDR above to restrict access.</p>
      </div>
      <table v-else class="data-table">
        <thead><tr>
          <th class="sortable" :class="{ active: sortBy === 'cidr' }" @click="toggleSort('cidr')">IP / CIDR <span class="sort-icon">{{ sortBy === 'cidr' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
          <th>Description</th>
          <th class="sortable" :class="{ active: sortBy === 'enabled' }" @click="toggleSort('enabled')">Status <span class="sort-icon">{{ sortBy === 'enabled' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
          <th class="sortable" :class="{ active: sortBy === 'created_by' }" @click="toggleSort('created_by')">Added by <span class="sort-icon">{{ sortBy === 'created_by' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
          <th class="sortable" :class="{ active: sortBy === 'created_at' }" @click="toggleSort('created_at')">Date <span class="sort-icon">{{ sortBy === 'created_at' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
          <th>Actions</th>
        </tr></thead>
        <tbody>
          <template v-if="loading">
            <tr v-for="i in 4" :key="i" class="skeleton-row">
              <td><div class="skeleton-bar skeleton-bar--md" /></td>
              <td><div class="skeleton-bar skeleton-bar--lg" /></td>
              <td><div class="skeleton-bar skeleton-bar--btn" /></td>
              <td><div class="skeleton-bar skeleton-bar--sm" /></td>
              <td><div class="skeleton-bar skeleton-bar--sm" /></td>
              <td><div class="skeleton-bar skeleton-bar--btn" /></td>
            </tr>
          </template>
          <template v-else>
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
          </template>
        </tbody>
      </table>
      <PaginationBar v-if="!loading && entries.length > 0" v-model:page="page" v-model:pageSize="pageSize" :total="entries.length" />
    </div>
  </div>
</template>

<style scoped>
.mode-banner { display: flex; align-items: center; gap: 0.75rem; padding: 0.9rem 1.25rem; border-radius: 10px; font-size: 0.9rem; }
.banner-open       { background: #f0fdf4; color: #166534; border: 1px solid #bbf7d0; }
.banner-restricted { background: #fffbeb; color: #92400e; border: 1px solid #fde68a; }
.banner-dot { width: 9px; height: 9px; border-radius: 50%; background: currentColor; flex-shrink: 0; }
.add-form { display: flex; align-items: flex-end; gap: 0.75rem; flex-wrap: wrap; }
.wl-field { display: flex; flex-direction: column; gap: 0.3rem; flex: 1; min-width: 160px; }
.wl-field label { font-size: 0.78rem; font-weight: 600; color: #64748b; }
.wl-field input { padding: 0.45rem 0.75rem; border: 1px solid #e2e8f0; border-radius: 6px; font-size: 0.9rem; outline: none; }
.wl-field input:focus { border-color: #3b82f6; }
.mono { font-family: monospace; }
.confirm-text { font-size: 0.83rem; color: #ef4444; font-weight: 600; }
.status-toggle { padding: 0.15rem 0.55rem; border-radius: 999px; border: none; cursor: pointer; font-size: 0.75rem; font-weight: 600; }
.toggle-on  { background: #dcfce7; color: #166534; }
.toggle-off { background: #f1f5f9; color: #64748b; }
.status-toggle:hover { opacity: 0.75; }
</style>
