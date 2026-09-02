import { computed } from 'vue'
import { useState } from '#imports'

const stateKey = 'wheelset-selection-assistant-modal-open-count'

export const useWheelsetSelectionAssistantModalState = () => {
  const openCount = useState<number>(stateKey, () => 0)

  const isOpen = computed(() => openCount.value > 0)

  const register = () => {
    openCount.value += 1
  }

  const unregister = () => {
    openCount.value = Math.max(0, openCount.value - 1)
  }

  return {
    isOpen,
    register,
    unregister,
  }
}
