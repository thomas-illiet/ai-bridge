<script setup lang="ts">
import { ref } from 'vue'

defineProps<{
  activeUrl: string
  provider:  'openai' | 'ollama'
}>()

const copied = ref<string | null>(null)
function copy(text: string, key: string) {
  navigator.clipboard.writeText(text)
  copied.value = key
  setTimeout(() => { if (copied.value === key) copied.value = null }, 1800)
}
</script>

<template>
  <div class="tab-content">
    <div class="tab-header">
      <h2>Open WebUI</h2>
      <span class="provider-badge" :class="provider">{{ provider === 'openai' ? 'OpenAI' : 'Ollama' }}</span>
    </div>
    <p>Open WebUI supports custom OpenAI-compatible endpoints via its admin settings.</p>
    <ol class="steps">
      <li>Open the admin panel → <em>Settings → Connections → OpenAI API</em>.</li>
      <li>
        Set the <strong>API URL</strong> to:
        <div class="code-block">
          <code>{{ activeUrl }}</code>
          <button class="copy-btn" @click="copy(activeUrl, 'owui-url')">{{ copied === 'owui-url' ? 'Copied!' : 'Copy' }}</button>
        </div>
      </li>
      <li>Set the <strong>API Key</strong> to your Personal Access Token.</li>
      <li>Click <em>Save</em> and refresh the model list.</li>
    </ol>
    <div v-if="provider === 'ollama'" class="tip">
      <strong>Ollama:</strong> Open WebUI also has native Ollama support — using the AI Bridge endpoint applies IP restrictions and usage tracking on top.
    </div>
    <div v-else class="tip">
      <strong>Tip:</strong> Add additional connections in Open WebUI to expose models from other AI Bridge providers simultaneously.
    </div>
  </div>
</template>

<style scoped>
.tab-content {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.tab-content p { color: #475569; font-size: 0.93rem; margin: 0; }
.tab-header { display: flex; align-items: center; gap: 0.75rem; }
.tab-header h2 { font-size: 1.2rem; font-weight: 700; margin: 0; }
.provider-badge { padding: 0.15rem 0.65rem; border-radius: 999px; font-size: 0.75rem; font-weight: 700; }
.provider-badge.openai { background: #d1fae5; color: #065f46; }
.provider-badge.ollama { background: #ede9fe; color: #5b21b6; }
.steps { padding-left: 1.25rem; display: flex; flex-direction: column; gap: 0.8rem; margin: 0; }
.steps li { color: #334155; font-size: 0.92rem; line-height: 1.5; }
.steps strong { color: #1e293b; }
.code-block {
  position: relative;
  margin-top: 0.4rem;
  background: #0f172a;
  border-radius: 8px;
  padding: 0.75rem 1rem;
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
}
.code-block code {
  font-family: 'Fira Code', 'Cascadia Code', monospace;
  font-size: 0.82rem;
  color: #e2e8f0;
  margin: 0;
  white-space: pre;
  overflow-x: auto;
  flex: 1;
}
.copy-btn {
  flex-shrink: 0;
  padding: 0.2rem 0.6rem;
  font-size: 0.75rem;
  font-weight: 600;
  border: 1px solid #334155;
  background: #1e293b;
  color: #94a3b8;
  border-radius: 5px;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
  align-self: center;
}
.copy-btn:hover { background: #334155; color: #f1f5f9; }
.tip {
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  border-radius: 8px;
  padding: 0.75rem 1rem;
  font-size: 0.87rem;
  color: #1e40af;
}
.tip strong { font-weight: 700; }
.tip code { background: #dbeafe; border-radius: 4px; padding: 0.1rem 0.3rem; font-family: monospace; }
</style>
