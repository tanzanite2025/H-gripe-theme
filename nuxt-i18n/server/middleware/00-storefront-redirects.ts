import { defineEventHandler, getMethod, getRequestURL, sendRedirect } from 'h3'

interface PublishedRedirectRule {
  source_path: string
  target_path: string
  status_code: 301 | 308
}

interface RedirectRuleResponse {
  data?: PublishedRedirectRule[]
}

const cacheLifetimeMS = 15_000
let cachedRules = new Map<string, PublishedRedirectRule>()
let cachedAt = 0
let refreshPromise: Promise<Map<string, PublishedRedirectRule>> | null = null

const bypassPath = (pathname: string): boolean => (
  pathname.startsWith('/api/')
  || pathname.startsWith('/_nuxt/')
  || pathname.startsWith('/_internal/')
  || pathname.startsWith('/icons/')
  || pathname.startsWith('/images/')
)

const normalizePath = (value: string): string => {
  if (value === '/') return '/'
  return `/${value.replace(/^\/+/, '')}`.replace(/\/+$/, '')
}

const appendQuery = (targetPath: string, query: string): string => {
  const normalizedQuery = String(query || '').replace(/^\?/, '')
  if (!normalizedQuery) return targetPath
  return `${targetPath}${targetPath.includes('?') ? '&' : '?'}${normalizedQuery}`
}

const loadPublishedRules = async (): Promise<Map<string, PublishedRedirectRule>> => {
  const now = Date.now()
  if (cachedAt > 0 && now - cachedAt < cacheLifetimeMS) return cachedRules
  if (refreshPromise) return refreshPromise

  refreshPromise = (async () => {
    const config = useRuntimeConfig()
    const origin = String(config.apiInternalOrigin || '').replace(/\/+$/, '')
    if (!origin) return cachedRules

    try {
      const response = await $fetch<RedirectRuleResponse>(`${origin}/api/v1/storefront/redirects`)
      const nextRules = new Map<string, PublishedRedirectRule>()
      for (const rule of response?.data || []) {
        if (!rule?.source_path || !rule.target_path) continue
        nextRules.set(normalizePath(rule.source_path), rule)
      }
      cachedRules = nextRules
      cachedAt = Date.now()
    } catch {
      // Keep the last successful policy in memory during a transient API failure.
    } finally {
      refreshPromise = null
    }
    return cachedRules
  })()

  return refreshPromise
}

export default defineEventHandler(async (event) => {
  const method = getMethod(event)
  if (method !== 'GET' && method !== 'HEAD') return

  const requestURL = getRequestURL(event)
  if (bypassPath(requestURL.pathname)) return

  const rule = (await loadPublishedRules()).get(normalizePath(requestURL.pathname))
  if (rule) {
    return sendRedirect(event, appendQuery(rule.target_path, requestURL.search), rule.status_code)
  }
})
