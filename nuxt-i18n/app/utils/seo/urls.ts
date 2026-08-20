const trimTrailingSlash = (value: string): string => value.replace(/\/+$/, '')

export const toAbsoluteSeoUrl = (origin: string, path: string): string => {
  const base = trimTrailingSlash(String(origin || '').trim())
  const target = String(path || '').trim()

  if (!base) return target

  try {
    return new URL(target || '/', `${base}/`).toString()
  } catch {
    return target
  }
}

export const buildProductPath = (slug: string): string => {
  const normalizedSlug = String(slug || '').trim()
  if (!normalizedSlug) return ''
  return `/shop/${encodeURIComponent(normalizedSlug)}`
}
