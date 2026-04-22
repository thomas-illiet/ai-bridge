<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { adminListAccessRequests, adminApproveRequest, adminRejectRequest } from '@/services/api'
import type { AccessRequest } from '@/services/api'
import { useMinLoad } from '@/composables/useMinLoad'

const auth = useAuthStore()

const requests    = ref<AccessRequest[]>([])
const pendingCount = ref(0)
const { loading, withLoad } = useMinLoad(300, true)
const statusFilter = ref('pending')
const sortBy      = ref('created_at')
const sortDir     = ref<'asc' | 'desc'>('desc')

function toggleSort(col: string) {
  if (sortBy.value === col) { sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc' }
  else { sortBy.value = col; sortDir.value = 'desc' }
  load()
}

const approveModal = ref<AccessRequest | null>(null)
const rejectModal  = ref<AccessRequest | null>(null)
const approveRole  = ref('user')
const approveExpiry = ref('')
const rejectNote   = ref('')
const saving       = ref(false)

async function load() {
  await withLoad(async () => {
    const res = await adminListAccessRequests(statusFilter.value || undefined, sortBy.value, sortDir.value)
    requests.value   = res.data.requests ?? []
    pendingCount.value = res.data.pendingCount
  })
}

async function approve() {
  if (!approveModal.value) return
  saving.value = true
  try {
    await adminApproveRequest(approveModal.value.id, approveRole.value, approveExpiry.value || undefined)
    approveModal.value = null
    approveRole.value = 'user'
    approveExpiry.value = ''
    load()
  } finally { saving.value = false }
}

async function reject() {
  if (!rejectModal.value || !rejectNote.value.trim()) return
  saving.value = true
  try {
    await adminRejectRequest(rejectModal.value.id, rejectNote.value.trim())
    rejectModal.value = null
    rejectNote.value = ''
    load()
  } finally { saving.value = false }
}

function openApprove(req: AccessRequest) {
  approveModal.value = req
  approveRole.value = 'user'
  approveExpiry.value = ''
}

function openReject(req: AccessRequest) {
  rejectModal.value = req
  rejectNote.value = ''
}

function statusLabel(s: string) {
  return { pending: 'Pending', approved: 'Approved', rejected: 'Rejected' }[s] ?? s
}

function formatDate(d: string | null) {
  if (!d) return '—'
  return new Date(d).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
}

onMounted(load)
</script>

<template>
  <div class="tab-content">
    <div class="card">
    <div class="card-header">
      <h2 class="card-title">Access Requests</h2>
      <div class="header-actions">
        <p class="sub">{{ pendingCount }} pending request{{ pendingCount !== 1 ? 's' : '' }}.</p>
        <select v-model="statusFilter" class="role-select" @change="load">
          <option value="">All statuses</option>
          <option value="pending">Pending</option>
          <option value="approved">Approved</option>
          <option value="rejected">Rejected</option>
        </select>
      </div>
    </div>

    <div v-if="!loading && requests.length === 0" class="empty-card">
      <div class="empty-icon">
        <svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
          <polyline points="14 2 14 8 20 8"/>
          <line x1="9" y1="13" x2="15" y2="13"/>
          <line x1="9" y1="17" x2="11" y2="17"/>
        </svg>
      </div>
      <p class="empty-title">No requests found</p>
      <p class="empty-sub">{{ statusFilter ? 'Try changing the status filter.' : 'When users request access, they will appear here.' }}</p>
    </div>

    <table v-else class="data-table">
      <thead>
        <tr>
          <th>User</th>
          <th>Email</th>
          <th>Reason</th>
          <th class="sortable" :class="{ active: sortBy === 'created_at' }" @click="toggleSort('created_at')">Submitted <span class="sort-icon">{{ sortBy === 'created_at' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
          <th class="sortable" :class="{ active: sortBy === 'status' }" @click="toggleSort('status')">Status <span class="sort-icon">{{ sortBy === 'status' ? (sortDir === 'asc' ? '↑' : '↓') : '↕' }}</span></th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <template v-if="loading">
          <tr v-for="i in 5" :key="i" class="skeleton-row">
            <td><div class="skeleton-bar skeleton-bar--sm" /></td>
            <td><div class="skeleton-bar skeleton-bar--lg" /></td>
            <td><div class="skeleton-bar skeleton-bar--md" /></td>
            <td><div class="skeleton-bar skeleton-bar--sm" /></td>
            <td><div class="skeleton-bar skeleton-bar--pill" /></td>
            <td class="actions-cell"><div class="skeleton-bar skeleton-bar--btn" /></td>
          </tr>
        </template>
        <template v-else>
          <tr v-for="req in requests" :key="req.id">
            <td class="username-cell">{{ req.user?.username ?? req.userId }}</td>
            <td class="muted">{{ req.user?.email ?? '—' }}</td>
            <td class="reason-cell" :title="req.reason">{{ req.reason }}</td>
            <td class="muted">{{ formatDate(req.createdAt) }}</td>
            <td>
              <span class="status-badge" :class="req.status">{{ statusLabel(req.status) }}</span>
            </td>
            <td class="actions-cell">
              <template v-if="req.status === 'pending'">
                <button class="btn-approve" @click="openApprove(req)">Approve</button>
                <button class="btn-reject"  @click="openReject(req)">Reject</button>
              </template>
              <span v-else-if="req.status === 'rejected' && req.reviewNote" class="review-note-tip" :title="req.reviewNote">
                Note
              </span>
              <span v-else class="muted">—</span>
            </td>
          </tr>
        </template>
      </tbody>
    </table>
    </div>

    <!-- Approve modal -->
    <Teleport to="body">
      <div v-if="approveModal" class="modal-backdrop" @click.self="approveModal = null">
        <div class="modal">
          <h3>Approve Request</h3>
          <p class="modal-sub">User: <strong>{{ approveModal.user?.username }}</strong></p>

          <div class="form-group">
            <label>Grant role</label>
            <select v-model="approveRole" class="role-select full">
              <option value="user">User</option>
              <option v-if="auth.isAdmin" value="admin">Admin</option>
            </select>
          </div>

          <div class="form-group">
            <label>Expires on <span class="optional">(optional)</span></label>
            <input type="date" v-model="approveExpiry" class="date-input" />
          </div>

          <div class="modal-actions">
            <button class="btn-cancel" @click="approveModal = null">Cancel</button>
            <button class="btn-approve" :disabled="saving" @click="approve">
              {{ saving ? 'Approving…' : 'Approve' }}
            </button>
          </div>
        </div>
      </div>

      <!-- Reject modal -->
      <div v-if="rejectModal" class="modal-backdrop" @click.self="rejectModal = null">
        <div class="modal">
          <h3>Reject Request</h3>
          <p class="modal-sub">User: <strong>{{ rejectModal.user?.username }}</strong></p>

          <div class="form-group">
            <label>Reason for rejection</label>
            <textarea v-model="rejectNote" rows="4" placeholder="Explain why this request is being rejected…" class="note-textarea" />
          </div>

          <div class="modal-actions">
            <button class="btn-cancel" @click="rejectModal = null">Cancel</button>
            <button class="btn-reject" :disabled="saving || !rejectNote.trim()" @click="reject">
              {{ saving ? 'Rejecting…' : 'Reject' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.role-select.full { width: 100%; padding: 0.45rem 0.6rem; font-size: 0.9rem; }
.username-cell { font-weight: 600; color: #1e293b; }
.reason-cell { max-width: 280px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: #475569; }
.actions-cell { white-space: nowrap; }
.status-badge { display: block; padding: 0.15rem 0.6rem; border-radius: 999px; font-size: 0.75rem; font-weight: 600; text-transform: capitalize; text-align: center; }
.status-badge.pending  { background: #fef3c7; color: #92400e; }
.status-badge.approved { background: #d1fae5; color: #065f46; }
.status-badge.rejected { background: #fee2e2; color: #991b1b; }
.btn-approve { padding: 0.2rem 0.65rem; border: none; border-radius: 6px; background: #d1fae5; color: #065f46; font-size: 0.8rem; font-weight: 600; cursor: pointer; margin-right: 0.4rem; }
.btn-approve:hover:not(:disabled) { background: #6ee7b7; }
.btn-approve:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-reject { padding: 0.2rem 0.65rem; border: none; border-radius: 6px; background: #fee2e2; color: #991b1b; font-size: 0.8rem; font-weight: 600; cursor: pointer; }
.btn-reject:hover:not(:disabled) { background: #fca5a5; }
.btn-reject:disabled { opacity: 0.5; cursor: not-allowed; }
.review-note-tip { font-size: 0.75rem; color: #64748b; cursor: help; text-decoration: underline dotted; }
.note-textarea { padding: 0.55rem 0.75rem; border: 1px solid #d1d5db; border-radius: 8px; font-size: 0.9rem; font-family: inherit; resize: vertical; outline: none; }
.note-textarea:focus { border-color: #3b82f6; box-shadow: 0 0 0 2px #bfdbfe; }
</style>
