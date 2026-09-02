<template>
  <ClientOnly>
    <component
      :is="behaviorAttributionComponent"
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
let behaviorAttributionModulePromise: Promise<{ default: Component }> | null = null

const loadBehaviorAttribution = () => {
  behaviorAttributionModulePromise ??= import('./BehaviorAttributionBootstrap.vue')
  return behaviorAttributionModulePromise
}

const behaviorAttributionComponent = defineAsyncComponent(loadBehaviorAttribution)

const mountBehaviorAttribution = () => {
  mounted.value = true
  cancelDeferredMount?.()
  cancelDeferredMount = null
}

onMounted(() => {
  cancelDeferredMount = scheduleDeferredClientWork(mountBehaviorAttribution, STOREFRONT_IDLE_CLIENT_WORK)
})

onBeforeUnmount(() => {
  cancelDeferredMount?.()
  cancelDeferredMount = null
})
</script>
