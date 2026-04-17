<script setup lang="ts">
import { computed, ref, watch, onMounted } from 'vue'
import { getModels, getStatus } from '@/services/api'

// ── all known providers ───────────────────────────────────────────────────
const allProviders = [
  { id: 'openai' as const, label: 'OpenAI', color: '#10b981' },
  { id: 'ollama' as const, label: 'Ollama', color: '#6366f1' },
]

// ── available providers (filtered by status) ──────────────────────────────
const availableProviders = ref<typeof allProviders>([])
const statusLoading = ref(true)

async function loadStatus() {
  statusLoading.value = true
  try {
    const res = await getStatus()
    const enabled = new Set(
      res.data.services
        .filter(s => s.status !== 'disabled')
        .map(s => s.name)
    )
    availableProviders.value = allProviders.filter(p => enabled.has(p.id))
    if (availableProviders.value.length > 0) {
      provider.value = availableProviders.value[0].id
    }
  } catch (_) {
    availableProviders.value = []
  } finally {
    statusLoading.value = false
  }
}

// ── provider selector ─────────────────────────────────────────────────────
const provider = ref<'openai' | 'ollama'>('openai')

// ── models ────────────────────────────────────────────────────────────────
const modelList = ref<string[]>([])
const modelLoading = ref(false)
const selectedModel = ref('')

const defaultModels: Record<string, string> = { openai: 'gpt-4o', ollama: 'llama3.2' }

async function loadModels() {
  modelLoading.value = true
  modelList.value = []
  try {
    const res = await getModels(provider.value)
    modelList.value = res.data.models ?? []
    selectedModel.value = modelList.value.length > 0 ? modelList.value[0] : defaultModels[provider.value]
  } catch (_) {
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

// ── dynamic values ────────────────────────────────────────────────────────
const baseURL    = computed(() => window.location.origin)
const openaiURL  = computed(() => `${baseURL.value}/openai/v1`)
const ollamaURL  = computed(() => `${baseURL.value}/ollama/v1`)
const activeURL  = computed(() => provider.value === 'openai' ? openaiURL.value : ollamaURL.value)
const activeModel = computed(() => selectedModel.value || defaultModels[provider.value])

// ── tabs ──────────────────────────────────────────────────────────────────
const activeTab = ref<'openwebui' | 'opencode' | 'n8n' | 'python' | 'curl'>('openwebui')

const tabs = [
  { id: 'openwebui', label: 'Open WebUI' },
  { id: 'opencode',  label: 'OpenCode' },
  { id: 'n8n',       label: 'n8N' },
  { id: 'python',    label: 'Python' },
  { id: 'curl',      label: 'cURL' },
] as const

// ── code snippets ─────────────────────────────────────────────────────────
const opencodeSnippetDisplay = computed(() =>
`{
  "providers": {
    "aibridge": {
      "name": "AI Bridge",
      "apiKey": "YOUR_PAT_HERE",
      "baseURL": "${activeURL.value}"
    }
  }
}`)

const opencodeSnippet = computed(() => opencodeSnippetDisplay.value)

const pythonSnippetDisplay = computed(() =>
`from openai import OpenAI

client = OpenAI(
    api_key="YOUR_PAT_HERE",
    base_url="${activeURL.value}",
)

response = client.chat.completions.create(
    model="${activeModel.value}",
    messages=[{"role": "user", "content": "Hello!"}],
)
print(response.choices[0].message.content)`)

const pythonSnippet = computed(() => pythonSnippetDisplay.value)

const curlModelsDisplay  = computed(() =>
`curl ${activeURL.value}/models \\
  -H "Authorization: Bearer YOUR_PAT_HERE"`)
const curlModelsSnippet  = computed(() => curlModelsDisplay.value)

const curlChatDisplay = computed(() =>
`curl ${activeURL.value}/chat/completions \\
  -H "Authorization: Bearer YOUR_PAT_HERE" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "${activeModel.value}",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'`)
const curlChatSnippet = computed(() => curlChatDisplay.value)

const curlStreamDisplay = computed(() =>
`curl ${activeURL.value}/chat/completions \\
  -H "Authorization: Bearer YOUR_PAT_HERE" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "${activeModel.value}",
    "stream": true,
    "messages": [{"role": "user", "content": "Hello!"}]
  }'`)
const curlStreamSnippet = computed(() => curlStreamDisplay.value)

// ── copy helper ───────────────────────────────────────────────────────────
const copied = ref<string | null>(null)
function copy(text: string, key: string) {
  navigator.clipboard.writeText(text)
  copied.value = key
  setTimeout(() => { if (copied.value === key) copied.value = null }, 1800)
}
</script>

<template>
  <div class="help-page">
    <div class="page-header">
      <div>
        <h1>Integration Guide</h1>
        <p class="subtitle">Connect your tools to AI Bridge using the OpenAI-compatible API.</p>
      </div>

      <!-- provider selector: only when 2+ providers are available -->
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

    <!-- no provider configured -->
    <div v-if="!statusLoading && availableProviders.length === 0" class="no-provider">
      <div class="no-provider-icon">⚙️</div>
      <h2>No provider configured</h2>
      <p>The administrator has not configured any AI provider yet.<br>
         Set <code>OPENAI_API_KEY</code> or <code>OLLAMA_BASE_URL</code> on the server to enable the integration.</p>
    </div>

    <template v-if="!statusLoading && availableProviders.length > 0">

    <!-- endpoint banner -->
    <div class="info-banner">
      <div class="info-row">
        <span class="info-label">Active endpoint</span>
        <code class="info-value">{{ activeURL }}</code>
        <button class="copy-btn" @click="copy(activeURL, 'active-url')">
          {{ copied === 'active-url' ? 'Copied!' : 'Copy' }}
        </button>
      </div>
      <div class="info-row">
        <span class="info-label">Model</span>
        <span v-if="modelLoading" class="info-value muted">Loading…</span>
        <template v-else-if="modelList.length > 0">
          <select class="model-select" v-model="selectedModel">
            <option v-for="m in modelList" :key="m" :value="m">{{ m }}</option>
          </select>
        </template>
        <code v-else class="info-value">{{ activeModel }}</code>
      </div>
      <div class="info-row">
        <span class="info-label">API Key</span>
        <span class="info-value muted">Generate a Personal Access Token in <RouterLink to="/tokens">Tokens</RouterLink></span>
      </div>
    </div>

    <div class="tabs">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        class="tab-btn"
        :class="{ active: activeTab === tab.id }"
        @click="activeTab = tab.id"
      >{{ tab.label }}</button>
    </div>

    <!-- Open WebUI -->
    <div v-if="activeTab === 'openwebui'" class="tab-content">
      <div class="tab-header">
        <h2>Open WebUI</h2>
        <span class="provider-badge" :class="provider">{{ provider === 'openai' ? 'OpenAI' : 'Ollama' }}</span>
      </div>
      <p>Open WebUI supports custom OpenAI-compatible endpoints via its admin settings.</p>
      <ol class="steps">
        <li>
          Open the admin panel → <em>Settings → Connections → OpenAI API</em>.
        </li>
        <li>
          Set the <strong>API URL</strong> to:
          <div class="code-block">
            <code>{{ activeURL }}</code>
            <button class="copy-btn" @click="copy(activeURL, 'owui-url')">{{ copied === 'owui-url' ? 'Copied!' : 'Copy' }}</button>
          </div>
        </li>
        <li>Set the <strong>API Key</strong> to your Personal Access Token.</li>
        <li>Click <em>Save</em> and refresh the model list.</li>
      </ol>
      <div v-if="provider === 'ollama'" class="tip">
        <strong>Ollama:</strong> Open WebUI also has native Ollama support — using the AI Bridge Ollama endpoint lets you apply IP restrictions and usage tracking on top.
      </div>
      <div v-else class="tip">
        <strong>Tip:</strong> Add a second connection with <code>{{ ollamaURL }}</code> to have both OpenAI and Ollama models available simultaneously.
      </div>
    </div>

    <!-- OpenCode -->
    <div v-if="activeTab === 'opencode'" class="tab-content">
      <div class="tab-header">
        <h2>OpenCode</h2>
        <span class="provider-badge" :class="provider">{{ provider === 'openai' ? 'OpenAI' : 'Ollama' }}</span>
      </div>
      <p>OpenCode reads its provider configuration from <code>~/.config/opencode/config.json</code>.</p>
      <ol class="steps">
        <li>
          Generate a PAT in AI Bridge (<RouterLink to="/tokens">Tokens</RouterLink>) and copy it.
        </li>
        <li>
          Add the AI Bridge provider to your config:
          <div class="code-block">
            <pre>{{ opencodeSnippetDisplay }}</pre>
            <button class="copy-btn top" @click="copy(opencodeSnippet, 'opencode-cfg')">{{ copied === 'opencode-cfg' ? 'Copied!' : 'Copy' }}</button>
          </div>
        </li>
        <li>Restart OpenCode — the provider will appear in the model picker.</li>
      </ol>
      <div class="tip">
        <strong>Model name:</strong> Set the model field to <code>{{ activeModel }}</code> or any model available through this provider.
      </div>
    </div>

    <!-- n8N -->
    <div v-if="activeTab === 'n8n'" class="tab-content">
      <div class="tab-header">
        <h2>n8N</h2>
        <span class="provider-badge" :class="provider">{{ provider === 'openai' ? 'OpenAI' : 'Ollama' }}</span>
      </div>
      <p>Use n8N's built-in <em>OpenAI</em> credential with a custom base URL to route requests through AI Bridge.</p>
      <ol class="steps">
        <li>In n8N, go to <em>Credentials → New → OpenAI</em>.</li>
        <li>Set <strong>API Key</strong> to your Personal Access Token.</li>
        <li>
          Expand <strong>Base URL</strong> and enter:
          <div class="code-block">
            <code>{{ activeURL }}</code>
            <button class="copy-btn" @click="copy(activeURL, 'n8n-url')">{{ copied === 'n8n-url' ? 'Copied!' : 'Copy' }}</button>
          </div>
        </li>
        <li>Save and use it in any <em>OpenAI Chat Model</em> or <em>OpenAI</em> node.</li>
        <li>Set the model to <code>{{ activeModel }}</code>.</li>
      </ol>
      <div class="tip">
        <strong>HTTP Request node:</strong> Alternatively, POST to <code>{{ activeURL }}/chat/completions</code> with <code>Authorization: Bearer YOUR_PAT</code>.
      </div>
    </div>

    <!-- Python -->
    <div v-if="activeTab === 'python'" class="tab-content">
      <div class="tab-header">
        <h2>Python</h2>
        <span class="provider-badge" :class="provider">{{ provider === 'openai' ? 'OpenAI' : 'Ollama' }}</span>
      </div>
      <p>Use the <code>openai</code> Python SDK — just point <code>base_url</code> at AI Bridge{{ provider === 'ollama' ? ' (Ollama is OpenAI-compatible)' : '' }}.</p>
      <ol class="steps">
        <li>
          Install the SDK:
          <div class="code-block">
            <code>pip install openai</code>
            <button class="copy-btn" @click="copy('pip install openai', 'py-install')">{{ copied === 'py-install' ? 'Copied!' : 'Copy' }}</button>
          </div>
        </li>
        <li>
          Use AI Bridge as the backend:
          <div class="code-block">
            <pre>{{ pythonSnippetDisplay }}</pre>
            <button class="copy-btn top" @click="copy(pythonSnippet, 'py-code')">{{ copied === 'py-code' ? 'Copied!' : 'Copy' }}</button>
          </div>
        </li>
      </ol>
      <div class="tip">
        <strong>Streaming:</strong> Pass <code>stream=True</code> to <code>chat.completions.create</code> — AI Bridge forwards the SSE stream transparently.
      </div>
    </div>

    <!-- cURL -->
    <div v-if="activeTab === 'curl'" class="tab-content">
      <div class="tab-header">
        <h2>cURL</h2>
        <span class="provider-badge" :class="provider">{{ provider === 'openai' ? 'OpenAI' : 'Ollama' }}</span>
      </div>
      <p>Test directly from your terminal. Replace <code>YOUR_PAT_HERE</code> with a token from <RouterLink to="/tokens">Tokens</RouterLink>.</p>
      <ol class="steps">
        <li>
          <strong>List available models</strong>:
          <div class="code-block">
            <pre>{{ curlModelsDisplay }}</pre>
            <button class="copy-btn top" @click="copy(curlModelsSnippet, 'curl-models')">{{ copied === 'curl-models' ? 'Copied!' : 'Copy' }}</button>
          </div>
        </li>
        <li>
          <strong>Chat completion</strong>:
          <div class="code-block">
            <pre>{{ curlChatDisplay }}</pre>
            <button class="copy-btn top" @click="copy(curlChatSnippet, 'curl-chat')">{{ copied === 'curl-chat' ? 'Copied!' : 'Copy' }}</button>
          </div>
        </li>
        <li>
          <strong>Streaming response</strong>:
          <div class="code-block">
            <pre>{{ curlStreamDisplay }}</pre>
            <button class="copy-btn top" @click="copy(curlStreamSnippet, 'curl-stream')">{{ copied === 'curl-stream' ? 'Copied!' : 'Copy' }}</button>
          </div>
        </li>
      </ol>
      <div class="tip">
        <strong>Pretty-print:</strong> Pipe through <code>jq '.choices[0].message.content'</code> to extract just the reply text.
      </div>
    </div>

    </template><!-- end v-if availableProviders -->
  </div>
</template>

<style scoped>
.help-page { display: flex; flex-direction: column; gap: 1.5rem; }

.page-header {
  display: flex; align-items: flex-start;
  justify-content: space-between; gap: 1rem; flex-wrap: wrap;
}
.page-header h1 { font-size: 1.75rem; font-weight: 700; margin: 0 0 0.25rem; }
.subtitle { color: #64748b; font-size: 0.95rem; margin: 0; }

/* provider selector */
.provider-selector { display: flex; align-items: center; gap: 0.6rem; flex-shrink: 0; }
.provider-label { font-size: 0.8rem; font-weight: 600; color: #64748b; }
.provider-toggle { display: flex; border: 1px solid #e2e8f0; border-radius: 8px; overflow: hidden; }
.provider-btn {
  padding: 0.4rem 1rem; border: none; cursor: pointer;
  font-size: 0.85rem; font-weight: 600; color: #64748b;
  background: white; transition: background 0.15s, color 0.15s;
}
.provider-btn:not(:last-child) { border-right: 1px solid #e2e8f0; }
.provider-btn.active { color: white; }
.provider-btn:not(.active):hover { background: #f8fafc; color: #1e293b; }

/* tab header */
.tab-header { display: flex; align-items: center; gap: 0.75rem; }
.tab-header h2 { font-size: 1.2rem; font-weight: 700; margin: 0; }

.provider-badge {
  padding: 0.15rem 0.65rem; border-radius: 999px;
  font-size: 0.75rem; font-weight: 700;
}
.provider-badge.openai { background: #d1fae5; color: #065f46; }
.provider-badge.ollama { background: #ede9fe; color: #5b21b6; }

/* endpoint banner */
.info-banner {
  background: #f8fafc; border: 1px solid #e2e8f0;
  border-radius: 10px; padding: 1rem 1.25rem;
  display: flex; flex-direction: column; gap: 0.6rem;
}
.info-row { display: flex; align-items: center; gap: 0.75rem; flex-wrap: wrap; }
.info-label { font-size: 0.8rem; font-weight: 600; color: #64748b; min-width: 130px; }
.info-value {
  font-size: 0.85rem; background: #fff; border: 1px solid #e2e8f0;
  border-radius: 5px; padding: 0.2rem 0.5rem; font-family: monospace;
}
.info-value.muted {
  background: none; border: none; color: #475569;
  font-family: inherit; padding: 0;
}
.info-value.muted a { color: #3b82f6; text-decoration: none; }
.info-value.muted a:hover { text-decoration: underline; }
/* no provider */
.no-provider {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  gap: 0.75rem; padding: 3rem 2rem; text-align: center;
  background: #fff; border: 1px solid #e2e8f0; border-radius: 12px;
}
.no-provider-icon { font-size: 2.5rem; }
.no-provider h2 { font-size: 1.15rem; font-weight: 700; color: #1e293b; margin: 0; }
.no-provider p { color: #64748b; font-size: 0.92rem; margin: 0; line-height: 1.6; }
.no-provider code { background: #f1f5f9; border-radius: 4px; padding: 0.1rem 0.35rem; font-family: monospace; color: #0f172a; }

.model-select {
  font-size: 0.85rem; background: #fff; border: 1px solid #e2e8f0;
  border-radius: 5px; padding: 0.2rem 0.5rem; font-family: monospace;
  color: #1e293b; cursor: pointer; outline: none;
  min-width: 220px;
}
.model-select:focus { border-color: #3b82f6; }

/* tabs */
.tabs { display: flex; gap: 0.5rem; border-bottom: 1px solid #e2e8f0; }
.tab-btn {
  padding: 0.5rem 1.1rem; border: none; background: none;
  cursor: pointer; font-size: 0.9rem; font-weight: 500; color: #64748b;
  border-bottom: 2px solid transparent; margin-bottom: -1px;
  transition: color 0.15s, border-color 0.15s;
}
.tab-btn:hover { color: #1e293b; }
.tab-btn.active { color: #3b82f6; border-bottom-color: #3b82f6; }

/* tab content */
.tab-content {
  background: #fff; border: 1px solid #e2e8f0; border-radius: 10px;
  padding: 1.5rem; display: flex; flex-direction: column; gap: 1rem;
}
.tab-content p { color: #475569; font-size: 0.93rem; margin: 0; }

/* steps */
.steps { padding-left: 1.25rem; display: flex; flex-direction: column; gap: 0.8rem; margin: 0; }
.steps li { color: #334155; font-size: 0.92rem; line-height: 1.5; }
.steps strong { color: #1e293b; }

/* code blocks */
.code-block {
  position: relative; margin-top: 0.4rem; background: #0f172a;
  border-radius: 8px; padding: 0.75rem 1rem;
  display: flex; align-items: flex-start; gap: 0.75rem;
}
.code-block code,
.code-block pre {
  font-family: 'Fira Code', 'Cascadia Code', monospace;
  font-size: 0.82rem; color: #e2e8f0; margin: 0;
  white-space: pre; overflow-x: auto; flex: 1;
}
.copy-btn {
  flex-shrink: 0; padding: 0.2rem 0.6rem; font-size: 0.75rem; font-weight: 600;
  border: 1px solid #334155; background: #1e293b; color: #94a3b8;
  border-radius: 5px; cursor: pointer; transition: background 0.15s, color 0.15s;
  align-self: center;
}
.copy-btn.top { align-self: flex-start; }
.copy-btn:hover { background: #334155; color: #f1f5f9; }
.info-row .copy-btn { background: #fff; border-color: #e2e8f0; color: #64748b; }
.info-row .copy-btn:hover { background: #f1f5f9; color: #1e293b; }

/* tip */
.tip {
  background: #eff6ff; border: 1px solid #bfdbfe; border-radius: 8px;
  padding: 0.75rem 1rem; font-size: 0.87rem; color: #1e40af;
}
.tip strong { font-weight: 700; }
.tip code { background: #dbeafe; border-radius: 4px; padding: 0.1rem 0.3rem; font-family: monospace; }
</style>
