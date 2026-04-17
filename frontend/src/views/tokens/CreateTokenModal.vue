<script setup lang="ts">
import { ref, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useTokenStore } from '@/stores/tokens'
import type { CreateTokenResponse } from '@/services/api'

const emit = defineEmits<{ close: []; created: [result: CreateTokenResponse] }>()

const auth  = useAuthStore()
const store = useTokenStore()

const name     = ref('')
const duration = ref(7)
const creating = ref(false)
const error    = ref<string | null>(null)

const durationOptions = computed(() => {
  const base = [
    { label: '1 day', value: 1 },
    { label: '7 days', value: 7 },
    { label: '30 days', value: 30 },
  ]
  if (auth.isAdmin) {
    base.push(
      { label: '90 days', value: 90 },
      { label: '6 months', value: 180 },
      { label: '12 months', value: 365 },
    )
  }
  return base
})

async function submit() {
  if (!name.value.trim()) return
  creating.value = true; error.value = null
  try {
    const result = await store.generateToken(name.value.trim(), duration.value)
    emit('created', result)
    name.value = ''; duration.value = 7
  } catch (e: unknown) {
    const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error
    error.value = msg || 'Failed to create token'
  } finally { creating.value = false }
}
</script>

<template>
  <div class="modal-overlay" @click.self="emit('close')">
    <div class="modal">
      <h2>Create Token</h2>
      <form @submit.prevent="submit">
        <div class="field">
          <label for="token-name">Token name</label>
          <input id="token-name" v-model="name" type="text" placeholder="e.g. CI Pipeline, Local Dev" maxlength="100" autofocus />
        </div>
        <div class="field">
          <label for="token-duration">Expiration</label>
          <select id="token-duration" v-model="duration">
            <option v-for="opt in durationOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
          </select>
          <p class="field-hint">Max duration: {{ auth.isAdmin ? '12 months' : '30 days' }}</p>
        </div>
        <div v-if="error" class="error-msg">{{ error }}</div>
        <div class="modal-actions">
          <button type="button" class="btn btn-outline" @click="emit('close')">Cancel</button>
          <button type="submit" class="btn btn-primary" :disabled="creating || !name.trim()">
            {{ creating ? 'Creating…' : 'Create Token' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.4); display: flex; align-items: center; justify-content: center; z-index: 100; }
.modal { background: white; border-radius: 12px; padding: 2rem; width: 100%; max-width: 440px; display: flex; flex-direction: column; gap: 1.25rem; }
.modal h2 { font-size: 1.25rem; font-weight: 700; margin: 0; }
.field { display: flex; flex-direction: column; gap: 0.4rem; }
.field label { font-size: 0.875rem; font-weight: 500; color: #374151; }
.field input, .field select { width: 100%; padding: 0.5rem 0.75rem; border: 1px solid #d1d5db; border-radius: 6px; font-size: 0.95rem; background: white; box-sizing: border-box; }
.field input:focus, .field select:focus { outline: none; border-color: #3b82f6; box-shadow: 0 0 0 2px #bfdbfe; }
.field-hint { font-size: 0.78rem; color: #94a3b8; margin: 0; }
.error-msg { color: #ef4444; font-size: 0.875rem; }
.modal-actions { display: flex; justify-content: flex-end; gap: 0.75rem; }
.btn { padding: 0.4rem 1rem; border-radius: 6px; border: none; cursor: pointer; font-size: 0.9rem; font-weight: 500; }
.btn-primary { background: #3b82f6; color: white; }
.btn-primary:hover:not(:disabled) { background: #2563eb; }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-outline { background: transparent; color: #475569; border: 1px solid #cbd5e1; }
.btn-outline:hover { background: #f1f5f9; }
</style>
