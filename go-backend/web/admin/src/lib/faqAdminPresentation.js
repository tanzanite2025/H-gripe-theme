export const FAQ_STATUS_FILTER_OPTIONS = [
  { label: '全部状态', value: 'all' },
  { label: '已发布', value: 'published' },
  { label: '草稿', value: 'draft' }
]

export const FAQ_LOCALE_FILTER_OPTIONS = [
  { label: '全部语言', value: 'all' },
  { label: '中文', value: 'zh' },
  { label: 'English', value: 'en' }
]

export const FAQ_STRUCTURE_LOCALES = [
  { label: '中文', value: 'zh' },
  { label: 'English', value: 'en' }
]

export const localeName = (locale) => ({ zh: '中文', en: 'English' })[locale] || locale || '-'
export const statusName = (status) => ({ published: '已发布', draft: '草稿', active: '启用', hidden: '隐藏' })[status] || status || '-'
export const statusTone = (status) => ({ published: 'green', draft: 'gray', active: 'green', hidden: 'gray' })[status] || 'gray'
export const visibilityName = (status) => statusName(status)
export const visibilityTone = (status) => statusTone(status)
export const domainName = (domain) => ({ products: 'PRODUCTS', guides: 'GUIDES', support: 'SUPPORT', company: 'COMPANY' })[domain] || (domain || 'GENERAL').toUpperCase()
export const formatDate = (dateString) => dateString ? new Date(dateString).toLocaleString('zh-CN') : '-'
export const plainTextFromHTML = (value) => String(value || '').replace(/<[^>]*>/g, ' ').replace(/\s+/g, ' ').trim()

const pagesForLocale = (faqStructures, locale) => (
  locale !== 'all'
    ? (faqStructures[locale] || [])
    : Object.values(faqStructures).flat()
)

const uniquePageOptions = (pages, includeID = false) => {
  const seen = new Set()
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

export const buildStructurePageOptions = (activePages, allPages) => (
  uniquePageOptions(activePages.length > 0 ? activePages : allPages, true)
)

export const buildFAQPageOptions = (faqStructures, locale) => (
  uniquePageOptions(faqStructures[locale] || [], true)
)

export const findAvailableFAQCategories = (faqStructures, locale, pageID) => {
  const page = (faqStructures[locale] || []).find((item) => item.page_id === pageID)
  return page?.categories?.filter((category) => category.status !== 'hidden') || []
}

export const buildPageFilterOptions = (faqStructures, locale) => [
  { label: '全部页面', value: 'all' },
  ...uniquePageOptions(pagesForLocale(faqStructures, locale))
]

export const buildCategoryFilterOptions = (faqStructures, locale, pageID) => {
  const seen = new Set()
  return [
    { label: '全部分类', value: 'all' },
    ...pagesForLocale(faqStructures, locale)
      .flatMap((page) => (page.categories || []).map((category) => ({ ...category, page_id: page.page_id })))
      .filter((category) => pageID === 'all' || category.page_id === pageID)
      .filter((category) => {
        if (seen.has(category.category_key)) return false
        seen.add(category.category_key)
        return true
      })
      .map((category) => ({ label: category.name || category.category_key, value: category.category_key }))
      .sort((left, right) => left.label.localeCompare(right.label))
  ]
}

const structureKey = (pageID, locale) => `${pageID || ''}\u0000${locale || ''}`
const categoryKey = (pageID, locale, category) => `${pageID || ''}\u0000${locale || ''}\u0000${category || ''}`

export const buildFAQLabelMaps = (pages) => {
  const pageTitles = new Map()
  const categoryLabels = new Map()

  for (const page of pages) {
    pageTitles.set(structureKey(page.page_id, page.locale), page.title || page.page_id)
    for (const category of page.categories || []) {
      categoryLabels.set(categoryKey(page.page_id, page.locale, category.category_key), category.name || category.category_key)
    }
  }

  return { pageTitles, categoryLabels }
}

export const pageTitleForFAQ = (pageTitles, faq) => (
  pageTitles.get(structureKey(faq.page_id, faq.locale)) || faq.page_id || '-'
)

export const categoryLabelForFAQ = (categoryLabels, faq) => (
  categoryLabels.get(categoryKey(faq.page_id, faq.locale, faq.category)) || faq.category || '-'
)
