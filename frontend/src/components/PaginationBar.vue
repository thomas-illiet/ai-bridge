<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  page: number
  pageSize: number
  total: number
  pageSizeOptions?: number[]
}>()

const emit = defineEmits<{
  'update:page': [value: number]
  'update:pageSize': [value: number]
}>()

const options = computed(() => props.pageSizeOptions ?? [10, 20, 50, 100])
const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))

function pages(): (number | '…')[] {
  const tp = totalPages.value
  const p  = props.page
  if (tp <= 7) return Array.from({ length: tp }, (_, i) => i + 1)
  const set = new Set([1, tp, p - 1, p, p + 1].filter(n => n >= 1 && n <= tp))
  const sorted = [...set].sort((a, b) => a - b)
  const result: (number | '…')[] = []
  for (let i = 0; i < sorted.length; i++) {
    if (i > 0 && (sorted[i] as number) - (sorted[i - 1] as number) > 1) result.push('…')
    result.push(sorted[i])
  }
  return result
}

function setPage(p: number) {
  if (p < 1 || p > totalPages.value || p === props.page) return
  emit('update:page', p)
}

function setPageSize(e: Event) {
  emit('update:pageSize', Number((e.target as HTMLSelectElement).value))
  emit('update:page', 1)
}
</script>

<template>
  <div class="pagination-bar">
    <div class="pg-left">
      <label class="pg-size-label">Rows per page</label>
      <select class="pg-size-select" :value="pageSize" @change="setPageSize">
        <option v-for="o in options" :key="o" :value="o">{{ o }}</option>
      </select>
      <span class="pg-total">{{ total }} total</span>
    </div>

    <div class="pg-pages">
      <button class="pg-btn" :disabled="page === 1" @click="setPage(page - 1)">‹</button>
      <template v-for="(p, i) in pages()" :key="i">
        <span v-if="p === '…'" class="pg-ellipsis">…</span>
        <button
          v-else
          class="pg-btn"
          :class="{ active: p === page }"
          @click="setPage(p as number)"
        >{{ p }}</button>
      </template>
      <button class="pg-btn" :disabled="page === totalPages" @click="setPage(page + 1)">›</button>
    </div>
  </div>
</template>

<style scoped>
.pagination-bar {
  display: flex; align-items: center; justify-content: space-between;
  flex-wrap: wrap; gap: 0.75rem;
}

.pg-left { display: flex; align-items: center; gap: 0.6rem; }
.pg-size-label { font-size: 0.8rem; color: #64748b; }
.pg-size-select {
  padding: 0.25rem 0.5rem; border: 1px solid #e2e8f0; border-radius: 6px;
  font-size: 0.82rem; background: white; cursor: pointer; color: #374151; outline: none;
}
.pg-size-select:focus { border-color: #3b82f6; }
.pg-total { font-size: 0.8rem; color: #94a3b8; }

.pg-pages { display: flex; align-items: center; gap: 0.3rem; }
.pg-btn {
  min-width: 32px; height: 32px; padding: 0 0.5rem;
  border: 1px solid #e2e8f0; border-radius: 6px;
  background: white; color: #374151; font-size: 0.85rem; font-weight: 500;
  cursor: pointer; transition: background 0.12s, border-color 0.12s;
}
.pg-btn:hover:not(:disabled) { background: #f1f5f9; border-color: #cbd5e1; }
.pg-btn.active { background: #3b82f6; color: white; border-color: #3b82f6; }
.pg-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.pg-ellipsis { padding: 0 0.25rem; color: #94a3b8; font-size: 0.9rem; }
</style>
