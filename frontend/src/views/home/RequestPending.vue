<script setup lang="ts">
import type { AccessRequest } from '@/services/api'

defineProps<{ request: AccessRequest }>()

function relativeDate(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime()
  const mins = Math.floor(diff / 60_000)
  if (mins < 1)  return 'just now'
  if (mins < 60) return `${mins} minute${mins > 1 ? 's' : ''} ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24)  return `${hrs} hour${hrs > 1 ? 's' : ''} ago`
  const days = Math.floor(hrs / 24)
  return `${days} day${days > 1 ? 's' : ''} ago`
}
</script>

<template>
  <div class="state-card pending">
    <div class="spinner" />

    <div class="pending-text">
      <h2>Your request is being reviewed</h2>
      <p>
        Our team has received your request and is reviewing it.
        You'll be automatically redirected here as soon as a decision is made.
      </p>
      <p class="submitted-at">
        Submitted {{ relativeDate(request.createdAt) }}
      </p>
    </div>

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
  align-items: center;
  gap: 1.25rem;
  text-align: center;
}
.state-card h2 { font-size: 1.25rem; font-weight: 700; margin: 0; color: #1e293b; }
.state-card p  { color: #475569; font-size: 0.95rem; margin: 0; line-height: 1.6; }
.state-card.pending { border-color: #bfdbfe; background: #eff6ff; }

@keyframes spin { to { transform: rotate(360deg); } }
.spinner {
  width: 40px; height: 40px;
  border: 3px solid #bfdbfe;
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 0.9s linear infinite;
  flex-shrink: 0;
}

.pending-text { display: flex; flex-direction: column; gap: 0.5rem; }
.submitted-at { font-size: 0.8rem; color: #94a3b8; margin-top: 0.25rem; }
</style>
