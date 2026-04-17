<script setup lang="ts">
import type { ChatMessage } from '@/stores/chat'

defineProps<{ message: ChatMessage; isStreaming: boolean }>()
</script>

<template>
  <div class="msg-row" :class="message.role">
    <div class="bubble" :class="message.role">
      <!-- animated typing dots while waiting for first token -->
      <span v-if="message.role === 'assistant' && isStreaming && message.content === ''" class="dots">
        <span /><span /><span />
      </span>
      <span v-else class="text">{{ message.content }}</span>
    </div>
  </div>
</template>

<style scoped>
.msg-row {
  display: flex;
  margin-bottom: 0.6rem;
}
.msg-row.user      { justify-content: flex-end; }
.msg-row.assistant,
.msg-row.error     { justify-content: flex-start; }

.bubble {
  max-width: 78%;
  padding: 0.55rem 0.85rem;
  border-radius: 14px;
  font-size: 0.88rem;
  line-height: 1.55;
  word-break: break-word;
  white-space: pre-wrap;
}
.bubble.user      { background: #3b82f6; color: #fff; border-bottom-right-radius: 4px; }
.bubble.assistant { background: #f1f5f9; color: #1e293b; border-bottom-left-radius: 4px; }
.bubble.error     { background: #fee2e2; color: #dc2626; border-bottom-left-radius: 4px; font-size: 0.82rem; }

/* 3-dot animated spinner */
.dots {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 0;
}
.dots span {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #94a3b8;
  animation: bounce 1.2s ease-in-out infinite;
}
.dots span:nth-child(2) { animation-delay: 0.2s; }
.dots span:nth-child(3) { animation-delay: 0.4s; }

@keyframes bounce {
  0%, 80%, 100% { transform: scale(0.7); opacity: 0.5; }
  40%           { transform: scale(1);   opacity: 1;   }
}
</style>
