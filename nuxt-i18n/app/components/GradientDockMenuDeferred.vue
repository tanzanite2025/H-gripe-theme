<template>
  <ClientOnly>
    <component
      :is="gradientDockMenuComponent"
      v-if="mounted"
    />
    <GradientDockMenuShell
      v-else
      @intent="handleShellIntent"
    />
    <template #fallback>
      <GradientDockMenuShell />
    </template>
  </ClientOnly>
</template>

<script setup lang="ts">
import { defineAsyncComponent, nextTick, onBeforeUnmount, onMounted, ref, type Component } from 'vue'
import GradientDockMenuShell from '~/components/GradientDockMenuShell.vue'
import { activateStorefrontClientOverlays } from '~/utils/clientOverlays'
import { scheduleDeferredClientWork } from '~/utils/clientDeferredWork'
import { STOREFRONT_IDLE_CLIENT_WORK } from '~/utils/storefrontLoadingPolicy'
import { useSidePanelState } from '~/composables/useSidePanelState'

type GradientDockMenuShellIntent = 'sidebar' | 'chat' | 'quick-buy' | 'cart'

const mounted = ref(false)
const { openLeft } = useSidePanelState()
let cancelDeferredMount: (() => void) | null = null
let dockModulePromise: Promise<{ default: Component }> | null = null

const loadGradientDockMenu = () => {
  dockModulePromise ??= import('./GradientDockMenu.vue')
  return dockModulePromise
}

const gradientDockMenuComponent = defineAsyncComponent(loadGradientDockMenu)

const activateDock = async () => {
  await loadGradientDockMenu()

  if (!mounted.value) {
    mounted.value = true
  }
  cancelDeferredMount?.()
  cancelDeferredMount = null

  activateStorefrontClientOverlays()

  await nextTick()
}

const dispatchDockIntent = (intent: GradientDockMenuShellIntent) => {
  if (typeof window === 'undefined') return

  if (intent === 'sidebar') {
    openLeft()
    return
  }

  const eventName = intent === 'quick-buy'
    ? 'dock:open-quick-buy'
    : intent === 'cart'
      ? 'dock:open-cart'
      : 'dock:open-chat'
  window.dispatchEvent(new Event(eventName))
}

const handleShellIntent = async (intent: GradientDockMenuShellIntent) => {
  if (intent === 'sidebar') {
    openLeft()
    return
  }

  await activateDock()
  await nextTick()
  dispatchDockIntent(intent)
}

onMounted(() => {
  cancelDeferredMount = scheduleDeferredClientWork(() => {
    void activateDock()
  }, STOREFRONT_IDLE_CLIENT_WORK)
})

onBeforeUnmount(() => {
  cancelDeferredMount?.()
  cancelDeferredMount = null
})
</script>
