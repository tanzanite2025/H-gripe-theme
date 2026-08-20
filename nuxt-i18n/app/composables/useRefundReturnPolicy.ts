import { computed, type Ref } from 'vue'
import { useAsyncData, useI18n } from '#imports'
import { usePublicApiBase } from '~/composables/usePublicApiBase'
import { refundReturnPolicyFallback } from '~/data/refundReturnPolicyFallback'
import type {
  RefundReturnPolicy,
  RefundReturnPolicyImage,
  RefundReturnPolicyResponse,
  RefundReturnPolicySection,
} from '~/types/refundReturnPolicy'

type LocaleInput = Ref<string> | string | undefined

const normalizeLocale = (value: unknown): string => {
  const raw = String(value || '').trim().toLowerCase().replace(/-/g, '_')
  if (!raw) return 'en'
  if (raw === 'zh' || raw === 'zh_cn' || raw === 'zh_hans') return 'zh_cn'
  return raw.split('_')[0] || 'en'
}

const asString = (value: unknown): string => typeof value === 'string' ? value.trim() : ''

const normalizeImage = (value: unknown): RefundReturnPolicyImage | undefined => {
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

const normalizeSection = (value: unknown, index: number): RefundReturnPolicySection | null => {
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

export const normalizeRefundReturnPolicy = (value: unknown): RefundReturnPolicy => {
  const source = value && typeof value === 'object' && !Array.isArray(value)
    ? value as Record<string, unknown>
    : {}
  const sections = Array.isArray(source.sections)
    ? source.sections
      .map(normalizeSection)
      .filter((section): section is RefundReturnPolicySection => Boolean(section))
    : []

  return {
    title: asString(source.title) || refundReturnPolicyFallback.title,
    intro: asString(source.intro),
    sections,
    contact_label: asString(source.contact_label),
    contact_url: asString(source.contact_url),
    updated_at: asString(source.updated_at),
  }
}

export function useRefundReturnPolicy(localeInput?: LocaleInput) {
  const apiBase = usePublicApiBase()
  const { locale } = useI18n()
  const requestedLocale = computed(() => normalizeLocale(
    typeof localeInput === 'string'
      ? localeInput
      : localeInput?.value ?? locale.value,
  ))

  const { data, pending, error, refresh } = useAsyncData<RefundReturnPolicyResponse | null>(
    () => `refund-return-policy-${requestedLocale.value}`,
    () => $fetch<RefundReturnPolicyResponse>(
      `${apiBase.value}/settings/refund-return-policy`,
      {
        headers: { accept: 'application/json' },
        query: { locale: requestedLocale.value },
      },
    ),
    {
      default: () => null,
      watch: [requestedLocale],
    },
  )

  const remotePolicy = computed<RefundReturnPolicy | null>(() => (
    data.value?.policy
      ? normalizeRefundReturnPolicy(data.value.policy)
      : null
  ))

  const policy = computed<RefundReturnPolicy>(() => remotePolicy.value || refundReturnPolicyFallback)
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
