import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface ChatMessage {
  id:      string
  role:    'user' | 'assistant' | 'error'
  content: string
}

let _id = 0
export function newId() { return String(++_id) }

export const useChatStore = defineStore('chat', () => {
  const messages  = ref<ChatMessage[]>([])
  const provider  = ref<string>('')
  const model     = ref('')
  const streaming = ref(false)

  function clear() { messages.value = [] }

  function addMessage(msg: ChatMessage) {
    messages.value.push(msg)
  }

  return { messages, provider, model, streaming, clear, addMessage }
})
