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

const mounted = ref(false)
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
}

onMounted(() => {
  window.addEventListener(STOREFRONT_CLIENT_OVERLAYS_EVENT, activate)
})

onBeforeUnmount(() => {
  window.removeEventListener(STOREFRONT_CLIENT_OVERLAYS_EVENT, activate)
})
</script>
