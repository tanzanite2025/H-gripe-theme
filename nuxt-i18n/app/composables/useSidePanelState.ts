import { computed, ref } from 'vue'
import { useOverlayBackStack } from '~/composables/useOverlayBackStack'
import { activateStorefrontClientOverlays } from '~/utils/clientOverlays'

const leftOpen = ref(false)
const leftEverOpened = ref(false)

export const useSidePanelState = () => {
  const overlayBackStack = useOverlayBackStack()

  const closeLeftState = () => {
    leftOpen.value = false
  }

  const openLeft = () => {
    activateStorefrontClientOverlays()
    leftEverOpened.value = true
    leftOpen.value = true
    overlayBackStack.open('account-sidebar', closeLeftState)
  }

  const closeLeft = () => {
    void overlayBackStack.close('account-sidebar')
    closeLeftState()
  }

  const toggleLeft = () => {
    if (leftOpen.value) {
      closeLeft()
    } else {
      openLeft()
    }
  }

  return {
    leftOpen: computed(() => leftOpen.value),
    leftEverOpened: computed(() => leftEverOpened.value),
    openLeft,
    closeLeft,
    toggleLeft,
  }
}
