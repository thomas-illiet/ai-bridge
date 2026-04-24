<script setup lang="ts">
import { ref } from 'vue'
import type { AccessRequest } from '@/services/api'

defineProps<{ request: AccessRequest; submitting: boolean; error: string }>()
const emit = defineEmits<{ submit: [reason: string] }>()

const reason = ref('')

function handleSubmit() {
  if (!reason.value.trim()) return
  emit('submit', reason.value.trim())
  reason.value = ''
}
</script>

<template>
  <div class="state-card rejected">

    <div class="rejected-text">
      <h2>Request not approved</h2>
      <p>Your request was reviewed but could not be approved at this time.</p>
    </div>

    <div v-if="request.reviewNote" class="review-note">
      <span class="reason-label">Reviewer's note</span>
      <p>{{ request.reviewNote }}</p>
    </div>

    <p class="muted small">You may submit a new request below.</p>

    <form class="request-form" @submit.prevent="handleSubmit">
      <label for="reason-retry">Explain your need for access</label>
      <textarea
        id="reason-retry"
        v-model="reason"
        rows="4"
        placeholder="Explain your use case and what you plan to do with AI Bridge…"
        :disabled="submitting"
      />
      <p v-if="error" class="form-error">{{ error }}</p>
      <button type="submit" class="btn btn-primary" :disabled="submitting || !reason.trim()">
        <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg>
        {{ submitting ? 'Submitting…' : 'Re-submit Request' }}
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
  align-items: flex-start;
  gap: 1.25rem;
}
.state-card h2 { font-size: 1.25rem; font-weight: 700; margin: 0; color: #1e293b; }
.state-card p  { color: #475569; font-size: 0.95rem; margin: 0; line-height: 1.6; }
.state-card.rejected { border-color: #fca5a5; background: #fff1f2; }

.rejected-text { display: flex; flex-direction: column; gap: 0.4rem; width: 100%; }
.submitted-at { font-size: 0.8rem; color: #94a3b8; margin-top: 0.15rem; }

.review-note {
  width: 100%;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 10px;
  padding: 0.9rem 1rem;
  box-sizing: border-box;
}
.reason-label {
  display: block;
  font-size: 0.7rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: #94a3b8;
  margin-bottom: 0.35rem;
}
.review-note p { font-size: 0.9rem; color: #374151; margin: 0; line-height: 1.5; }

.muted { color: #64748b; }
.small { font-size: 0.85rem; }

.request-form { display: flex; flex-direction: column; gap: 0.75rem; width: 100%; }
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
