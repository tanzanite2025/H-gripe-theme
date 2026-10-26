import { createApp } from 'vue'
import { createPinia } from 'pinia'
import './styles/admin.css'
import 'vue-sonner/style.css'

import App from './App.vue'
import router from './router'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

async function mountAdminApplicationAfterBuiltInFontLoads(): Promise<void> {
  try {
    // font-display: block can reveal a fallback after about three seconds.
    // Load the registered face explicitly before mounting any page or portal.
    await document.fonts.load('400 16px "MapleUICJK"')
  } catch (error) {
    // A failed font request must not prevent administrators from signing in.
    console.warn('[Admin] Built-in font could not load; continuing startup.', error)
  }

  app.mount('#app')
}

void mountAdminApplicationAfterBuiltInFontLoads()
