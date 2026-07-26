import type { PageFaqData } from './types'
import { normalizeFaqRoutePath } from './routing'

function hasFaqContent(page?: PageFaqData): page is PageFaqData {
  return Boolean(page?.categories?.some(category => category.items.length > 0))
}

function hasAnyFaqContent(pages?: PageFaqData[]): pages is PageFaqData[] {
  return Boolean(pages?.some(page => hasFaqContent(page)))
}

function getBackendFaqLocale() {
  try {
    const { locale } = useI18n()
    const currentLocale = String(locale.value || 'en').trim().toLowerCase().replace(/-/g, '_')
    if (['zh', 'zh_cn', 'zh_hans', 'zh_sg'].includes(currentLocale)) return 'zh_cn'
    return currentLocale || 'en'
  } catch {
    return 'en'
  }
}

function getFaqApiBase() {
  const config = useRuntimeConfig()
  return (config.public as { apiBase?: string }).apiBase || '/api/v1'
}

function getFetchStatusCode(error: unknown): number | undefined {
  const candidate = error as {
    status?: number
    statusCode?: number
    response?: { status?: number; statusCode?: number }
    data?: { status?: number; statusCode?: number }
  } | null

  return candidate?.statusCode
    || candidate?.status
    || candidate?.response?.statusCode
    || candidate?.response?.status
    || candidate?.data?.statusCode
    || candidate?.data?.status
}

function logFaqFetchError(message: string, error: unknown) {
  if (getFetchStatusCode(error) === 404) return
  console.error(message, error)
}

/**
 * Fetch FAQ data for a specific page from Go backend.
 */
export async function fetchFaqData(pageId: string): Promise<PageFaqData | null> {
  try {
    const structured = await $fetch<{ page?: PageFaqData }>(`${getFaqApiBase()}/content/faq-pages/${pageId}`, {
      query: { locale: getBackendFaqLocale() }
    })
    if (hasFaqContent(structured.page)) return structured.page
  } catch (error) {
    logFaqFetchError('Failed to fetch structured FAQs from Go backend:', error)
  }

  return null
}

export async function fetchFaqDataByRoutePath(routePath: string): Promise<PageFaqData | null> {
  const normalizedPath = normalizeFaqRoutePath(routePath)

  try {
    const structured = await $fetch<{ page?: PageFaqData }>(`${getFaqApiBase()}/content/faq-pages/by-route`, {
      query: { route_path: normalizedPath, locale: getBackendFaqLocale() }
    })
    if (hasFaqContent(structured.page)) return structured.page
  } catch (error) {
    logFaqFetchError('Failed to fetch structured FAQ by route from Go backend:', error)
  }

  return null
}

/**
 * Fetch all FAQ data from Go backend.
 */
export async function fetchAllFaqData(): Promise<PageFaqData[]> {
  try {
    const structured = await $fetch<{ pages?: PageFaqData[] }>(`${getFaqApiBase()}/content/faq-pages`, {
      query: { locale: getBackendFaqLocale() }
    })
    if (hasAnyFaqContent(structured.pages)) return structured.pages
  } catch (error) {
    logFaqFetchError('Failed to fetch structured FAQ pages from Go backend:', error)
  }

  return []
}
