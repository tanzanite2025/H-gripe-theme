import {
  buildLanguageOptions,
  buildLocaleFilterOptions as buildSupportedLocaleFilterOptions,
  localeNameFromLanguages
} from '@/lib/languages'
import type { AdminLanguage, LanguageOption } from '@/lib/languages'

export type FAQStatusTone = 'green' | 'gray'

export type FAQID = string | number

export interface FAQItemLike {
  id?: FAQID | null
  question?: string | null
  answer?: string | null
  answer_image_url?: string | null
  answer_image_alt?: string | null
  answer_image_width?: number | string | null
  answer_image_height?: number | string | null
  page_id?: string | null
  locale?: string | null
  status?: string | null
  order?: number | string | null
  sort_order?: number | string | null
}

export interface FAQStructurePage {
  page_id: string
  locale: string
  route_path?: string | null
  domain?: string | null
  title?: string | null
  subtitle?: string | null
  status?: string | null
  sort_order?: number | string | null
  faqs?: FAQItemLike[] | null
}

export type FAQStructureMap = Record<string, FAQStructurePage[] | undefined>

export const FAQ_STATUS_FILTER_OPTIONS: LanguageOption[] = [
  { label: '全部状态', value: 'all' },
  { label: '已发布', value: 'published' },
  { label: '草稿', value: 'draft' }
]

export const buildLocaleFilterOptions = buildSupportedLocaleFilterOptions

export const buildStructureLocaleOptions = buildLanguageOptions

export const localeName = localeNameFromLanguages
export const statusName = (status?: string | null): string => ({ published: '已发布', draft: '草稿', active: '启用', hidden: '隐藏' })[status || ''] || status || '-'
export const statusTone = (status?: string | null): FAQStatusTone => ({ published: 'green', draft: 'gray', active: 'green', hidden: 'gray' } as Record<string, FAQStatusTone>)[status || ''] || 'gray'
export const visibilityName = (status?: string | null): string => statusName(status)
export const visibilityTone = (status?: string | null): FAQStatusTone => statusTone(status)
export const domainName = (domain?: string | null): string => ({ products: 'PRODUCTS', guides: 'GUIDES', support: 'SUPPORT', company: 'COMPANY' })[domain || ''] || (domain || 'GENERAL').toUpperCase()
export const formatDate = (dateString?: string | number | Date | null): string => dateString ? new Date(dateString).toLocaleString('zh-CN') : '-'
export const plainTextFromHTML = (value?: string | null): string => String(value || '').replace(/<[^>]*>/g, ' ').replace(/\s+/g, ' ').trim()

const pagesForLocale = (faqStructures: FAQStructureMap, locale: string): FAQStructurePage[] => (
  locale !== 'all'
    ? (faqStructures[locale] || [])
    : Object.values(faqStructures).flat().filter(Boolean) as FAQStructurePage[]
)

const uniquePageOptions = (pages: FAQStructurePage[], includeID = false): LanguageOption[] => {
  const seen = new Set<string>()
  return pages
    .filter((page) => {
      if (seen.has(page.page_id)) return false
      seen.add(page.page_id)
      return true
    })
    .map((page) => ({
      label: includeID ? `${page.title || page.page_id} · ${page.page_id}` : (page.title || page.page_id),
      value: page.page_id
    }))
}

export const buildStructurePageOptions = (activePages: FAQStructurePage[] = [], allPages: FAQStructurePage[] = []): LanguageOption[] => (
  uniquePageOptions(activePages.length > 0 ? activePages : allPages, true)
)

export const buildFAQPageOptions = (faqStructures: FAQStructureMap, locale: string): LanguageOption[] => (
  uniquePageOptions(faqStructures[locale] || [], true)
)

export const buildPageFilterOptions = (faqStructures: FAQStructureMap, locale: string): LanguageOption[] => [
  { label: '全部页面', value: 'all' },
  ...uniquePageOptions(pagesForLocale(faqStructures, locale))
]

const structureKey = (pageID?: string | null, locale?: string | null): string => `${pageID || ''}\u0000${locale || ''}`

export const buildFAQLabelMaps = (pages: FAQStructurePage[]): { pageTitles: Map<string, string> } => {
  const pageTitles = new Map<string, string>()

  for (const page of pages) {
    pageTitles.set(structureKey(page.page_id, page.locale), page.title || page.page_id)
  }

  return { pageTitles }
}

export const pageTitleForFAQ = (pageTitles: Map<string, string>, faq: FAQItemLike): string => (
  pageTitles.get(structureKey(faq.page_id, faq.locale)) || faq.page_id || '-'
)

export type { AdminLanguage }
