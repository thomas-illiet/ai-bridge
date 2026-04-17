<script setup lang="ts">
import NavBar from '@/components/NavBar.vue'
</script>

<template>
  <div class="app">
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
}

.app {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
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
  border-top: 1px solid #e2e8f0;
  background: #f8fafc;
  font-size: 0.8rem;
  color: #94a3b8;
}

.footer-slogan { letter-spacing: 0.01em; }
.footer-highlight { color: #3b82f6; font-weight: 600; }
.footer-sep { color: #cbd5e1; }
.footer-copy { font-variant-numeric: tabular-nums; }
</style>
