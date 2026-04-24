<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
</script>

<template>
  <section class="hero">
    <h1>AI Br<span class="fake-i"><span class="dot"></span></span>dge</h1>
    <p>Designed for intelligent interactions</p>

    <div class="hero-actions">
      <RouterLink v-if="auth.authenticated && auth.dbRole !== 'none'" to="/dashboard" class="btn btn-primary">
        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/></svg>
        Go to Dashboard
      </RouterLink>
      <button v-else-if="!auth.authenticated" class="btn btn-primary" @click="auth.login()">
        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"/><polyline points="10 17 15 12 10 7"/><line x1="15" y1="12" x2="3" y2="12"/></svg>
        Sign in to continue
      </button>
    </div>
  </section>
</template>

<style scoped>
.hero { text-align: center; }
.hero h1 { font-size: 3rem; font-weight: 800; color: #1e293b; margin-bottom: 0.5rem; }
.hero p  { font-size: 1.2rem; color: #64748b; margin-bottom: 2rem; }
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
.btn-primary:hover { background: #2563eb; }


.fake-i {
  display: inline-block;
  width: 8px;
  height: 28px;
  background: black;
  position: relative;
  margin: 0 2px;
}

.dot {
  position: absolute;
  top: -12px;
  left: 50%;
  width: 10px;
  height: 10px;
  background: black;
  border-radius: 50%;
  transform: translateX(-50%);
  animation: bounce 1s infinite;
}

@keyframes bounce {
  0%, 100% { transform: translate(-50%, 0); }
  50% { transform: translate(-50%, -10px); }
}
</style>
