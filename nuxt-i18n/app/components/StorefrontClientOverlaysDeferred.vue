<template>
  <ClientOnly>
    <component
      :is="storefrontClientOverlaysComponent"
      v-if="mounted"
    />
  </ClientOnly>
</template>

<script setup lang="ts">
import { defineAsyncComponent, onBeforeUnmount, onMounted, ref, type Component } from 'vue'
import { STOREFRONT_CLIENT_OVERLAYS_EVENT } from '~/utils/clientOverlays'
import { scheduleDeferredClientWork } from '~/utils/clientDeferredWork'
import { STOREFRONT_IDLE_CLIENT_WORK } from '~/utils/storefrontLoadingPolicy'

const mounted = ref(false)
let cancelDeferredMount: (() => void) | null = null
let storefrontClientOverlaysModulePromise: Promise<{ default: Component }> | null = null

const loadStorefrontClientOverlays = () => {
  storefrontClientOverlaysModulePromise ??= import('./StorefrontClientOverlays.client.vue')
  return storefrontClientOverlaysModulePromise
}

const storefrontClientOverlaysComponent = defineAsyncComponent(loadStorefrontClientOverlays)

const activate = () => {
  if (mounted.value) return
  void loadStorefrontClientOverlays()
  mounted.value = true
  cancelDeferredMount?.()
  cancelDeferredMount = null
}

onMounted(() => {
  cancelDeferredMount = scheduleDeferredClientWork(activate, STOREFRONT_IDLE_CLIENT_WORK)

  window.addEventListener(STOREFRONT_CLIENT_OVERLAYS_EVENT, activate)
})

onBeforeUnmount(() => {
  cancelDeferredMount?.()
  cancelDeferredMount = null
  window.removeEventListener(STOREFRONT_CLIENT_OVERLAYS_EVENT, activate)
})
</script>
