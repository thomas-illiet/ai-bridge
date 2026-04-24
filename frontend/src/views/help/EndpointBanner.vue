<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{
  activeUrl:    string
  activeModel:  string
  modelList:    string[]
  modelLoading: boolean
  selectedModel: string
}>()
const emit = defineEmits<{ 'update:selectedModel': [value: string] }>()

const copied = ref(false)
function copy() {
  navigator.clipboard.writeText(props.activeUrl)
  copied.value = true
  setTimeout(() => { copied.value = false }, 1800)
}
</script>

<template>
  <div class="endpoint-card">
    <div class="card-row">
      <div class="row-icon endpoint-icon">
        <svg viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M12.316 3.051a1 1 0 01.633 1.265l-4 12a1 1 0 11-1.898-.632l4-12a1 1 0 011.265-.633zM5.707 6.293a1 1 0 010 1.414L3.414 10l2.293 2.293a1 1 0 11-1.414 1.414l-3-3a1 1 0 010-1.414l3-3a1 1 0 011.414 0zm8.586 0a1 1 0 011.414 0l3 3a1 1 0 010 1.414l-3 3a1 1 0 11-1.414-1.414L16.586 10l-2.293-2.293a1 1 0 010-1.414z" clip-rule="evenodd" />
        </svg>
      </div>
      <div class="row-body">
        <span class="row-label">Endpoint URL</span>
        <div class="row-value-group">
          <code class="row-value url-value">{{ activeUrl }}</code>
          <button class="copy-btn" :class="{ copied }" @click="copy">
            <svg v-if="!copied" viewBox="0 0 20 20" fill="currentColor">
              <path d="M8 3a1 1 0 011-1h2a1 1 0 110 2H9a1 1 0 01-1-1z" />
              <path d="M6 3a2 2 0 00-2 2v11a2 2 0 002 2h8a2 2 0 002-2V5a2 2 0 00-2-2 3 3 0 01-3 3H9a3 3 0 01-3-3z" />
            </svg>
            <svg v-else viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
            </svg>
            {{ copied ? 'Copied!' : 'Copy' }}
          </button>
        </div>
      </div>
    </div>

    <div class="divider" />

    <div class="card-row">
      <div class="row-icon model-icon">
        <svg viewBox="0 0 20 20" fill="currentColor">
          <path d="M13 7H7v6h6V7z" />
          <path fill-rule="evenodd" d="M7 2a1 1 0 012 0v1h2V2a1 1 0 112 0v1h2a2 2 0 012 2v2h1a1 1 0 110 2h-1v2h1a1 1 0 110 2h-1v2a2 2 0 01-2 2h-2v1a1 1 0 11-2 0v-1H9v1a1 1 0 11-2 0v-1H5a2 2 0 01-2-2v-2H2a1 1 0 110-2h1V9H2a1 1 0 110-2h1V5a2 2 0 012-2h2V2zM5 5h10v10H5V5z" clip-rule="evenodd" />
        </svg>
      </div>
      <div class="row-body">
        <span class="row-label">Model</span>
        <div class="row-value-group">
          <span v-if="modelLoading" class="row-value muted">Loading…</span>
          <select
            v-else-if="modelList.length > 0"
            class="model-select"
            :value="selectedModel"
            @change="emit('update:selectedModel', ($event.target as HTMLSelectElement).value)"
          >
            <option v-for="m in modelList" :key="m" :value="m">{{ m }}</option>
          </select>
          <code v-else class="row-value">{{ activeModel }}</code>
        </div>
      </div>
    </div>

    <div class="divider" />

    <div class="card-row">
      <div class="row-icon key-icon">
        <svg viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M18 8a6 6 0 01-7.743 5.743L10 14l-1 1-1 1H6v2H2v-4l4.257-4.257A6 6 0 1118 8zm-6-4a1 1 0 100 2 2 2 0 012 2 1 1 0 102 0 4 4 0 00-4-4z" clip-rule="evenodd" />
        </svg>
      </div>
      <div class="row-body">
        <span class="row-label">API Key</span>
        <span class="row-value muted">
          Generate a Personal Access Token in
          <RouterLink to="/tokens" class="token-link">Tokens</RouterLink>
          and use it as the Bearer token.
        </span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.endpoint-card {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0,0,0,0.06);
}

.card-row {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  padding: 0.9rem 1.25rem;
}

.divider { height: 1px; background: #f1f5f9; }

.row-icon {
  flex-shrink: 0;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: 2px;
}
.row-icon svg { width: 16px; height: 16px; }

.endpoint-icon { background: #eff6ff; color: #3b82f6; }
.model-icon    { background: #f0fdf4; color: #16a34a; }
.key-icon      { background: #fefce8; color: #ca8a04; }

.row-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  min-width: 0;
}

.row-label {
  font-size: 0.72rem;
  font-weight: 600;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.row-value-group {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  flex-wrap: wrap;
}

.row-value {
  font-size: 0.875rem;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  padding: 0.25rem 0.6rem;
  font-family: 'Fira Code', 'Cascadia Code', monospace;
  color: #1e293b;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 100%;
}
.row-value.muted {
  background: none;
  border: none;
  font-family: inherit;
  color: #64748b;
  font-size: 0.875rem;
  padding: 0;
  white-space: normal;
}
.url-value { max-width: 520px; }

.model-select {
  font-size: 0.875rem;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  padding: 0.25rem 0.6rem;
  font-family: 'Fira Code', 'Cascadia Code', monospace;
  color: #1e293b;
  cursor: pointer;
  outline: none;
  min-width: 220px;
}
.model-select:focus { border-color: #3b82f6; box-shadow: 0 0 0 2px #bfdbfe; }

.copy-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  flex-shrink: 0;
  padding: 0.25rem 0.65rem;
  font-size: 0.75rem;
  font-weight: 600;
  border: 1px solid #e2e8f0;
  background: #fff;
  color: #64748b;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s;
}
.copy-btn svg { width: 13px; height: 13px; }
.copy-btn:hover { background: #f1f5f9; color: #1e293b; border-color: #cbd5e1; }
.copy-btn.copied { background: #f0fdf4; color: #16a34a; border-color: #bbf7d0; }

.token-link { color: #3b82f6; text-decoration: none; font-weight: 500; }
.token-link:hover { text-decoration: underline; }
</style>
