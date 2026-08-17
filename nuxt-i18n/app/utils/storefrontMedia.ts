export interface StorefrontMediaRuntimeConfig {
  apiInternalOrigin?: unknown
  imageInternalOrigin?: unknown
  additionalOrigins?: unknown[]
  public?: {
    apiBase?: unknown
    siteUrl?: unknown
  }
}

export interface StorefrontMediaContext {
  knownOrigins: ReadonlySet<string>
}

const syntheticBaseUrl = 'https://storefront.invalid/'

const normalizeOrigin = (value: unknown): string | null => {
  const candidate = String(value || '').trim()
  if (!candidate) return null

  try {
    const url = new URL(candidate)
    if (!['http:', 'https:'].includes(url.protocol)) return null
    if (url.username || url.password || url.hostname.includes('*')) return null
    return url.origin
  } catch {
    return null
  }
}

const relativeUrl = (url: URL): string => `${url.pathname}${url.search}${url.hash}`

const normalizeMediaRecord = (
  value: unknown,
  fields: string[],
  context: StorefrontMediaContext,
): Record<string, unknown> | null => {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return null

  const record = { ...(value as Record<string, unknown>) }
  for (const field of fields) {
    if (field in record) {
      record[field] = normalizeStorefrontMediaUrl(record[field], context)
    }
  }
  return record
}

export const createStorefrontMediaContext = (
  runtimeConfig: StorefrontMediaRuntimeConfig,
  browserOrigin = '',
): StorefrontMediaContext => {
  const origins = [
    runtimeConfig.public?.apiBase,
    runtimeConfig.public?.siteUrl,
    runtimeConfig.apiInternalOrigin,
    runtimeConfig.imageInternalOrigin,
    ...(runtimeConfig.additionalOrigins || []),
    browserOrigin,
  ]
    .map(normalizeOrigin)
    .filter((origin): origin is string => Boolean(origin))

  return {
    knownOrigins: new Set(origins),
  }
}

/**
 * Backend media contracts use /uploads as the first-party asset path.
 * Normalize those URLs before they reach SSR HTML, even when the backend
 * serialized an internal Docker/API origin into the absolute URL.
 */
export const normalizeStorefrontMediaUrl = (
  value: unknown,
  context: StorefrontMediaContext = { knownOrigins: new Set<string>() },
): string => {
  const candidate = String(value || '').trim()
  if (!candidate) return ''
  if (/^(?:data|blob):/i.test(candidate)) return candidate

  try {
    const url = new URL(candidate, syntheticBaseUrl)
    if (url.origin === 'null') return candidate
    if (url.origin === 'https://storefront.invalid') return relativeUrl(url)
    if (!['http:', 'https:'].includes(url.protocol)) return candidate
    if (url.username || url.password) return candidate

    if (context.knownOrigins.has(url.origin)) {
      return relativeUrl(url)
    }

    return url.href
  } catch {
    return candidate
  }
}

/**
 * Normalize the media-bearing fields in a product API response before it is
 * retained by SSR state or passed to a product view.
 */
export const normalizeStorefrontProductMedia = <T extends object>(
  value: T,
  context: StorefrontMediaContext,
): T => {
  const record = normalizeMediaRecord(
    value,
    ['thumbnail', 'featured_image'],
    context,
  )
  if (!record) return value

  if (Array.isArray(record.media)) {
    record.media = record.media
      .map(item => normalizeMediaRecord(item, ['url', 'thumbnail_url', 'poster_url'], context))
      .filter((item): item is Record<string, unknown> => Boolean(item))
  }

  if (Array.isArray(record.variant_option_values)) {
    record.variant_option_values = record.variant_option_values
      .map(item => normalizeMediaRecord(item, ['swatch_url'], context))
      .filter((item): item is Record<string, unknown> => Boolean(item))
  }

  if (record.brand && typeof record.brand === 'object' && !Array.isArray(record.brand)) {
    record.brand = normalizeMediaRecord(record.brand, ['logo_url'], context)
  }

  if (
    record.product_specification_template
    && typeof record.product_specification_template === 'object'
    && !Array.isArray(record.product_specification_template)
  ) {
    record.product_specification_template = normalizeMediaRecord(
      record.product_specification_template,
      ['image_url'],
      context,
    )
  }

  return record as T
}
