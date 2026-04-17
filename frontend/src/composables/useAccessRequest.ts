import { ref, watch, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { getMyAccessRequest } from '@/services/api'
import type { AccessRequest } from '@/services/api'

export function useAccessRequest() {
  const auth    = useAuthStore()
  const router  = useRouter()
  const request = ref<AccessRequest | null>(null)
  const loading = ref(false)
  let pollTimer: ReturnType<typeof setInterval> | null = null

  async function fetchRequest() {
    try {
      const res = await getMyAccessRequest()
      request.value = res.data
      if (res.data?.status === 'approved') {
        stopPolling()
        await auth.fetchRole()
        router.push('/dashboard')
      }
    } catch {
      request.value = null
    }
  }

  function startPolling() {
    if (pollTimer !== null) return
    pollTimer = setInterval(fetchRequest, 30_000)
  }

  function stopPolling() {
    if (pollTimer !== null) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  }

  watch(request, (val) => {
    if (val?.status === 'pending') startPolling()
    else stopPolling()
  })

  onUnmounted(stopPolling)

  return { request, loading, fetchRequest, startPolling, stopPolling }
}
