import { createApp } from 'vue'
import './styles/shared.css'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { initOidc } from './services/oidc'
import { useAuthStore } from './stores/auth'
import { useConnectivityStore } from './stores/connectivity'

async function bootstrap() {
  const app = createApp(App)
  const pinia = createPinia()

  app.use(pinia)

  const connectivity = useConnectivityStore()

  try {
    const res = await fetch('/health', { signal: AbortSignal.timeout(5000) })
    if (!res.ok) connectivity.markDown()
  } catch {
    connectivity.markDown()
  }

  if (!connectivity.backendDown) {
    try {
      await initOidc()
      const auth = useAuthStore()
      auth.sync()
      await auth.fetchRole()
    } catch (e) {
      console.error('OIDC init failed, continuing unauthenticated', e)
    }
  }

  app.use(router)

  app.mount('#app')
}

bootstrap()
