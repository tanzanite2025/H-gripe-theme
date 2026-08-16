import type { ResolvedPageFaqData } from './helpers'

export interface GlobalAllFaqFlatItem {
  id: string
  pageId: string
  pageTitle: string
  category: string
  question: string
  answer: string
  answerImageUrl?: string
  answerImageAlt?: string
  answerImageWidth?: number
  answerImageHeight?: number
  tags?: string[]
}

export interface GlobalAllFaqsDisplayGroup {
  pageId: string
  pageTitle: string
  items: GlobalAllFaqFlatItem[]
}

export interface GlobalAllFaqSearchTopic {
  id: string
  label: string
  count: number
  items: GlobalAllFaqFlatItem[]
}

export function flattenGlobalAllFaqItems(
  pages: ResolvedPageFaqData[],
): GlobalAllFaqFlatItem[] {
  const items: GlobalAllFaqFlatItem[] = []

  for (const page of pages) {
    for (const category of page.categories) {
      for (const item of category.items) {
        items.push({
          id: `${page.pageId}-${item.id}`,
          pageId: page.pageId,
          pageTitle: page.title || page.pageId,
          category: category.name,
          question: item.question,
          answer: item.answer,
          answerImageUrl: item.answerImageUrl,
          answerImageAlt: item.answerImageAlt,
          answerImageWidth: item.answerImageWidth,
          answerImageHeight: item.answerImageHeight,
          tags: item.tags,
        })
      }
    }
  }

  return items
}

export function groupGlobalAllFaqTopics(
  items: GlobalAllFaqFlatItem[],
): GlobalAllFaqSearchTopic[] {
  const groups = new Map<string, GlobalAllFaqSearchTopic>()

  for (const item of items) {
    const label = item.category.trim() || item.pageTitle
    const id = label.toLocaleLowerCase()
    const existing = groups.get(id)

    if (existing) {
      existing.items.push(item)
      existing.count += 1
      continue
    }

    groups.set(id, {
      id,
      label,
      count: 1,
      items: [item],
    })
  }

  return Array.from(groups.values()).sort((left, right) => (
    right.count - left.count || left.label.localeCompare(right.label)
  ))
}

export function filterGlobalAllFaqItems(
  items: GlobalAllFaqFlatItem[],
  searchQuery: string,
  activePageId = 'all',
): GlobalAllFaqFlatItem[] {
  let filteredItems = activePageId === 'all'
    ? items
    : items.filter(item => item.pageId === activePageId)

  const normalizedQuery = searchQuery.trim().toLowerCase()
  if (!normalizedQuery) return filteredItems

  filteredItems = filteredItems.filter(item => (
    item.question.toLowerCase().includes(normalizedQuery)
    || item.answer.toLowerCase().includes(normalizedQuery)
    || item.category.toLowerCase().includes(normalizedQuery)
    || Boolean(item.tags?.some(tag => tag.toLowerCase().includes(normalizedQuery)))
  ))

  return filteredItems
}

export function groupGlobalAllFaqItemsByPage(
  items: GlobalAllFaqFlatItem[],
): GlobalAllFaqsDisplayGroup[] {
  const groups = new Map<string, GlobalAllFaqsDisplayGroup>()

  for (const item of items) {
    let group = groups.get(item.pageId)

    if (!group) {
      group = {
        pageId: item.pageId,
        pageTitle: item.pageTitle,
        items: [],
      }
      groups.set(item.pageId, group)
    }

    group.items.push(item)
  }

  return Array.from(groups.values())
}
