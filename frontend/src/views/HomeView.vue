<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { getMyAccessRequest, createAccessRequest } from '@/services/api'
import type { AccessRequest } from '@/services/api'

const auth    = useAuthStore()
const router  = useRouter()
const request = ref<AccessRequest | null>(null)
const reason  = ref('')
const loading = ref(false)
const submitting = ref(false)
const error   = ref('')

let pollTimer: ReturnType<typeof setInterval> | null = null

async function fetchRequest() {
  try {
    const res = await getMyAccessRequest()
    request.value = res.data

    // If approved, refresh the role then redirect to dashboard
    if (res.data?.status === 'approved') {
      stopPolling()
      await auth.fetchRole()
      router.push('/dashboard')
    }
  } catch {
    request.value = null
  }
}

function startPolling() {
  pollTimer = setInterval(fetchRequest, 30_000)
}

function stopPolling() {
  if (pollTimer !== null) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

onMounted(async () => {
  if (auth.authenticated && auth.dbRole === 'none') {
    loading.value = true
    await fetchRequest()
    loading.value = false
    if (request.value?.status === 'pending') {
      startPolling()
    }
  }
})

onUnmounted(stopPolling)

// Start polling when request becomes pending (e.g. after first submit)
watch(request, (val) => {
  if (val?.status === 'pending' && !pollTimer) {
    startPolling()
  } else if (val?.status !== 'pending') {
    stopPolling()
  }
})

async function submit() {
  if (!reason.value.trim()) return
  submitting.value = true
  error.value = ''
  try {
    const res = await createAccessRequest(reason.value.trim())
    request.value = res.data
    reason.value = ''
  } catch (e: any) {
    error.value = e?.response?.data?.error ?? 'An error occurred. Please try again.'
  } finally {
    submitting.value = false
  }
}

function relativeDate(iso: string): string {
  const d = new Date(iso)
  const diff = Date.now() - d.getTime()
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
  <div class="home">
    <section class="hero">
      <h1>AI Bridge</h1>
      <p>A secure platform powered by Keycloak SSO</p>

      <div class="hero-actions">
        <RouterLink v-if="auth.authenticated && auth.dbRole !== 'none'" to="/dashboard" class="btn btn-primary">
          Go to Dashboard
        </RouterLink>
        <button v-else-if="!auth.authenticated" class="btn btn-primary" @click="auth.login()">
          Sign In with Keycloak
        </button>
      </div>
    </section>

    <!-- Access request section for role=none users -->
    <section v-if="auth.authenticated && auth.dbRole === 'none'" class="request-section">

      <div v-if="loading" class="state-card">
        <div class="spinner" />
        <p class="muted">Loading your request status…</p>
      </div>

      <!-- Pending -->
      <div v-else-if="request?.status === 'pending'" class="state-card pending">
        <div class="spinner spinner--amber" />

        <div class="pending-text">
          <h2>Your request is being reviewed</h2>
          <p>
            Our team has received your request and is reviewing it.
            You'll be automatically redirected here as soon as a decision is made —
            no need to refresh the page.
          </p>
          <p class="submitted-at">
            Submitted {{ relativeDate(request.createdAt) }}
            <span class="dot">·</span>
            Checking for updates every 30 s
          </p>
        </div>

        <div class="reason-preview">
          <span class="reason-label">Your reason</span>
          <p>{{ request.reason }}</p>
        </div>
      </div>

      <!-- Rejected -->
      <div v-else-if="request?.status === 'rejected'" class="state-card rejected">
        <div class="reject-icon">✕</div>
        <div class="pending-text">
          <h2>Request not approved</h2>
          <p>Your request was reviewed but could not be approved at this time.</p>
          <p class="submitted-at">Reviewed {{ relativeDate(request.reviewedAt ?? request.createdAt) }}</p>
        </div>
        <div v-if="request.reviewNote" class="review-note">
          <span class="reason-label">Reviewer's note</span>
          <p>{{ request.reviewNote }}</p>
        </div>
        <p class="muted small">You may submit a new request below.</p>
        <form class="request-form" @submit.prevent="submit">
          <label for="reason">Explain your need for access</label>
          <textarea
            id="reason"
            v-model="reason"
            rows="4"
            placeholder="Explain your use case and what you plan to do with AI Bridge…"
            :disabled="submitting"
          />
          <p v-if="error" class="form-error">{{ error }}</p>
          <button type="submit" class="btn btn-primary" :disabled="submitting || !reason.trim()">
            {{ submitting ? 'Submitting…' : 'Re-submit Request' }}
          </button>
        </form>
      </div>

      <!-- No request yet -->
      <div v-else class="state-card form-card">
        <h2>Request Access</h2>
        <p class="muted">Your account is pending activation. Tell us why you need access to AI Bridge and we'll review your request.</p>
        <form class="request-form" @submit.prevent="submit">
          <label for="reason">Why do you need access?</label>
          <textarea
            id="reason"
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

    </section>
  </div>
</template>

<style scoped>
.home {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2rem;
  padding-top: 4rem;
}

.hero { text-align: center; }
.hero h1 { font-size: 3rem; font-weight: 800; color: #1e293b; margin-bottom: 0.5rem; }
.hero p { font-size: 1.2rem; color: #64748b; margin-bottom: 2rem; }
.hero-actions { display: flex; justify-content: center; gap: 1rem; }

.btn {
  padding: 0.75rem 1.75rem;
  border-radius: 8px;
  border: none;
  cursor: pointer;
  font-size: 1rem;
  font-weight: 600;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  transition: background 0.15s;
}
.btn-primary { background: #3b82f6; color: white; }
.btn-primary:hover:not(:disabled) { background: #2563eb; }
.btn:disabled { opacity: 0.55; cursor: not-allowed; }

.request-section { width: 100%; max-width: 560px; }

/* Card base */
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

.state-card.pending  { border-color: #fcd34d; background: #fffbeb; }
.state-card.rejected { border-color: #fca5a5; background: #fff1f2; align-items: flex-start; text-align: left; }
.state-card.form-card { text-align: left; align-items: flex-start; }

/* Spinner */
@keyframes spin { to { transform: rotate(360deg); } }
.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #e2e8f0;
  border-top-color: #6366f1;
  border-radius: 50%;
  animation: spin 0.9s linear infinite;
  flex-shrink: 0;
}
.spinner--amber {
  border-color: #fde68a;
  border-top-color: #f59e0b;
}

/* Pending body */
.pending-text { display: flex; flex-direction: column; gap: 0.5rem; }
.submitted-at { font-size: 0.8rem; color: #94a3b8; margin-top: 0.25rem; }
.dot { margin: 0 0.4rem; }

/* Reason / review note box */
.reason-preview,
.review-note {
  width: 100%;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 0.9rem 1rem;
  text-align: left;
  box-sizing: border-box;
}
.state-card.pending .reason-preview { background: #fefce8; border-color: #fde68a; }
.review-note { background: #fef2f2; border-color: #fecaca; }
.reason-label {
  display: block;
  font-size: 0.7rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: #94a3b8;
  margin-bottom: 0.35rem;
}
.reason-preview p,
.review-note p { font-size: 0.9rem; color: #374151; margin: 0; line-height: 1.5; }

/* Reject icon */
.reject-icon {
  width: 40px; height: 40px;
  border-radius: 50%;
  background: #fee2e2;
  color: #dc2626;
  font-size: 1.1rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  align-self: center;
}

/* Form */
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
.muted { color: #64748b; }
.small { font-size: 0.8rem; }
</style>
