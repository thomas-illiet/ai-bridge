<script setup lang="ts">
import { ref } from 'vue'
import UsersTab    from './admin/UsersTab.vue'
import WhitelistTab from './admin/WhitelistTab.vue'
import TokensTab   from './admin/TokensTab.vue'
import HistoryTab  from './admin/HistoryTab.vue'

const activeTab = ref<'users' | 'whitelist' | 'tokens' | 'history'>('users')
</script>

<template>
  <div class="admin-page">
    <div class="page-header">
      <h1>Admin</h1>
      <p class="subtitle">Manage users and access restrictions.</p>
    </div>

    <div class="tabs">
      <button class="tab-btn" :class="{ active: activeTab === 'users' }"     @click="activeTab = 'users'">Users</button>
      <button class="tab-btn" :class="{ active: activeTab === 'whitelist' }" @click="activeTab = 'whitelist'">IP Whitelist</button>
      <button class="tab-btn" :class="{ active: activeTab === 'tokens' }"    @click="activeTab = 'tokens'">Tokens</button>
      <button class="tab-btn" :class="{ active: activeTab === 'history' }"   @click="activeTab = 'history'">History</button>
    </div>

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
.tab-btn { padding: 0.5rem 1.1rem; border: none; background: none; cursor: pointer; font-size: 0.9rem; font-weight: 500; color: #64748b; border-bottom: 2px solid transparent; margin-bottom: -1px; transition: color 0.15s, border-color 0.15s; }
.tab-btn:hover { color: #1e293b; }
.tab-btn.active { color: #3b82f6; border-bottom-color: #3b82f6; }
</style>
