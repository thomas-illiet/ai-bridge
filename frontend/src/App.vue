<script setup lang="ts">
import NavBar from '@/components/NavBar.vue'
import ChatWidget from '@/views/chat/ChatWidget.vue'
import { useAuthStore } from '@/stores/auth'
import { useConnectivityStore } from '@/stores/connectivity'

const auth = useAuthStore()
const connectivity = useConnectivityStore()
function reload() { location.reload() }
</script>

<template>
  <div class="app">
    <template v-if="connectivity.backendDown">
      <div class="error-page">
        <div class="error-card">
          <div class="error-icon">
            <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M9.172 16.172a4 4 0 0 1 5.656 0"/><path d="M2 8.82a15 15 0 0 1 20 0"/><path d="M5 12.859a10 10 0 0 1 14 0"/><line x1="12" y1="20" x2="12.01" y2="20"/>
              <line x1="2" y1="2" x2="22" y2="22"/>
            </svg>
          </div>
          <h1 class="error-title">Backend unreachable</h1>
          <p class="error-sub">AI Bridge cannot connect to the server. Check that the backend is running and try again.</p>
          <button class="btn-retry" @click="reload">Retry</button>
        </div>
      </div>
    </template>
    <template v-else>
      <NavBar />
      <main class="main-content">
        <RouterView />
      </main>
      <footer class="footer">
        <span class="footer-slogan">
          One bridge. Every model. Infinite possibilities —
          <span class="footer-highlight">AI Bridge</span>
        </span>
        <span class="footer-sep">·</span>
        <span class="footer-copy">{{ new Date().getFullYear() }}</span>
      </footer>
      <ChatWidget v-if="auth.authenticated && auth.dbRole !== 'none'" />
    </template>
  </div>
</template>

<style>
*,
*::before,
*::after {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  font-family: Inter, system-ui, sans-serif;
  background: #f8fafc;
  color: #1e293b;
  min-height: 100vh;
  overflow-y: scroll;
}

.app {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  overflow-x: hidden;
}

.main-content {
  flex: 1;
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
}

.footer {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.6rem;
  padding: 1rem 2rem;
  border-top: 1px solid #334155;
  background: #1e293b;
  font-size: 0.8rem;
  color: #94a3b8;
}

.footer-slogan { letter-spacing: 0.01em; color: #94a3b8; }
.footer-highlight { color: #f1f5f9; font-weight: 700; }
.footer-sep { color: #334155; }
.footer-copy { font-variant-numeric: tabular-nums; color: #64748b; }

.error-page {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: #f8fafc;
  padding: 2rem;
}

.error-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  padding: 3rem 2.5rem;
  max-width: 440px;
  width: 100%;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  box-shadow: 0 4px 24px rgba(0,0,0,0.06);
}

.error-icon {
  color: #94a3b8;
  margin-bottom: 0.5rem;
}

.error-title {
  font-size: 1.4rem;
  font-weight: 700;
  color: #0f172a;
}

.error-sub {
  font-size: 0.9rem;
  color: #64748b;
  line-height: 1.6;
}

.btn-retry {
  margin-top: 0.5rem;
  padding: 0.6rem 1.75rem;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s;
}

.btn-retry:hover { background: #2563eb; }
</style>
