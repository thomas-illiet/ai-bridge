<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { createAccessRequest } from '@/services/api'
import { useAccessRequest } from '@/composables/useAccessRequest'
import HeroSection from '@/views/home/HeroSection.vue'
import RequestPending from '@/views/home/RequestPending.vue'
import RequestRejected from '@/views/home/RequestRejected.vue'
import RequestForm from '@/views/home/RequestForm.vue'

const auth = useAuthStore()
const { request, loading, fetchRequest } = useAccessRequest()

const submitting = ref(false)
const error = ref('')

onMounted(async () => {
  if (auth.authenticated && auth.dbRole === 'none') {
    loading.value = true
    await fetchRequest()
    loading.value = false
  }
})

async function submit(reason: string) {
  submitting.value = true
  error.value = ''
  try {
    const res = await createAccessRequest(reason)
    request.value = res.data
  } catch (e: any) {
    error.value = e?.response?.data?.error ?? 'An error occurred. Please try again.'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="home">
    <HeroSection />

    <section v-if="auth.authenticated && auth.dbRole === 'none'" class="request-section">
      <div v-if="loading" class="state-card loading-card">
        <div class="spinner" />
        <p class="muted">Loading your request status…</p>
      </div>

      <RequestPending
        v-else-if="request?.status === 'pending'"
        :request="request"
      />
      <RequestRejected
        v-else-if="request?.status === 'rejected'"
        :request="request"
        :submitting="submitting"
        :error="error"
        @submit="submit"
      />
      <RequestForm
        v-else
        :submitting="submitting"
        :error="error"
        @submit="submit"
      />
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

.request-section { width: 100%; max-width: 560px; }

.loading-card {
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
.muted { color: #64748b; font-size: 0.95rem; margin: 0; }

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
</style>
