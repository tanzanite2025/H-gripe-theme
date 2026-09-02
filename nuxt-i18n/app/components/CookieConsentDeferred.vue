<template>
  <ClientOnly>
    <component
      :is="cookieConsentComponent"
      v-if="mounted"
    />
  </ClientOnly>
</template>

<script setup lang="ts">
import { defineAsyncComponent, onBeforeUnmount, onMounted, ref, type Component } from 'vue'
import { scheduleDeferredClientWork } from '~/utils/clientDeferredWork'
import { STOREFRONT_IDLE_CLIENT_WORK } from '~/utils/storefrontLoadingPolicy'

const mounted = ref(false)
let cancelDeferredMount: (() => void) | null = null
let cookieConsentModulePromise: Promise<{ default: Component }> | null = null

const loadCookieConsent = () => {
  cookieConsentModulePromise ??= import('./CookieConsent.vue')
  return cookieConsentModulePromise
}

const cookieConsentComponent = defineAsyncComponent(loadCookieConsent)

const mountCookieConsent = () => {
  mounted.value = true
  cancelDeferredMount?.()
  cancelDeferredMount = null
}

onMounted(() => {
  cancelDeferredMount = scheduleDeferredClientWork(mountCookieConsent, STOREFRONT_IDLE_CLIENT_WORK)
})

onBeforeUnmount(() => {
  cancelDeferredMount?.()
  cancelDeferredMount = null
})
</script>
