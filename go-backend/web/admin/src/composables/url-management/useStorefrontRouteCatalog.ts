import { computed, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { storefrontRouteCatalogApi } from '@/modules/url-management/routeCatalog'
import { storefrontURLIssuesApi, type StorefrontURLIssueStats } from '@/modules/url-management/urlIssues'
import type { SEOResourcePagination } from '@/modules/seo/types'
import type {
  StorefrontRouteCatalogEntry,
  StorefrontRouteCatalogListParams,
  StorefrontRouteCatalogStats,
  StorefrontRouteCheckResult,
} from '@/modules/url-management/routeCatalogTypes'
import { checkLabel } from '@/modules/url-management/routeCatalogPresentation'
import { useURLOperationStore } from '@/stores/urlOperation'

export type RouteCatalogMode = 'catalog' | 'canonical'

export interface StorefrontRouteCatalogFilters {
  search: string
  locale: string
  source_type: string
  entry_status: string
  check_status: string
  searchable: string
  search_profile_status: string
  includeAliases: boolean
}

export const defaultStorefrontRouteCatalogStats = (): StorefrontRouteCatalogStats => ({
  total: 0,
  active: 0,
  alias: 0,
  duplicate: 0,
  stale: 0,
  needs_attention: 0,
  checked: 0,
  unchecked: 0,
  ok: 0,
  redirects: 0,
  not_found: 0,
  server_errors: 0,
  canonical_mismatch: 0,
  errors: 0,
  searchable: 0,
  checkable: 0,
  indexable: 0,
  sitemap_eligible: 0,
  last_synced_at: null,
  manifest_version: '',
})

const defaultPagination = (pageSize: number): SEOResourcePagination => ({
  page: 1,
  page_size: pageSize,
  total: 0,
  total_pages: 0,
})

const errorMessage = (error: unknown): string => {
  if (error && typeof error === 'object') {
    const responseData = (error as { response?: { data?: unknown } }).response?.data
    if (responseData && typeof responseData === 'object') {
      const record = responseData as Record<string, unknown>
      for (const key of ['error', 'message']) {
        if (typeof record[key] === 'string' && record[key].trim()) {
          return record[key].trim()
        }
      }
    }
  }

  return error instanceof Error ? error.message : ''
}

export function useStorefrontRouteCatalog(canEdit: boolean) {
  const urlOperationStore = useURLOperationStore()
  const stats = ref<StorefrontRouteCatalogStats>(defaultStorefrontRouteCatalogStats())
  const issueStats = ref<StorefrontURLIssueStats>({
    active: 0,
    open: 0,
    acknowledged: 0,
    resolved: 0,
    verified: 0,
    suppressed: 0,
    critical: 0,
    high: 0,
  })
  const items = ref<StorefrontRouteCatalogEntry[]>([])
  const loading = ref(false)
  const statsLoading = ref(false)
  const syncing = ref(false)
  const checking = computed(() => urlOperationStore.running)
  const checkingLocale = computed(() => urlOperationStore.locale)
  const checkProgress = computed(() => urlOperationStore.progress)
  const detailLoading = ref(false)
  const historyLoading = ref(false)
  const checkingSelected = ref(false)
  const detailOpen = ref(false)
  const selectedEntry = ref<StorefrontRouteCatalogEntry | null>(null)
  const historyItems = ref<StorefrontRouteCheckResult[]>([])

  const filters = reactive<StorefrontRouteCatalogFilters>({
    search: '',
    // The view resolves the default from the supported-language registry;
    // keeping this empty avoids silently scoping operations to Chinese.
    locale: '',
    source_type: 'all',
    entry_status: 'all',
    check_status: 'all',
    searchable: 'all',
    search_profile_status: 'all',
    includeAliases: false,
  })
  const mode = ref<RouteCatalogMode>('catalog')

  const pagination = reactive<SEOResourcePagination>(defaultPagination(50))
  const historyPagination = reactive<SEOResourcePagination>(defaultPagination(10))
  const latestHistoryItem = computed(() => historyItems.value[0] || null)

  const listParams = (): StorefrontRouteCatalogListParams => ({
    page: pagination.page,
    page_size: pagination.page_size,
    ...(filters.search.trim() ? { search: filters.search.trim() } : {}),
    ...(filters.locale && filters.locale !== 'all' ? { locale: filters.locale } : {}),
    ...(filters.source_type !== 'all' && !(mode.value === 'canonical' && filters.source_type === 'alias')
      ? { source_type: filters.source_type }
      : {}),
    ...(filters.entry_status !== 'all' && !(mode.value === 'canonical' && filters.entry_status === 'alias')
      ? { entry_status: filters.entry_status }
      : {}),
    ...(filters.check_status !== 'all' ? { check_status: filters.check_status } : {}),
    ...(filters.searchable !== 'all' ? { searchable: filters.searchable } : {}),
    ...(filters.search_profile_status !== 'all' ? { search_profile_status: filters.search_profile_status } : {}),
    ...(mode.value === 'canonical' ? { problem_scope: 'canonical' as const } : {}),
    include_aliases: filters.includeAliases,
  })

  const loadStats = async (): Promise<void> => {
    statsLoading.value = true
    try {
      const locale = filters.locale !== 'all' ? filters.locale : undefined
      stats.value = {
        ...defaultStorefrontRouteCatalogStats(),
        ...(await storefrontRouteCatalogApi.stats(locale, mode.value === 'canonical' ? 'canonical' : undefined)),
      }
    } catch (error) {
      console.error('Failed to load storefront route catalog stats:', error)
      toast.error('URL 台账统计加载失败')
    } finally {
      statsLoading.value = false
    }
  }

  const load = async (): Promise<void> => {
    loading.value = true
    try {
      const response = await storefrontRouteCatalogApi.list(listParams())
      items.value = response.items
      Object.assign(pagination, response.pagination)
    } catch (error) {
      console.error('Failed to load storefront route catalog:', error)
      toast.error('URL 台账加载失败')
    } finally {
      loading.value = false
    }
  }

  const loadIssueStats = async (): Promise<void> => {
    try {
      // The issue queue is the human workflow source of truth for "待处理".
      // Route observations remain useful for the other health counters, but
      // must not make this card drift after an issue is suppressed/resolved.
      issueStats.value = { ...issueStats.value, ...(await storefrontURLIssuesApi.summary()) }
    } catch (error) {
      console.error('Failed to load storefront URL issue stats:', error)
      toast.error('URL 问题统计加载失败')
    }
  }

  const refreshAll = async (): Promise<void> => {
    await Promise.all([loadStats(), load(), loadIssueStats()])
  }

  const applyFilters = (): void => {
    pagination.page = 1
    void load()
  }

  const resetFilters = (): void => {
    filters.search = ''
    filters.source_type = 'all'
    filters.entry_status = 'all'
    filters.check_status = 'all'
    filters.searchable = 'all'
    filters.search_profile_status = 'all'
    filters.includeAliases = false
    applyPreset(mode.value, false)
    pagination.page = 1
    void load()
  }

  const applyPreset = (nextMode: RouteCatalogMode, reload = true): void => {
    mode.value = nextMode
    filters.entry_status = 'all'
    filters.check_status = 'all'
    if (nextMode === 'canonical' && filters.source_type === 'alias') {
      filters.source_type = 'all'
    }
    // Canonical conflicts are defined on canonical/duplicate entries; alias
    // rows are intentionally excluded so an alias status cannot create an
    // impossible `entry_status=alias AND problem_scope=canonical` query.
    filters.includeAliases = false
    pagination.page = 1
    if (reload) void refreshAll()
  }

  const updatePage = (page: number): void => {
    pagination.page = page
    void load()
  }

  const updatePageSize = (pageSize: number): void => {
    pagination.page_size = pageSize
    pagination.page = 1
    void load()
  }

  const syncCatalog = async (): Promise<void> => {
    if (!canEdit || syncing.value) return
    syncing.value = true
    try {
      const summary = await storefrontRouteCatalogApi.sync()
      toast.success(`URL 已同步：${summary.entries || 0} 条，重复 ${summary.duplicates || 0} 条`)
      pagination.page = 1
      await refreshAll()
    } catch (error) {
      console.error('Failed to sync storefront route catalog:', error)
      const detail = errorMessage(error)
      toast.error(detail ? `URL 同步失败：${detail}` : 'URL 同步失败，请检查 STOREFRONT_INTERNAL_ORIGIN 和 manifest')
    } finally {
      syncing.value = false
    }
  }

  const checkCatalog = async (): Promise<void> => {
    if (!canEdit || checking.value) return
    const completed = await urlOperationStore.run(
      { ...listParams(), limit: 200 },
      filters.locale !== 'all' ? filters.locale : '',
    )
    if (completed) await refreshAll()
  }

  const loadDetail = async (id: number): Promise<void> => {
    detailLoading.value = true
    historyLoading.value = true
    try {
      const [entry, history] = await Promise.all([
        storefrontRouteCatalogApi.get(id),
        storefrontRouteCatalogApi.history(id, {
          page: historyPagination.page,
          page_size: historyPagination.page_size,
        }),
      ])
      selectedEntry.value = entry
      historyItems.value = history.items
      Object.assign(historyPagination, history.pagination)
    } catch (error) {
      console.error('Failed to load storefront route detail:', error)
      toast.error('URL 详情加载失败')
    } finally {
      detailLoading.value = false
      historyLoading.value = false
    }
  }

  const openDetail = async (item: StorefrontRouteCatalogEntry): Promise<void> => {
    selectedEntry.value = item
    detailOpen.value = true
    historyPagination.page = 1
    await loadDetail(item.id)
  }

  const checkSelected = async (): Promise<void> => {
    if (!selectedEntry.value || !canEdit || checkingSelected.value) return
    checkingSelected.value = true
    try {
      const result = await storefrontRouteCatalogApi.checkOne(selectedEntry.value.id)
      toast.success(`检查完成：${checkLabel(result.status, result.http_status)}`)
      await Promise.all([loadStats(), load(), loadDetail(selectedEntry.value.id)])
    } catch (error) {
      console.error('Failed to check storefront route:', error)
      toast.error('URL 检查失败')
    } finally {
      checkingSelected.value = false
    }
  }

  const updateHistoryPage = (page: number): void => {
    if (!selectedEntry.value) return
    historyPagination.page = page
    void loadDetail(selectedEntry.value.id)
  }

  const updateHistoryPageSize = (pageSize: number): void => {
    if (!selectedEntry.value) return
    historyPagination.page_size = pageSize
    historyPagination.page = 1
    void loadDetail(selectedEntry.value.id)
  }

  return {
    mode,
    stats,
    issueStats,
    items,
    loading,
    statsLoading,
    syncing,
    checking,
    checkingLocale,
    checkProgress,
    detailLoading,
    historyLoading,
    checkingSelected,
    detailOpen,
    selectedEntry,
    historyItems,
    filters,
    pagination,
    historyPagination,
    latestHistoryItem,
    applyPreset,
    refreshAll,
    applyFilters,
    resetFilters,
    updatePage,
    updatePageSize,
    syncCatalog,
    checkCatalog,
    openDetail,
    checkSelected,
    updateHistoryPage,
    updateHistoryPageSize,
  }
}
