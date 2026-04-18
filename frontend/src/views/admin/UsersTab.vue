<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { listUsers, updateUserRole } from '@/services/api'
import { formatDate, fmtNum } from '@/utils/format'
import PaginationBar from '@/components/PaginationBar.vue'
import UserStatsModal from './UserStatsModal.vue'
import DeleteUserModal from './DeleteUserModal.vue'

export interface RegisteredUser {
  id: string; username: string; email: string; role: 'admin' | 'manager' | 'user' | 'none' | 'service'
  roleExpiresAt: string | null
  createdAt: string; totalRequests: number; totalInput: number; totalOutput: number
}

const auth = useAuthStore()

const users      = ref<RegisteredUser[]>([])
const loading    = ref(true)
const error      = ref<string | null>(null)
const saving     = ref(false)

async function loadUsers() {
  try {
    const res = await listUsers()
    users.value = res.data.users ?? []
  } catch { error.value = 'Failed to load users' }
  finally { loading.value = false }
}

// ── edit role modal ───────────────────────────────────────────────────────
const editUser    = ref<RegisteredUser | null>(null)
const editRole    = ref<string>('user')
const editExpiry  = ref<string>('')

function openEditModal(u: RegisteredUser) {
  editUser.value   = u
  editRole.value   = u.role
  editExpiry.value = u.roleExpiresAt ? u.roleExpiresAt.slice(0, 10) : ''
  closeMenus()
}

async function saveEdit() {
  if (!editUser.value) return
  saving.value = true
  try {
    await updateUserRole(editUser.value.id, editRole.value, editExpiry.value || undefined)
    editUser.value.role = editRole.value as RegisteredUser['role']
    editUser.value.roleExpiresAt = editExpiry.value
      ? new Date(editExpiry.value + 'T23:59:59Z').toISOString()
      : null
    editUser.value = null
  } catch (e: any) {
    error.value = e?.response?.data?.error ?? 'Failed to update'
  } finally { saving.value = false }
}

// ── helpers ────────────────────────────────────────────────────────────────
function expiryLabel(iso: string | null): string {
  if (!iso) return 'Never'
  return new Date(iso).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
}

function isExpired(iso: string | null): boolean {
  return !!iso && new Date(iso) < new Date()
}

// ── pagination ────────────────────────────────────────────────────────────
const page     = ref(1)
const pageSize = ref(10)
const pagedUsers = computed(() => {
  const s = (page.value - 1) * pageSize.value
  return users.value.slice(s, s + pageSize.value)
})

// ── action dropdown ───────────────────────────────────────────────────────
const openMenuId = ref<string | null>(null)
function toggleMenu(id: string) { openMenuId.value = openMenuId.value === id ? null : id }
function closeMenus() { openMenuId.value = null }

function onDocClick() { closeMenus() }
onMounted(() => { document.addEventListener('click', onDocClick); loadUsers() })
onBeforeUnmount(() => document.removeEventListener('click', onDocClick))

// ── modals ────────────────────────────────────────────────────────────────
const statsUser  = ref<RegisteredUser | null>(null)
const deleteUser = ref<RegisteredUser | null>(null)

function onDeleted(id: string) {
  users.value = users.value.filter(u => u.id !== id)
  deleteUser.value = null
}

function roleBadgeClass(role: string) {
  return { 'badge-admin': role === 'admin', 'badge-manager': role === 'manager', 'badge-user': role === 'user', 'badge-none': role === 'none', 'badge-service': role === 'service' }
}
</script>

<template>
  <div class="tab-content">
    <div v-if="loading" class="state-msg">Loading…</div>
    <div v-else-if="error" class="state-msg error">{{ error }}</div>
    <table v-else class="data-table">
      <thead>
        <tr>
          <th>Username</th>
          <th>Email</th>
          <th>Role</th>
          <th>Expires</th>
          <th class="num">Requests</th>
          <th class="num">Input tokens</th>
          <th class="num">Output tokens</th>
          <th>Registered</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="u in pagedUsers" :key="u.id">
          <td class="bold">{{ u.username }}</td>
          <td class="muted">{{ u.email || '—' }}</td>
          <td><span class="badge" :class="roleBadgeClass(u.role)">{{ u.role }}</span></td>
          <td>
            <span class="expiry-label" :class="{ expired: isExpired(u.roleExpiresAt) }">
              {{ expiryLabel(u.roleExpiresAt) }}
            </span>
          </td>
          <td class="num">{{ fmtNum(u.totalRequests) }}</td>
          <td class="num">{{ fmtNum(u.totalInput) }}</td>
          <td class="num">{{ fmtNum(u.totalOutput) }}</td>
          <td class="muted">{{ formatDate(u.createdAt) }}</td>
          <td class="actions">
            <div class="action-menu">
              <button class="btn-action-trigger" @click.stop="toggleMenu(u.id)">
                Actions <span class="chevron-down">▾</span>
              </button>
              <div v-if="openMenuId === u.id" class="action-dropdown">
                <button class="action-item" @click="statsUser = u; closeMenus()">View stats</button>
                <button
                  class="action-item"
                  :disabled="u.id === auth.tokenParsed?.sub || u.role === 'service' || (auth.isManager && u.role === 'admin')"
                  :title="u.role === 'service' ? 'Manage service accounts in the Service Accounts tab' : u.id === auth.tokenParsed?.sub ? 'Cannot edit your own role' : (auth.isManager && u.role === 'admin') ? 'Managers cannot edit admin accounts' : ''"
                  @click="u.role !== 'service' && u.id !== auth.tokenParsed?.sub && !(auth.isManager && u.role === 'admin') && openEditModal(u)"
                >Edit role &amp; expiry</button>
                <div class="action-divider" />
                <button
                  class="action-item danger"
                  :disabled="u.id === auth.tokenParsed?.sub || u.role === 'service' || (auth.isManager && u.role === 'admin')"
                  :title="u.role === 'service' ? 'Delete service accounts in the Service Accounts tab' : u.id === auth.tokenParsed?.sub ? 'Cannot delete your own account' : (auth.isManager && u.role === 'admin') ? 'Managers cannot delete admin accounts' : ''"
                  @click="u.role !== 'service' && u.id !== auth.tokenParsed?.sub && !(auth.isManager && u.role === 'admin') && (deleteUser = u, closeMenus())"
                >Delete user</button>
              </div>
            </div>
          </td>
        </tr>
      </tbody>
    </table>

    <PaginationBar
      v-if="users.length > 0"
      v-model:page="page"
      v-model:pageSize="pageSize"
      :total="users.length"
    />

    <UserStatsModal  :user="statsUser"  @close="statsUser = null" />
    <DeleteUserModal :user="deleteUser" @deleted="onDeleted" @close="deleteUser = null" />

    <!-- Edit role & expiry modal -->
    <Teleport to="body">
      <div v-if="editUser" class="modal-backdrop" @click.self="editUser = null">
        <div class="modal">
          <h3>Edit role &amp; expiry</h3>
          <p class="modal-sub">User: <strong>{{ editUser.username }}</strong></p>

          <div class="form-group">
            <label>Role</label>
            <select v-model="editRole" class="role-select-full">
              <option v-if="auth.isAdmin" value="admin">admin</option>
              <option v-if="auth.isAdmin" value="manager">manager</option>
              <option value="user">user</option>
              <option value="none">none</option>
            </select>
          </div>

          <div class="form-group">
            <label>Role expires on <span class="optional">(leave empty = never)</span></label>
            <input type="date" v-model="editExpiry" class="date-input" />
          </div>

          <div class="modal-actions">
            <button class="btn-cancel" @click="editUser = null">Cancel</button>
            <button class="btn-primary" :disabled="saving" @click="saveEdit">
              {{ saving ? 'Saving…' : 'Save' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.tab-content { display: flex; flex-direction: column; gap: 1.25rem; }
.state-msg { color: #64748b; font-size: 0.9rem; }
.state-msg.error { color: #ef4444; }
.data-table { width: 100%; border-collapse: collapse; font-size: 0.88rem; background: white; border: 1px solid #e2e8f0; border-radius: 10px; }
.data-table th { text-align: left; padding: 0.55rem 0.9rem; font-size: 0.75rem; font-weight: 600; color: #64748b; text-transform: uppercase; letter-spacing: 0.03em; border-bottom: 1px solid #e2e8f0; background: #f8fafc; white-space: nowrap; }
.data-table th.num { text-align: right; }
.data-table thead tr:first-child th:first-child { border-radius: 10px 0 0 0; }
.data-table thead tr:first-child th:last-child  { border-radius: 0 10px 0 0; }
.data-table tbody tr:last-child td:first-child   { border-radius: 0 0 0 10px; }
.data-table tbody tr:last-child td:last-child    { border-radius: 0 0 10px 0; }
.data-table td { padding: 0.65rem 0.9rem; border-bottom: 1px solid #f1f5f9; }
.data-table tr:last-child td { border-bottom: none; }
.bold { font-weight: 600; color: #1e293b; }
.muted { color: #64748b; }
.num { text-align: right; font-variant-numeric: tabular-nums; color: #334155; font-weight: 500; }
.actions { display: flex; align-items: center; gap: 0.4rem; white-space: nowrap; }
.badge { padding: 0.18rem 0.55rem; border-radius: 999px; font-size: 0.75rem; font-weight: 600; }
.badge-admin    { background: #ede9fe; color: #6d28d9; }
.badge-manager  { background: #fef3c7; color: #92400e; }
.badge-user     { background: #dcfce7; color: #166534; }
.badge-none     { background: #f1f5f9; color: #64748b; }
.badge-service  { background: #dbeafe; color: #1d4ed8; }

.expiry-label { font-size: 0.83rem; color: #475569; }
.expiry-label.expired { color: #dc2626; font-weight: 600; }

.action-menu { position: relative; display: inline-block; }
.btn-action-trigger { display: flex; align-items: center; gap: 0.3rem; padding: 0.25rem 0.65rem; border: 1px solid #e2e8f0; border-radius: 6px; background: white; color: #374151; font-size: 0.82rem; font-weight: 500; cursor: pointer; white-space: nowrap; }
.btn-action-trigger:hover { background: #f1f5f9; border-color: #cbd5e1; }
.chevron-down { font-size: 0.7rem; color: #94a3b8; }
.action-dropdown { position: absolute; right: 0; top: calc(100% + 4px); z-index: 200; min-width: 160px; background: white; border: 1px solid #e2e8f0; border-radius: 8px; box-shadow: 0 6px 20px rgba(0,0,0,0.12); padding: 0.25rem 0; }
.action-item { display: block; width: 100%; padding: 0.5rem 0.9rem; background: none; border: none; text-align: left; font-size: 0.85rem; font-weight: 500; color: #374151; cursor: pointer; }
.action-item:hover:not(:disabled) { background: #f8fafc; }
.action-item.danger { color: #dc2626; }
.action-item.danger:hover:not(:disabled) { background: #fef2f2; }
.action-item:disabled { opacity: 0.4; cursor: not-allowed; }
.action-divider { height: 1px; background: #f1f5f9; margin: 0.2rem 0; }

/* Modal */
.modal-backdrop { position: fixed; inset: 0; background: rgba(0,0,0,0.45); display: flex; align-items: center; justify-content: center; z-index: 200; }
.modal { background: white; border-radius: 12px; padding: 1.75rem; width: 100%; max-width: 380px; display: flex; flex-direction: column; gap: 1rem; }
.modal h3 { font-size: 1.1rem; font-weight: 700; margin: 0; }
.modal-sub { font-size: 0.9rem; color: #475569; margin: 0; }
.form-group { display: flex; flex-direction: column; gap: 0.35rem; }
.form-group label { font-size: 0.85rem; font-weight: 600; color: #374151; }
.optional { font-weight: 400; color: #94a3b8; }
.role-select-full { width: 100%; padding: 0.45rem 0.6rem; border: 1px solid #d1d5db; border-radius: 6px; font-size: 0.9rem; background: white; outline: none; }
.role-select-full:focus { border-color: #3b82f6; box-shadow: 0 0 0 2px #bfdbfe; }
.date-input { padding: 0.45rem 0.6rem; border: 1px solid #d1d5db; border-radius: 6px; font-size: 0.9rem; outline: none; }
.date-input:focus { border-color: #3b82f6; box-shadow: 0 0 0 2px #bfdbfe; }
.modal-actions { display: flex; justify-content: flex-end; gap: 0.6rem; }
.btn-cancel { padding: 0.4rem 1rem; border: 1px solid #e2e8f0; border-radius: 6px; background: white; color: #374151; font-size: 0.88rem; cursor: pointer; }
.btn-cancel:hover { background: #f8fafc; }
.btn-primary { padding: 0.4rem 1rem; border: none; border-radius: 6px; background: #3b82f6; color: white; font-size: 0.88rem; font-weight: 600; cursor: pointer; }
.btn-primary:hover:not(:disabled) { background: #2563eb; }
.btn-primary:disabled { opacity: 0.55; cursor: not-allowed; }
</style>
