import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useConnectivityStore = defineStore('connectivity', () => {
  const backendDown = ref(false)

  function markDown() { backendDown.value = true }
  function markUp()   { backendDown.value = false }

  return { backendDown, markDown, markUp }
})
