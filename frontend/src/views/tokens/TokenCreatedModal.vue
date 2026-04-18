<script setup lang="ts">
import { ref } from 'vue'
import type { CreateTokenResponse } from '@/services/api'

const props = defineProps<{ result: CreateTokenResponse }>()
const emit = defineEmits<{ close: [] }>()

const copied = ref(false)

async function copy() {
  await navigator.clipboard.writeText(props.result.rawToken)
  copied.value = true
  setTimeout(() => { copied.value = false }, 2000)
}
</script>

<template>
  <div class="modal-overlay">
    <div class="modal">
      <div class="modal-icon">
        <svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4"/>
        </svg>
      </div>
      <h2>Token created</h2>
      <p class="warning-text">
        Copy your token now — it will <strong>not</strong> be shown again.
      </p>
      <div class="token-box">
        <code class="token-value">{{ result.rawToken }}</code>
        <button class="btn btn-sm" :class="{ 'btn-copied': copied }" @click="copy">
          {{ copied ? 'Copied!' : 'Copy' }}
        </button>
      </div>
      <div class="modal-actions">
        <button class="btn btn-primary" @click="emit('close')">I have saved my token</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.45);
  display: flex; align-items: center; justify-content: center; z-index: 100;
}
.modal {
  background: white; border-radius: 14px; padding: 2rem 2rem 1.75rem;
  width: 100%; max-width: 480px; display: flex; flex-direction: column;
  align-items: center; gap: 1rem; text-align: center;
  box-shadow: 0 20px 60px rgba(0,0,0,0.18);
}
.modal-icon {
  width: 56px; height: 56px; border-radius: 50%;
  background: #fef9c3; display: flex; align-items: center; justify-content: center;
  color: #ca8a04;
}
h2 { font-size: 1.25rem; font-weight: 700; margin: 0; }
.warning-text { font-size: 0.9rem; color: #64748b; margin: 0; }
.token-box {
  width: 100%; display: flex; align-items: center; gap: 0.6rem;
  background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 8px;
  padding: 0.6rem 0.75rem; text-align: left;
}
.token-value {
  flex: 1; font-family: monospace; font-size: 0.78rem;
  word-break: break-all; color: #1e293b;
}
.modal-actions { width: 100%; display: flex; justify-content: center; padding-top: 0.25rem; }
.btn-copied { background: #16a34a !important; color: white !important; border-color: #16a34a !important; }
</style>
