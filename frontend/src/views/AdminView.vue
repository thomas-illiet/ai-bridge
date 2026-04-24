<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import OverviewTab         from './admin/OverviewTab.vue'
import UsersTab           from './admin/UsersTab.vue'
import FirewallTab         from './admin/FirewallTab.vue'
import TokensTab           from './admin/TokensTab.vue'
import HistoryTab          from './admin/HistoryTab.vue'
import RequestsTab         from './admin/RequestsTab.vue'
import ServiceAccountsTab  from './admin/ServiceAccountsTab.vue'
import ProvidersTab        from './admin/ProvidersTab.vue'
import MCPTab              from './admin/MCPTab.vue'
import { adminListAccessRequests } from '@/services/api'

const auth = useAuthStore()

type Tab = 'overview' | 'requests' | 'users' | 'firewall' | 'tokens' | 'service-accounts' | 'history' | 'providers' | 'mcp'
const activeTab    = ref<Tab>('overview')
const pendingCount = ref(0)

onMounted(async () => {
  try {
    const res = await adminListAccessRequests('pending')
    pendingCount.value = res.data.pendingCount
  } catch { /* ignore */ }
})

interface NavItem { id: Tab; label: string; icon: string; adminOnly?: boolean }
interface NavGroup { label: string; items: NavItem[] }

const navGroups: NavGroup[] = [
  {
    label: 'Overview',
    items: [
      { id: 'overview', label: 'Dashboard', icon: 'M3 3h7v7H3zM14 3h7v7h-7zM3 14h7v7H3zM14 14h7v7h-7z' },
    ],
  },
  {
    label: 'Management',
    items: [
      { id: 'requests', label: 'Requests',    icon: 'M9 5H7a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2h-2M9 5a2 2 0 0 0 2 2h2a2 2 0 0 0 2-2M9 5a2 2 0 0 1 2-2h2a2 2 0 0 1 2 2m-6 9 2 2 4-4' },
      { id: 'users',    label: 'Users',        icon: 'M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2M9 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8zm8 0c1.66 0 3-1.34 3-3s-1.34-3-3-3M23 21v-2c0-1.45-.98-2.67-2.33-3.08' },
    ],
  },
  {
    label: 'Security',
    items: [
      { id: 'firewall', label: 'Firewall', icon: 'M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z', adminOnly: true },
      { id: 'providers', label: 'Providers',    icon: 'M22 12h-4l-3 9L9 3l-3 9H2',               adminOnly: true },
      { id: 'mcp',       label: 'MCP Servers',  icon: 'M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71', adminOnly: true },
    ],
  },
  {
    label: 'Monitoring',
    items: [
      { id: 'tokens',           label: 'Tokens',          icon: 'M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0 3 3L22 7l-3-3m-3.5 3.5L19 4' },
      { id: 'service-accounts', label: 'Service Accounts', icon: 'M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2M12 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8z', adminOnly: true },
      { id: 'history',          label: 'History',          icon: 'M12 8v4l3 3m6-3a9 9 0 1 1-18 0 9 9 0 0 1 18 0z' },
    ],
  },
]

function visibleItems(items: NavItem[]) {
  return items.filter(i => !i.adminOnly || auth.isAdmin)
}
</script>

<template>
  <div class="admin-shell">
    <!-- ── Sidebar ───────────────────────────────────────────────────── -->
    <aside class="admin-sidebar">
      <nav class="sidebar-nav">
        <template v-for="group in navGroups" :key="group.label">
          <template v-if="visibleItems(group.items).length">
            <p class="nav-group-label">{{ group.label }}</p>
            <button
              v-for="item in visibleItems(group.items)"
              :key="item.id"
              class="nav-item"
              :class="{ active: activeTab === item.id }"
              @click="activeTab = item.id"
            >
              <svg class="nav-icon" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path :d="item.icon"/>
              </svg>
              <span>{{ item.label }}</span>
              <span v-if="item.id === 'requests' && pendingCount > 0" class="nav-badge">{{ pendingCount }}</span>
            </button>
          </template>
        </template>
      </nav>
    </aside>

    <!-- ── Content ──────────────────────────────────────────────────── -->
    <main class="admin-content">
      <div class="content-header">
        <div>
          <h1 class="content-title" style="display:flex;align-items:center;gap:0.6rem">
            {{ navGroups.flatMap(g => g.items).find(i => i.id === activeTab)?.label }}
          </h1>
          <p class="content-sub">
            <template v-if="activeTab === 'overview'">Global activity metrics, service status, and platform health.</template>
            <template v-else-if="activeTab === 'requests'">Review and approve or reject user access requests.</template>
            <template v-else-if="activeTab === 'users'">Manage registered users and their roles.</template>
            <template v-else-if="activeTab === 'firewall'">Define allow/deny firewall rules to control proxy access by IP address or CIDR range.</template>
            <template v-else-if="activeTab === 'providers'">Configure AI provider backends available through the proxy.</template>
            <template v-else-if="activeTab === 'mcp'">Manage MCP servers for centralized tool injection into proxied AI requests.</template>
            <template v-else-if="activeTab === 'tokens'">View and revoke all user API tokens.</template>
            <template v-else-if="activeTab === 'service-accounts'">Manage service accounts and their long-lived tokens.</template>
            <template v-else-if="activeTab === 'history'">Browse all proxied AI requests across every user.</template>
          </p>
        </div>
        <div id="admin-search-portal" class="content-portal" />
      </div>

      <OverviewTab        v-if="activeTab === 'overview'" />
      <RequestsTab        v-if="activeTab === 'requests'" />
      <UsersTab           v-if="activeTab === 'users'" />
      <FirewallTab        v-if="activeTab === 'firewall' && auth.isAdmin" />
      <TokensTab          v-if="activeTab === 'tokens'" />
      <ServiceAccountsTab v-if="activeTab === 'service-accounts' && auth.isAdmin" />
      <ProvidersTab       v-if="activeTab === 'providers' && auth.isAdmin" />
      <MCPTab             v-if="activeTab === 'mcp' && auth.isAdmin" />
      <HistoryTab         v-if="activeTab === 'history'" />
    </main>
  </div>
</template>

<style scoped>
/* ── Shell layout ──────────────────────────────────────────────────────────── */
.admin-shell {
  display: flex;
  gap: 0;
  /* escape .main-content padding + max-width centering */
  width: 100vw;
  position: relative;
  left: calc(-50vw + 50%);
  margin-top: -2rem;
  margin-bottom: -2rem;
  min-height: calc(100vh - 60px - 49px);
}

/* ── Sidebar ───────────────────────────────────────────────────────────────── */
.admin-sidebar {
  width: 220px;
  flex-shrink: 0;
  background: #1e293b;
  border-right: 1px solid #334155;
  display: flex;
  flex-direction: column;
  padding: 0.75rem 0 1.5rem;
}

/* ── Nav groups ────────────────────────────────────────────────────────────── */
.sidebar-nav {
  display: flex;
  flex-direction: column;
  padding: 0 0.75rem;
  gap: 0.1rem;
}

.nav-group-label {
  font-size: 0.68rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: #475569;
  padding: 0.85rem 0.6rem 0.3rem;
  margin: 0;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  width: 100%;
  padding: 0.55rem 0.75rem;
  border: none;
  border-radius: 7px;
  background: none;
  color: #94a3b8;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  text-align: left;
  transition: background 0.12s, color 0.12s;
  position: relative;
}

.nav-item:hover {
  background: #334155;
  color: #e2e8f0;
}

.nav-item.active {
  background: #1d4ed8;
  color: #fff;
}

.nav-item.active .nav-icon { opacity: 1; }

.nav-icon {
  flex-shrink: 0;
  opacity: 0.7;
}

.nav-item.active .nav-icon,
.nav-item:hover .nav-icon { opacity: 1; }

.nav-badge {
  margin-left: auto;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  border-radius: 999px;
  background: #ef4444;
  color: white;
  font-size: 0.68rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* ── Content area ──────────────────────────────────────────────────────────── */
.admin-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  padding: 2rem;
  background: #f1f5f9;
  min-width: 0;
}

.content-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
}
.content-portal { display: flex; align-items: center; gap: 0.5rem; }

.content-title {
  font-size: 1.4rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0 0 0.2rem;
}

.content-sub {
  font-size: 0.875rem;
  color: #64748b;
  margin: 0;
}
</style>
