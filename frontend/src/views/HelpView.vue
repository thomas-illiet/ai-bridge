<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { getAvailableProviders, getModels } from '@/services/api'
import type { ProviderInfo } from '@/services/api'
import EndpointBanner from '@/views/help/EndpointBanner.vue'
import TabOpenWebUI from '@/views/help/TabOpenWebUI.vue'
import TabOpenCode from '@/views/help/TabOpenCode.vue'
import TabN8n from '@/views/help/TabN8n.vue'
import TabPython from '@/views/help/TabPython.vue'
import TabCurl from '@/views/help/TabCurl.vue'

const providers      = ref<ProviderInfo[]>([])
const loading        = ref(true)
const selectedProvider = ref<ProviderInfo | null>(null)

async function loadProviders() {
  loading.value = true
  try {
    const res = await getAvailableProviders()
    providers.value = res.data.providers ?? []
    if (providers.value.length > 0) selectedProvider.value = providers.value[0]
  } catch {
    providers.value = []
  } finally {
    loading.value = false
  }
}

const modelList     = ref<string[]>([])
const modelLoading  = ref(false)
const selectedModel = ref('')
const defaultModels: Record<string, string> = { openai: 'gpt-4o', ollama: 'llama3.2' }

async function loadModels() {
  if (!selectedProvider.value) return
  modelLoading.value = true
  modelList.value = []
  try {
    const res = await getModels(selectedProvider.value.name)
    modelList.value = res.data.models ?? []
    selectedModel.value = modelList.value.length > 0 ? modelList.value[0] : defaultModels[selectedProvider.value.type] ?? ''
  } catch {
    selectedModel.value = defaultModels[selectedProvider.value?.type ?? 'openai'] ?? ''
  } finally {
    modelLoading.value = false
  }
}

watch(selectedProvider, () => loadModels())
onMounted(async () => {
  await loadProviders()
  if (selectedProvider.value) loadModels()
})

const baseURL      = computed(() => window.location.origin)
const activeURL    = computed(() => selectedProvider.value ? `${baseURL.value}/${selectedProvider.value.name}/v1` : '')
const activeModel  = computed(() => selectedModel.value || defaultModels[selectedProvider.value?.type ?? 'openai'])
const providerType = computed(() => selectedProvider.value?.type ?? 'openai')

const activeTab = ref<'openwebui' | 'opencode' | 'n8n' | 'python' | 'curl'>('openwebui')
const tabs = [
  { id: 'openwebui', label: 'Open WebUI' },
  { id: 'opencode',  label: 'OpenCode' },
  { id: 'n8n',       label: 'n8N' },
  { id: 'python',    label: 'Python' },
  { id: 'curl',      label: 'cURL' },
] as const
</script>

<template>
  <div class="help-page">
    <div class="page-header">
      <div>
        <h1>Integration Guide</h1>
        <p class="subtitle">Connect your tools to AI Bridge.</p>
      </div>

      <div v-if="!loading && providers.length > 0" class="provider-selector">
        <span class="provider-label">Provider</span>
        <select
          class="provider-select"
          :value="selectedProvider?.name"
          @change="selectedProvider = providers.find(p => p.name === ($event.target as HTMLSelectElement).value) ?? null"
        >
          <option v-for="p in providers" :key="p.name" :value="p.name">{{ p.name }} — {{ p.type }}</option>
        </select>
      </div>
    </div>

    <div v-if="loading" class="loading-state">
      <div class="loading-spinner" />
      <span>Loading providers…</span>
    </div>

    <div v-else-if="providers.length === 0" class="empty-state">
      <div class="empty-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M21 7.5l-9-5.25L3 7.5m18 0l-9 5.25m9-5.25v9l-9 5.25M3 7.5l9 5.25M3 7.5v9l9 5.25m0-9v9" />
        </svg>
      </div>
      <div class="empty-body">
        <h2>No provider configured</h2>
        <p>An administrator must add at least one AI provider before integration guides become available.</p>
        <div class="empty-steps">
          <div class="empty-step">
            <span class="step-num">1</span>
            <span>Sign in as an admin and open <strong>Admin → Providers</strong>.</span>
          </div>
          <div class="empty-step">
            <span class="step-num">2</span>
            <span>Click <strong>Add provider</strong> and enter the connection details.</span>
          </div>
          <div class="empty-step">
            <span class="step-num">3</span>
            <span>Return here — the integration guides will appear automatically.</span>
          </div>
        </div>
      </div>
    </div>

    <template v-if="!loading && providers.length > 0 && selectedProvider">
      <EndpointBanner
        :active-url="activeURL"
        :active-model="activeModel"
        :model-list="modelList"
        :model-loading="modelLoading"
        v-model:selected-model="selectedModel"
      />

      <div class="tabs">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          class="tab-btn"
          :class="{ active: activeTab === tab.id }"
          @click="activeTab = tab.id"
        >{{ tab.label }}</button>
      </div>

      <TabOpenWebUI v-if="activeTab === 'openwebui'" :active-url="activeURL" :provider="providerType" />
      <TabOpenCode  v-if="activeTab === 'opencode'"  :active-url="activeURL" :active-model="activeModel" :provider="providerType" />
      <TabN8n       v-if="activeTab === 'n8n'"       :active-url="activeURL" :active-model="activeModel" :provider="providerType" />
      <TabPython    v-if="activeTab === 'python'"    :active-url="activeURL" :active-model="activeModel" :provider="providerType" />
      <TabCurl      v-if="activeTab === 'curl'"      :active-url="activeURL" :active-model="activeModel" :provider="providerType" />
    </template>
  </div>
</template>

<style scoped>
.help-page { display: flex; flex-direction: column; gap: 1.5rem; }

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
}
.page-header h1 { font-size: 1.75rem; font-weight: 700; margin: 0 0 0.25rem; }
.subtitle { color: #64748b; font-size: 0.95rem; margin: 0; }

.provider-selector { display: flex; align-items: center; gap: 0.6rem; flex-shrink: 0; }
.provider-label { font-size: 0.8rem; font-weight: 600; color: #64748b; white-space: nowrap; }

.provider-select {
  font-size: 0.85rem;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 0.4rem 0.75rem;
  color: #1e293b;
  background: white;
  cursor: pointer;
  outline: none;
}
.provider-select:focus { border-color: #3b82f6; }

/* Loading */
.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  padding: 3rem 2rem;
  color: #64748b;
  font-size: 0.92rem;
}
.loading-spinner {
  width: 20px;
  height: 20px;
  border: 2px solid #e2e8f0;
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* Empty state */
.empty-state {
  display: flex;
  gap: 1.5rem;
  padding: 2rem;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
}
.empty-icon {
  flex-shrink: 0;
  width: 48px;
  height: 48px;
  border-radius: 10px;
  background: #f1f5f9;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #94a3b8;
}
.empty-icon svg { width: 24px; height: 24px; }
.empty-body { display: flex; flex-direction: column; gap: 0.75rem; }
.empty-body h2 { font-size: 1.05rem; font-weight: 700; color: #1e293b; margin: 0; }
.empty-body p { color: #64748b; font-size: 0.9rem; margin: 0; line-height: 1.5; }
.empty-steps { display: flex; flex-direction: column; gap: 0.5rem; }
.empty-step {
  display: flex;
  align-items: baseline;
  gap: 0.6rem;
  font-size: 0.88rem;
  color: #475569;
}
.empty-step strong { color: #1e293b; }
.step-num {
  flex-shrink: 0;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: #e2e8f0;
  color: #475569;
  font-size: 0.72rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Tabs */
.tabs { display: flex; gap: 0.5rem; border-bottom: 1px solid #e2e8f0; }
.tab-btn {
  padding: 0.5rem 1.1rem;
  border: none;
  background: none;
  cursor: pointer;
  font-size: 0.9rem;
  font-weight: 500;
  color: #64748b;
  border-bottom: 2px solid transparent;
  margin-bottom: -1px;
  transition: color 0.15s, border-color 0.15s;
}
.tab-btn:hover { color: #1e293b; }
.tab-btn.active { color: #3b82f6; border-bottom-color: #3b82f6; }
</style>
