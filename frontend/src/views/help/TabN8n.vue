<script setup lang="ts">
import { ref } from 'vue'

defineProps<{
  activeUrl:   string
  activeModel: string
  provider:    'openai' | 'ollama' | 'anthropic'
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
      <div class="header-left">
        <h2>N8N</h2>
        <span class="provider-badge" :class="provider">{{ provider === 'openai' ? 'OpenAI' : 'Ollama' }}</span>
      </div>
    </div>

    <p class="tab-desc">
      Use n8N's built-in <em>OpenAI</em> credential with a custom base URL to route requests through AI Bridge.
    </p>

    <div class="steps">
      <div class="step">
        <div class="step-num">1</div>
        <div class="step-body">
          <p>In n8N, go to <em>Credentials → New → OpenAI</em>.</p>
        </div>
      </div>

      <div class="step">
        <div class="step-num">2</div>
        <div class="step-body">
          <p>Set <strong>API Key</strong> to your Personal Access Token.</p>
        </div>
      </div>

      <div class="step">
        <div class="step-num">3</div>
        <div class="step-body">
          <p>Expand <strong>Base URL</strong> and enter:</p>
          <div class="code-block">
            <div class="code-header">
              <span class="code-lang">url</span>
              <button class="code-copy" :class="{ copied: copied === 'n8n-url' }" @click="copy(activeUrl, 'n8n-url')">
                {{ copied === 'n8n-url' ? '✓ Copied!' : 'Copy' }}
              </button>
            </div>
            <code>{{ activeUrl }}</code>
          </div>
        </div>
      </div>

      <div class="step">
        <div class="step-num">4</div>
        <div class="step-body">
          <p>Save and use it in any <em>OpenAI Chat Model</em> or <em>OpenAI</em> node.</p>
        </div>
      </div>

      <div class="step">
        <div class="step-num">5</div>
        <div class="step-body">
          <p>Set the model to <code>{{ activeModel }}</code>.</p>
        </div>
      </div>
    </div>

    <div class="tip">
      <div class="tip-icon">💡</div>
      <div>
        <strong>HTTP Request node:</strong> Alternatively, POST to
        <code>{{ activeUrl }}/chat/completions</code> with <code>Authorization: Bearer YOUR_PAT</code>.
      </div>
    </div>
  </div>
</template>

<style scoped src="./tab-shared.css" />
