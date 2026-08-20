export const adminApiBaseUrl = (): string => {
  const configured = String(import.meta.env.VITE_API_BASE_URL || '').trim().replace(/\/+$/, '')
  if (/^https?:\/\//i.test(configured)) {
    try {
      return new URL(configured).origin
    } catch {
      return configured
    }
  }
  if (typeof window !== 'undefined' && window.location?.origin) return window.location.origin
  return ''
}

export const adminApiUrl = (path: string): string => {
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  const base = adminApiBaseUrl()
  return base ? `${base}${normalizedPath}` : normalizedPath
}
