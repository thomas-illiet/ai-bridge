<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { listUsers, updateUserRole } from '@/services/api'
import { formatDate, fmtNum } from '@/utils/format'
import PaginationBar from '@/components/PaginationBar.vue'
import UserStatsModal from './UserStatsModal.vue'
import DeleteUserModal from './DeleteUserModal.vue'
import { useMinLoad } from '@/composables/useMinLoad'

export interface RegisteredUser {
  id: string; username: string; email: string; role: 'admin' | 'manager' | 'user' | 'none' | 'service'
  roleExpiresAt: string | null
  createdAt: string; totalRequests: number; totalInput: number; totalOutput: number
}

const auth = useAuthStore()

const users      = ref<RegisteredUser[]>([])
const { loading, withLoad } = useMinLoad(300, true)
const error      = ref<string | null>(null)
const saving     = ref(false)
const sortBy     = ref('created_at')
const sortDir    = ref<'asc' | 'desc'>('asc')

function sortIcon(col: string) {
  if (sortBy.value !== col) return '↕'
  return sortDir.value === 'asc' ? '↑' : '↓'
}

function toggleSort(col: string) {
  if (sortBy.value === col) { sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc' }
  else { sortBy.value = col; sortDir.value = 'desc' }
  page.value = 1
  loadUsers()
}

async function loadUsers() {
  await withLoad(async () => {
    try {
      const res = await listUsers(sortBy.value, sortDir.value)
      users.value = res.data.users ?? []
    } catch { error.value = 'Failed to load users' }
  })
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
    <div class="card">
    <h2 class="card-title">Users</h2>
    <div v-if="!loading && error" class="state-msg error">{{ error }}</div>
    <div v-else-if="!loading && !error && users.length === 0" class="empty-card">
      <p class="empty-title">No users found</p>
    </div>
    <table v-else class="data-table">
      <thead>
        <tr>
          <th class="sortable" :class="{ active: sortBy === 'username' }" @click="toggleSort('username')">Username <span class="sort-icon">{{ sortIcon('username') }}</span></th>
          <th class="sortable" :class="{ active: sortBy === 'email' }" @click="toggleSort('email')">Email <span class="sort-icon">{{ sortIcon('email') }}</span></th>
          <th class="col-center sortable" :class="{ active: sortBy === 'role' }" @click="toggleSort('role')">Role <span class="sort-icon">{{ sortIcon('role') }}</span></th>
          <th class="sortable" :class="{ active: sortBy === 'role_expires_at' }" @click="toggleSort('role_expires_at')">Expires <span class="sort-icon">{{ sortIcon('role_expires_at') }}</span></th>
          <th class="num sortable" :class="{ active: sortBy === 'total_requests' }" @click="toggleSort('total_requests')">Requests <span class="sort-icon">{{ sortIcon('total_requests') }}</span></th>
          <th class="num sortable" :class="{ active: sortBy === 'total_input' }" @click="toggleSort('total_input')">Input <span class="sort-icon">{{ sortIcon('total_input') }}</span></th>
          <th class="num sortable" :class="{ active: sortBy === 'total_output' }" @click="toggleSort('total_output')">Output <span class="sort-icon">{{ sortIcon('total_output') }}</span></th>
          <th class="sortable" :class="{ active: sortBy === 'created_at' }" @click="toggleSort('created_at')">Registered <span class="sort-icon">{{ sortIcon('created_at') }}</span></th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <template v-if="loading">
          <tr v-for="i in 6" :key="i" class="skeleton-row">
            <td><div class="skeleton-bar skeleton-bar--md" /></td>
            <td><div class="skeleton-bar skeleton-bar--lg" /></td>
            <td class="col-center"><div class="skeleton-bar skeleton-bar--pill" style="margin:auto" /></td>
            <td><div class="skeleton-bar skeleton-bar--sm" /></td>
            <td class="num"><div class="skeleton-bar skeleton-bar--xs" style="margin:auto" /></td>
            <td class="num"><div class="skeleton-bar skeleton-bar--xs" style="margin:auto" /></td>
            <td class="num"><div class="skeleton-bar skeleton-bar--xs" style="margin:auto" /></td>
            <td><div class="skeleton-bar skeleton-bar--sm" /></td>
            <td><div class="skeleton-bar skeleton-bar--btn" /></td>
          </tr>
        </template>
        <template v-else>
        <tr v-for="u in pagedUsers" :key="u.id">
          <td class="bold">{{ u.username }}</td>
          <td class="muted">{{ u.email || '—' }}</td>
          <td class="col-center"><span class="badge" :class="roleBadgeClass(u.role)">{{ u.role }}</span></td>
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
        </template>
      </tbody>
    </table>

    <PaginationBar
      v-if="!loading && users.length > 0"
      v-model:page="page"
      v-model:pageSize="pageSize"
      :total="users.length"
    />
    </div>

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
.expiry-label { font-size: 0.83rem; color: #475569; }
.expiry-label.expired { color: #dc2626; font-weight: 600; }
.role-select-full { width: 100%; padding: 0.45rem 0.6rem; border: 1px solid #d1d5db; border-radius: 6px; font-size: 0.9rem; background: white; outline: none; }
.role-select-full:focus { border-color: #3b82f6; box-shadow: 0 0 0 2px #bfdbfe; }
</style>
