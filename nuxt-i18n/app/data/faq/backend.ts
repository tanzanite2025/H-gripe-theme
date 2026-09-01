import type { PageFaqData } from './types'
import { normalizeFaqRoutePath } from './routing'
import { useApiRequest } from '~/composables/useApiRequest'
import { normalizeStorefrontLocaleCode } from '~/utils/storefrontLocales'
import {
  createStorefrontMediaContext,
  normalizeStorefrontMediaUrl,
} from '~/utils/storefrontMedia'

function hasFaqContent(page?: PageFaqData): page is PageFaqData {
  return Boolean(page?.categories?.some(category => category.items.length > 0))
}

function hasAnyFaqContent(pages?: PageFaqData[]): pages is PageFaqData[] {
  return Boolean(pages?.some(page => hasFaqContent(page)))
}

function getBackendFaqLocale() {
  try {
    const { locale } = useI18n()
    return normalizeStorefrontLocaleCode(locale.value) || 'en'
  } catch {
    return 'en'
  }
}

function normalizeFaqPageMedia(
  page: PageFaqData,
  mediaContext: ReturnType<typeof createStorefrontMediaContext>,
): PageFaqData {
  return {
    ...page,
    categories: page.categories.map(category => ({
      ...category,
      items: category.items.map(item => ({
        ...item,
        answerImageUrl: normalizeStorefrontMediaUrl(item.answerImageUrl, mediaContext) || undefined,
      })),
    })),
  }
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
    const { request } = useApiRequest()
    const mediaContext = createStorefrontMediaContext(useRuntimeConfig())
    const structured = await request<{ page?: PageFaqData }>(
      `/content/faq-pages/${pageId}`,
      {
        query: { locale: getBackendFaqLocale() },
        headers: { accept: 'application/json' },
      },
      'Failed to fetch structured FAQs',
    )
    const page = structured.page
      ? normalizeFaqPageMedia(structured.page, mediaContext)
      : undefined
    if (hasFaqContent(page)) return page
  } catch (error) {
    logFaqFetchError('Failed to fetch structured FAQs from Go backend:', error)
  }

  return null
}

export async function fetchFaqDataByRoutePath(routePath: string): Promise<PageFaqData | null> {
  const normalizedPath = normalizeFaqRoutePath(routePath)

  try {
    const { request } = useApiRequest()
    const mediaContext = createStorefrontMediaContext(useRuntimeConfig())
    const structured = await request<{ page?: PageFaqData }>(
      '/content/faq-pages/by-route',
      {
        query: { route_path: normalizedPath, locale: getBackendFaqLocale() },
        headers: { accept: 'application/json' },
      },
      'Failed to fetch structured FAQ by route',
    )
    const page = structured.page
      ? normalizeFaqPageMedia(structured.page, mediaContext)
      : undefined
    if (hasFaqContent(page)) return page
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
    const { request } = useApiRequest()
    const mediaContext = createStorefrontMediaContext(useRuntimeConfig())
    const structured = await request<{ pages?: PageFaqData[] }>(
      '/content/faq-pages',
      {
        query: { locale: getBackendFaqLocale() },
        headers: { accept: 'application/json' },
      },
      'Failed to fetch structured FAQ pages',
    )
    const pages = (structured.pages || []).map(page => normalizeFaqPageMedia(page, mediaContext))
    if (hasAnyFaqContent(pages)) return pages
  } catch (error) {
    logFaqFetchError('Failed to fetch structured FAQ pages from Go backend:', error)
  }

  return []
}
