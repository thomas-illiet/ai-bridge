<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{
  tokenName: string
  action?: 'revoke' | 'unrevoke'
  onConfirm: () => Promise<void>
}>()

const emit = defineEmits<{ done: []; close: [] }>()

const loading = ref(false)
const error   = ref<string | null>(null)

const isRevoke = () => (props.action ?? 'revoke') === 'revoke'

async function confirm() {
  loading.value = true
  error.value = null
  try {
    await props.onConfirm()
    emit('done')
  } catch (e: unknown) {
    error.value = (e as { response?: { data?: { error?: string } } })?.response?.data?.error
      ?? (isRevoke() ? 'Failed to revoke token' : 'Failed to restore token')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <Teleport to="body">
    <div class="modal-backdrop" @click.self="emit('close')">
      <div class="modal">
        <div class="modal-header">
          <h2>{{ isRevoke() ? 'Revoke token' : 'Unrevoke token' }}</h2>
          <button class="modal-close" @click="emit('close')">✕</button>
        </div>
        <p class="confirm-text">
          <template v-if="isRevoke()">
            Are you sure you want to revoke <strong>{{ tokenName }}</strong>?
            Any application using this token will immediately lose access. This action cannot be undone.
          </template>
          <template v-else>
            Are you sure you want to restore <strong>{{ tokenName }}</strong>? The token will become active again.
          </template>
        </p>
        <p v-if="error" class="error-msg">{{ error }}</p>
        <div class="modal-actions">
          <button class="btn btn-outline" @click="emit('close')">
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            Cancel
          </button>
          <button
            class="btn"
            :class="isRevoke() ? 'btn-danger-solid' : 'btn-primary'"
            :disabled="loading"
            @click="confirm"
          >
            <svg v-if="isRevoke()" xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="4.93" y1="4.93" x2="19.07" y2="19.07"/></svg>
            <svg v-else xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 .49-3.51"/></svg>
            {{ loading ? (isRevoke() ? 'Revoking…' : 'Restoring…') : (isRevoke() ? 'Revoke token' : 'Unrevoke token') }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.modal-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 1rem; }
.modal-close { background: none; border: none; font-size: 1.1rem; cursor: pointer; color: #64748b; padding: 0.2rem 0.4rem; border-radius: 5px; }
.modal-close:hover { background: #f1f5f9; }
.confirm-text { font-size: 0.92rem; color: #475569; line-height: 1.6; margin: 0; }
.confirm-text strong { color: #1e293b; }
</style>
