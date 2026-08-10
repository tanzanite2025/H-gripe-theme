import { computed } from 'vue'
import { useRuntimeConfig } from '#imports'

const normalizeBaseUrl = (value?: string) => (value ? value.replace(/\/$/, '') : '')

export function usePublicApiBase() {
  const config = useRuntimeConfig()

  return computed(() => {
    const publicBase = normalizeBaseUrl((config.public as { apiBase?: string }).apiBase || '')
    if (publicBase) return publicBase

    const internalOrigin = normalizeBaseUrl((config as { apiInternalOrigin?: string }).apiInternalOrigin || '')
    return internalOrigin ? `${internalOrigin}/api/v1` : '/api/v1'
  })
}
