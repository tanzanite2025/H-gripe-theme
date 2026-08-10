import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { normalizeShopProduct, type ShopProductsResult } from '~/composables/useShopProducts'
import type {
  QuickBuyFlow,
  QuickBuySession,
  QuickBuySessionItem,
  QuickBuySessionSelectionInput,
  QuickBuySessionValidationIssue,
  QuickBuySessionValidationResult,
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

const asNumber = (value: unknown, fallback = 0) => {
  const numberValue = Number(value)
  return Number.isFinite(numberValue) ? numberValue : fallback
}

const extractPayload = (response: unknown): unknown => {
  let current = response
  for (let depth = 0; depth < 3; depth += 1) {
    const record = asObject(current)
    if (!record || !('data' in record)) return current
    current = record.data
  }
  return current
}

const normalizeValidationIssue = (value: unknown): QuickBuySessionValidationIssue | null => {
  const record = asObject(value)
  if (!record) return null
  const code = asString(record.code).trim()
  const message = asString(record.message).trim()
  if (!code || !message) return null

  return {
    severity: asString(record.severity).trim() || 'error',
    code,
    message,
    stepKey: asString(record.stepKey ?? record.step_key).trim() || undefined,
    productId: asNumber(record.productId ?? record.product_id) || undefined,
    variantId: asNumber(record.variantId ?? record.variant_id) || null,
  }
}

const normalizeValidation = (value: unknown): QuickBuySessionValidationResult | undefined => {
  const record = asObject(value)
  if (!record) return undefined
  return {
    valid: Boolean(record.valid),
    issues: asArray(record.issues)
      .map(normalizeValidationIssue)
      .filter((item): item is QuickBuySessionValidationIssue => Boolean(item)),
  }
}

const normalizeSessionItem = (value: unknown): QuickBuySessionItem | null => {
  const record = asObject(value)
  if (!record) return null
  const id = asNumber(record.id)
  const stepKey = asString(record.stepKey ?? record.step_key).trim()
  const productId = asNumber(record.productId ?? record.product_id)
  if (!id || !stepKey || !productId) return null

  return {
    id,
    stepKey,
    productId,
    variantId: asNumber(record.variantId ?? record.variant_id) || null,
    quantity: asNumber(record.quantity, 1),
    unitPriceSnapshot: asNumber(record.unitPriceSnapshot ?? record.unit_price_snapshot),
    currencySnapshot: asString(record.currencySnapshot ?? record.currency_snapshot).trim(),
    weightSnapshotG: asNumber(record.weightSnapshotG ?? record.weight_snapshot_g),
    productSnapshot: asObject(record.productSnapshot ?? record.product_snapshot) || undefined,
    variantSnapshot: asObject(record.variantSnapshot ?? record.variant_snapshot) || undefined,
    sortOrder: asNumber(record.sortOrder ?? record.sort_order),
  }
}

const normalizeSession = (value: unknown): QuickBuySession | null => {
  const record = asObject(value)
  if (!record) return null
  const sessionToken = asString(record.sessionToken ?? record.session_token).trim()
  if (!sessionToken) return null
  const flowRecord = asObject(record.flow)

  return {
    sessionToken,
    flowId: asNumber(record.flowId ?? record.flow_id),
    flowVersionId: asNumber(record.flowVersionId ?? record.flow_version_id),
    locale: asString(record.locale).trim(),
    marketCountry: asString(record.marketCountry ?? record.market_country).trim(),
    currency: asString(record.currency).trim(),
    status: asString(record.status).trim(),
    validationStatus: asString(record.validationStatus ?? record.validation_status).trim(),
    subtotalSnapshot: asNumber(record.subtotalSnapshot ?? record.subtotal_snapshot),
    weightSnapshotG: asNumber(record.weightSnapshotG ?? record.weight_snapshot_g),
    expiresAt: asString(record.expiresAt ?? record.expires_at).trim() || undefined,
    flow: flowRecord ? flowRecord as unknown as QuickBuyFlow : undefined,
    items: asArray(record.items)
      .map(normalizeSessionItem)
      .filter((item): item is QuickBuySessionItem => Boolean(item)),
    validation: normalizeValidation(record.validation),
  }
}

const selectionPayload = (selection: QuickBuySessionSelectionInput) => ({
  step_key: selection.stepKey,
  product_id: selection.productId,
  variant_id: selection.variantId || null,
  quantity: selection.quantity,
})

export function useQuickBuySession(surface = 'dock') {
  const { request } = useApiRequest()
  const { locale } = useI18n()
  const { countryCode, displayCurrency, baseCurrency } = useStorefrontContext()
  const session = ref<QuickBuySession | null>(null)
  const pending = ref(false)
  const error = ref<string | null>(null)
  let createPromise: Promise<QuickBuySession | null> | null = null

  const sessionToken = computed(() => session.value?.sessionToken || '')
  const storefrontHeaders = () => {
    const headers: Record<string, string> = { Accept: 'application/json' }
    if (locale.value) headers['Accept-Language'] = String(locale.value)
    if (displayCurrency.value) headers['X-Display-Currency'] = displayCurrency.value
    if (countryCode.value && countryCode.value !== 'ZZ') headers['X-Market-Country'] = countryCode.value
    return headers
  }

  const createSession = async () => {
    if (session.value) return session.value
    if (createPromise) return createPromise

    pending.value = true
    error.value = null
    createPromise = request<unknown>('/quick-buy/sessions', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
      body: JSON.stringify({
        surface,
        locale: locale.value,
        market_country: countryCode.value !== 'ZZ' ? countryCode.value : '',
        currency: displayCurrency.value || 'USD',
      }),
    }, 'Unable to create QUICK session')
      .then((response) => {
        const nextSession = normalizeSession(extractPayload(response))
        session.value = nextSession
        return nextSession
      })
      .catch((err) => {
        error.value = err instanceof Error ? err.message : String(err)
        return null
      })
      .finally(() => {
        pending.value = false
        createPromise = null
      })

    return createPromise
  }

  const updateSelections = async (selections: QuickBuySessionSelectionInput[]) => {
    const currentSession = await createSession()
    if (!currentSession?.sessionToken) return null

    pending.value = true
    error.value = null
    try {
      const response = await request<unknown>(
        `/quick-buy/sessions/${encodeURIComponent(currentSession.sessionToken)}/selections`,
        {
          method: 'PATCH',
          headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
          body: JSON.stringify({ selections: selections.map(selectionPayload) }),
        },
        'Unable to update QUICK session',
      )
      const nextSession = normalizeSession(extractPayload(response))
      session.value = nextSession
      return nextSession
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err)
      return null
    } finally {
      pending.value = false
    }
  }

  const fetchStepCandidates = async (
    stepKey: string,
    params: { keyword?: string, page?: number, pageSize?: number } = {},
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

    pending.value = true
    error.value = null
    try {
      const response = await request<unknown>(
        `/quick-buy/sessions/${encodeURIComponent(currentSession.sessionToken)}/steps/${encodeURIComponent(stepKey)}/candidates?${query.toString()}`,
        { method: 'GET', headers: storefrontHeaders() },
        'Unable to load QUICK candidates',
      )
      return {
        items: extractCandidateProducts(response).map((item) => normalizeShopProduct(item, baseCurrency.value)),
        raw: response,
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err)
      return null
    } finally {
      pending.value = false
    }
  }

  const validateSession = async () => {
    const currentSession = await createSession()
    if (!currentSession?.sessionToken) return null

    pending.value = true
    error.value = null
    try {
      const response = await request<unknown>(
        `/quick-buy/sessions/${encodeURIComponent(currentSession.sessionToken)}/validate`,
        { method: 'POST', headers: { Accept: 'application/json' } },
        'Unable to validate QUICK session',
      )
      const nextSession = normalizeSession(extractPayload(response))
      session.value = nextSession
      return nextSession
    } catch (err) {
      error.value = err instanceof Error ? err.message : String(err)
      return null
    } finally {
      pending.value = false
    }
  }

  return {
    session,
    sessionToken,
    pending,
    error,
    createSession,
    fetchStepCandidates,
    updateSelections,
    validateSession,
  }
}

const extractCandidateProducts = (response: unknown): unknown[] => {
  const payload = extractPayload(response)
  if (Array.isArray(payload)) return payload
  const record = asObject(payload)
  if (!record) return []
  if (Array.isArray(record.products)) return record.products
  if (Array.isArray(record.items)) return record.items
  if (Array.isArray(record.data)) return record.data
  return []
}
