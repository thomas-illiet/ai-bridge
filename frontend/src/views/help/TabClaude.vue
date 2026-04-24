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
      <div class="header-left">
        <h2>Claude Code</h2>
        <span class="provider-badge anthropic">Anthropic</span>
      </div>
    </div>

    <p class="tab-desc">
      Configure the Claude client to route through AI Bridge by overriding
      <code>ANTHROPIC_BASE_URL</code> and <code>ANTHROPIC_AUTH_TOKEN</code>.
      Use a Personal Access Token from <RouterLink to="/tokens">Tokens</RouterLink> as the auth token.
    </p>

    <div class="steps">
      <div class="step">
        <div class="step-num">1</div>
        <div class="step-body">
          <p><strong>Environment variables</strong> — export these in your shell or <code>.env</code> file before running any Claude-based tool.</p>
          <div class="code-block">
            <div class="code-header">
              <span class="code-lang">bash</span>
              <button class="code-copy" :class="{ copied: copied === 'env' }" @click="copy(envSnippet, 'env')">
                {{ copied === 'env' ? '✓ Copied!' : 'Copy' }}
              </button>
            </div>
            <pre>{{ envSnippet }}</pre>
          </div>
        </div>
      </div>

      <div class="step">
        <div class="step-num">2</div>
        <div class="step-body">
          <p><strong>Claude Code — <code>settings.json</code></strong> — add this block to <code>~/.claude/settings.json</code> or the project-level <code>.claude/settings.json</code>.</p>
          <div class="code-block">
            <div class="code-header">
              <span class="code-lang">json</span>
              <button class="code-copy" :class="{ copied: copied === 'settings' }" @click="copy(settingsJson, 'settings')">
                {{ copied === 'settings' ? '✓ Copied!' : 'Copy' }}
              </button>
            </div>
            <pre>{{ settingsJson }}</pre>
          </div>
        </div>
      </div>
    </div>

    <div class="tip">
      <div class="tip-icon">💡</div>
      <div><strong>API_TIMEOUT_MS</strong> is optional but recommended for long-running requests.</div>
    </div>
  </div>
</template>

<style scoped src="./tab-shared.css" />
