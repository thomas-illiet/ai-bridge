<script setup lang="ts">
import { ref, computed } from 'vue'

const props = defineProps<{
  activeUrl:   string
  activeModel: string
  provider:    'openai' | 'ollama' | 'anthropic'
}>()

const snippet = computed(() =>
`from openai import OpenAI

client = OpenAI(
    api_key="YOUR_PAT_HERE",
    base_url="${props.activeUrl}",
)

response = client.chat.completions.create(
    model="${props.activeModel}",
    messages=[{"role": "user", "content": "Hello!"}],
)
print(response.choices[0].message.content)`)

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
      <div class="header-left">
        <h2>Python</h2>
        <span class="provider-badge" :class="provider">{{ provider === 'openai' ? 'OpenAI' : 'Ollama' }}</span>
      </div>
    </div>

    <p class="tab-desc">
      Use the <code>openai</code> Python SDK — just point <code>base_url</code> at AI Bridge{{ provider === 'ollama' ? ' (Ollama is OpenAI-compatible)' : '' }}.
    </p>

    <div class="steps">
      <div class="step">
        <div class="step-num">1</div>
        <div class="step-body">
          <p>Install the SDK:</p>
          <div class="code-block">
            <div class="code-header">
              <span class="code-lang">bash</span>
              <button class="code-copy" :class="{ copied: copied === 'install' }" @click="copy('pip install openai', 'install')">
                {{ copied === 'install' ? '✓ Copied!' : 'Copy' }}
              </button>
            </div>
            <code>pip install openai</code>
          </div>
        </div>
      </div>

      <div class="step">
        <div class="step-num">2</div>
        <div class="step-body">
          <p>Use AI Bridge as the backend:</p>
          <div class="code-block">
            <div class="code-header">
              <span class="code-lang">python</span>
              <button class="code-copy" :class="{ copied: copied === 'py-code' }" @click="copy(snippet, 'py-code')">
                {{ copied === 'py-code' ? '✓ Copied!' : 'Copy' }}
              </button>
            </div>
            <pre>{{ snippet }}</pre>
          </div>
        </div>
      </div>
    </div>

    <div class="tip">
      <div class="tip-icon">💡</div>
      <div><strong>Streaming:</strong> Pass <code>stream=True</code> to <code>chat.completions.create</code> — AI Bridge forwards the SSE stream transparently.</div>
    </div>
  </div>
</template>

<style scoped src="./tab-shared.css" />
