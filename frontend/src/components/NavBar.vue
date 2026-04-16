<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { RouterLink } from 'vue-router'

const auth = useAuthStore()
</script>

<template>
  <nav class="navbar">
    <div class="navbar-brand">
      <RouterLink to="/">AI Bridge</RouterLink>
    </div>

    <div class="navbar-links">
      <RouterLink to="/">Home</RouterLink>
      <RouterLink v-if="auth.authenticated" to="/dashboard">Dashboard</RouterLink>
      <RouterLink v-if="auth.authenticated" to="/profile">Profile</RouterLink>
      <RouterLink v-if="auth.authenticated" to="/tokens">Tokens</RouterLink>
    </div>

    <div class="navbar-auth">
      <template v-if="auth.authenticated">
        <span class="navbar-user">{{ auth.username }}</span>
        <button class="btn btn-outline" @click="auth.logout()">Logout</button>
      </template>
      <template v-else>
        <button class="btn btn-primary" @click="auth.login()">Login</button>
      </template>
    </div>
  </nav>
</template>

<style scoped>
.navbar {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  padding: 0 2rem;
  height: 60px;
  background: #1e293b;
  color: #f1f5f9;
}

.navbar-brand a {
  font-size: 1.25rem;
  font-weight: 700;
  color: #f1f5f9;
  text-decoration: none;
}

.navbar-links {
  display: flex;
  gap: 1rem;
  flex: 1;
}

.navbar-links a {
  color: #94a3b8;
  text-decoration: none;
  font-size: 0.95rem;
  transition: color 0.15s;
}

.navbar-links a:hover,
.navbar-links a.router-link-active {
  color: #f1f5f9;
}

.navbar-auth {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.navbar-user {
  font-size: 0.9rem;
  color: #94a3b8;
}

.btn {
  padding: 0.4rem 1rem;
  border-radius: 6px;
  border: none;
  cursor: pointer;
  font-size: 0.9rem;
  font-weight: 500;
  transition: background 0.15s;
}

.btn-primary {
  background: #3b82f6;
  color: white;
}

.btn-primary:hover {
  background: #2563eb;
}

.btn-outline {
  background: transparent;
  color: #94a3b8;
  border: 1px solid #475569;
}

.btn-outline:hover {
  background: #334155;
  color: #f1f5f9;
}
</style>
