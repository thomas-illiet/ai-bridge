<script setup lang="ts">
import { ref, computed } from 'vue'

const props = defineProps<{
  activeUrl:  string
  activeModel: string
}>()

const anthropicBaseUrl = computed(() => props.activeUrl.replace(/\/v1$/, ''))

const envSnippet = computed(() =>
`export ANTHROPIC_BASE_URL="${anthropicBaseUrl.value}"
export ANTHROPIC_AUTH_TOKEN="YOUR_PAT_HERE"`)

const settingsJson = computed(() =>
JSON.stringify({
  env: {
    ANTHROPIC_AUTH_TOKEN: 'YOUR_PAT_HERE',
    ANTHROPIC_BASE_URL: anthropicBaseUrl.value,
    API_TIMEOUT_MS: '3000000',
  },
}, null, 4))

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
      <h2>Claude</h2>
      <span class="provider-badge anthropic">Anthropic</span>
    </div>
    <p>
      Configure the Claude client to route through AI Bridge by overriding
      <code>ANTHROPIC_BASE_URL</code> and <code>ANTHROPIC_AUTH_TOKEN</code>.
      Use a Personal Access Token from <RouterLink to="/tokens">Tokens</RouterLink> as the auth token.
    </p>

    <h3>Environment variables</h3>
    <p>Export these in your shell or <code>.env</code> file before running any Claude-based tool.</p>
    <div class="code-block">
      <pre>{{ envSnippet }}</pre>
      <button class="copy-btn top" @click="copy(envSnippet, 'env')">{{ copied === 'env' ? 'Copied!' : 'Copy' }}</button>
    </div>

    <h3>Claude Code — <code>settings.json</code></h3>
    <p>
      Add the following block to your Claude Code
      <code>~/.claude/settings.json</code> (or the project-level <code>.claude/settings.json</code>).
    </p>
    <div class="code-block">
      <pre>{{ settingsJson }}</pre>
      <button class="copy-btn top" @click="copy(settingsJson, 'settings')">{{ copied === 'settings' ? 'Copied!' : 'Copy' }}</button>
    </div>

    <div class="tip">
      <strong>API_TIMEOUT_MS</strong> is optional but recommended for long-running requests.
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
.tab-header { display: flex; align-items: center; gap: 0.75rem; }
.tab-header h2 { font-size: 1.2rem; font-weight: 700; margin: 0; }
.provider-badge { padding: 0.15rem 0.65rem; border-radius: 999px; font-size: 0.75rem; font-weight: 700; }
.provider-badge.anthropic { background: #fef3c7; color: #92400e; }
.tab-content p { color: #475569; font-size: 0.93rem; margin: 0; }
.tab-content p code,
.tab-content h3 code {
  background: #f1f5f9;
  border-radius: 4px;
  padding: 0.1rem 0.35rem;
  font-family: monospace;
  color: #0f172a;
  font-size: 0.85rem;
}
.tab-content p a { color: #3b82f6; text-decoration: none; }
.tab-content p a:hover { text-decoration: underline; }
h3 { font-size: 0.95rem; font-weight: 700; color: #1e293b; margin: 0; }
.code-block {
  position: relative;
  background: #0f172a;
  border-radius: 8px;
  padding: 0.75rem 1rem;
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
}
.code-block pre {
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
.copy-btn.top { align-self: flex-start; }
.copy-btn:hover { background: #334155; color: #f1f5f9; }
.tip {
  background: #fffbeb;
  border: 1px solid #fde68a;
  border-radius: 8px;
  padding: 0.75rem 1rem;
  font-size: 0.87rem;
  color: #92400e;
}
.tip strong { font-weight: 700; }
</style>
