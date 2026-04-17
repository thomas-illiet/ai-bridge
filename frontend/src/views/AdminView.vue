<script setup lang="ts">
import { ref, onMounted } from 'vue'
import UsersTab    from './admin/UsersTab.vue'
import WhitelistTab from './admin/WhitelistTab.vue'
import TokensTab   from './admin/TokensTab.vue'
import HistoryTab  from './admin/HistoryTab.vue'
import RequestsTab from './admin/RequestsTab.vue'
import { adminListAccessRequests } from '@/services/api'

const activeTab   = ref<'requests' | 'users' | 'whitelist' | 'tokens' | 'history'>('requests')
const pendingCount = ref(0)

onMounted(async () => {
  try {
    const res = await adminListAccessRequests('pending')
    pendingCount.value = res.data.pendingCount
  } catch { /* ignore */ }
})
</script>

<template>
  <div class="admin-page">
    <div class="page-header">
      <h1>Admin</h1>
      <p class="subtitle">Manage users and access restrictions.</p>
    </div>

    <div class="tabs">
      <button class="tab-btn" :class="{ active: activeTab === 'requests' }" @click="activeTab = 'requests'">
        Requests
        <span v-if="pendingCount > 0" class="badge">{{ pendingCount }}</span>
      </button>
      <button class="tab-btn" :class="{ active: activeTab === 'users' }"     @click="activeTab = 'users'">Users</button>
      <button class="tab-btn" :class="{ active: activeTab === 'whitelist' }" @click="activeTab = 'whitelist'">IP Whitelist</button>
      <button class="tab-btn" :class="{ active: activeTab === 'tokens' }"    @click="activeTab = 'tokens'">Tokens</button>
      <button class="tab-btn" :class="{ active: activeTab === 'history' }"   @click="activeTab = 'history'">History</button>
    </div>

    <RequestsTab  v-if="activeTab === 'requests'" />
    <UsersTab     v-if="activeTab === 'users'" />
    <WhitelistTab v-if="activeTab === 'whitelist'" />
    <TokensTab    v-if="activeTab === 'tokens'" />
    <HistoryTab   v-if="activeTab === 'history'" />
  </div>
</template>

<style scoped>
.admin-page { display: flex; flex-direction: column; gap: 1.5rem; }
.page-header h1 { font-size: 1.75rem; font-weight: 700; margin: 0 0 0.2rem; }
.subtitle { color: #64748b; font-size: 0.9rem; margin: 0; }
.tabs { display: flex; gap: 0.5rem; border-bottom: 1px solid #e2e8f0; }
.tab-btn { display: flex; align-items: center; gap: 0.4rem; padding: 0.5rem 1.1rem; border: none; background: none; cursor: pointer; font-size: 0.9rem; font-weight: 500; color: #64748b; border-bottom: 2px solid transparent; margin-bottom: -1px; transition: color 0.15s, border-color 0.15s; }
.tab-btn:hover { color: #1e293b; }
.tab-btn.active { color: #3b82f6; border-bottom-color: #3b82f6; }
.badge { display: inline-flex; align-items: center; justify-content: center; min-width: 18px; height: 18px; padding: 0 5px; border-radius: 999px; background: #ef4444; color: white; font-size: 0.7rem; font-weight: 700; }
</style>
