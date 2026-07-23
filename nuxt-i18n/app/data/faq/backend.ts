import type { FaqCategory, PageFaqData } from './types'
import { getAllFaqData, getFaqData, getFaqDataByRoutePath } from './registry'
import { normalizeFaqRoutePath } from './routing'

type LegacyFaqItem = {
  id: string | number
  page_id?: string
  category?: string
  question: string
  answer: string
  answer_image_url?: string
  answer_image_alt?: string
  answer_image_width?: number
  answer_image_height?: number
  answerImageUrl?: string
  answerImageAlt?: string
  answerImageWidth?: number
  answerImageHeight?: number
}

function hasFaqContent(page?: PageFaqData): page is PageFaqData {
  return Boolean(page?.categories?.some(category => category.items.length > 0))
}

function hasAnyFaqContent(pages?: PageFaqData[]): pages is PageFaqData[] {
  return Boolean(pages?.some(page => hasFaqContent(page)))
}

function getBackendFaqLocale() {
  try {
    const { locale } = useI18n()
    const currentLocale = String(locale.value || 'en').toLowerCase().replace('_', '-')
    if (currentLocale.startsWith('zh-')) return 'zh'
    return currentLocale || 'en'
  } catch {
    return 'en'
  }
}

function getFaqApiBase() {
  const config = useRuntimeConfig()
  return (config.public as { apiBase?: string }).apiBase || '/api/v1'
}

function buildFallbackPageTitle(pageId: string) {
  return `${pageId.split('-').map(s => s.charAt(0).toUpperCase() + s.slice(1)).join(' ')} FAQs`
}

function legacyItemToFaqItem(item: LegacyFaqItem) {
  return {
    id: item.id.toString(),
    question: item.question,
    answer: item.answer,
    answerImageUrl: item.answer_image_url || item.answerImageUrl,
    answerImageAlt: item.answer_image_alt || item.answerImageAlt,
    answerImageWidth: item.answer_image_width || item.answerImageWidth,
    answerImageHeight: item.answer_image_height || item.answerImageHeight,
    tags: []
  }
}

function categoryIdFromName(name: string) {
  return name.toLowerCase().replace(/\s+/g, '-')
}

function buildPageFromLegacyItems(pageId: string, items: LegacyFaqItem[]): PageFaqData {
  const categoriesMap = new Map<string, FaqCategory>()

  items.forEach(item => {
    const catName = item.category || 'General'
    if (!categoriesMap.has(catName)) {
      categoriesMap.set(catName, {
        id: categoryIdFromName(catName),
        name: catName,
        items: []
      })
    }
    categoriesMap.get(catName)!.items.push(legacyItemToFaqItem(item))
  })

  return {
    pageId,
    title: buildFallbackPageTitle(pageId),
    categories: Array.from(categoriesMap.values())
  }
}

function buildPagesFromLegacyItems(items: LegacyFaqItem[]): PageFaqData[] {
  const pagesMap = new Map<string, PageFaqData>()

  items.forEach(item => {
    const pageId = item.page_id || 'general'
    if (!pagesMap.has(pageId)) {
      pagesMap.set(pageId, {
        pageId,
        title: buildFallbackPageTitle(pageId),
        categories: []
      })
    }

    const pageData = pagesMap.get(pageId)!
    const catName = item.category || 'General'

    let category = pageData.categories.find(c => c.name === catName)
    if (!category) {
      category = {
        id: categoryIdFromName(catName),
        name: catName,
        items: []
      }
      pageData.categories.push(category)
    }

    category.items.push(legacyItemToFaqItem(item))
  })

  return Array.from(pagesMap.values())
}

/**
 * Fetch FAQ data for a specific page from Go backend.
 */
export async function fetchFaqData(pageId: string): Promise<PageFaqData | undefined> {
  const fallback = getFaqData(pageId)

  try {
    const structured = await $fetch<{ page?: PageFaqData }>(`${getFaqApiBase()}/content/faq-pages/${pageId}`, {
      query: { locale: getBackendFaqLocale() }
    })
    if (hasFaqContent(structured.page)) return structured.page
  } catch (error) {
    console.error('Failed to fetch structured FAQs from Go backend:', error)
  }

  try {
    const res = await $fetch<{ data: LegacyFaqItem[] }>(`${getFaqApiBase()}/content/faqs`, {
      query: { page_id: pageId, page_size: 100, locale: getBackendFaqLocale() }
    })

    if (!res.data) throw new Error('[CRITICAL] FAQ data missing')
    if (res.data.length === 0) return fallback

    return buildPageFromLegacyItems(pageId, res.data)
  } catch (error) {
    console.error('Failed to fetch FAQs from Go backend:', error)
    return fallback
  }
}

export async function fetchFaqDataByRoutePath(routePath: string): Promise<PageFaqData | undefined> {
  const normalizedPath = normalizeFaqRoutePath(routePath)
  const fallback = getFaqDataByRoutePath(normalizedPath)

  try {
    const structured = await $fetch<{ page?: PageFaqData }>(`${getFaqApiBase()}/content/faq-pages/by-route`, {
      query: { route_path: normalizedPath, locale: getBackendFaqLocale() }
    })
    if (hasFaqContent(structured.page)) return structured.page
  } catch (error) {
    console.error('Failed to fetch structured FAQ by route from Go backend:', error)
  }

  return fallback
}

/**
 * Fetch all FAQ data from Go backend.
 */
export async function fetchAllFaqData(): Promise<PageFaqData[]> {
  const fallback = getAllFaqData()

  try {
    const structured = await $fetch<{ pages?: PageFaqData[] }>(`${getFaqApiBase()}/content/faq-pages`, {
      query: { locale: getBackendFaqLocale() }
    })
    if (hasAnyFaqContent(structured.pages)) return structured.pages
  } catch (error) {
    console.error('Failed to fetch structured FAQ pages from Go backend:', error)
  }

  try {
    const res = await $fetch<{ data: LegacyFaqItem[] }>(`${getFaqApiBase()}/content/faqs`, {
      query: { page_size: 1000, locale: getBackendFaqLocale() }
    })

    if (!res.data) throw new Error('[CRITICAL] FAQ data missing')
    if (res.data.length === 0) return fallback

    return buildPagesFromLegacyItems(res.data)
  } catch (error) {
    console.error('Failed to fetch all FAQs from Go backend:', error)
    return fallback
  }
}
