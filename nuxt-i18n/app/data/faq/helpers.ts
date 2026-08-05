import type { PageFaqData } from './types'

export type ResolvedPageFaqData = PageFaqData & {
  pageId: string
}

export function getPageFaqId(page: PageFaqData): string {
  return page.pageId || page.id || ''
}

export function resolvePageFaqData(page: PageFaqData): ResolvedPageFaqData | null {
  const pageId = getPageFaqId(page)
  return pageId ? { ...page, pageId } : null
}

export function resolvePageFaqDataList(pages: PageFaqData[]): ResolvedPageFaqData[] {
  return pages
    .map(resolvePageFaqData)
    .filter((page): page is ResolvedPageFaqData => Boolean(page))
}
