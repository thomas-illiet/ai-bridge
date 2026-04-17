<script setup lang="ts">
import { ref, watch, nextTick, computed } from 'vue'
import { useChatStore, newId } from '@/stores/chat'
import { getValidToken } from '@/services/keycloak'
import { getModels, getStatus } from '@/services/api'
import ChatMessage from './ChatMessage.vue'

const store = useChatStore()
const open  = ref(false)
const input = ref('')
const messagesEl = ref<HTMLElement | null>(null)

// ── provider / model setup ──────────────────────────────────────────────────
const allProviders = [
  { id: 'openai' as const, label: 'OpenAI' },
  { id: 'ollama' as const, label: 'Ollama' },
]
const availableProviders = ref([...allProviders])
const modelList          = ref<string[]>([])
const modelLoading       = ref(false)
const initialized        = ref(false)

async function initProviders() {
  if (initialized.value) return
  initialized.value = true
  try {
    const res = await getStatus()
    const enabled = new Set(res.data.services.filter(s => s.status !== 'disabled').map(s => s.name))
    const filtered = allProviders.filter(p => enabled.has(p.id))
    if (filtered.length > 0) availableProviders.value = filtered
    if (!store.provider || !availableProviders.value.find(p => p.id === store.provider)) {
      store.provider = availableProviders.value[0].id
    }
  } catch { /* keep allProviders */ }
  await loadModels()
}

async function loadModels() {
  modelLoading.value = true
  modelList.value = []
  try {
    const res = await getModels(store.provider)
    modelList.value = res.data.models ?? []
    if (modelList.value.length > 0) store.model = modelList.value[0]
  } catch {
    store.model = store.provider === 'openai' ? 'gpt-4o' : 'llama3.2'
  } finally {
    modelLoading.value = false
  }
}

watch(() => store.provider, loadModels)

function toggleOpen() {
  open.value = !open.value
  if (open.value) initProviders()
}

// ── auto-scroll ─────────────────────────────────────────────────────────────
function scrollToBottom() {
  if (messagesEl.value) messagesEl.value.scrollTop = messagesEl.value.scrollHeight
}
async function scrollToBottomNextTick() {
  await nextTick()
  scrollToBottom()
}
watch(() => store.messages.length, scrollToBottomNextTick)

let scrollRafId = 0
watch(() => store.streaming, (streaming) => {
  if (streaming) {
    const tick = () => {
      scrollToBottom()
      if (store.streaming) scrollRafId = requestAnimationFrame(tick)
    }
    scrollRafId = requestAnimationFrame(tick)
  } else {
    cancelAnimationFrame(scrollRafId)
    scrollToBottom()
  }
})

// ── streaming send ───────────────────────────────────────────────────────────
async function send() {
  const text = input.value.trim()
  if (!text || store.streaming) return
  input.value = ''

  store.addMessage({ id: newId(), role: 'user', content: text })
  store.addMessage({ id: newId(), role: 'assistant' as const, content: '' })
  store.streaming = true

  // Access through the store's reactive proxy so Vue tracks mutations
  const liveMsg = store.messages[store.messages.length - 1]

  const history = store.messages
    .filter(m => m.role !== 'error')
    .slice(0, -1)
    .map(m => ({ role: m.role, content: m.content }))

  try {
    const token = await getValidToken()
    const base = import.meta.env.VITE_API_BASE_URL ?? ''
    const url = store.provider === 'openai'
      ? `${base}/openai/v1/chat/completions`
      : `${base}/ollama/v1/chat/completions`

    const res = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: JSON.stringify({
        model:    store.model,
        messages: history,
        stream:   true,
      }),
    })

    if (!res.ok) {
      liveMsg.content = `Error ${res.status}: ${res.statusText}`
      liveMsg.role = 'error' as any
      return
    }

    const reader  = res.body!.getReader()
    const decoder = new TextDecoder()

    while (true) {
      const { done, value } = await reader.read()
      if (done) break
      for (const line of decoder.decode(value).split('\n')) {
        if (!line.startsWith('data: ')) continue
        const data = line.slice(6).trim()
        if (data === '[DONE]') continue
        try {
          const delta = JSON.parse(data).choices?.[0]?.delta?.content
          if (delta) liveMsg.content += delta
        } catch { /* malformed chunk — skip */ }
      }
    }
  } catch (e: any) {
    liveMsg.content = e?.message ?? 'An error occurred.'
    liveMsg.role = 'error' as any
  } finally {
    store.streaming = false
  }
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    send()
  }
}

const isLastStreaming = computed(() =>
  store.streaming &&
  store.messages.length > 0 &&
  store.messages[store.messages.length - 1].role === 'assistant'
)
</script>

<template>
  <!-- floating action button -->
  <button class="fab" :class="{ active: open }" @click="toggleOpen" aria-label="Open chat">
    <svg v-if="!open" xmlns="http://www.w3.org/2000/svg" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
    </svg>
    <svg v-else xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
      <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
    </svg>
  </button>

  <!-- chat panel -->
  <Transition name="panel">
    <div v-if="open" class="chat-panel">
      <!-- header -->
      <div class="chat-header">
        <span class="chat-title">AI Chat</span>

        <div class="header-controls">
          <!-- engine dropdown -->
          <select
            v-model="store.provider"
            class="engine-select"
            :disabled="availableProviders.length <= 1"
          >
            <option v-for="p in availableProviders" :key="p.id" :value="p.id">{{ p.label }}</option>
          </select>

          <!-- model dropdown -->
          <select
            v-model="store.model"
            class="model-select"
            :disabled="modelLoading || modelList.length === 0"
          >
            <option v-if="modelLoading" value="">Loading…</option>
            <option v-for="m in modelList" :key="m" :value="m">{{ m }}</option>
          </select>

          <!-- clear -->
          <button class="icon-btn" title="Clear chat" @click="store.clear()">
            <svg xmlns="http://www.w3.org/2000/svg" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14H6L5 6"/><path d="M10 11v6"/><path d="M14 11v6"/><path d="M9 6V4h6v2"/>
            </svg>
          </button>

          <!-- close -->
          <button class="icon-btn" title="Close" @click="open = false">
            <svg xmlns="http://www.w3.org/2000/svg" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
              <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
      </div>

      <!-- messages -->
      <div ref="messagesEl" class="chat-messages">
        <div v-if="store.messages.length === 0" class="empty-state">
          Start a conversation — your history is kept until you close the page.
        </div>
        <ChatMessage
          v-for="msg in store.messages"
          :key="msg.id"
          :message="msg"
          :is-streaming="isLastStreaming && msg.id === store.messages[store.messages.length - 1].id"
        />
      </div>

      <!-- input -->
      <div class="chat-input-area">
        <textarea
          v-model="input"
          class="chat-input"
          placeholder="Message… (Shift+Enter for newline)"
          rows="1"
          :disabled="store.streaming"
          @keydown="handleKeydown"
        />
        <button
          class="send-btn"
          :disabled="store.streaming || !input.trim()"
          @click="send"
        >
          <svg v-if="!store.streaming" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/>
          </svg>
          <span v-else class="send-spinner" />
        </button>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
/* ── FAB ── */
.fab {
  position: fixed;
  bottom: 1.75rem;
  right: 1.75rem;
  z-index: 1000;
  width: 52px;
  height: 52px;
  border-radius: 50%;
  border: none;
  background: #3b82f6;
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 16px rgba(59,130,246,0.45);
  transition: background 0.15s, transform 0.15s, box-shadow 0.15s;
}
.fab:hover { background: #2563eb; transform: scale(1.07); box-shadow: 0 6px 20px rgba(59,130,246,0.5); }
.fab.active { background: #1d4ed8; }

/* ── panel transition ── */
.panel-enter-active,
.panel-leave-active { transition: opacity 0.18s ease, transform 0.18s ease; }
.panel-enter-from,
.panel-leave-to { opacity: 0; transform: translateY(12px) scale(0.97); }

/* ── panel ── */
.chat-panel {
  position: fixed;
  bottom: 5.5rem;
  right: 1.75rem;
  z-index: 999;
  width: 400px;
  height: 560px;
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  box-shadow: 0 12px 40px rgba(0,0,0,0.14);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* ── header ── */
.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.7rem 0.9rem;
  border-bottom: 1px solid #f1f5f9;
  background: #f8fafc;
  gap: 0.5rem;
  flex-shrink: 0;
}
.chat-title { font-size: 0.9rem; font-weight: 700; color: #1e293b; white-space: nowrap; }
.header-controls { display: flex; align-items: center; gap: 0.4rem; flex-wrap: wrap; justify-content: flex-end; }

.engine-select,
.model-select {
  font-size: 0.72rem;
  border: 1px solid #e2e8f0;
  border-radius: 5px;
  padding: 0.22rem 0.4rem;
  color: #374151;
  background: white;
  cursor: pointer;
  outline: none;
  max-width: 120px;
  transition: border-color 0.12s;
}
.engine-select:focus,
.model-select:focus { border-color: #3b82f6; }
.engine-select:disabled,
.model-select:disabled { opacity: 0.6; cursor: not-allowed; }

.icon-btn {
  width: 26px;
  height: 26px;
  border: none;
  background: none;
  color: #94a3b8;
  cursor: pointer;
  border-radius: 5px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.12s, color 0.12s;
}
.icon-btn:hover { background: #f1f5f9; color: #475569; }

/* ── messages ── */
.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 0.85rem;
  display: flex;
  flex-direction: column;
  scroll-behavior: smooth;
}
.empty-state {
  margin: auto;
  text-align: center;
  color: #94a3b8;
  font-size: 0.82rem;
  line-height: 1.6;
  padding: 1.5rem;
}

/* ── input area ── */
.chat-input-area {
  display: flex;
  align-items: flex-end;
  gap: 0.5rem;
  padding: 0.65rem 0.75rem;
  border-top: 1px solid #f1f5f9;
  flex-shrink: 0;
}
.chat-input {
  flex: 1;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 0.5rem 0.75rem;
  font-size: 0.88rem;
  font-family: inherit;
  resize: none;
  outline: none;
  line-height: 1.45;
  max-height: 120px;
  overflow-y: auto;
  transition: border-color 0.12s;
}
.chat-input:focus { border-color: #3b82f6; box-shadow: 0 0 0 2px #bfdbfe; }
.chat-input:disabled { background: #f8fafc; }

.send-btn {
  width: 34px;
  height: 34px;
  border-radius: 8px;
  border: none;
  background: #3b82f6;
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: background 0.12s, opacity 0.12s;
}
.send-btn:hover:not(:disabled) { background: #2563eb; }
.send-btn:disabled { opacity: 0.45; cursor: not-allowed; }

/* send spinner */
@keyframes spin { to { transform: rotate(360deg); } }
.send-spinner {
  display: block;
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255,255,255,0.4);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

/* ── responsive ── */
@media (max-width: 480px) {
  .chat-panel { width: calc(100vw - 2rem); right: 1rem; left: 1rem; }
}
</style>
