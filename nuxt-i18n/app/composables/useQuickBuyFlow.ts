import { computed } from 'vue'
import { useAsyncData } from '#imports'
import { useI18n } from 'vue-i18n'
import { usePublicApiBase } from '~/composables/usePublicApiBase'
import { extractQuickBuyPayload, normalizeQuickBuyFlow } from '~/utils/quickBuy/normalize'
import type {
  QuickBuyConfig,
  QuickBuyFlow,
} from '~/utils/quickBuy/types'

export function useQuickBuyFlow(surface = 'dock') {
  const apiBase = usePublicApiBase()
  const { locale } = useI18n()
  let refreshPromise: Promise<void> | null = null

  const { data, pending, error, refresh } = useAsyncData<QuickBuyFlow | null>(
    `mytheme-quick-buy-flow:${surface}`,
    async () => {
      if (!apiBase.value) return null
      try {
        const response = await $fetch<unknown>(`${apiBase.value}/quick-buy/flows/current`, {
          headers: locale.value ? { 'Accept-Language': String(locale.value), accept: 'application/json' } : { accept: 'application/json' },
          params: {
            surface,
            locale: locale.value,
          }
        })
        return normalizeQuickBuyFlow(extractQuickBuyPayload(response))
      } catch (fetchError) {
        console.warn('Failed to load quick buy flow:', fetchError)
        return null
      }
    },
    {
      server: false,
      default: () => null,
      watch: [apiBase, locale],
      dedupe: 'defer',
    }
  )

  const quickBuyFlow = computed<QuickBuyFlow | null>(() => data.value)
  const quickBuyFlowConfig = computed<QuickBuyConfig | null>(() => {
    if (!quickBuyFlow.value || quickBuyFlow.value.steps.length === 0) return null
    return {
      enabled: quickBuyFlow.value.isEnabled !== false,
      steps: quickBuyFlow.value.steps,
      flowHelpText: quickBuyFlow.value.helpText,
    }
  })
  const refreshQuickBuyFlow = async () => {
    if (refreshPromise) return refreshPromise
    refreshPromise = refresh({ dedupe: 'defer' }).finally(() => {
      refreshPromise = null
    })
    return refreshPromise
  }

  return {
    quickBuyFlow,
    quickBuyFlowConfig,
    pending,
    error,
    refresh: refreshQuickBuyFlow,
  }
}
