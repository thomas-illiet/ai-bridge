<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { getModels, getStatus } from '@/services/api'
import EndpointBanner from '@/views/help/EndpointBanner.vue'
import TabOpenWebUI from '@/views/help/TabOpenWebUI.vue'
import TabOpenCode from '@/views/help/TabOpenCode.vue'
import TabN8n from '@/views/help/TabN8n.vue'
import TabPython from '@/views/help/TabPython.vue'
import TabCurl from '@/views/help/TabCurl.vue'

const allProviders = [
  { id: 'openai' as const, label: 'OpenAI', color: '#10b981' },
  { id: 'ollama' as const, label: 'Ollama', color: '#6366f1' },
]

const availableProviders = ref<typeof allProviders>([])
const statusLoading      = ref(true)
const provider           = ref<'openai' | 'ollama'>('openai')

async function loadStatus() {
  statusLoading.value = true
  try {
    const res = await getStatus()
    const enabled = new Set(
      res.data.services.filter(s => s.status !== 'disabled').map(s => s.name)
    )
    availableProviders.value = allProviders.filter(p => enabled.has(p.id))
    if (availableProviders.value.length > 0) provider.value = availableProviders.value[0].id
  } catch {
    availableProviders.value = []
  } finally {
    statusLoading.value = false
  }
}

const modelList     = ref<string[]>([])
const modelLoading  = ref(false)
const selectedModel = ref('')
const defaultModels: Record<string, string> = { openai: 'gpt-4o', ollama: 'llama3.2' }

async function loadModels() {
  modelLoading.value = true
  modelList.value = []
  try {
    const res = await getModels(provider.value)
    modelList.value = res.data.models ?? []
    selectedModel.value = modelList.value.length > 0 ? modelList.value[0] : defaultModels[provider.value]
  } catch {
    selectedModel.value = defaultModels[provider.value]
  } finally {
    modelLoading.value = false
  }
}

watch(provider, () => loadModels())
onMounted(async () => {
  await loadStatus()
  if (availableProviders.value.length > 0) loadModels()
})

const baseURL     = computed(() => window.location.origin)
const openaiURL   = computed(() => `${baseURL.value}/openai/v1`)
const ollamaURL   = computed(() => `${baseURL.value}/ollama/v1`)
const activeURL   = computed(() => provider.value === 'openai' ? openaiURL.value : ollamaURL.value)
const activeModel = computed(() => selectedModel.value || defaultModels[provider.value])

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
        <p class="subtitle">Connect your tools to AI Bridge using the OpenAI-compatible API.</p>
      </div>

      <div v-if="availableProviders.length > 1" class="provider-selector">
        <span class="provider-label">Provider</span>
        <div class="provider-toggle">
          <button
            v-for="p in availableProviders"
            :key="p.id"
            class="provider-btn"
            :class="{ active: provider === p.id }"
            :style="provider === p.id ? { background: p.color, borderColor: p.color } : {}"
            @click="provider = p.id"
          >{{ p.label }}</button>
        </div>
      </div>
    </div>

    <div v-if="!statusLoading && availableProviders.length === 0" class="no-provider">
      <div class="no-provider-icon">⚙️</div>
      <h2>No provider configured</h2>
      <p>
        The administrator has not configured any AI provider yet.<br>
        Set <code>OPENAI_API_KEY</code> or <code>OLLAMA_BASE_URL</code> on the server to enable the integration.
      </p>
    </div>

    <template v-if="!statusLoading && availableProviders.length > 0">
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

      <TabOpenWebUI v-if="activeTab === 'openwebui'" :active-url="activeURL" :ollama-url="ollamaURL" :provider />
      <TabOpenCode  v-if="activeTab === 'opencode'"  :active-url="activeURL" :active-model="activeModel" :provider />
      <TabN8n       v-if="activeTab === 'n8n'"       :active-url="activeURL" :active-model="activeModel" :provider />
      <TabPython    v-if="activeTab === 'python'"    :active-url="activeURL" :active-model="activeModel" :provider />
      <TabCurl      v-if="activeTab === 'curl'"      :active-url="activeURL" :active-model="activeModel" :provider />
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
.provider-label { font-size: 0.8rem; font-weight: 600; color: #64748b; }
.provider-toggle { display: flex; border: 1px solid #e2e8f0; border-radius: 8px; overflow: hidden; }
.provider-btn {
  padding: 0.4rem 1rem;
  border: none;
  cursor: pointer;
  font-size: 0.85rem;
  font-weight: 600;
  color: #64748b;
  background: white;
  transition: background 0.15s, color 0.15s;
}
.provider-btn:not(:last-child) { border-right: 1px solid #e2e8f0; }
.provider-btn.active { color: white; }
.provider-btn:not(.active):hover { background: #f8fafc; color: #1e293b; }

.no-provider {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  padding: 3rem 2rem;
  text-align: center;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
}
.no-provider-icon { font-size: 2.5rem; }
.no-provider h2 { font-size: 1.15rem; font-weight: 700; color: #1e293b; margin: 0; }
.no-provider p { color: #64748b; font-size: 0.92rem; margin: 0; line-height: 1.6; }
.no-provider code { background: #f1f5f9; border-radius: 4px; padding: 0.1rem 0.35rem; font-family: monospace; color: #0f172a; }

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
