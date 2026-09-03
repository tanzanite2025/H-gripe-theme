import { computed, type Ref } from 'vue'
import { useAsyncData, useI18n } from '#imports'
import { useApiRequest } from '~/composables/useApiRequest'
import { refundCancellationPolicyFallback } from '~/data/refundCancellationPolicyFallback'
import type {
  RefundCancellationPolicy,
  RefundCancellationPolicyImage,
  RefundCancellationPolicyResponse,
  RefundCancellationPolicySection,
} from '~/types/refundCancellationPolicy'
import { normalizeStorefrontLocaleCode } from '~/utils/storefrontLocales'

type LocaleInput = Ref<string> | string | undefined

const normalizeLocale = (value: unknown): string => {
  return normalizeStorefrontLocaleCode(value) || String(value || '').trim().toLowerCase().replace(/-/g, '_') || 'en'
}

const asString = (value: unknown): string => typeof value === 'string' ? value.trim() : ''

const normalizeImage = (value: unknown): RefundCancellationPolicyImage | undefined => {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return undefined
  const record = value as Record<string, unknown>
  const url = asString(record.url)
  if (!url) return undefined
  return {
    url,
    alt: asString(record.alt),
    caption: asString(record.caption),
  }
}

const normalizeSection = (value: unknown, index: number): RefundCancellationPolicySection | null => {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return null
  const record = value as Record<string, unknown>
  const title = asString(record.title)
  if (!title) return null
  const bullets = Array.isArray(record.bullets)
    ? record.bullets.map(asString).filter(Boolean)
    : []
  return {
    id: asString(record.id) || `section-${index + 1}`,
    title,
    body: asString(record.body),
    bullets,
    image: normalizeImage(record.image),
  }
}

export const normalizeRefundCancellationPolicy = (value: unknown): RefundCancellationPolicy => {
  const source = value && typeof value === 'object' && !Array.isArray(value)
    ? value as Record<string, unknown>
    : {}
  const sections = Array.isArray(source.sections)
    ? source.sections
      .map(normalizeSection)
      .filter((section): section is RefundCancellationPolicySection => Boolean(section))
    : []

  return {
    title: asString(source.title) || refundCancellationPolicyFallback.title,
    intro: asString(source.intro),
    sections,
    contact_label: asString(source.contact_label),
    contact_url: asString(source.contact_url),
    updated_at: asString(source.updated_at),
  }
}

export function useRefundCancellationPolicy(localeInput?: LocaleInput) {
  const { request } = useApiRequest()
  const { locale } = useI18n()
  const requestedLocale = computed(() => normalizeLocale(
    typeof localeInput === 'string'
      ? localeInput
      : localeInput?.value ?? locale.value,
  ))

  const { data, pending, error, refresh } = useAsyncData<RefundCancellationPolicyResponse | null>(
    () => `refund-cancellation-policy-${requestedLocale.value}`,
    () => request<RefundCancellationPolicyResponse>(
      '/content/refund-cancellation-policy',
      {
        headers: { accept: 'application/json' },
        query: { locale: requestedLocale.value },
      },
      'Failed to load refund cancellation policy',
    ),
    {
      default: () => null,
      watch: [requestedLocale],
    },
  )

  const remotePolicy = computed<RefundCancellationPolicy | null>(() => (
    data.value?.policy
      ? normalizeRefundCancellationPolicy(data.value.policy)
      : null
  ))

  const policy = computed<RefundCancellationPolicy>(() => remotePolicy.value || refundCancellationPolicyFallback)
  const hasRemotePolicy = computed(() => Boolean(remotePolicy.value))
  const isFallbackPolicy = computed(() => Boolean(data.value?.fallback))

  return {
    policy,
    pending,
    error,
    refresh,
    requestedLocale,
    hasRemotePolicy,
    isFallbackPolicy,
  }
}
