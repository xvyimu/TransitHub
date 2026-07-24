import { createApp } from 'vue'
import { createPinia } from 'pinia'
import Antd from 'antdv-next'
import App from './App.vue'
import router from './router'
import i18n from './i18n'
import { setUnauthorizedHandler } from './api/http'
import { useAuthStore } from './stores/auth'
import './styles/main.css'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(i18n)
app.use(Antd)

const auth = useAuthStore(pinia)
setUnauthorizedHandler(() => {
  auth.clearSession()
  if (router.currentRoute.value.name !== 'login') {
    void router.replace({
      name: 'login',
      query: { redirect: router.currentRoute.value.fullPath },
    })
  }
})

app.mount('#app')