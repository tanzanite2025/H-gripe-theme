import { computed } from 'vue'
import { useAsyncData, useRuntimeConfig } from '#imports'
import { useI18n } from 'vue-i18n'
import { useApiRequest } from '~/composables/useApiRequest'
import { extractQuickBuyPayload, normalizeQuickBuyFlow } from '~/utils/quickBuy/normalize'
import { createStorefrontMediaContext } from '~/utils/storefrontMedia'
import type {
  QuickBuyConfig,
  QuickBuyFlow,
} from '~/utils/quickBuy/types'

interface QuickBuyFlowOptions {
  immediate?: boolean
}

export function useQuickBuyFlow(surface = 'dock', options: QuickBuyFlowOptions = {}) {
  const runtimeConfig = useRuntimeConfig()
  const mediaContext = createStorefrontMediaContext(runtimeConfig)
  const { request } = useApiRequest()
  const { locale } = useI18n()
  const immediate = options.immediate ?? true
  let refreshPromise: Promise<void> | null = null

  const { data, pending, error, refresh } = useAsyncData<QuickBuyFlow | null>(
    `mytheme-quick-buy-flow:${surface}`,
    async () => {
      try {
        const response = await request<unknown>('/quick-buy/flows/current', {
          headers: locale.value ? { 'Accept-Language': String(locale.value), accept: 'application/json' } : { accept: 'application/json' },
          params: {
            surface,
            locale: locale.value,
          }
        }, 'Failed to load quick buy flow')
        return normalizeQuickBuyFlow(extractQuickBuyPayload(response), mediaContext)
      } catch (fetchError) {
        console.warn('Failed to load quick buy flow:', fetchError)
        return null
      }
    },
    {
      server: false,
      default: () => null,
      immediate,
      watch: immediate ? [locale] : undefined,
      dedupe: 'defer',
    }
  )

  const quickBuyFlow = computed<QuickBuyFlow | null>(() => data.value ?? null)
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
