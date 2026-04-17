<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { getMyAccessRequest, createAccessRequest } from '@/services/api'
import type { AccessRequest } from '@/services/api'

const auth    = useAuthStore()
const request = ref<AccessRequest | null>(null)
const reason  = ref('')
const loading = ref(false)
const submitting = ref(false)
const error   = ref('')

onMounted(async () => {
  if (auth.authenticated && auth.dbRole === 'none') {
    loading.value = true
    try {
      const res = await getMyAccessRequest()
      request.value = res.data
    } catch {
      request.value = null
    } finally {
      loading.value = false
    }
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
        <p class="muted">Loading…</p>
      </div>

      <!-- Pending -->
      <div v-else-if="request?.status === 'pending'" class="state-card pending">
        <div class="state-icon">⏳</div>
        <h2>Request Pending</h2>
        <p>Your access request is under review. Our team will get back to you shortly.</p>
        <p class="muted small">Submitted on {{ new Date(request.createdAt).toLocaleDateString() }}</p>
      </div>

      <!-- Rejected -->
      <div v-else-if="request?.status === 'rejected'" class="state-card rejected">
        <div class="state-icon">✗</div>
        <h2>Request Rejected</h2>
        <p v-if="request.reviewNote" class="review-note">
          <strong>Reason:</strong> {{ request.reviewNote }}
        </p>
        <p class="muted small">You may submit a new request below.</p>
        <form class="request-form" @submit.prevent="submit">
          <label for="reason">Describe why you need access</label>
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

.request-section {
  width: 100%;
  max-width: 560px;
}

.state-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  padding: 2.5rem 2rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  text-align: center;
}
.state-card h2 { font-size: 1.3rem; font-weight: 700; margin: 0; }
.state-card p { color: #475569; font-size: 0.95rem; margin: 0; }

.state-card.pending { border-color: #fbbf24; background: #fffbeb; }
.state-card.rejected { border-color: #fca5a5; background: #fff1f2; }
.state-card.form-card { text-align: left; align-items: flex-start; }

.state-icon { font-size: 2.5rem; line-height: 1; }

.review-note {
  background: #fee2e2;
  border-left: 3px solid #f87171;
  padding: 0.6rem 0.9rem;
  border-radius: 0 6px 6px 0;
  font-size: 0.9rem;
  text-align: left;
  width: 100%;
  box-sizing: border-box;
}

.request-form {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  width: 100%;
  margin-top: 0.5rem;
}
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
.request-form .btn { align-self: flex-start; }

.form-error { color: #dc2626; font-size: 0.85rem; margin: 0; }
.muted { color: #64748b; }
.small { font-size: 0.8rem; }
</style>
