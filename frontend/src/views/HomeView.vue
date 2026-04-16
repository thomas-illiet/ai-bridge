<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { RouterLink } from 'vue-router'

const auth = useAuthStore()
</script>

<template>
  <div class="home">
    <section class="hero">
      <h1>AI Bridge</h1>
      <p>A secure platform powered by Keycloak SSO</p>

      <div class="hero-actions">
        <RouterLink v-if="auth.authenticated" to="/dashboard" class="btn btn-primary">
          Go to Dashboard
        </RouterLink>
        <button v-else class="btn btn-primary" @click="auth.login()">
          Sign In with Keycloak
        </button>
      </div>
    </section>

    <section v-if="auth.authenticated" class="welcome-card">
      <h2>Welcome back, {{ auth.fullName || auth.username }}</h2>
      <p>{{ auth.email }}</p>
      <p v-if="auth.roles.length">
        Roles: <span class="role-badge" v-for="role in auth.roles" :key="role">{{ role }}</span>
      </p>
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

.hero {
  text-align: center;
}

.hero h1 {
  font-size: 3rem;
  font-weight: 800;
  color: #1e293b;
  margin-bottom: 0.5rem;
}

.hero p {
  font-size: 1.2rem;
  color: #64748b;
  margin-bottom: 2rem;
}

.hero-actions {
  display: flex;
  justify-content: center;
  gap: 1rem;
}

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

.btn-primary {
  background: #3b82f6;
  color: white;
}

.btn-primary:hover {
  background: #2563eb;
}

.welcome-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1.5rem 2rem;
  width: 100%;
  max-width: 480px;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.welcome-card h2 {
  font-size: 1.2rem;
  font-weight: 600;
}

.welcome-card p {
  color: #64748b;
  font-size: 0.9rem;
}

.role-badge {
  display: inline-block;
  background: #dbeafe;
  color: #1d4ed8;
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.1rem 0.5rem;
  border-radius: 4px;
  margin-left: 0.25rem;
}
</style>
