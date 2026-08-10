import { computed } from 'vue'
import { useAsyncData } from '#imports'
import { useI18n } from 'vue-i18n'
import { usePublicApiBase } from '~/composables/usePublicApiBase'
import type {
  QuickBuyConfig,
  QuickBuyFlow,
  QuickBuyFlowVersion,
  QuickBuyProductType,
  QuickBuyStep,
} from '~/utils/quickBuy/types'

type RawRecord = Record<string, unknown>

const asObject = (value: unknown): RawRecord | null => {
  return value && typeof value === 'object' && !Array.isArray(value) ? value as RawRecord : null
}

const asArray = (value: unknown): unknown[] => {
  return Array.isArray(value) ? value : []
}

const asString = (value: unknown) => {
  if (typeof value === 'string') return value
  if (value === null || value === undefined) return ''
  return String(value)
}

const asBoolean = (value: unknown, fallback = false) => {
  if (typeof value === 'boolean') return value
  if (typeof value === 'string') return value.toLowerCase() === 'true'
  if (value === null || value === undefined) return fallback
  return Boolean(value)
}

const asNumber = (value: unknown, fallback = 0) => {
  const numberValue = Number(value)
  return Number.isFinite(numberValue) ? numberValue : fallback
}

const normalizeProductType = (value: unknown): QuickBuyProductType | null => {
  const record = asObject(value)
  if (!record) return null
  const id = asNumber(record.id)
  const slug = asString(record.slug).trim()
  const name = asString(record.name || slug).trim()
  if (!id || !slug || !name) return null

  return {
    id,
    slug,
    name,
    imageUrl: asString(record.imageUrl ?? record.image_url).trim() || undefined,
    primary: asBoolean(record.primary)
  }
}

const normalizeStep = (value: unknown, index: number): QuickBuyStep | null => {
  const record = asObject(value)
  if (!record) return null
  const id = asNumber(record.id, index + 1)
  const stepKey = asString(record.stepKey ?? record.step_key ?? record.key).trim()
  const productTypes = asArray(record.productTypes ?? record.product_types)
    .map(normalizeProductType)
    .filter((item): item is QuickBuyProductType => Boolean(item))
  const primaryProductType = productTypes.find(item => item.primary) || productTypes[0] || null
  const slug = asString(record.slug || primaryProductType?.slug || stepKey || `step-${index + 1}`).trim()
  const name = asString(record.name || record.label || record.title || primaryProductType?.name || slug).trim()
  if (!Number.isFinite(id) || !slug || !name) return null

  return {
    id,
    slug,
    name,
    stepKey: stepKey || slug,
    description: asString(record.description).trim() || undefined,
    helpText: asString(record.helpText ?? record.help_text).trim() || undefined,
    sortOrder: asNumber(record.sortOrder ?? record.sort_order, (index + 1) * 10),
    selectionMode: asString((record.selectionMode ?? record.selection_mode) || 'single'),
    isRequired: asBoolean(record.isRequired ?? record.is_required, true),
    minSelect: asNumber(record.minSelect ?? record.min_select),
    maxSelect: asNumber(record.maxSelect ?? record.max_select, 1),
    defaultQuantity: asNumber(record.defaultQuantity ?? record.default_quantity, 1),
    allowSkip: asBoolean(record.allowSkip ?? record.allow_skip),
    productTypes
  }
}

const normalizeVersion = (value: unknown): QuickBuyFlowVersion | undefined => {
  const record = asObject(value)
  if (!record) return undefined
  const id = asNumber(record.id)
  if (!id) return undefined

  return {
    id,
    versionNumber: asNumber(record.versionNumber ?? record.version_number, 1),
    status: asString(record.status),
    publishedAt: asString(record.publishedAt ?? record.published_at).trim() || undefined,
    startsAt: asString(record.startsAt ?? record.starts_at).trim() || undefined,
    endsAt: asString(record.endsAt ?? record.ends_at).trim() || undefined,
  }
}

const extractPayload = (response: unknown): unknown => {
  const record = asObject(response)
  if (!record) return response
  return record.data ?? response
}

const normalizeFlow = (value: unknown): QuickBuyFlow | null => {
  const record = asObject(value)
  if (!record) return null
  const id = asNumber(record.id)
  const slug = asString(record.slug).trim()
  const name = asString(record.name || slug).trim()
  if (!id || !slug || !name) return null

  const steps = asArray(record.steps)
    .map(normalizeStep)
    .filter((item): item is QuickBuyStep => Boolean(item))

  return {
    id,
    slug,
    name,
    description: asString(record.description).trim() || undefined,
    entrySurface: asString(record.entrySurface ?? record.entry_surface).trim() || undefined,
    isEnabled: asBoolean(record.isEnabled ?? record.is_enabled, true),
    version: normalizeVersion(record.version),
    steps,
  }
}

export function useQuickBuyFlow(surface = 'dock') {
  const apiBase = usePublicApiBase()
  const { locale } = useI18n()

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
        return normalizeFlow(extractPayload(response))
      } catch (fetchError) {
        console.warn('Failed to load quick buy flow:', fetchError)
        return null
      }
    },
    {
      server: false,
      default: () => null,
      watch: [apiBase, locale],
    }
  )

  const quickBuyFlow = computed<QuickBuyFlow | null>(() => data.value)
  const quickBuyFlowConfig = computed<QuickBuyConfig | null>(() => {
    if (!quickBuyFlow.value || quickBuyFlow.value.steps.length === 0) return null
    return {
      enabled: quickBuyFlow.value.isEnabled !== false,
      steps: quickBuyFlow.value.steps,
    }
  })

  return {
    quickBuyFlow,
    quickBuyFlowConfig,
    pending,
    error,
    refresh,
  }
}
