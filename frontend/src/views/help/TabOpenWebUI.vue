<script setup lang="ts">
import { ref } from 'vue'

defineProps<{
  activeUrl: string
  provider:  'openai' | 'ollama' | 'anthropic'
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
        <h2>Open WebUI</h2>
        <span class="provider-badge" :class="provider">{{ provider === 'openai' ? 'OpenAI' : 'Ollama' }}</span>
      </div>
    </div>

    <p class="tab-desc">Open WebUI supports custom OpenAI-compatible endpoints via its admin settings.</p>

    <div class="steps">
      <div class="step">
        <div class="step-num">1</div>
        <div class="step-body">
          <p>Open the admin panel → <em>Settings → Connections → OpenAI API</em>.</p>
        </div>
      </div>

      <div class="step">
        <div class="step-num">2</div>
        <div class="step-body">
          <p>Set the <strong>API URL</strong> to:</p>
          <div class="code-block">
            <div class="code-header">
              <span class="code-lang">url</span>
              <button class="code-copy" :class="{ copied: copied === 'owui-url' }" @click="copy(activeUrl, 'owui-url')">
                {{ copied === 'owui-url' ? '✓ Copied!' : 'Copy' }}
              </button>
            </div>
            <code>{{ activeUrl }}</code>
          </div>
        </div>
      </div>

      <div class="step">
        <div class="step-num">3</div>
        <div class="step-body">
          <p>Set the <strong>API Key</strong> to your Personal Access Token.</p>
        </div>
      </div>

      <div class="step">
        <div class="step-num">4</div>
        <div class="step-body">
          <p>Click <em>Save</em> and refresh the model list.</p>
        </div>
      </div>
    </div>

    <div v-if="provider === 'ollama'" class="tip">
      <div class="tip-icon">💡</div>
      <div><strong>Ollama:</strong> Open WebUI also has native Ollama support — using the AI Bridge endpoint applies IP restrictions and usage tracking on top.</div>
    </div>
    <div v-else class="tip">
      <div class="tip-icon">💡</div>
      <div><strong>Tip:</strong> Add additional connections in Open WebUI to expose models from other AI Bridge providers simultaneously.</div>
    </div>
  </div>
</template>

<style scoped src="./tab-shared.css" />
