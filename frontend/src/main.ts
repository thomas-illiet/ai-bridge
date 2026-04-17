import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { initKeycloak } from './services/keycloak'
import { useAuthStore } from './stores/auth'

async function bootstrap() {
  const app = createApp(App)
  const pinia = createPinia()

  app.use(pinia)

  try {
    await initKeycloak()
    const auth = useAuthStore()
    auth.sync()
    await auth.fetchRole()
  } catch (e) {
    console.error('Keycloak init failed, continuing unauthenticated', e)
  }

  app.use(router)

  app.mount('#app')
}

bootstrap()
