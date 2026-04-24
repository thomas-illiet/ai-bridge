<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { CreateTokenResponse } from '@/services/api'

const props = defineProps<{ result: CreateTokenResponse }>()
const emit = defineEmits<{ close: [] }>()

const ready = ref(false)
const copied = ref(false)

onMounted(() => {
  setTimeout(() => { ready.value = true }, 900)
})

async function copy() {
  await navigator.clipboard.writeText(props.result.rawToken)
  copied.value = true
  setTimeout(() => { copied.value = false }, 2000)
}
</script>

<template>
  <div class="modal-overlay">
    <div class="modal">
      <Transition name="fade" mode="out-in">
        <div v-if="!ready" class="spinner-wrap" key="spinner">
          <div class="spinner" />
          <p class="spinner-label">Generating token…</p>
        </div>

        <div v-else class="modal-content" key="content">
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
            <button class="copy-btn" :class="{ 'copy-btn--copied': copied }" @click="copy">
              <svg v-if="!copied" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
                <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
              </svg>
              <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="20 6 9 17 4 12"/>
              </svg>
              <span>{{ copied ? 'Copied!' : 'Copy' }}</span>
            </button>
          </div>
          <div class="modal-actions">
            <button class="btn btn-primary" @click="emit('close')">
              <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
              I have saved my token
            </button>
          </div>
        </div>
      </Transition>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.45);
  display: flex; align-items: center; justify-content: center; z-index: 100;
}
.modal {
  background: white; border-radius: 14px; padding: 2.25rem 2.25rem 2rem;
  width: 100%; max-width: 620px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.18);
  min-height: 220px; display: flex; align-items: center; justify-content: center;
}

/* ── spinner state ── */
.spinner-wrap {
  display: flex; flex-direction: column; align-items: center; gap: 1rem;
  padding: 1rem 0;
}
.spinner {
  width: 42px; height: 42px; border-radius: 50%;
  border: 3px solid #e2e8f0; border-top-color: #3b82f6;
  animation: spin 0.75s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
.spinner-label { font-size: 0.9rem; color: #64748b; margin: 0; }

/* ── content state ── */
.modal-content {
  width: 100%; display: flex; flex-direction: column;
  align-items: center; gap: 1rem; text-align: center;
}
.modal-icon {
  width: 56px; height: 56px; border-radius: 50%;
  background: #fef9c3; display: flex; align-items: center; justify-content: center;
  color: #ca8a04;
}
h2 { font-size: 1.3rem; font-weight: 700; margin: 0; }
.warning-text { font-size: 0.9rem; color: #64748b; margin: 0; }

/* ── token box ── */
.token-box {
  width: 100%; display: flex; align-items: stretch;
  background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 8px;
  overflow: hidden; text-align: left;
}
.token-value {
  flex: 1; font-family: monospace; font-size: 0.78rem;
  word-break: break-all; color: #1e293b;
  padding: 0.65rem 0.85rem; line-height: 1.5;
}
.copy-btn {
  display: flex; align-items: center; justify-content: center; gap: 0.4rem;
  padding: 0 1.1rem;
  border: none; border-left: 1px solid #e2e8f0;
  background: #f1f5f9; color: #475569;
  font-size: 0.82rem; font-weight: 600; cursor: pointer;
  transition: background 0.15s, color 0.15s;
  white-space: nowrap;
}
.copy-btn:hover { background: #e2e8f0; }
.copy-btn--copied { background: #16a34a !important; color: white !important; border-color: #16a34a !important; }

/* ── actions ── */
.modal-actions { width: 100%; display: flex; justify-content: center; padding-top: 0.25rem; }

/* ── fade transition ── */
.fade-enter-active, .fade-leave-active { transition: opacity 0.25s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
