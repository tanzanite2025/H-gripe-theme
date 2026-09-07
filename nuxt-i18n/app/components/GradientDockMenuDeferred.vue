<template>
  <ClientOnly>
    <component
      :is="gradientDockMenuComponent"
      v-if="mounted"
    />
    <GradientDockMenuShell
      v-else
      @expand="activateDock"
    />
    <template #fallback>
      <GradientDockMenuShell />
    </template>
  </ClientOnly>
</template>

<script setup lang="ts">
import { defineAsyncComponent, nextTick, onBeforeUnmount, onMounted, ref, type Component } from 'vue'
import GradientDockMenuShell from '~/components/GradientDockMenuShell.vue'
import { useQuickBuyOpenRequestState } from '~/composables/useQuickBuyOpenRequestState'

const mounted = ref(false)
const { requestOpen: requestQuickBuyOpen } = useQuickBuyOpenRequestState()
let dockModulePromise: Promise<{ default: Component }> | null = null

const loadGradientDockMenu = () => {
  dockModulePromise ??= import('./GradientDockMenu.vue')
  return dockModulePromise
}

const gradientDockMenuComponent = defineAsyncComponent(loadGradientDockMenu)

const activateDock = async () => {
  if (mounted.value) return
  await loadGradientDockMenu()
  mounted.value = true
  await nextTick()
}

const openQuickBuyFromGlobalEvent = () => {
  requestQuickBuyOpen()
  void activateDock()
}

onMounted(() => {
  window.addEventListener('quickbuy:open-entry', openQuickBuyFromGlobalEvent)
})

onBeforeUnmount(() => {
  window.removeEventListener('quickbuy:open-entry', openQuickBuyFromGlobalEvent)
})
</script>
