import { computed, ref } from 'vue'
import { useApiRequest } from '~/composables/useApiRequest'
import type { CheckoutPaymentOption, PaymentMethodAvailability } from '~/types/payment'

type PaymentMethodsEnvelope = {
  data?: PaymentMethodAvailability[] | { data?: PaymentMethodAvailability[] }
}

const isRecord = (value: unknown): value is Record<string, unknown> =>
  Boolean(value) && typeof value === 'object' && !Array.isArray(value)

const unwrapPaymentMethods = (payload: unknown): PaymentMethodAvailability[] => {
  if (Array.isArray(payload)) return payload as PaymentMethodAvailability[]
  if (!isRecord(payload)) return []

  const envelope = payload as PaymentMethodsEnvelope
  if (Array.isArray(envelope.data)) {
    return envelope.data
  }
  if (isRecord(envelope.data) && Array.isArray(envelope.data.data)) {
    return envelope.data.data
  }
  return []
}

const normalizePaymentMethodReason = (method: PaymentMethodAvailability) =>
  String(method.unavailableReason || method.unavailable_reason || '').trim()

const methodAvailabilityKey = (method: PaymentMethodAvailability) =>
  String(method.provider || method.code || '').trim().toLowerCase()

const methodCodeKey = (method: PaymentMethodAvailability) =>
  String(method.code || '').trim().toLowerCase()

const toCheckoutPaymentOption = (method: PaymentMethodAvailability): CheckoutPaymentOption => {
  const id = String(method.code || method.provider || method.id || '').trim()
  return {
    id,
    code: method.code,
    provider: method.provider,
    icon: method.icon,
    title: method.name || id,
    subtitle: method.provider || method.code || '',
    description: method.description || '',
    enabled: method.enabled,
    available: method.available,
    unavailableReason: normalizePaymentMethodReason(method) || undefined,
  }
}

export const usePaymentMethods = () => {
  const { request } = useApiRequest()
  const paymentMethods = ref<PaymentMethodAvailability[]>([])
  const paymentMethodsLoading = ref(false)
  const paymentMethodsError = ref('')

  const paymentMethodOptions = computed<CheckoutPaymentOption[]>(() =>
    paymentMethods.value
      .filter((method) => method && (method.code || method.provider || method.id))
      .map(toCheckoutPaymentOption)
  )

  const availablePaymentMethods = computed(() =>
    paymentMethods.value.filter((method) => method.enabled !== false && method.available !== false)
  )

  const availabilityByCode = computed(() => {
    const entries: Record<string, { available: boolean; reason: string }> = {}
    for (const method of paymentMethods.value) {
      const keys = [methodAvailabilityKey(method), methodCodeKey(method)].filter(Boolean)
      for (const key of keys) {
        entries[key] = {
          available: method.enabled !== false && method.available !== false,
          reason: normalizePaymentMethodReason(method),
        }
      }
    }
    return entries
  })

  const isPaymentMethodAvailable = (codeOrProvider: string) => {
    const key = String(codeOrProvider || '').trim().toLowerCase()
    if (!key) return false
    const availability = availabilityByCode.value[key]
    return availability ? availability.available : true
  }

  const loadPaymentMethods = async (country?: string) => {
    if (paymentMethodsLoading.value) return paymentMethods.value

    paymentMethodsLoading.value = true
    paymentMethodsError.value = ''
    try {
      const query = new URLSearchParams()
      if (country?.trim()) {
        query.set('country', country.trim().toUpperCase())
      }
      const path = `/payment/methods${query.toString() ? `?${query.toString()}` : ''}`
      const response = await request<unknown>(path, { method: 'GET' }, 'Failed to load payment methods')
      paymentMethods.value = unwrapPaymentMethods(response)
      return paymentMethods.value
    } catch (error) {
      paymentMethodsError.value = error instanceof Error ? error.message : 'Failed to load payment methods'
      paymentMethods.value = []
      return paymentMethods.value
    } finally {
      paymentMethodsLoading.value = false
    }
  }

  return {
    paymentMethods,
    paymentMethodOptions,
    availablePaymentMethods,
    availabilityByCode,
    paymentMethodsLoading,
    paymentMethodsError,
    loadPaymentMethods,
    isPaymentMethodAvailable,
  }
}
