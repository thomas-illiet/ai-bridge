import { ref } from 'vue'

export function useMinLoad(minMs = 300, initial = false) {
  const loading = ref(initial)

  async function withLoad(fn: () => Promise<void>): Promise<void> {
    loading.value = true
    try {
      await Promise.all([fn(), new Promise<void>(r => setTimeout(r, minMs))])
    } finally {
      loading.value = false
    }
  }

  return { loading, withLoad }
}
