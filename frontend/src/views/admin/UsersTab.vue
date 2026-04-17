<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { listUsers, updateUserRole } from '@/services/api'
import { formatDate, fmtNum } from '@/utils/format'
import PaginationBar from '@/components/PaginationBar.vue'
import UserStatsModal from './UserStatsModal.vue'
import DeleteUserModal from './DeleteUserModal.vue'

export interface RegisteredUser {
  id: string; username: string; email: string; role: 'admin' | 'user' | 'none'
  createdAt: string; totalRequests: number; totalInput: number; totalOutput: number
}

const auth = useAuthStore()

const users      = ref<RegisteredUser[]>([])
const loading    = ref(true)
const error      = ref<string | null>(null)
const updatingId = ref<string | null>(null)

async function loadUsers() {
  try {
    const res = await listUsers()
    users.value = res.data.users ?? []
  } catch { error.value = 'Failed to load users' }
  finally { loading.value = false }
}

async function changeRole(user: RegisteredUser, role: string) {
  updatingId.value = user.id
  try {
    await updateUserRole(user.id, role)
    user.role = role as RegisteredUser['role']
  } catch (e: any) {
    error.value = e?.response?.data?.error ?? 'Failed to update role'
  } finally { updatingId.value = null }
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
  return { 'badge-admin': role === 'admin', 'badge-user': role === 'user', 'badge-none': role === 'none' }
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
          <th class="num">Requests</th>
          <th class="num">Input tokens</th>
          <th class="num">Output tokens</th>
          <th>Registered</th>
          <th>Change role</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="u in pagedUsers" :key="u.id">
          <td class="bold">{{ u.username }}</td>
          <td class="muted">{{ u.email || '—' }}</td>
          <td><span class="badge" :class="roleBadgeClass(u.role)">{{ u.role }}</span></td>
          <td class="num">{{ fmtNum(u.totalRequests) }}</td>
          <td class="num">{{ fmtNum(u.totalInput) }}</td>
          <td class="num">{{ fmtNum(u.totalOutput) }}</td>
          <td class="muted">{{ formatDate(u.createdAt) }}</td>
          <td>
            <select
              :value="u.role"
              :disabled="u.id === auth.tokenParsed?.sub || updatingId === u.id"
              class="role-select"
              @change="changeRole(u, ($event.target as HTMLSelectElement).value)"
            >
              <option value="admin">admin</option>
              <option value="user">user</option>
              <option value="none">none</option>
            </select>
          </td>
          <td class="actions">
            <div class="action-menu">
              <button class="btn-action-trigger" @click.stop="toggleMenu(u.id)">
                Actions <span class="chevron-down">▾</span>
              </button>
              <div v-if="openMenuId === u.id" class="action-dropdown">
                <button class="action-item" @click="statsUser = u; closeMenus()">View stats</button>
                <div class="action-divider" />
                <button
                  class="action-item danger"
                  :disabled="u.id === auth.tokenParsed?.sub"
                  :title="u.id === auth.tokenParsed?.sub ? 'Cannot delete your own account' : ''"
                  @click="u.id !== auth.tokenParsed?.sub && (deleteUser = u, closeMenus())"
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
  </div>
</template>

<style scoped>
.tab-content { display: flex; flex-direction: column; gap: 1.25rem; }
.state-msg { color: #64748b; font-size: 0.9rem; }
.state-msg.error { color: #ef4444; }
.data-table { width: 100%; border-collapse: collapse; font-size: 0.88rem; background: white; border: 1px solid #e2e8f0; border-radius: 10px; }
.data-table th { text-align: left; padding: 0.55rem 0.9rem; font-size: 0.75rem; font-weight: 600; color: #64748b; text-transform: uppercase; letter-spacing: 0.03em; border-bottom: 1px solid #e2e8f0; background: #f8fafc; }
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
.badge-admin { background: #ede9fe; color: #6d28d9; }
.badge-user  { background: #dcfce7; color: #166534; }
.badge-none  { background: #f1f5f9; color: #64748b; }
.role-select { padding: 0.3rem 0.5rem; border: 1px solid #e2e8f0; border-radius: 6px; font-size: 0.85rem; background: white; cursor: pointer; }
.role-select:disabled { opacity: 0.4; cursor: not-allowed; }
.action-menu { position: relative; display: inline-block; }
.btn-action-trigger { display: flex; align-items: center; gap: 0.3rem; padding: 0.25rem 0.65rem; border: 1px solid #e2e8f0; border-radius: 6px; background: white; color: #374151; font-size: 0.82rem; font-weight: 500; cursor: pointer; white-space: nowrap; }
.btn-action-trigger:hover { background: #f1f5f9; border-color: #cbd5e1; }
.chevron-down { font-size: 0.7rem; color: #94a3b8; }
.action-dropdown { position: absolute; right: 0; top: calc(100% + 4px); z-index: 200; min-width: 148px; background: white; border: 1px solid #e2e8f0; border-radius: 8px; box-shadow: 0 6px 20px rgba(0,0,0,0.12); padding: 0.25rem 0; }
.action-item { display: block; width: 100%; padding: 0.5rem 0.9rem; background: none; border: none; text-align: left; font-size: 0.85rem; font-weight: 500; color: #374151; cursor: pointer; }
.action-item:hover { background: #f8fafc; }
.action-item.danger { color: #dc2626; }
.action-item.danger:hover:not(:disabled) { background: #fef2f2; }
.action-item:disabled { opacity: 0.4; cursor: not-allowed; }
.action-divider { height: 1px; background: #f1f5f9; margin: 0.2rem 0; }
</style>
