<script setup lang="ts">
import { ref } from 'vue'
import { useTokenStore } from '@/stores/tokens'
import type { ClientToken } from '@/services/api'

const props = defineProps<{ token: ClientToken }>()
const emit = defineEmits<{ close: [] }>()

const store = useTokenStore()

const name        = ref(props.token.name)
const description = ref(props.token.description ?? '')
const saving      = ref(false)
const error       = ref<string | null>(null)

async function submit() {
  if (!name.value.trim()) return
  saving.value = true; error.value = null
  try {
    await store.updateToken(props.token.id, name.value.trim(), description.value.trim())
    emit('close')
  } catch (e: unknown) {
    const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error
    error.value = msg || 'Failed to update token'
  } finally { saving.value = false }
}
</script>

<template>
  <div class="modal-overlay" @click.self="emit('close')">
    <div class="modal">
      <h2>Edit Token</h2>
      <form @submit.prevent="submit">
        <div class="field">
          <label for="edit-token-name">Token name</label>
          <input id="edit-token-name" v-model="name" type="text" maxlength="100" autofocus />
        </div>
        <div class="field">
          <label for="edit-token-description">Description <span class="optional">(optional)</span></label>
          <textarea id="edit-token-description" v-model="description" maxlength="255" rows="2" />
        </div>
        <div v-if="error" class="error-msg">{{ error }}</div>
        <div class="modal-actions">
          <button type="button" class="btn btn-outline" @click="emit('close')">
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            Cancel
          </button>
          <button type="submit" class="btn btn-primary" :disabled="saving || !name.trim()">
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
            {{ saving ? 'Saving…' : 'Save changes' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped></style>
