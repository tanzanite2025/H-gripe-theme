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
  // Attribution must run during the initial landing-page interaction window;
  // deferring it for 15s loses UTM/GCLID data when users bounce or navigate.
  cancelDeferredMount = scheduleDeferredClientWork(mountBehaviorAttribution, {
    delayMs: 250,
    idleTimeoutMs: 1000,
  })
})

onBeforeUnmount(() => {
  cancelDeferredMount?.()
  cancelDeferredMount = null
})
</script>
