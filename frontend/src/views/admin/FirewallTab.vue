<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { listFirewallRules, addFirewallRule, deleteFirewallRule, toggleFirewallRule, reloadFirewall, moveFirewallRulePriority } from '@/services/api'
import { formatDate } from '@/utils/format'
import PaginationBar from '@/components/PaginationBar.vue'
import LoadingOverlay from '@/components/LoadingOverlay.vue'
import { useMinLoad } from '@/composables/useMinLoad'

interface FirewallEntry {
  id: string; cidr: string; description: string
  action: 'allow' | 'deny'; priority: number
  enabled: boolean; createdAt: string
}

const entries      = ref<FirewallEntry[]>([])
const { loading, withLoad } = useMinLoad(300, true)
const { loading: reloading, withLoad: withReload } = useMinLoad(1200)
const error        = ref<string | null>(null)
const saving       = ref(false)
const newCIDR      = ref('')
const newDesc      = ref('')
const newAction    = ref<'allow' | 'deny'>('allow')
const newPriority  = ref(100)
const formError    = ref<string | null>(null)
const confirmDelId = ref<string | null>(null)
const movingId = ref<string | null>(null)

const isEmpty   = computed(() => entries.value.length === 0)
const noEnabled = computed(() => entries.value.filter(e => e.enabled).length === 0)

const search = ref('')
let searchTimer: ReturnType<typeof setTimeout> | null = null
watch(search, () => {
  page.value = 1
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => loadRules(), 300)
})

const page     = ref(1)
const pageSize = ref(10)
watch(pageSize, () => { page.value = 1 })
const pagedEntries = computed(() => {
  const s = (page.value - 1) * pageSize.value
  return entries.value.slice(s, s + pageSize.value)
})

const sortBy  = ref('priority')
const sortDir = ref<'asc' | 'desc'>('asc')

function toggleSort(col: string) {
  if (sortBy.value === col) { sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc' }
  else { sortBy.value = col; sortDir.value = 'asc' }
  page.value = 1
  loadRules()
}

async function loadRules() {
  error.value = null
  await withLoad(async () => {
    try { const res = await listFirewallRules(sortBy.value, sortDir.value, search.value); entries.value = res.data.entries ?? [] }
    catch { error.value = 'Failed to load firewall rules' }
  })
}

async function addEntry() {
  if (!newCIDR.value.trim()) { formError.value = 'IP or CIDR is required'; return }
  formError.value = null; saving.value = true
  try {
    await addFirewallRule(newCIDR.value.trim(), newDesc.value.trim(), newAction.value, newPriority.value)
    newCIDR.value = ''; newDesc.value = ''; newAction.value = 'allow'; newPriority.value = 100
    await loadRules()
  } catch (e: any) { formError.value = e?.response?.data?.error ?? 'Failed to add rule' }
  finally { saving.value = false }
}

async function toggle(entry: FirewallEntry) {
  try { await toggleFirewallRule(entry.id, !entry.enabled); entry.enabled = !entry.enabled }
  catch { error.value = 'Failed to toggle rule' }
}

async function remove(id: string) {
  try { await deleteFirewallRule(id); entries.value = entries.value.filter(e => e.id !== id) }
  catch { error.value = 'Failed to delete rule' }
  finally { confirmDelId.value = null }
}

const entryIndex = (id: string) => entries.value.findIndex(e => e.id === id)

async function movePriority(entry: FirewallEntry, direction: 'up' | 'down') {
  movingId.value = entry.id
  try {
    const orderedIds = entries.value.map(e => e.id)
    await moveFirewallRulePriority(entry.id, direction, orderedIds)
    await loadRules()
  } catch (e: any) { error.value = e?.response?.data?.error ?? 'Failed to move rule' }
  finally { movingId.value = null }
}

async function forceReload() {
  await withReload(async () => {
    try { await reloadFirewall() }
    catch (e: any) { error.value = e?.response?.data?.error ?? 'Failed to reload firewall' }
  })
}

onMounted(loadRules)
</script>

<template>
  <Teleport defer to="#admin-search-portal">
    <div class="portal-search-wrap">
      <svg class="portal-search-icon" xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
      <input v-model="search" type="text" placeholder="Search IP or description…" class="portal-input" />
    </div>
  </Teleport>

  <div class="tab-content">
    <LoadingOverlay :visible="reloading" message="Reloading firewall…" />

    <div class="mode-banner" :class="noEnabled ? 'banner-open' : 'banner-restricted'">
      <span class="banner-dot" />
      <span v-if="noEnabled"><strong>Open mode</strong> — no enabled rules, all authenticated users can access the proxy.</span>
      <span v-else><strong>Rules active</strong> — firewall rules are enforced. First matching rule wins.</span>
    </div>

    <div class="card">
      <h2 class="card-title">Add Firewall Rule</h2>
      <div class="add-form">
        <div class="wl-field wl-field--cidr">
          <label>IP address or CIDR</label>
          <input v-model="newCIDR" placeholder="192.168.1.10  or  10.0.0.0/8" @keyup.enter="addEntry" />
        </div>
        <div class="wl-field wl-field--desc">
          <label>Description <span class="optional">(optional)</span></label>
          <input v-model="newDesc" placeholder="e.g. Office network" @keyup.enter="addEntry" />
        </div>
        <div class="wl-field wl-field--action">
          <label>Action</label>
          <select v-model="newAction">
            <option value="allow">Allow</option>
            <option value="deny">Deny</option>
          </select>
        </div>
        <div class="wl-field wl-field--priority">
          <label>Priority</label>
          <input v-model.number="newPriority" type="number" min="1" placeholder="100" @keyup.enter="addEntry" />
        </div>
        <button class="btn btn-primary" :disabled="saving" @click="addEntry">
          <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          {{ saving ? 'Adding…' : 'Add' }}
        </button>
      </div>
      <p v-if="formError" class="form-error">{{ formError }}</p>
    </div>

    <div class="card">
      <div class="card-header">
        <h2 class="card-title">Firewall rules <span class="title-count">{{ entries.length }}</span></h2>
        <button v-if="!isEmpty" class="btn btn-secondary btn-sm" :disabled="reloading" @click="forceReload">
          <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
          {{ reloading ? 'Reloading…' : 'Force Reload' }}
        </button>
      </div>
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
        <p class="empty-title">No firewall rules yet</p>
        <p class="empty-sub">All authenticated users can currently access the proxy. Add a rule above to restrict or block access.</p>
      </div>
      <table v-else class="data-table">
        <thead><tr>
          <th class="sortable" :class="{ active: sortBy === 'priority' }" @click="toggleSort('priority')">Priority <span class="sort-icon">{{ sortBy === 'priority' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
          <th class="sortable" :class="{ active: sortBy === 'action' }" @click="toggleSort('action')">Action <span class="sort-icon">{{ sortBy === 'action' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
          <th class="sortable" :class="{ active: sortBy === 'cidr' }" @click="toggleSort('cidr')">IP / CIDR <span class="sort-icon">{{ sortBy === 'cidr' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
          <th>Description</th>
          <th class="sortable" :class="{ active: sortBy === 'enabled' }" @click="toggleSort('enabled')">Status <span class="sort-icon">{{ sortBy === 'enabled' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
          <th class="sortable" :class="{ active: sortBy === 'created_at' }" @click="toggleSort('created_at')">Date <span class="sort-icon">{{ sortBy === 'created_at' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
          <th>Actions</th>
        </tr></thead>
        <tbody>
          <template v-if="loading">
            <tr v-for="i in 4" :key="i" class="skeleton-row">
              <td><div class="skeleton-bar skeleton-bar--sm" /></td>
              <td><div class="skeleton-bar skeleton-bar--btn" /></td>
              <td><div class="skeleton-bar skeleton-bar--md" /></td>
              <td><div class="skeleton-bar skeleton-bar--lg" /></td>
              <td><div class="skeleton-bar skeleton-bar--btn" /></td>
              <td><div class="skeleton-bar skeleton-bar--sm" /></td>
              <td><div class="skeleton-bar skeleton-bar--btn" /></td>
            </tr>
          </template>
          <template v-else>
            <tr v-for="entry in pagedEntries" :key="entry.id">
              <td class="mono priority-cell">{{ entry.priority }}</td>
              <td>
                <span class="action-badge" :class="entry.action === 'allow' ? 'action-allow' : 'action-deny'">
                  {{ entry.action === 'allow' ? 'Allow' : 'Deny' }}
                </span>
              </td>
              <td class="mono">{{ entry.cidr }}</td>
              <td class="muted">{{ entry.description || '—' }}</td>
              <td>
                <button class="status-toggle" :class="entry.enabled ? 'toggle-on' : 'toggle-off'" @click="toggle(entry)">
                  {{ entry.enabled ? 'Enabled' : 'Disabled' }}
                </button>
              </td>
              <td class="muted">{{ formatDate(entry.createdAt) }}</td>
              <td class="actions">
                <div class="priority-arrows">
                  <button class="btn-arrow" :disabled="movingId === entry.id || entryIndex(entry.id) === 0" @click="movePriority(entry, 'up')" title="Move up (lower priority number)">
                    <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="18 15 12 9 6 15"/></svg>
                  </button>
                  <button class="btn-arrow" :disabled="movingId === entry.id || entryIndex(entry.id) === entries.length - 1" @click="movePriority(entry, 'down')" title="Move down (higher priority number)">
                    <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"/></svg>
                  </button>
                </div>
                <template v-if="confirmDelId === entry.id">
                  <span class="confirm-text">Delete?</span>
                  <button class="btn btn-sm btn-danger" @click="remove(entry.id)">
                    <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"/></svg>
                    Yes
                  </button>
                  <button class="btn-link" @click="confirmDelId = null">No</button>
                </template>
                <button v-else class="btn btn-sm btn-danger" @click="confirmDelId = entry.id">
                  <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"/></svg>
                  Delete
                </button>
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
.wl-field { display: flex; flex-direction: column; gap: 0.3rem; }
.wl-field--cidr     { flex: 2; min-width: 160px; }
.wl-field--desc     { flex: 2; min-width: 140px; }
.wl-field--action   { flex: 0 0 100px; }
.wl-field--priority { flex: 0 0 80px; }
.wl-field label { font-size: 0.78rem; font-weight: 600; color: #64748b; }
.wl-field input, .wl-field select { padding: 0.45rem 0.75rem; border: 1px solid #e2e8f0; border-radius: 6px; font-size: 0.9rem; outline: none; background: #fff; }
.wl-field input:focus, .wl-field select:focus { border-color: #3b82f6; }
.mono { font-family: monospace; }
.priority-cell { font-weight: 600; color: #475569; }
.action-badge { display: inline-block; padding: 0.15rem 0.6rem; border-radius: 999px; font-size: 0.75rem; font-weight: 700; }
.action-allow { background: #dcfce7; color: #166534; }
.action-deny  { background: #fee2e2; color: #991b1b; }
.confirm-text { font-size: 0.83rem; color: #ef4444; font-weight: 600; }
.status-toggle { padding: 0.15rem 0.55rem; border-radius: 999px; border: none; cursor: pointer; font-size: 0.75rem; font-weight: 600; }
.toggle-on  { background: #dcfce7; color: #166534; }
.toggle-off { background: #f1f5f9; color: #64748b; }
.status-toggle:hover { opacity: 0.75; }
.priority-arrows { display: inline-flex; flex-direction: column; gap: 1px; margin-right: 6px; vertical-align: middle; }
.btn-arrow { display: flex; align-items: center; justify-content: center; width: 20px; height: 18px; padding: 0; border: 1px solid #e2e8f0; border-radius: 4px; background: #f8fafc; color: #475569; cursor: pointer; line-height: 1; }
.btn-arrow:hover:not(:disabled) { background: #e2e8f0; color: #1e293b; }
.btn-arrow:disabled { opacity: 0.3; cursor: not-allowed; }
</style>
