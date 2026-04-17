<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { RouterLink } from 'vue-router'

const auth = useAuthStore()
const menuOpen = ref(false)
const menuRef = ref<HTMLElement | null>(null)

function closeMenu() { menuOpen.value = false }

function onClickOutside(e: MouseEvent) {
  if (menuRef.value && !menuRef.value.contains(e.target as Node)) {
    closeMenu()
  }
}

onMounted(() => document.addEventListener('mousedown', onClickOutside))
onUnmounted(() => document.removeEventListener('mousedown', onClickOutside))
</script>

<template>
  <nav class="navbar">
    <div class="navbar-brand">
      <RouterLink to="/">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32" class="brand-logo" aria-hidden="true">
          <defs>
            <linearGradient id="nav-bg" x1="0%" y1="0%" x2="100%" y2="100%">
              <stop offset="0%" stop-color="#6366f1"/>
              <stop offset="100%" stop-color="#3b82f6"/>
            </linearGradient>
          </defs>
          <rect width="32" height="32" rx="7" fill="url(#nav-bg)"/>
          <path d="M5 21 Q16 7 27 21" fill="none" stroke="white" stroke-width="2.5" stroke-linecap="round"/>
          <line x1="9" y1="21" x2="9" y2="27" stroke="white" stroke-width="2.5" stroke-linecap="round"/>
          <line x1="23" y1="21" x2="23" y2="27" stroke="white" stroke-width="2.5" stroke-linecap="round"/>
          <line x1="5" y1="21" x2="27" y2="21" stroke="white" stroke-width="2" stroke-linecap="round" opacity="0.7"/>
          <circle cx="5" cy="21" r="2.2" fill="white"/>
          <circle cx="27" cy="21" r="2.2" fill="white"/>
        </svg>
        <span>AI Bridge</span>
      </RouterLink>
    </div>

    <div class="navbar-links">
      <RouterLink v-if="!auth.authenticated" to="/">Home</RouterLink>
      <RouterLink v-if="auth.authenticated" to="/dashboard">Dashboard</RouterLink>
      <RouterLink v-if="auth.authenticated && (auth.hasRole('user') || auth.isAdmin)" to="/tokens">Tokens</RouterLink>
      <RouterLink v-if="auth.authenticated && (auth.hasRole('user') || auth.isAdmin)" to="/history">History</RouterLink>
      <RouterLink v-if="auth.isAdmin" to="/admin">Admin</RouterLink>
      <RouterLink to="/help">Help</RouterLink>
    </div>

    <div class="navbar-auth">
      <template v-if="auth.authenticated">
        <div class="user-menu" ref="menuRef">
          <button class="user-trigger" @click="menuOpen = !menuOpen">
            <span class="user-avatar">{{ auth.username?.charAt(0).toUpperCase() }}</span>
            <span class="user-name">{{ auth.username }}</span>
            <span class="chevron" :class="{ open: menuOpen }">▾</span>
          </button>

          <div v-if="menuOpen" class="dropdown">
            <RouterLink to="/profile" class="dropdown-item" @click="closeMenu">
              <span class="item-icon">👤</span> Profile
            </RouterLink>
            <div class="dropdown-divider" />
            <button class="dropdown-item danger" @click="auth.logout(); closeMenu()">
              <span class="item-icon">→</span> Logout
            </button>
          </div>
        </div>
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
  display: flex;
  align-items: center;
  gap: 0.6rem;
  font-size: 1.25rem;
  font-weight: 700;
  color: #f1f5f9;
  text-decoration: none;
}

.brand-logo {
  width: 28px;
  height: 28px;
  flex-shrink: 0;
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
}

/* user menu */
.user-menu {
  position: relative;
}

.user-trigger {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.3rem 0.5rem;
  border-radius: 6px;
  transition: background 0.15s;
  color: #f1f5f9;
}

.user-trigger:hover { background: #334155; }

.user-avatar {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: #3b82f6;
  color: #fff;
  font-size: 0.85rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.user-name {
  font-size: 0.9rem;
  font-weight: 500;
  color: #f1f5f9;
}

.chevron {
  font-size: 0.75rem;
  color: #94a3b8;
  transition: transform 0.15s;
  display: inline-block;
}
.chevron.open { transform: rotate(180deg); }

/* dropdown */
.dropdown {
  position: absolute;
  top: calc(100% + 6px);
  right: 0;
  min-width: 160px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 8px;
  box-shadow: 0 8px 24px rgba(0,0,0,0.35);
  z-index: 100;
  overflow: hidden;
  padding: 0.3rem 0;
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  width: 100%;
  padding: 0.55rem 1rem;
  font-size: 0.88rem;
  font-weight: 500;
  color: #cbd5e1;
  text-decoration: none;
  background: none;
  border: none;
  cursor: pointer;
  transition: background 0.12s, color 0.12s;
  text-align: left;
}

.dropdown-item:hover {
  background: #334155;
  color: #f1f5f9;
}

.dropdown-item.danger { color: #f87171; }
.dropdown-item.danger:hover { background: #450a0a; color: #fca5a5; }

.item-icon { font-size: 0.85rem; width: 16px; text-align: center; }

.dropdown-divider {
  height: 1px;
  background: #334155;
  margin: 0.3rem 0;
}

/* login button */
.btn {
  padding: 0.4rem 1rem;
  border-radius: 6px;
  border: none;
  cursor: pointer;
  font-size: 0.9rem;
  font-weight: 500;
  transition: background 0.15s;
}

.btn-primary { background: #3b82f6; color: white; }
.btn-primary:hover { background: #2563eb; }
</style>
