<script setup lang="ts">
import { ref } from 'vue'

defineProps<{ submitting: boolean; error: string }>()
const emit = defineEmits<{ submit: [reason: string] }>()

const reason = ref('')

function handleSubmit() {
  if (!reason.value.trim()) return
  emit('submit', reason.value.trim())
  reason.value = ''
}
</script>

<template>
  <div class="state-card">
    <h2>Request Access</h2>
    <p class="muted">Your account is pending activation. Tell us why you need access to AI Bridge and we'll review your request.</p>

    <form class="request-form" @submit.prevent="handleSubmit">
      <label for="reason-new">Why do you need access?</label>
      <textarea
        id="reason-new"
        v-model="reason"
        rows="5"
        placeholder="Explain your use case and what you plan to do with AI Bridge…"
        :disabled="submitting"
      />
      <p v-if="error" class="form-error">{{ error }}</p>
      <button type="submit" class="btn btn-primary" :disabled="submitting || !reason.trim()">
        {{ submitting ? 'Submitting…' : 'Submit Request' }}
      </button>
    </form>
  </div>
</template>

<style scoped>
.state-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  padding: 2.5rem 2rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}
.state-card h2 { font-size: 1.25rem; font-weight: 700; margin: 0; color: #1e293b; }
.muted { color: #64748b; font-size: 0.95rem; margin: 0; }

.request-form { display: flex; flex-direction: column; gap: 0.75rem; }
.request-form label { font-size: 0.9rem; font-weight: 600; color: #374151; }
.request-form textarea {
  width: 100%;
  padding: 0.65rem 0.85rem;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 0.9rem;
  resize: vertical;
  font-family: inherit;
  outline: none;
  box-sizing: border-box;
}
.request-form textarea:focus { border-color: #3b82f6; box-shadow: 0 0 0 2px #bfdbfe; }
.form-error { color: #dc2626; font-size: 0.85rem; margin: 0; }

.btn {
  padding: 0.75rem 1.75rem;
  border-radius: 8px;
  border: none;
  cursor: pointer;
  font-size: 1rem;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  transition: background 0.15s;
}
.btn-primary { background: #3b82f6; color: white; }
.btn-primary:hover:not(:disabled) { background: #2563eb; }
.btn:disabled { opacity: 0.55; cursor: not-allowed; }
</style>
