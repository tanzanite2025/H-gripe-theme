import { computed } from 'vue'
import { useState } from '#imports'

const stateKey = 'quick-buy-open-request-count'

export const useQuickBuyOpenRequestState = () => {
  const openRequestCount = useState<number>(stateKey, () => 0)

  const hasPendingOpenRequest = computed(() => openRequestCount.value > 0)

  const requestOpen = () => {
    openRequestCount.value += 1
  }

  const consumeOpenRequest = () => {
    openRequestCount.value = Math.max(0, openRequestCount.value - 1)
  }

  return {
    openRequestCount,
    hasPendingOpenRequest,
    requestOpen,
    consumeOpenRequest,
  }
}
