import { ref } from 'vue'
import axios from '@/utils/axios'

let configPromise = null

export const useAdminGoogleAuth = () => {
  const clientId = ref('')
  const isLoading = ref(false)
  const error = ref('')

  const loadConfig = async () => {
    if (clientId.value) return clientId.value
    if (!configPromise) {
      configPromise = axios
        .get('/api/admin/auth/config')
        .then((response) => {
          clientId.value = String(
            response.data?.google_client_id
              || response.data?.data?.google_client_id
              || ''
          ).trim()
          return clientId.value
        })
        .finally(() => {
          configPromise = null
        })
    }
    return configPromise
  }

  const loadGsiScript = () => new Promise((resolve, reject) => {
    if (typeof window === 'undefined') {
      reject(new Error('Google Sign-In is only available in a browser'))
      return
    }

    if (window.google?.accounts?.id) {
      resolve()
      return
    }

    const existingScript = document.querySelector('script[src*="accounts.google.com/gsi/client"]')
    if (existingScript) {
      existingScript.addEventListener('load', resolve, { once: true })
      existingScript.addEventListener('error', () => reject(new Error('Failed to load Google Sign-In SDK')), { once: true })
      return
    }

    const script = document.createElement('script')
    script.src = 'https://accounts.google.com/gsi/client'
    script.async = true
    script.defer = true
    script.onload = resolve
    script.onerror = () => reject(new Error('Failed to load Google Sign-In SDK'))
    document.head.appendChild(script)
  })

  const initialize = async (callback) => {
    isLoading.value = true
    error.value = ''

    try {
      await loadConfig()
      if (!clientId.value) {
        throw new Error('Google login is not configured')
      }

      await loadGsiScript()
      if (!window.google?.accounts?.id) {
        throw new Error('Google Sign-In is unavailable')
      }

      window.google.accounts.id.initialize({
        client_id: clientId.value,
        callback,
        auto_select: false,
        cancel_on_tap_outside: true,
        context: 'signin',
        ux_mode: 'popup',
        itp_support: true
      })
      return true
    } catch (caught) {
      error.value = caught instanceof Error ? caught.message : 'Google Sign-In initialization failed'
      return false
    } finally {
      isLoading.value = false
    }
  }

  const prompt = () => {
    if (!window.google?.accounts?.id) {
      error.value = 'Google Sign-In is not initialized'
      return
    }
    window.google.accounts.id.prompt((notification) => {
      if (notification.isNotDisplayed()) {
        error.value = 'Google 登录弹窗未显示，请检查浏览器设置后重试'
      }
    })
  }

  return {
    clientId,
    isLoading,
    error,
    loadConfig,
    initialize,
    prompt
  }
}
