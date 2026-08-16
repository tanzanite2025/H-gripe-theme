import { nextTick, ref, watch, type Ref } from 'vue'
import type { RouteLocationNormalizedLoaded } from 'vue-router'
import type { GlobalAllFaqFlatItem } from '~/data/faq'
import type { ResolvedPageFaqData } from '~/data/faq'

interface UseFaqDeepLinkOptions {
  enabled: boolean
  route: RouteLocationNormalizedLoaded
  allPages: Readonly<Ref<ResolvedPageFaqData[]>>
  allItems: Readonly<Ref<GlobalAllFaqFlatItem[]>>
  activePageId: Ref<string>
  showAllGroups: () => void
  showGroupsThrough: (groupIndex: number) => void
  expandItem: (itemId: string) => void
  resetExpandedItems: () => void
}

const queryValue = (value: unknown) => {
  if (Array.isArray(value)) return String(value[0] || '').trim()
  return String(value || '').trim()
}

export function useFaqDeepLink(options: UseFaqDeepLinkOptions) {
  const applyingDeepLink = ref(false)
  const route = options.route

  const applyDeepLinkQuery = async () => {
    if (!options.allPages.value.length) return

    const requestedPageId = queryValue(route.query.page)
    const requestedFaqId = queryValue(route.query.faq)
    const target = requestedFaqId
      ? options.allItems.value.find((item) => {
          if (requestedPageId && item.pageId !== requestedPageId) return false
          return item.id === requestedFaqId || item.id === `${item.pageId}-${requestedFaqId}`
        })
      : null
    const pageId = target?.pageId
      || (requestedPageId && options.allPages.value.some((page) => page.pageId === requestedPageId)
        ? requestedPageId
        : '')

    applyingDeepLink.value = true
    try {
      options.activePageId.value = pageId || 'all'
      if (options.activePageId.value === 'all') {
        const pageIndex = target
          ? options.allPages.value.findIndex((page) => page.pageId === target.pageId)
          : -1
        options.showGroupsThrough(pageIndex)
      } else {
        options.showAllGroups()
      }

      await nextTick()
      if (target) {
        options.expandItem(target.id)
      } else {
        options.resetExpandedItems()
      }
    } finally {
      applyingDeepLink.value = false
    }
  }

  if (options.enabled) {
    watch(
      [options.allPages, () => route.query.page, () => route.query.faq],
      () => {
        void applyDeepLinkQuery()
      },
      { immediate: true },
    )
  }

  return {
    applyingDeepLink,
  }
}
