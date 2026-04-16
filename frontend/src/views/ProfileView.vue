<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getMe } from '@/services/api'

interface UserProfile {
  id: string
  username: string
  email: string
  firstName: string
  lastName: string
  roles: string[]
}

const profile = ref<UserProfile | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)

onMounted(async () => {
  try {
    const res = await getMe()
    profile.value = res.data
  } catch (e) {
    error.value = 'Failed to load profile'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="profile">
    <h1>Profile</h1>

    <div v-if="loading" class="state-message">Loading...</div>
    <div v-else-if="error" class="state-message error">{{ error }}</div>
    <div v-else-if="profile" class="card">
      <div class="profile-row">
        <span class="label">Username</span>
        <span>{{ profile.username }}</span>
      </div>
      <div class="profile-row">
        <span class="label">Email</span>
        <span>{{ profile.email }}</span>
      </div>
      <div class="profile-row">
        <span class="label">Name</span>
        <span>{{ profile.firstName }} {{ profile.lastName }}</span>
      </div>
      <div class="profile-row">
        <span class="label">Subject</span>
        <span class="mono">{{ profile.id }}</span>
      </div>
      <div class="profile-row" v-if="profile.roles.length">
        <span class="label">Roles</span>
        <span>
          <span class="role-badge" v-for="role in profile.roles" :key="role">{{ role }}</span>
        </span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.profile {
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
  gap: 0.75rem;
  max-width: 560px;
}

.profile-row {
  display: flex;
  gap: 1rem;
  font-size: 0.95rem;
}

.label {
  width: 100px;
  font-weight: 600;
  color: #64748b;
  flex-shrink: 0;
}

.mono {
  font-family: monospace;
  font-size: 0.85rem;
  color: #475569;
  word-break: break-all;
}

.role-badge {
  display: inline-block;
  background: #dbeafe;
  color: #1d4ed8;
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.1rem 0.5rem;
  border-radius: 4px;
  margin-right: 0.25rem;
}

.state-message {
  color: #64748b;
}

.state-message.error {
  color: #ef4444;
}
</style>
