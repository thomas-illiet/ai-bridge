<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { getAvailableProviders, getModels } from '@/services/api'
import type { ProviderInfo } from '@/services/api'
import { getConfig } from '@/services/config'
import EndpointBanner from '@/views/help/EndpointBanner.vue'
import TabOpenWebUI from '@/views/help/TabOpenWebUI.vue'
import TabOpenCode from '@/views/help/TabOpenCode.vue'
import TabN8n from '@/views/help/TabN8n.vue'
import TabPython from '@/views/help/TabPython.vue'
import TabCurl from '@/views/help/TabCurl.vue'
import TabClaude from '@/views/help/TabClaude.vue'

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
const defaultModels: Record<string, string> = { openai: 'gpt-4o', ollama: 'llama3.2', anthropic: 'claude-sonnet-4-6' }

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

const baseURL      = computed(() => getConfig().apiBaseUrl)
const activeURL    = computed(() => selectedProvider.value ? `${baseURL.value}/${selectedProvider.value.name}/v1` : '')
const activeModel  = computed(() => selectedModel.value || defaultModels[selectedProvider.value?.type ?? 'openai'])
const providerType = computed(() => selectedProvider.value?.type ?? 'openai')

const activeTab = ref<'openwebui' | 'opencode' | 'n8n' | 'python' | 'curl' | 'claude'>('openwebui')

const tabs = computed(() =>
  providerType.value === 'anthropic'
    ? [{ id: 'claude' as const, label: 'Claude Code', icon: '◆' }]
    : [
        { id: 'openwebui' as const, label: 'Open WebUI', icon: '◉' },
        { id: 'opencode'  as const, label: 'OpenCode',   icon: '◈' },
        { id: 'n8n'       as const, label: 'n8N',        icon: '⬡' },
        { id: 'python'    as const, label: 'Python',     icon: '⬢' },
        { id: 'curl'      as const, label: 'cURL',       icon: '▶' },
      ]
)

watch(providerType, (type) => {
  activeTab.value = type === 'anthropic' ? 'claude' : 'openwebui'
})
</script>

<template>
  <div class="help-page">
    <div class="page-header">
      <div class="header-text">
        <h1>Integration Guide</h1>
        <p class="subtitle">Connect your tools and apps to AI Bridge in minutes.</p>
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
      <div class="provider-section">
        <span class="section-label">Provider</span>
        <div class="provider-pills">
          <button
            v-for="p in providers"
            :key="p.name"
            class="provider-pill"
            :class="[`type-${p.type}`, { active: selectedProvider?.name === p.name }]"
            @click="selectedProvider = p"
          >
            <span class="pill-name">{{ p.name }}</span>
            <span class="pill-type">{{ p.type }}</span>
          </button>
        </div>
      </div>

      <EndpointBanner
        :active-url="activeURL"
        :active-model="activeModel"
        :model-list="modelList"
        :model-loading="modelLoading"
        v-model:selected-model="selectedModel"
      />

      <div class="tabs-wrapper">
        <div class="tabs">
          <button
            v-for="tab in tabs"
            :key="tab.id"
            class="tab-btn"
            :class="{ active: activeTab === tab.id }"
            @click="activeTab = tab.id"
          >
            <span class="tab-icon">{{ tab.icon }}</span>
            {{ tab.label }}
          </button>
        </div>
      </div>

      <TabClaude    v-if="activeTab === 'claude'"    :active-url="activeURL" :active-model="activeModel" />
      <TabOpenWebUI v-if="activeTab === 'openwebui'" :active-url="activeURL" :provider="providerType" />
      <TabOpenCode  v-if="activeTab === 'opencode'"  :active-url="activeURL" :active-model="activeModel" :provider="providerType" />
      <TabN8n       v-if="activeTab === 'n8n'"       :active-url="activeURL" :active-model="activeModel" :provider="providerType" />
      <TabPython    v-if="activeTab === 'python'"    :active-url="activeURL" :active-model="activeModel" :provider="providerType" />
      <TabCurl      v-if="activeTab === 'curl'"      :active-url="activeURL" :active-model="activeModel" :provider="providerType" />
    </template>
  </div>
</template>

<style scoped>
.help-page { display: flex; flex-direction: column; gap: 1.25rem; }

/* Header */
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
}
.page-header h1 { font-size: 1.75rem; font-weight: 700; margin: 0 0 0.25rem; color: #0f172a; }
.subtitle { color: #64748b; font-size: 0.95rem; margin: 0; }

/* Provider pills */
.provider-section { display: flex; align-items: center; gap: 0.75rem; flex-wrap: wrap; }
.section-label { font-size: 0.8rem; font-weight: 600; color: #94a3b8; text-transform: uppercase; letter-spacing: 0.05em; white-space: nowrap; }
.provider-pills { display: flex; gap: 0.5rem; flex-wrap: wrap; }

.provider-pill {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4rem 0.85rem;
  border: 1.5px solid #e2e8f0;
  border-radius: 999px;
  background: #fff;
  cursor: pointer;
  transition: all 0.15s;
  font-size: 0.875rem;
  color: #475569;
  font-weight: 500;
}
.provider-pill:hover { border-color: #cbd5e1; background: #f8fafc; color: #1e293b; }
.provider-pill.active { border-color: #3b82f6; background: #eff6ff; color: #1d4ed8; }
.provider-pill.active .pill-type { background: #bfdbfe; color: #1d4ed8; }

.pill-name { font-weight: 600; }
.pill-type {
  font-size: 0.7rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  background: #f1f5f9;
  color: #64748b;
  border-radius: 999px;
  padding: 0.1rem 0.45rem;
}
.type-anthropic.active { border-color: #f59e0b; background: #fffbeb; color: #92400e; }
.type-anthropic.active .pill-type { background: #fde68a; color: #92400e; }
.type-ollama.active { border-color: #8b5cf6; background: #f5f3ff; color: #5b21b6; }
.type-ollama.active .pill-type { background: #ddd6fe; color: #5b21b6; }

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
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: #3b82f6;
  color: #fff;
  font-size: 0.72rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Tabs */
.tabs-wrapper {
  border-bottom: 1px solid #e2e8f0;
}
.tabs { display: flex; gap: 0.25rem; }
.tab-btn {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.55rem 1.1rem;
  border: none;
  background: none;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  color: #64748b;
  border-bottom: 2px solid transparent;
  margin-bottom: -1px;
  border-radius: 6px 6px 0 0;
  transition: color 0.15s, background 0.15s, border-color 0.15s;
}
.tab-btn:hover { color: #1e293b; background: #f8fafc; }
.tab-btn.active { color: #2563eb; border-bottom-color: #3b82f6; background: #eff6ff; }
.tab-icon { font-size: 0.7rem; opacity: 0.7; }
</style>
