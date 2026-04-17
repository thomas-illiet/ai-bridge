<script setup lang="ts">
import type { InterceptionDetail } from '@/services/api'
import { formatDate, fmtNum, interceptionDuration, providerColor } from '@/utils/format'

defineProps<{
  detail: InterceptionDetail | null
  loading: boolean
}>()

defineEmits<{ close: [] }>()
</script>

<template>
  <Teleport to="body">
    <div v-if="detail || loading" class="modal-backdrop" @click.self="$emit('close')">
      <div class="modal">
        <div class="modal-header">
          <div v-if="detail">
            <h2>{{ detail.model }}</h2>
            <span class="prov-badge" :style="{ background: providerColor(detail.provider) + '22', color: providerColor(detail.provider) }">
              {{ detail.provider }}
            </span>
            <span class="meta-info">{{ formatDate(detail.startedAt) }} · {{ interceptionDuration(detail) }}</span>
          </div>
          <div v-else class="muted">Loading…</div>
          <button class="modal-close" @click="$emit('close')">✕</button>
        </div>

        <div v-if="loading" class="state-msg">Loading…</div>

        <template v-else-if="detail">
          <div class="token-row">
            <div class="token-card">
              <span class="token-label">Input tokens</span>
              <span class="token-val">{{ fmtNum(detail.inputTokens) }}</span>
            </div>
            <div class="token-card">
              <span class="token-label">Output tokens</span>
              <span class="token-val">{{ fmtNum(detail.outputTokens) }}</span>
            </div>
          </div>

          <div class="prompts-section">
            <h3 class="prompts-title">Prompts</h3>
            <div v-if="!detail.prompts?.length" class="state-msg">No prompts recorded.</div>
            <div v-else class="prompt-list">
              <div v-for="(p, i) in detail.prompts" :key="i" class="prompt-card">
                <span class="prompt-index">#{{ i + 1 }}</span>
                <pre class="prompt-text">{{ p }}</pre>
              </div>
            </div>
          </div>
        </template>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.modal-backdrop {
  position: fixed; inset: 0; background: rgba(0,0,0,0.45);
  display: flex; align-items: center; justify-content: center;
  z-index: 200; padding: 1rem;
}
.modal {
  background: white; border-radius: 14px; padding: 1.75rem;
  width: 100%; max-width: 720px; max-height: 88vh; overflow-y: auto;
  display: flex; flex-direction: column; gap: 1.25rem;
  box-shadow: 0 20px 60px rgba(0,0,0,0.2);
}
.modal-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 1rem; }
.modal-header h2 { font-size: 1.1rem; font-weight: 700; margin: 0 0 0.3rem; font-family: monospace; }
.modal-close { background: none; border: none; font-size: 1.1rem; cursor: pointer; color: #64748b; padding: 0.2rem 0.4rem; border-radius: 5px; flex-shrink: 0; }
.modal-close:hover { background: #f1f5f9; }
.muted { color: #64748b; }
.meta-info { font-size: 0.8rem; color: #94a3b8; margin-left: 0.5rem; }
.prov-badge { display: inline-block; padding: 0.15rem 0.55rem; border-radius: 999px; font-size: 0.75rem; font-weight: 600; text-transform: capitalize; margin-left: 0.4rem; }
.token-row { display: flex; gap: 0.75rem; }
.token-card { flex: 1; background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 8px; padding: 0.8rem 1rem; display: flex; flex-direction: column; gap: 0.2rem; }
.token-label { font-size: 0.72rem; font-weight: 600; color: #64748b; text-transform: uppercase; letter-spacing: 0.04em; }
.token-val { font-size: 1.5rem; font-weight: 700; color: #0f172a; }
.prompts-section { display: flex; flex-direction: column; gap: 0.75rem; }
.prompts-title { font-size: 0.85rem; font-weight: 700; color: #475569; text-transform: uppercase; letter-spacing: 0.04em; margin: 0; }
.prompt-list { display: flex; flex-direction: column; gap: 0.75rem; max-height: 420px; overflow-y: auto; }
.prompt-card { background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 8px; padding: 0.85rem 1rem; position: relative; }
.prompt-index { position: absolute; top: 0.5rem; right: 0.75rem; font-size: 0.7rem; color: #94a3b8; font-weight: 600; }
.prompt-text { margin: 0; font-family: monospace; font-size: 0.82rem; color: #1e293b; white-space: pre-wrap; word-break: break-word; line-height: 1.55; }
.state-msg { color: #64748b; font-size: 0.9rem; }
</style>
