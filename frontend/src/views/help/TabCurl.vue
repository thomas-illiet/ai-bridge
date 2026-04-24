<script setup lang="ts">
import { ref, computed } from 'vue'

const props = defineProps<{
  activeUrl:   string
  activeModel: string
  provider:    'openai' | 'ollama' | 'anthropic'
}>()

const curlModels = computed(() =>
`curl ${props.activeUrl}/models \\
  -H "Authorization: Bearer YOUR_PAT_HERE"`)

const curlChat = computed(() =>
`curl ${props.activeUrl}/chat/completions \\
  -H "Authorization: Bearer YOUR_PAT_HERE" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "${props.activeModel}",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'`)

const curlStream = computed(() =>
`curl ${props.activeUrl}/chat/completions \\
  -H "Authorization: Bearer YOUR_PAT_HERE" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "${props.activeModel}",
    "stream": true,
    "messages": [{"role": "user", "content": "Hello!"}]
  }'`)

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
        <h2>cURL</h2>
        <span class="provider-badge" :class="provider">{{ provider === 'openai' ? 'OpenAI' : 'Ollama' }}</span>
      </div>
    </div>

    <p class="tab-desc">
      Test directly from your terminal. Replace <code>YOUR_PAT_HERE</code> with a token from
      <RouterLink to="/tokens">Tokens</RouterLink>.
    </p>

    <div class="steps">
      <div class="step">
        <div class="step-num">1</div>
        <div class="step-body">
          <p><strong>List available models</strong></p>
          <div class="code-block">
            <div class="code-header">
              <span class="code-lang">bash</span>
              <button class="code-copy" :class="{ copied: copied === 'models' }" @click="copy(curlModels, 'models')">
                {{ copied === 'models' ? '✓ Copied!' : 'Copy' }}
              </button>
            </div>
            <pre>{{ curlModels }}</pre>
          </div>
        </div>
      </div>

      <div class="step">
        <div class="step-num">2</div>
        <div class="step-body">
          <p><strong>Chat completion</strong></p>
          <div class="code-block">
            <div class="code-header">
              <span class="code-lang">bash</span>
              <button class="code-copy" :class="{ copied: copied === 'chat' }" @click="copy(curlChat, 'chat')">
                {{ copied === 'chat' ? '✓ Copied!' : 'Copy' }}
              </button>
            </div>
            <pre>{{ curlChat }}</pre>
          </div>
        </div>
      </div>

      <div class="step">
        <div class="step-num">3</div>
        <div class="step-body">
          <p><strong>Streaming response</strong></p>
          <div class="code-block">
            <div class="code-header">
              <span class="code-lang">bash</span>
              <button class="code-copy" :class="{ copied: copied === 'stream' }" @click="copy(curlStream, 'stream')">
                {{ copied === 'stream' ? '✓ Copied!' : 'Copy' }}
              </button>
            </div>
            <pre>{{ curlStream }}</pre>
          </div>
        </div>
      </div>
    </div>

    <div class="tip">
      <div class="tip-icon">💡</div>
      <div><strong>Pretty-print:</strong> Pipe through <code>jq '.choices[0].message.content'</code> to extract just the reply text.</div>
    </div>
  </div>
</template>

<style scoped src="./tab-shared.css" />
