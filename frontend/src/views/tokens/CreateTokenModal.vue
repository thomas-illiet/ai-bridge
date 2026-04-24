<script setup lang="ts">
import { ref, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useTokenStore } from '@/stores/tokens'
import type { CreateTokenResponse } from '@/services/api'

const emit = defineEmits<{ close: []; created: [result: CreateTokenResponse] }>()

const auth  = useAuthStore()
const store = useTokenStore()

const name        = ref('')
const description = ref('')
const duration    = ref(auth.isElevated ? 7 : 5)
const creating    = ref(false)
const error       = ref<string | null>(null)

const durationOptions = computed(() => {
  if (auth.isElevated) {
    return [
      { label: '1 day', value: 1 },
      { label: '7 days', value: 7 },
      { label: '30 days', value: 30 },
    ]
  }
  return [
    { label: '1 day', value: 1 },
    { label: '5 days', value: 5 },
  ]
})

async function submit() {
  if (!name.value.trim()) return
  creating.value = true; error.value = null
  try {
    const result = await store.generateToken(name.value.trim(), duration.value, description.value.trim())
    emit('created', result)
    name.value = ''; description.value = ''; duration.value = auth.isElevated ? 7 : 5
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
          <label for="token-description">Description <span class="optional">(optional)</span></label>
          <textarea id="token-description" v-model="description" placeholder="What is this token used for?" maxlength="255" rows="2" />
        </div>
        <div class="field">
          <label for="token-duration">Expiration</label>
          <select id="token-duration" v-model="duration">
            <option v-for="opt in durationOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
          </select>
          <p class="field-hint">Max duration: {{ auth.isElevated ? '30 days' : '5 days' }}</p>
        </div>
        <div v-if="error" class="error-msg">{{ error }}</div>
        <div class="modal-actions">
          <button type="button" class="btn btn-outline" @click="emit('close')">
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            Cancel
          </button>
          <button type="submit" class="btn btn-primary" :disabled="creating || !name.trim()">
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            {{ creating ? 'Creating…' : 'Create Token' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped></style>
