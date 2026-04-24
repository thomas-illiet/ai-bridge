<script setup lang="ts">
import { ref, computed } from 'vue'

const props = defineProps<{
  activeUrl:   string
  activeModel: string
  provider:    'openai' | 'ollama' | 'anthropic'
}>()

const snippet = computed(() =>
JSON.stringify({
  providers: {
    aibridge: {
      name: 'AI Bridge',
      apiKey: 'YOUR_PAT_HERE',
      baseURL: props.activeUrl,
    },
  },
}, null, 2))

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
        <h2>OpenCode</h2>
        <span class="provider-badge" :class="provider">{{ provider === 'openai' ? 'OpenAI' : 'Ollama' }}</span>
      </div>
    </div>

    <p class="tab-desc">
      OpenCode reads its provider configuration from <code>~/.config/opencode/config.json</code>.
    </p>

    <div class="steps">
      <div class="step">
        <div class="step-num">1</div>
        <div class="step-body">
          <p>Generate a PAT in AI Bridge (<RouterLink to="/tokens">Tokens</RouterLink>) and copy it.</p>
        </div>
      </div>

      <div class="step">
        <div class="step-num">2</div>
        <div class="step-body">
          <p>Add the AI Bridge provider to your config:</p>
          <div class="code-block">
            <div class="code-header">
              <span class="code-lang">json</span>
              <button class="code-copy" :class="{ copied: copied === 'cfg' }" @click="copy(snippet, 'cfg')">
                {{ copied === 'cfg' ? '✓ Copied!' : 'Copy' }}
              </button>
            </div>
            <pre>{{ snippet }}</pre>
          </div>
        </div>
      </div>

      <div class="step">
        <div class="step-num">3</div>
        <div class="step-body">
          <p>Restart OpenCode — the provider will appear in the model picker.</p>
        </div>
      </div>
    </div>

    <div class="tip">
      <div class="tip-icon">💡</div>
      <div><strong>Model name:</strong> Set the model field to <code>{{ activeModel }}</code> or any model available through this provider.</div>
    </div>
  </div>
</template>

<style scoped src="./tab-shared.css" />
