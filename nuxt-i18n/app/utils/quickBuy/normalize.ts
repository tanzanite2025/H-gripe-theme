import type {
  QuickBuyFlow,
  QuickBuyFlowVersion,
  QuickBuyFlowTranslation,
  QuickBuyProductType,
  QuickBuyStep,
} from '~/utils/quickBuy/types'

export type QuickBuyRawRecord = Record<string, unknown>

export const asQuickBuyRecord = (value: unknown): QuickBuyRawRecord | null => {
  return value && typeof value === 'object' && !Array.isArray(value) ? value as QuickBuyRawRecord : null
}

export const asQuickBuyArray = (value: unknown): unknown[] => {
  return Array.isArray(value) ? value : []
}

export const asQuickBuyString = (value: unknown) => {
  if (typeof value === 'string') return value
  if (value === null || value === undefined) return ''
  return String(value)
}

export const asQuickBuyBoolean = (value: unknown, fallback = false) => {
  if (typeof value === 'boolean') return value
  if (typeof value === 'string') return value.toLowerCase() === 'true'
  if (value === null || value === undefined) return fallback
  return Boolean(value)
}

export const asQuickBuyNumber = (value: unknown, fallback = 0) => {
  const numberValue = Number(value)
  return Number.isFinite(numberValue) ? numberValue : fallback
}

export const extractQuickBuyPayload = (response: unknown): unknown => {
  let current = response
  for (let depth = 0; depth < 3; depth += 1) {
    const record = asQuickBuyRecord(current)
    if (!record || !('data' in record)) return current
    current = record.data
  }
  return current
}

const normalizeQuickBuyProductType = (value: unknown): QuickBuyProductType | null => {
  const record = asQuickBuyRecord(value)
  if (!record) return null
  const id = asQuickBuyNumber(record.id)
  const slug = asQuickBuyString(record.slug).trim()
  const name = asQuickBuyString(record.name || slug).trim()
  if (!id || !slug || !name) return null

  return {
    id,
    slug,
    name,
    imageUrl: asQuickBuyString(record.imageUrl ?? record.image_url).trim() || undefined,
    primary: asQuickBuyBoolean(record.primary),
  }
}

const normalizeQuickBuyStep = (value: unknown, index: number): QuickBuyStep | null => {
  const record = asQuickBuyRecord(value)
  if (!record) return null
  const id = asQuickBuyNumber(record.id, index + 1)
  const stepKey = asQuickBuyString(record.stepKey ?? record.step_key ?? record.key).trim()
  const productTypes = asQuickBuyArray(record.productTypes ?? record.product_types)
    .map(normalizeQuickBuyProductType)
    .filter((item): item is QuickBuyProductType => Boolean(item))
  const primaryProductType = productTypes.find(item => item.primary) || productTypes[0] || null
  const slug = asQuickBuyString(record.slug || primaryProductType?.slug || stepKey || `step-${index + 1}`).trim()
  const name = asQuickBuyString(record.name || record.label || record.title || primaryProductType?.name || slug).trim()
  if (!Number.isFinite(id) || !slug || !name) return null

  return {
    id,
    slug,
    name,
    stepKey: stepKey || slug,
    sortOrder: asQuickBuyNumber(record.sortOrder ?? record.sort_order, (index + 1) * 10),
    productTypes,
  }
}

const normalizeQuickBuyVersion = (value: unknown): QuickBuyFlowVersion | undefined => {
  const record = asQuickBuyRecord(value)
  if (!record) return undefined
  const id = asQuickBuyNumber(record.id)
  if (!id) return undefined

  return {
    id,
    versionNumber: asQuickBuyNumber(record.versionNumber ?? record.version_number, 1),
    status: asQuickBuyString(record.status),
    publishedAt: asQuickBuyString(record.publishedAt ?? record.published_at).trim() || undefined,
    startsAt: asQuickBuyString(record.startsAt ?? record.starts_at).trim() || undefined,
    endsAt: asQuickBuyString(record.endsAt ?? record.ends_at).trim() || undefined,
  }
}

const normalizeQuickBuyFlowTranslation = (value: unknown): QuickBuyFlowTranslation | null => {
  const record = asQuickBuyRecord(value)
  if (!record) return null
  const locale = asQuickBuyString(record.locale).trim()
  if (!locale) return null

  return {
    id: asQuickBuyNumber(record.id) || undefined,
    locale,
    helpText: asQuickBuyString(record.helpText ?? record.help_text).trim() || undefined,
  }
}

export const normalizeQuickBuyFlow = (value: unknown): QuickBuyFlow | null => {
  const record = asQuickBuyRecord(value)
  if (!record) return null
  const id = asQuickBuyNumber(record.id)
  const slug = asQuickBuyString(record.slug).trim()
  const name = asQuickBuyString(record.name || slug).trim()
  if (!id || !slug || !name) return null

  const steps = asQuickBuyArray(record.steps)
    .map(normalizeQuickBuyStep)
    .filter((item): item is QuickBuyStep => Boolean(item))
  const translations = asQuickBuyArray(record.translations ?? record.flowTranslations ?? record.flow_translations)
    .map(normalizeQuickBuyFlowTranslation)
    .filter((item): item is QuickBuyFlowTranslation => Boolean(item))

  return {
    id,
    slug,
    name,
    description: asQuickBuyString(record.description).trim() || undefined,
    helpText: asQuickBuyString(record.helpText ?? record.help_text).trim() || undefined,
    translations,
    entrySurface: asQuickBuyString(record.entrySurface ?? record.entry_surface).trim() || undefined,
    isEnabled: asQuickBuyBoolean(record.isEnabled ?? record.is_enabled, true),
    version: normalizeQuickBuyVersion(record.version),
    steps,
  }
}
