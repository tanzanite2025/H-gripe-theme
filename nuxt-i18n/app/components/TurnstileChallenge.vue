<template>
  <div ref="container" class="sr-only" aria-hidden="true"></div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRuntimeConfig } from 'nuxt/app'
import { loadTurnstileScript } from '~/utils/security/trustedScriptUrl'

type TurnstileApi = {
  render: (
    element: HTMLElement,
    options: {
      sitekey: string
      size?: 'invisible'
      action?: string
      callback?: (token: string) => void
      'error-callback'?: () => void
      'expired-callback'?: () => void
    },
  ) => string | number
  execute: (widgetId: string | number) => void
  reset: (widgetId: string | number) => void
}

declare global {
  interface Window {
    turnstile?: TurnstileApi
  }
}

const props = withDefaults(
  defineProps<{
    action?: string
    required?: boolean
  }>(),
  {
    action: 'verification',
    required: true,
  },
)

const container = ref<HTMLElement | null>(null)
const runtimeConfig = useRuntimeConfig()
let widgetId: string | number | null = null
let loadPromise: Promise<void> | null = null
let pending: { resolve: (token: string) => void; reject: (error: Error) => void } | null = null

const loadTurnstile = () => {
  if (window.turnstile) return Promise.resolve()
  if (loadPromise) return loadPromise

  loadPromise = loadTurnstileScript().then(() => undefined)
  return loadPromise
}

const ensureWidget = async () => {
  const siteKey = String(runtimeConfig.public?.turnstileSiteKey || '').trim()
  if (!siteKey) {
    if (props.required) {
      throw new Error('Verification challenge is not configured')
    }
    return null
  }
  if (!container.value) {
    throw new Error('Verification challenge is not ready')
  }

  await loadTurnstile()
  if (!window.turnstile) {
    throw new Error('Verification challenge is unavailable')
  }
  if (widgetId === null) {
    widgetId = window.turnstile.render(container.value, {
      sitekey: siteKey,
      size: 'invisible',
      action: props.action,
      callback: token => {
        pending?.resolve(token)
        pending = null
      },
      'error-callback': () => {
        pending?.reject(new Error('Verification challenge failed'))
        pending = null
      },
      'expired-callback': () => {
        pending?.reject(new Error('Verification challenge expired'))
        pending = null
      },
    })
  }
  return widgetId
}

const execute = async (): Promise<string> => {
  if (!import.meta.client) return ''
  const id = await ensureWidget()
  if (id === null) return ''

  return new Promise<string>((resolve, reject) => {
    if (!window.turnstile) {
      reject(new Error('Verification challenge is unavailable'))
      return
    }
    pending = { resolve, reject }
    window.turnstile.reset(id)
    window.turnstile.execute(id)
  })
}

onMounted(() => {
  if (runtimeConfig.public?.turnstileSiteKey) {
    void ensureWidget().catch(() => undefined)
  }
})

defineExpose({ execute })
</script>
