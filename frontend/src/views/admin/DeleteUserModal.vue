<script setup lang="ts">
import { ref } from 'vue'
import { deleteUser } from '@/services/api'

interface RegisteredUser { id: string; username: string }

const props = defineProps<{ user: RegisteredUser | null }>()
const emit  = defineEmits<{ deleted: [id: string]; close: [] }>()

const deleting = ref(false)
const error    = ref<string | null>(null)

async function confirm() {
  if (!props.user) return
  deleting.value = true
  error.value = null
  try {
    await deleteUser(props.user.id)
    emit('deleted', props.user.id)
  } catch (e: any) {
    error.value = e?.response?.data?.error ?? 'Failed to delete user'
  } finally {
    deleting.value = false
  }
}

function cancel() {
  error.value = null
  emit('close')
}
</script>

<template>
  <Teleport to="body">
    <div v-if="user" class="modal-backdrop" @click.self="cancel">
      <div class="modal">
        <div class="modal-header">
          <h2>Delete user</h2>
          <button class="modal-close" @click="cancel">✕</button>
        </div>
        <p class="confirm-text">
          Are you sure you want to delete <strong>{{ user.username }}</strong>?
          This will also revoke all their tokens. This action cannot be undone.
        </p>
        <p v-if="error" class="error-msg">{{ error }}</p>
        <div class="modal-actions">
          <button class="btn btn-outline" @click="cancel">
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            Cancel
          </button>
          <button class="btn btn-danger-solid" :disabled="deleting" @click="confirm">
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"/></svg>
            {{ deleting ? 'Deleting…' : 'Delete' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.modal-backdrop { position: fixed; inset: 0; background: rgba(0,0,0,0.45); display: flex; align-items: center; justify-content: center; z-index: 200; padding: 1rem; }
.modal { background: white; border-radius: 14px; padding: 1.75rem; width: 100%; max-width: 420px; display: flex; flex-direction: column; gap: 1.25rem; box-shadow: 0 20px 60px rgba(0,0,0,0.25); }
.modal-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 1rem; }
.modal-header h2 { font-size: 1.15rem; font-weight: 700; margin: 0; }
.modal-close { background: none; border: none; font-size: 1.1rem; cursor: pointer; color: #64748b; padding: 0.2rem 0.4rem; border-radius: 5px; }
.modal-close:hover { background: #f1f5f9; }
.confirm-text { font-size: 0.92rem; color: #475569; line-height: 1.6; margin: 0; }
.confirm-text strong { color: #1e293b; }
.error-msg { color: #ef4444; font-size: 0.85rem; margin: 0; }
.modal-actions { display: flex; justify-content: flex-end; gap: 0.6rem; }
.btn { padding: 0.45rem 1rem; border-radius: 6px; border: none; cursor: pointer; font-size: 0.9rem; font-weight: 500; }
.btn-outline { background: transparent; color: #475569; border: 1px solid #cbd5e1; }
.btn-outline:hover { background: #f1f5f9; }
.btn-danger-solid { background: #dc2626; color: white; }
.btn-danger-solid:hover:not(:disabled) { background: #b91c1c; }
.btn-danger-solid:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
