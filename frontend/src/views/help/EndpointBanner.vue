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
  <div class="info-banner">
    <div class="info-row">
      <span class="info-label">Active endpoint</span>
      <code class="info-value">{{ activeUrl }}</code>
      <button class="copy-btn" @click="copy">{{ copied ? 'Copied!' : 'Copy' }}</button>
    </div>
    <div class="info-row">
      <span class="info-label">Model</span>
      <span v-if="modelLoading" class="info-value muted">Loading…</span>
      <template v-else-if="modelList.length > 0">
        <select
          class="model-select"
          :value="selectedModel"
          @change="emit('update:selectedModel', ($event.target as HTMLSelectElement).value)"
        >
          <option v-for="m in modelList" :key="m" :value="m">{{ m }}</option>
        </select>
      </template>
      <code v-else class="info-value">{{ activeModel }}</code>
    </div>
    <div class="info-row">
      <span class="info-label">API Key</span>
      <span class="info-value muted">
        Generate a Personal Access Token in <RouterLink to="/tokens">Tokens</RouterLink>
      </span>
    </div>
  </div>
</template>

<style scoped>
.info-banner {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 1rem 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}
.info-row { display: flex; align-items: center; gap: 0.75rem; flex-wrap: wrap; }
.info-label { font-size: 0.8rem; font-weight: 600; color: #64748b; min-width: 130px; }
.info-value {
  font-size: 0.85rem;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 5px;
  padding: 0.2rem 0.5rem;
  font-family: monospace;
}
.info-value.muted {
  background: none;
  border: none;
  color: #475569;
  font-family: inherit;
  padding: 0;
}
.info-value.muted a { color: #3b82f6; text-decoration: none; }
.info-value.muted a:hover { text-decoration: underline; }
.model-select {
  font-size: 0.85rem;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 5px;
  padding: 0.2rem 0.5rem;
  font-family: monospace;
  color: #1e293b;
  cursor: pointer;
  outline: none;
  min-width: 220px;
}
.model-select:focus { border-color: #3b82f6; }
.copy-btn {
  flex-shrink: 0;
  padding: 0.2rem 0.6rem;
  font-size: 0.75rem;
  font-weight: 600;
  border: 1px solid #e2e8f0;
  background: #fff;
  color: #64748b;
  border-radius: 5px;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}
.copy-btn:hover { background: #f1f5f9; color: #1e293b; }
</style>
