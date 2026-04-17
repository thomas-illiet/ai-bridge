<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { adminListAccessRequests, adminApproveRequest, adminRejectRequest } from '@/services/api'
import type { AccessRequest } from '@/services/api'

const requests    = ref<AccessRequest[]>([])
const pendingCount = ref(0)
const loading     = ref(false)
const statusFilter = ref('')

const approveModal = ref<AccessRequest | null>(null)
const rejectModal  = ref<AccessRequest | null>(null)
const approveRole  = ref('user')
const approveExpiry = ref('')
const rejectNote   = ref('')
const saving       = ref(false)

async function load() {
  loading.value = true
  try {
    const res = await adminListAccessRequests(statusFilter.value || undefined)
    requests.value   = res.data.requests ?? []
    pendingCount.value = res.data.pendingCount
  } finally { loading.value = false }
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
    <div class="toolbar">
      <p class="sub">{{ pendingCount }} pending request{{ pendingCount !== 1 ? 's' : '' }}.</p>
      <div class="filters">
        <select v-model="statusFilter" class="role-select" @change="load">
          <option value="">All statuses</option>
          <option value="pending">Pending</option>
          <option value="approved">Approved</option>
          <option value="rejected">Rejected</option>
        </select>
      </div>
    </div>

    <div v-if="loading && requests.length === 0" class="empty-card">
      <div class="empty-icon">
        <svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/>
        </svg>
      </div>
      <p class="empty-title">Loading requests…</p>
    </div>
    <div v-else-if="requests.length === 0" class="empty-card">
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

    <table v-else class="data-table" :class="{ 'table-loading': loading }">
      <thead>
        <tr>
          <th>User</th>
          <th>Email</th>
          <th>Reason</th>
          <th>Submitted</th>
          <th>Status</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
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
      </tbody>
    </table>

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
              <option value="admin">Admin</option>
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
.tab-content { display: flex; flex-direction: column; gap: 1.25rem; }
.toolbar { display: flex; align-items: center; justify-content: space-between; gap: 1rem; flex-wrap: wrap; }
.sub { font-size: 0.85rem; color: #64748b; margin: 0; }
.filters { display: flex; gap: 0.5rem; align-items: center; }
.role-select { padding: 0.3rem 0.5rem; border: 1px solid #e2e8f0; border-radius: 6px; font-size: 0.85rem; background: white; cursor: pointer; }
.role-select.full { width: 100%; padding: 0.45rem 0.6rem; font-size: 0.9rem; }
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
.data-table th { text-align: left; padding: 0.55rem 0.9rem; font-size: 0.75rem; font-weight: 600; color: #64748b; text-transform: uppercase; letter-spacing: 0.03em; border-bottom: 1px solid #e2e8f0; background: #f8fafc; white-space: nowrap; }
.data-table thead tr:first-child th:first-child { border-radius: 10px 0 0 0; }
.data-table thead tr:first-child th:last-child  { border-radius: 0 10px 0 0; }
.data-table tbody tr:last-child td:first-child   { border-radius: 0 0 0 10px; }
.data-table tbody tr:last-child td:last-child    { border-radius: 0 0 10px 0; }
.data-table td { padding: 0.65rem 0.9rem; border-bottom: 1px solid #f1f5f9; vertical-align: top; }
.data-table tr:last-child td { border-bottom: none; }
.data-table.table-loading { opacity: 0.6; pointer-events: none; }

.username-cell { font-weight: 600; color: #1e293b; }
.reason-cell { max-width: 280px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: #475569; }
.muted { color: #64748b; font-size: 0.85rem; }

.status-badge { display: inline-block; padding: 0.15rem 0.6rem; border-radius: 999px; font-size: 0.75rem; font-weight: 600; text-transform: capitalize; }
.status-badge.pending  { background: #fef3c7; color: #92400e; }
.status-badge.approved { background: #d1fae5; color: #065f46; }
.status-badge.rejected { background: #fee2e2; color: #991b1b; }

.actions-cell { white-space: nowrap; }
.btn-approve { padding: 0.2rem 0.65rem; border: none; border-radius: 6px; background: #d1fae5; color: #065f46; font-size: 0.8rem; font-weight: 600; cursor: pointer; margin-right: 0.4rem; }
.btn-approve:hover:not(:disabled) { background: #6ee7b7; }
.btn-approve:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-reject  { padding: 0.2rem 0.65rem; border: none; border-radius: 6px; background: #fee2e2; color: #991b1b; font-size: 0.8rem; font-weight: 600; cursor: pointer; }
.btn-reject:hover:not(:disabled) { background: #fca5a5; }
.btn-reject:disabled { opacity: 0.5; cursor: not-allowed; }
.review-note-tip { font-size: 0.75rem; color: #64748b; cursor: help; text-decoration: underline dotted; }

/* Modals */
.modal-backdrop { position: fixed; inset: 0; background: rgba(0,0,0,0.45); display: flex; align-items: center; justify-content: center; z-index: 200; }
.modal { background: white; border-radius: 12px; padding: 1.75rem; width: 100%; max-width: 420px; display: flex; flex-direction: column; gap: 1rem; }
.modal h3 { font-size: 1.1rem; font-weight: 700; margin: 0; }
.modal-sub { font-size: 0.9rem; color: #475569; margin: 0; }
.form-group { display: flex; flex-direction: column; gap: 0.35rem; }
.form-group label { font-size: 0.85rem; font-weight: 600; color: #374151; }
.optional { font-weight: 400; color: #94a3b8; }
.date-input { padding: 0.45rem 0.6rem; border: 1px solid #d1d5db; border-radius: 6px; font-size: 0.9rem; outline: none; }
.date-input:focus { border-color: #3b82f6; box-shadow: 0 0 0 2px #bfdbfe; }
.note-textarea { padding: 0.55rem 0.75rem; border: 1px solid #d1d5db; border-radius: 8px; font-size: 0.9rem; font-family: inherit; resize: vertical; outline: none; }
.note-textarea:focus { border-color: #3b82f6; box-shadow: 0 0 0 2px #bfdbfe; }
.modal-actions { display: flex; justify-content: flex-end; gap: 0.6rem; }
.btn-cancel { padding: 0.4rem 1rem; border: 1px solid #e2e8f0; border-radius: 6px; background: white; color: #374151; font-size: 0.88rem; cursor: pointer; }
.btn-cancel:hover { background: #f8fafc; }
</style>
