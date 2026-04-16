<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getDashboard } from '@/services/api'

interface DashboardData {
  message: string
  user: string
}

const data = ref<DashboardData | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)

onMounted(async () => {
  try {
    const res = await getDashboard()
    data.value = res.data
  } catch (e) {
    error.value = 'Failed to load dashboard data'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="dashboard">
    <h1>Dashboard</h1>

    <div v-if="loading" class="state-message">Loading...</div>
    <div v-else-if="error" class="state-message error">{{ error }}</div>
    <div v-else-if="data" class="card">
      <p>{{ data.message }}</p>
      <p class="muted">Logged in as: <strong>{{ data.user }}</strong></p>
    </div>
  </div>
</template>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

h1 {
  font-size: 1.75rem;
  font-weight: 700;
}

.card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.muted {
  color: #64748b;
  font-size: 0.9rem;
}

.state-message {
  color: #64748b;
  font-size: 1rem;
}

.state-message.error {
  color: #ef4444;
}
</style>
