import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { normalizeShopProduct, type ShopProductsResult } from '~/composables/useShopProducts'
import {
  asQuickBuyArray,
  asQuickBuyBoolean,
  asQuickBuyNumber,
  asQuickBuyRecord,
  asQuickBuyString,
  extractQuickBuyPayload,
  normalizeQuickBuyFlow,
} from '~/utils/quickBuy/normalize'
import type {
  QuickBuySpecFilter,
  QuickBuySession,
  QuickBuySessionItem,
  QuickBuySessionSelectionInput,
  QuickBuySessionValidationIssue,
  QuickBuySessionValidationResult,
} from '~/utils/quickBuy/types'

const normalizeQuickBuySpecFilter = (value: unknown): QuickBuySpecFilter | null => {
  const record = asQuickBuyRecord(value)
  if (!record) return null
  const slug = asQuickBuyString(record.slug).trim()
  const name = asQuickBuyString(record.name ?? record.slug).trim()
  if (!slug || !name) return null
  return {
    id: asQuickBuyNumber(record.id),
    name,
    slug,
    unit: asQuickBuyString(record.unit).trim() || undefined,
    fieldType: asQuickBuyString(record.fieldType ?? record.field_type).trim() || 'text',
    presentation: asQuickBuyString(record.presentation).trim() || 'text',
    isVariantOption: Boolean(record.isVariantOption ?? record.is_variant_option),
    multiple: record.multiple !== false,
    values: asQuickBuyArray(record.values)
      .map((item) => asQuickBuyString(item).trim())
      .filter(Boolean),
  }
}

const normalizeQuickBuySpecFilters = (value: unknown): QuickBuySpecFilter[] => (
  asQuickBuyArray(value)
    .map(normalizeQuickBuySpecFilter)
    .filter((item): item is QuickBuySpecFilter => Boolean(item))
)

const normalizeQuickBuySessionValidationIssue = (value: unknown): QuickBuySessionValidationIssue | null => {
  const record = asQuickBuyRecord(value)
  if (!record) return null
  const code = asQuickBuyString(record.code).trim()
  const message = asQuickBuyString(record.message).trim()
  if (!code || !message) return null

  return {
    severity: asQuickBuyString(record.severity).trim() || 'error',
    code,
    message,
    stepKey: asQuickBuyString(record.stepKey ?? record.step_key).trim() || undefined,
    productId: asQuickBuyNumber(record.productId ?? record.product_id) || undefined,
    variantId: asQuickBuyNumber(record.variantId ?? record.variant_id) || null,
  }
}

const normalizeQuickBuySessionValidation = (value: unknown): QuickBuySessionValidationResult | undefined => {
  const record = asQuickBuyRecord(value)
  if (!record) return undefined
  return {
    valid: Boolean(record.valid),
    issues: asQuickBuyArray(record.issues)
      .map(normalizeQuickBuySessionValidationIssue)
      .filter((item): item is QuickBuySessionValidationIssue => Boolean(item)),
  }
}

const normalizeQuickBuySessionItem = (value: unknown): QuickBuySessionItem | null => {
  const record = asQuickBuyRecord(value)
  if (!record) return null
  const id = asQuickBuyNumber(record.id)
  const stepKey = asQuickBuyString(record.stepKey ?? record.step_key).trim()
  const productId = asQuickBuyNumber(record.productId ?? record.product_id)
  if (!id || !stepKey || !productId) return null

  return {
    id,
    stepKey,
    productId,
    variantId: asQuickBuyNumber(record.variantId ?? record.variant_id) || null,
    quantity: asQuickBuyNumber(record.quantity, 1),
    unitPriceSnapshot: asQuickBuyNumber(record.unitPriceSnapshot ?? record.unit_price_snapshot),
    currencySnapshot: asQuickBuyString(record.currencySnapshot ?? record.currency_snapshot).trim(),
    weightSnapshotG: asQuickBuyNumber(record.weightSnapshotG ?? record.weight_snapshot_g),
    productSnapshot: asQuickBuyRecord(record.productSnapshot ?? record.product_snapshot) || undefined,
    variantSnapshot: asQuickBuyRecord(record.variantSnapshot ?? record.variant_snapshot) || undefined,
    sortOrder: asQuickBuyNumber(record.sortOrder ?? record.sort_order),
  }
}

const normalizeQuickBuySession = (value: unknown): QuickBuySession | null => {
  const record = asQuickBuyRecord(value)
  if (!record) return null
  const sessionToken = asQuickBuyString(record.sessionToken ?? record.session_token).trim()
  if (!sessionToken) return null
  return {
    sessionToken,
    flowId: asQuickBuyNumber(record.flowId ?? record.flow_id),
    flowVersionId: asQuickBuyNumber(record.flowVersionId ?? record.flow_version_id),
    locale: asQuickBuyString(record.locale).trim(),
    marketCountry: asQuickBuyString(record.marketCountry ?? record.market_country).trim(),
    currency: asQuickBuyString(record.currency).trim(),
    status: asQuickBuyString(record.status).trim(),
    validationStatus: asQuickBuyString(record.validationStatus ?? record.validation_status).trim(),
    subtotalSnapshot: asQuickBuyNumber(record.subtotalSnapshot ?? record.subtotal_snapshot),
    weightSnapshotG: asQuickBuyNumber(record.weightSnapshotG ?? record.weight_snapshot_g),
    expiresAt: asQuickBuyString(record.expiresAt ?? record.expires_at).trim() || undefined,
    flow: normalizeQuickBuyFlow(record.flow) || undefined,
    items: asQuickBuyArray(record.items)
      .map(normalizeQuickBuySessionItem)
      .filter((item): item is QuickBuySessionItem => Boolean(item)),
    validation: normalizeQuickBuySessionValidation(record.validation),
  }
}

const serializeQuickBuySelection = (selection: QuickBuySessionSelectionInput) => ({
  step_key: selection.stepKey,
  product_id: selection.productId,
  variant_id: selection.variantId || null,
  quantity: selection.quantity,
})

const QUICK_BUY_ANONYMOUS_ID_KEY = 'tz_quick_buy_anonymous_id'

const readQuickBuyAnonymousId = (): string => {
  if (!import.meta.client) return ''
  try {
    const stored = window.localStorage.getItem(QUICK_BUY_ANONYMOUS_ID_KEY)
    if (stored) return stored
    const next = globalThis.crypto?.randomUUID?.() || `${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 12)}`
    const anonymousId = `quick_${next}`
    window.localStorage.setItem(QUICK_BUY_ANONYMOUS_ID_KEY, anonymousId)
    return anonymousId
  } catch {
    return ''
  }
}

export function useQuickBuySession(surface = 'dock') {
  const { request } = useApiRequest()
  const { locale } = useI18n()
  const { countryCode, displayCurrency, baseCurrency } = useStorefrontContext()
  const session = ref<QuickBuySession | null>(null)
  const pendingRequestCount = ref(0)
  const pending = computed(() => pendingRequestCount.value > 0)
  const error = ref<string | null>(null)
  let createPromise: Promise<QuickBuySession | null> | null = null
  let requestSequence = 0
  let mutationQueue = Promise.resolve()
  let lastSelectionMutationKey = ''
  let lastSelectionMutationPromise: Promise<QuickBuySession | null> | null = null

  const sessionToken = computed(() => session.value?.sessionToken || '')
  const beginRequest = () => {
    pendingRequestCount.value += 1
    const requestId = ++requestSequence
    error.value = null
    return requestId
  }
  const endRequest = () => {
    pendingRequestCount.value = Math.max(0, pendingRequestCount.value - 1)
  }
  const setRequestError = (requestId: number, err: unknown, fallback: string) => {
    if (requestId !== requestSequence) return
    error.value = err instanceof Error ? err.message : String(err)
    if (!error.value) error.value = fallback
  }
  const commitSession = (
    nextSession: QuickBuySession | null,
    expectedSessionToken = '',
  ) => {
    if (!nextSession) return null
    if (expectedSessionToken && nextSession.sessionToken !== expectedSessionToken) return null
    session.value = nextSession
    return nextSession
  }
  const enqueueMutation = <T>(operation: (requestId: number) => Promise<T>) => {
    const requestId = beginRequest()
    const queuedOperation = mutationQueue.then(
      () => operation(requestId),
      () => operation(requestId),
    )
    mutationQueue = queuedOperation.then(() => undefined, () => undefined)
    return queuedOperation.finally(endRequest)
  }
  const storefrontHeaders = () => {
    const headers: Record<string, string> = { Accept: 'application/json' }
    const anonymousId = readQuickBuyAnonymousId()
    if (anonymousId) headers['X-Anonymous-ID'] = anonymousId
    if (locale.value) headers['Accept-Language'] = String(locale.value)
    if (displayCurrency.value) headers['X-Display-Currency'] = displayCurrency.value
    if (countryCode.value && countryCode.value !== 'ZZ') headers['X-Market-Country'] = countryCode.value
    return headers
  }

  const createSession = async () => {
    if (session.value) return session.value
    if (createPromise) return createPromise

    const requestId = beginRequest()
    createPromise = request<unknown>('/quick-buy/sessions', {
      method: 'POST',
      headers: { ...storefrontHeaders(), 'Content-Type': 'application/json' },
      body: JSON.stringify({
        surface,
        locale: locale.value,
        market_country: countryCode.value !== 'ZZ' ? countryCode.value : '',
        currency: displayCurrency.value || 'USD',
      }),
    }, 'Unable to create QUICK session')
      .then((response) => {
        const nextSession = normalizeQuickBuySession(extractQuickBuyPayload(response))
        return commitSession(nextSession)
      })
      .catch((err) => {
        setRequestError(requestId, err, 'Unable to create QUICK session')
        return null
      })
      .finally(() => {
        endRequest()
        createPromise = null
      })

    return createPromise
  }

  const updateSessionSelections = async (selections: QuickBuySessionSelectionInput[]) => {
    const serializedSelections = selections.map(serializeQuickBuySelection)
    const selectionMutationKey = JSON.stringify(
      serializedSelections
        .slice()
        .sort((left, right) =>
          `${left.step_key}:${left.product_id}:${left.variant_id || 0}:${left.quantity}`
            .localeCompare(`${right.step_key}:${right.product_id}:${right.variant_id || 0}:${right.quantity}`),
        ),
    )
    if (lastSelectionMutationPromise && lastSelectionMutationKey === selectionMutationKey) {
      return lastSelectionMutationPromise
    }

    const mutationPromise = enqueueMutation(async (requestId) => {
      const currentSession = await createSession()
      if (!currentSession?.sessionToken) return null

      try {
        const response = await request<unknown>(
          `/quick-buy/sessions/${encodeURIComponent(currentSession.sessionToken)}/selections`,
          {
            method: 'PATCH',
            headers: { ...storefrontHeaders(), 'Content-Type': 'application/json' },
            body: JSON.stringify({ selections: serializedSelections }),
          },
          'Unable to update QUICK session',
        )
        const nextSession = normalizeQuickBuySession(extractQuickBuyPayload(response))
        return commitSession(nextSession, currentSession.sessionToken)
      } catch (err) {
        setRequestError(requestId, err, 'Unable to update QUICK session')
        return null
      }
    })
    lastSelectionMutationKey = selectionMutationKey
    lastSelectionMutationPromise = mutationPromise
    void mutationPromise.then(() => {
      if (lastSelectionMutationPromise === mutationPromise) {
        lastSelectionMutationKey = ''
        lastSelectionMutationPromise = null
      }
    }, () => {
      if (lastSelectionMutationPromise === mutationPromise) {
        lastSelectionMutationKey = ''
        lastSelectionMutationPromise = null
      }
    })
    return mutationPromise
  }

  const fetchStepCandidates = async (
    stepKey: string,
    params: { keyword?: string, page?: number, pageSize?: number, specFilters?: Record<string, string[]> } = {},
  ): Promise<ShopProductsResult | null> => {
    const currentSession = await createSession()
    if (!currentSession?.sessionToken) return null

    const query = new URLSearchParams()
    const keyword = String(params.keyword || '').trim()
    if (keyword) query.set('keyword', keyword)
    query.set('page', String(params.page && params.page > 0 ? params.page : 1))
    query.set('page_size', String(params.pageSize && params.pageSize > 0 ? params.pageSize : 12))
    query.set('locale', String(locale.value || currentSession.locale || 'en'))
    if (displayCurrency.value || currentSession.currency) {
      query.set('currency', displayCurrency.value || currentSession.currency)
    }
    const specFilters = Object.fromEntries(
      Object.entries(params.specFilters || {})
        .map(([slug, values]) => [slug, (values || []).map(value => String(value || '').trim()).filter(Boolean)])
        .filter(([, values]) => Array.isArray(values) && values.length > 0),
    )
    if (Object.keys(specFilters).length > 0) {
      query.set('spec_filters', JSON.stringify(specFilters))
    }

    const requestId = beginRequest()
    try {
      const response = await request<unknown>(
        `/quick-buy/sessions/${encodeURIComponent(currentSession.sessionToken)}/steps/${encodeURIComponent(stepKey)}/candidates?${query.toString()}`,
        { method: 'GET', headers: storefrontHeaders() },
        'Unable to load QUICK candidates',
      )
      const payload = extractQuickBuyPayload(response)
      const responseRecord = asQuickBuyRecord(response)
      const payloadRecord = asQuickBuyRecord(payload)
      const pagination = payloadRecord || responseRecord
      const stepRecord = asQuickBuyRecord(responseRecord?.step ?? payloadRecord?.step)
      return {
        items: extractQuickBuyCandidateProducts(response).map((item) => normalizeShopProduct(item, baseCurrency.value)),
        raw: response,
        page: asQuickBuyNumber(pagination?.page, params.page || 1),
        pageSize: asQuickBuyNumber(pagination?.page_size, params.pageSize || 12),
        total: asQuickBuyNumber(pagination?.total),
        hasMore: asQuickBuyBoolean(pagination?.has_more),
        quickBuyFilters: normalizeQuickBuySpecFilters(stepRecord?.filters),
      }
    } catch (err) {
      setRequestError(requestId, err, 'Unable to load QUICK candidates')
      return null
    } finally {
      endRequest()
    }
  }

  const validateSession = async () => {
    return enqueueMutation(async (requestId) => {
      const currentSession = await createSession()
      if (!currentSession?.sessionToken) return null

      try {
        const response = await request<unknown>(
          `/quick-buy/sessions/${encodeURIComponent(currentSession.sessionToken)}/validate`,
          { method: 'POST', headers: storefrontHeaders() },
          'Unable to validate QUICK session',
        )
        const nextSession = normalizeQuickBuySession(extractQuickBuyPayload(response))
        return commitSession(nextSession, currentSession.sessionToken)
      } catch (err) {
        setRequestError(requestId, err, 'Unable to validate QUICK session')
        return null
      }
    })
  }

  return {
    session,
    sessionToken,
    pending,
    error,
    createSession,
    fetchStepCandidates,
    updateSessionSelections,
    validateSession,
  }
}

const extractQuickBuyCandidateProducts = (response: unknown): unknown[] => {
  const payload = extractQuickBuyPayload(response)
  if (Array.isArray(payload)) return payload
  const record = asQuickBuyRecord(payload)
  if (!record) return []
  if (Array.isArray(record.products)) return record.products
  if (Array.isArray(record.items)) return record.items
  if (Array.isArray(record.data)) return record.data
  return []
}
