<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { listProviders, createProvider, updateProvider, deleteProvider, reloadProviders, type AIProvider } from '@/services/api'
import { formatDate } from '@/utils/format'
import PaginationBar from '@/components/PaginationBar.vue'
import LoadingOverlay from '@/components/LoadingOverlay.vue'
import { useMinLoad } from '@/composables/useMinLoad'

const providers     = ref<AIProvider[]>([])
const { loading, withLoad } = useMinLoad(300, true)
const error         = ref<string | null>(null)
const { loading: reloading, withLoad: withReload } = useMinLoad(1200)

// ── Add/Edit modal ──────────────────────────────────────────────────────────
const showForm  = ref(false)
const saving    = ref(false)
const formError = ref<string | null>(null)
const editingId = ref<string | null>(null)

const form = ref({
  name:         '',
  display_name: '',
  type:         'openai' as 'openai' | 'ollama' | 'anthropic',
  base_url:     '',
  api_key:      '',
  enabled:      true,
})

function openAdd() {
  editingId.value = null
  form.value = { name: '', display_name: '', type: 'openai', base_url: '', api_key: '', enabled: true }
  formError.value = null
  showForm.value = true
}

function openEdit(p: AIProvider) {
  editingId.value = p.id
  form.value = { name: p.name, display_name: p.displayName ?? '', type: p.type, base_url: p.baseUrl, api_key: '', enabled: p.enabled }
  formError.value = null
  showForm.value = true
}

function closeForm() {
  showForm.value = false
  formError.value = null
}

async function submitForm() {
  if (!form.value.name.trim()) { formError.value = 'Name is required'; return }
  if (form.value.type === 'ollama' && !form.value.base_url.trim()) {
    formError.value = 'Base URL is required for Ollama providers'
    return
  }
  formError.value = null
  saving.value = true
  try {
    if (editingId.value) {
      const body: any = { name: form.value.name, display_name: form.value.display_name, base_url: form.value.base_url, enabled: form.value.enabled }
      if (form.value.api_key) body.api_key = form.value.api_key
      await updateProvider(editingId.value, body)
    } else {
      await createProvider({
        name:         form.value.name,
        display_name: form.value.display_name || undefined,
        type:         form.value.type,
        base_url:     form.value.base_url,
        api_key:      form.value.api_key || undefined,
        enabled:      form.value.enabled,
      })
    }
    closeForm()
    await loadProviders()
  } catch (e: any) {
    formError.value = e?.response?.data?.error ?? 'Failed to save provider'
  } finally {
    saving.value = false
  }
}

// ── Delete modal ─────────────────────────────────────────────────────────────
const delTarget  = ref<AIProvider | null>(null)
const deleting   = ref(false)
const delError   = ref<string | null>(null)

function openDelete(p: AIProvider) {
  delTarget.value = p
  delError.value  = null
}

function closeDelete() {
  delTarget.value = null
  delError.value  = null
}

async function confirmDelete() {
  if (!delTarget.value) return
  deleting.value = true
  delError.value = null
  try {
    await deleteProvider(delTarget.value.id)
    providers.value = providers.value.filter(p => p.id !== delTarget.value!.id)
    closeDelete()
  } catch (e: any) {
    delError.value = e?.response?.data?.error ?? 'Failed to delete provider'
  } finally {
    deleting.value = false
  }
}

// ── Search + sort + pagination ────────────────────────────────────────────────
const search = ref('')
let searchTimer: ReturnType<typeof setTimeout> | null = null
watch(search, () => {
  page.value = 1
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => loadProviders(), 300)
})

const page     = ref(1)
const pageSize = ref(10)
watch(pageSize, () => { page.value = 1 })
const pagedProviders = computed(() => {
  const s = (page.value - 1) * pageSize.value
  return providers.value.slice(s, s + pageSize.value)
})

const sortBy  = ref('created_at')
const sortDir = ref<'asc' | 'desc'>('desc')

function toggleSort(col: string) {
  if (sortBy.value === col) { sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc' }
  else { sortBy.value = col; sortDir.value = 'desc' }
  page.value = 1
  loadProviders()
}

const isEmpty = computed(() => providers.value.length === 0)

async function loadProviders() {
  error.value = null
  await withLoad(async () => {
    try {
      const res = await listProviders(search.value)
      let list = res.data.providers ?? []
      list = [...list].sort((a, b) => {
        const av = (a as any)[sortBy.value] ?? ''
        const bv = (b as any)[sortBy.value] ?? ''
        const cmp = String(av).localeCompare(String(bv))
        return sortDir.value === 'asc' ? cmp : -cmp
      })
      providers.value = list
    } catch {
      error.value = 'Failed to load providers'
    }
  })
}

async function toggleEnabled(p: AIProvider) {
  try {
    await updateProvider(p.id, { enabled: !p.enabled })
    p.enabled = !p.enabled
  } catch {
    error.value = 'Failed to toggle provider'
  }
}

async function forceReload() {
  await withReload(async () => {
    try {
      await reloadProviders()
    } catch (e: any) {
      error.value = e?.response?.data?.error ?? 'Failed to reload bridge'
    }
  })
}

onMounted(loadProviders)
</script>

<template>
  <Teleport defer to="#admin-search-portal">
    <div class="portal-search-wrap">
      <svg class="portal-search-icon" xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
      <input v-model="search" type="text" placeholder="Search providers…" class="portal-input" />
    </div>
  </Teleport>

  <div class="tab-content">

    <LoadingOverlay :visible="reloading" message="Reloading providers…" />

    <!-- ── Providers table card ─────────────────────────────────────────── -->
    <div class="card">
      <div class="card-header">
        <h2 class="card-title">Provider entries <span class="title-count">{{ providers.length }}</span></h2>
        <div class="header-actions">
          <button v-if="!isEmpty" class="btn btn-secondary btn-sm" :disabled="reloading" @click="forceReload">
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
            {{ reloading ? 'Reloading…' : 'Force Reload' }}
          </button>
          <button class="btn btn-primary btn-sm" @click="openAdd">
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            Add Provider
          </button>
        </div>
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
            <rect x="2" y="3" width="20" height="14" rx="2"/><path d="M8 21h8M12 17v4"/>
          </svg>
        </div>
        <p class="empty-title">No providers configured</p>
        <p class="empty-sub">Click "Add Provider" to configure an OpenAI or Ollama provider.</p>
      </div>

      <table v-else class="data-table">
        <thead><tr>
          <th class="sortable" :class="{ active: sortBy === 'name' }" @click="toggleSort('name')">
            Name <span class="sort-icon">{{ sortBy === 'name' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span>
          </th>
          <th>Type</th>
          <th>Base URL</th>
          <th>API Key</th>
          <th class="sortable" :class="{ active: sortBy === 'enabled' }" @click="toggleSort('enabled')">
            Status <span class="sort-icon">{{ sortBy === 'enabled' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span>
          </th>
          <th class="sortable" :class="{ active: sortBy === 'createdAt' }" @click="toggleSort('createdAt')">
            Created <span class="sort-icon">{{ sortBy === 'createdAt' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span>
          </th>
          <th>Actions</th>
        </tr></thead>
        <tbody>
          <template v-if="loading">
            <tr v-for="i in 3" :key="i" class="skeleton-row">
              <td><div class="skeleton-bar skeleton-bar--md" /></td>
              <td><div class="skeleton-bar skeleton-bar--sm" /></td>
              <td><div class="skeleton-bar skeleton-bar--lg" /></td>
              <td><div class="skeleton-bar skeleton-bar--sm" /></td>
              <td><div class="skeleton-bar skeleton-bar--btn" /></td>
              <td><div class="skeleton-bar skeleton-bar--sm" /></td>
              <td><div class="skeleton-bar skeleton-bar--btn" /></td>
            </tr>
          </template>
          <template v-else>
            <tr v-for="p in pagedProviders" :key="p.id">
              <td>
                <span class="mono">{{ p.name }}</span>
                <span v-if="p.displayName" class="display-name">{{ p.displayName }}</span>
              </td>
              <td><span class="badge" :class="'badge-' + p.type">{{ p.type }}</span></td>
              <td class="muted mono-sm">{{ p.baseUrl || '—' }}</td>
              <td>
                <span v-if="p.apiKeySet" class="badge badge-active">set</span>
                <span v-else class="badge badge-none">none</span>
              </td>
              <td>
                <button class="status-toggle" :class="p.enabled ? 'toggle-on' : 'toggle-off'" @click="toggleEnabled(p)">
                  {{ p.enabled ? 'Enabled' : 'Disabled' }}
                </button>
              </td>
              <td class="muted">{{ formatDate(p.createdAt) }}</td>
              <td class="actions">
                <button class="btn btn-sm btn-secondary" @click="openEdit(p)">
                  <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                  Edit
                </button>
                <button class="btn btn-sm btn-danger" @click="openDelete(p)">
                  <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"/></svg>
                  Delete
                </button>
              </td>
            </tr>
          </template>
        </tbody>
      </table>
      <PaginationBar v-if="!loading && providers.length > 0" v-model:page="page" v-model:pageSize="pageSize" :total="providers.length" />
    </div>

    <!-- ── Add / Edit modal ───────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showForm" class="modal-backdrop" @click.self="closeForm">
        <div class="modal">
          <div class="modal-header">
            <h2>{{ editingId ? 'Edit Provider' : 'Add Provider' }}</h2>
            <button class="modal-close" @click="closeForm">✕</button>
          </div>

          <div class="form-grid">
            <div class="pv-field">
              <label>Name</label>
              <input v-model="form.name" placeholder="e.g. openai, ollama-local" :disabled="!!editingId" @keyup.enter="submitForm" />
            </div>
            <div v-if="!editingId" class="pv-field pv-field--sm">
              <label>Type</label>
              <select v-model="form.type">
                <option value="openai">OpenAI</option>
                <option value="ollama">Ollama</option>
                <option value="anthropic">Anthropic</option>
              </select>
            </div>
            <div class="pv-field pv-field--full">
              <label>Display Name <span class="optional">(optional)</span></label>
              <input v-model="form.display_name" placeholder="e.g. Claude Anthropic, OpenAI Production" @keyup.enter="submitForm" />
            </div>
            <div class="pv-field pv-field--full">
              <label>Base URL <span class="optional">(optional for OpenAI &amp; Anthropic)</span></label>
              <input v-model="form.base_url" placeholder="e.g. http://ollama:11434" @keyup.enter="submitForm" />
            </div>
            <div class="pv-field pv-field--full">
              <label>API Key <span class="optional">{{ editingId ? '(leave blank to keep current)' : '(optional)' }}</span></label>
              <input v-model="form.api_key" type="password" placeholder="sk-..." @keyup.enter="submitForm" />
            </div>
            <div class="pv-field pv-field--toggle">
              <label>Enabled</label>
              <label class="toggle-label">
                <input type="checkbox" v-model="form.enabled" />
                <span class="toggle-slider" />
              </label>
            </div>
          </div>

          <p v-if="formError" class="form-error">{{ formError }}</p>

          <div class="modal-actions">
            <button class="btn btn-outline" @click="closeForm">
              <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              Cancel
            </button>
            <button class="btn btn-primary-solid" :disabled="saving" @click="submitForm">
              <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
              {{ saving ? 'Saving…' : (editingId ? 'Update' : 'Add') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ── Delete confirmation modal ──────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="delTarget" class="modal-backdrop" @click.self="closeDelete">
        <div class="modal">
          <div class="modal-header">
            <h2>Delete provider</h2>
            <button class="modal-close" @click="closeDelete">✕</button>
          </div>
          <p class="confirm-text">
            Are you sure you want to permanently delete <strong>{{ delTarget.name }}</strong>?
            This will remove it from the database and cannot be undone.
          </p>
          <p v-if="delError" class="form-error">{{ delError }}</p>
          <div class="modal-actions">
            <button class="btn btn-outline" @click="closeDelete">
              <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              Cancel
            </button>
            <button class="btn btn-danger-solid" :disabled="deleting" @click="confirmDelete">
              <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"/></svg>
              {{ deleting ? 'Deleting…' : 'Delete' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>

<style scoped>

.mono { font-family: monospace; font-size: 0.88rem; }
.mono-sm { font-family: monospace; font-size: 0.8rem; max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.muted { color: #94a3b8; }

.status-toggle { padding: 0.15rem 0.55rem; border-radius: 999px; border: none; cursor: pointer; font-size: 0.75rem; font-weight: 600; }
.toggle-on  { background: #dcfce7; color: #166534; }
.toggle-off { background: #f1f5f9; color: #64748b; }
.status-toggle:hover { opacity: 0.75; }

.badge-openai    { background: #dbeafe; color: #1d4ed8; }
.badge-ollama    { background: #f3e8ff; color: #7c3aed; }
.badge-anthropic { background: #fef3c7; color: #92400e; }
.badge-none      { background: #f1f5f9; color: #94a3b8; }

.display-name { display: block; font-size: 0.75rem; color: #64748b; margin-top: 0.1rem; }

.actions { display: flex; gap: 0.4rem; align-items: center; }

/* ── Modal shared ─────────────────────────────────────────────────────────── */
.modal-backdrop { position: fixed; inset: 0; background: rgba(0,0,0,0.45); display: flex; align-items: center; justify-content: center; z-index: 200; padding: 1rem; }
.modal { background: white; border-radius: 14px; padding: 1.75rem; width: 100%; max-width: 480px; display: flex; flex-direction: column; gap: 1.25rem; box-shadow: 0 20px 60px rgba(0,0,0,0.25); }
.modal-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 1rem; }
.modal-header h2 { font-size: 1.15rem; font-weight: 700; margin: 0; }
.modal-close { background: none; border: none; font-size: 1.1rem; cursor: pointer; color: #64748b; padding: 0.2rem 0.4rem; border-radius: 5px; }
.modal-close:hover { background: #f1f5f9; }
.modal-actions { display: flex; justify-content: flex-end; gap: 0.6rem; }

.confirm-text { font-size: 0.92rem; color: #475569; line-height: 1.6; margin: 0; }
.confirm-text strong { color: #1e293b; }

/* ── Form grid ────────────────────────────────────────────────────────────── */
.form-grid { display: flex; flex-wrap: wrap; gap: 0.75rem; }
.pv-field { display: flex; flex-direction: column; gap: 0.3rem; flex: 1; min-width: 140px; }
.pv-field--sm { flex: 0 0 110px; min-width: 100px; }
.pv-field--full { flex: 1 1 100%; }
.pv-field--toggle { align-items: flex-start; flex: 0 0 auto; }
.pv-field label { font-size: 0.78rem; font-weight: 600; color: #64748b; }
.pv-field input, .pv-field select { padding: 0.45rem 0.75rem; border: 1px solid #e2e8f0; border-radius: 6px; font-size: 0.9rem; outline: none; background: white; }
.pv-field input:focus, .pv-field select:focus { border-color: #3b82f6; }
.pv-field input:disabled { background: #f8fafc; color: #94a3b8; cursor: not-allowed; }
.optional { font-weight: 400; color: #94a3b8; }

.toggle-label { display: flex; align-items: center; gap: 0.5rem; cursor: pointer; padding: 0.45rem 0; }
.toggle-label input { display: none; }
.toggle-slider { width: 36px; height: 20px; background: #cbd5e1; border-radius: 999px; position: relative; transition: background 0.2s; flex-shrink: 0; }
.toggle-slider::after { content: ''; position: absolute; top: 2px; left: 2px; width: 16px; height: 16px; background: white; border-radius: 50%; transition: transform 0.2s; }
.toggle-label input:checked + .toggle-slider { background: #3b82f6; }
.toggle-label input:checked + .toggle-slider::after { transform: translateX(16px); }

.form-error { color: #ef4444; font-size: 0.85rem; margin: 0; }

/* ── Buttons (scoped — only what shared.css doesn't cover) ───────────────── */
.btn-primary-solid { display: inline-flex; align-items: center; gap: 0.35em; background: #3b82f6; color: white; border: none; border-radius: 6px; padding: 0.4rem 1rem; font-size: 0.9rem; font-weight: 500; cursor: pointer; }
.btn-primary-solid:hover:not(:disabled) { background: #2563eb; }
.btn-primary-solid:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
