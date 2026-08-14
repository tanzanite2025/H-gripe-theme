import { ref, unref, watch, type MaybeRefOrGetter, type Ref } from 'vue'
import { useLocalePath, useRoute, useRouter } from '#imports'
import {
  getPageSubNavigationTabFromPath,
  isPageSubNavigationTabId,
  pageSubNavigationChildPath,
  type PageSubNavigationTab,
} from '~/utils/pageSubNavigation'

type TabId<Tabs extends readonly PageSubNavigationTab[]> = Tabs[number]['id']

interface UsePageSubNavigationTabOptions<Tabs extends readonly PageSubNavigationTab[]> {
  tabs: Tabs
  basePath: string
  defaultValue: TabId<Tabs>
  syncWithUrl?: MaybeRefOrGetter<boolean>
}

const resolveBooleanOption = (value: MaybeRefOrGetter<boolean> | undefined, fallback: boolean) => {
  if (value === undefined) return fallback
  if (typeof value === 'function') return Boolean((value as () => boolean)())
  return Boolean(unref(value))
}

export const usePageSubNavigationTab = <Tabs extends readonly PageSubNavigationTab[]>({
  tabs,
  basePath,
  defaultValue,
  syncWithUrl = true,
}: UsePageSubNavigationTabOptions<Tabs>) => {
  const route = useRoute()
  const router = useRouter()
  const localePath = useLocalePath()
  const activeTab = ref(defaultValue) as Ref<TabId<Tabs>>

  const shouldSyncWithUrl = () => resolveBooleanOption(syncWithUrl, true)

  const localizedTabPath = (tabId: TabId<Tabs>) => localePath(pageSubNavigationChildPath(basePath, tabId))

  const replaceRouteWithTab = (tabId: TabId<Tabs>) => {
    if (!shouldSyncWithUrl() || typeof window === 'undefined') return

    const path = localizedTabPath(tabId)
    if (route.path === path) return

    void router.replace({ path, query: route.query }).catch(() => {})
  }

  const syncActiveTabFromRoute = () => {
    if (!shouldSyncWithUrl()) return

    const pathTab = getPageSubNavigationTabFromPath(tabs, basePath, route.path, { match: 'nested' })
    activeTab.value = pathTab || defaultValue
  }

  watch(
    () => route.fullPath,
    syncActiveTabFromRoute,
    { immediate: true }
  )

  const setActiveTab = (id: string) => {
    if (!isPageSubNavigationTabId(tabs, id)) return

    activeTab.value = id
    replaceRouteWithTab(id)
  }

  return {
    activeTab,
    localizedTabPath,
    setActiveTab,
    syncActiveTabFromRoute,
  }
}
