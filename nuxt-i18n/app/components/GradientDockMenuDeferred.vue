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
import { scheduleDeferredClientWork } from '~/utils/clientDeferredWork'
import { STOREFRONT_IDLE_CLIENT_WORK } from '~/utils/storefrontLoadingPolicy'
import { useSidePanelState } from '~/composables/useSidePanelState'
import { useQuickBuyOpenRequestState } from '~/composables/useQuickBuyOpenRequestState'

type GradientDockMenuShellIntent = 'sidebar' | 'chat' | 'quick-buy' | 'cart'

const mounted = ref(false)
const { openLeft } = useSidePanelState()
const { requestOpen: requestQuickBuyOpen } = useQuickBuyOpenRequestState()
let cancelDeferredMount: (() => void) | null = null
let dockModulePromise: Promise<{ default: Component }> | null = null

const loadGradientDockMenu = () => {
  dockModulePromise ??= import('./GradientDockMenu.vue')
  return dockModulePromise
}

const gradientDockMenuComponent = defineAsyncComponent(loadGradientDockMenu)

const activateDock = async () => {
  if (!mounted.value) {
    mounted.value = true
  }
  cancelDeferredMount?.()
  cancelDeferredMount = null

  await loadGradientDockMenu()
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

const openQuickBuyFromGlobalEvent = () => {
  requestQuickBuyOpen()
  void activateDock()
}

const handleShellIntent = async (intent: GradientDockMenuShellIntent) => {
  if (intent === 'sidebar') {
    openLeft()
    return
  }

  if (intent === 'quick-buy') {
    requestQuickBuyOpen()
    await activateDock()
    return
  }

  await activateDock()
  await nextTick()
  dispatchDockIntent(intent)
}

onMounted(() => {
  window.addEventListener('quickbuy:open-entry', openQuickBuyFromGlobalEvent)
  cancelDeferredMount = scheduleDeferredClientWork(() => {
    void activateDock()
  }, STOREFRONT_IDLE_CLIENT_WORK)
})

onBeforeUnmount(() => {
  window.removeEventListener('quickbuy:open-entry', openQuickBuyFromGlobalEvent)
  cancelDeferredMount?.()
  cancelDeferredMount = null
})
</script>
