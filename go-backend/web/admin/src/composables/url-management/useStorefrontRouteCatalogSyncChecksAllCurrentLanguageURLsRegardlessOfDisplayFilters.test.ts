import { beforeEach, describe, expect, it, vi } from 'vitest'
import { effectScope } from 'vue'
import { storefrontRouteCatalogApi } from '@/modules/url-management/routeCatalog'
import { defaultStorefrontRouteCatalogStats } from '@/modules/url-management/routeCatalogStatsDefaults'
import { useStorefrontRouteCatalog } from './useStorefrontRouteCatalog'

const operationStore = vi.hoisted(() => ({ running: false, locale: '', progress: {}, run: vi.fn() }))
vi.mock('@/stores/urlOperation', () => ({ useURLOperationStore: () => operationStore }))
vi.mock('@/modules/url-management/routeCatalog', () => ({ storefrontRouteCatalogApi: { stats: vi.fn(), list: vi.fn() } }))
vi.mock('vue-sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }))

beforeEach(() => {
  vi.resetAllMocks()
  operationStore.running = false
  operationStore.run.mockResolvedValue(true)
  vi.mocked(storefrontRouteCatalogApi.stats).mockResolvedValue(defaultStorefrontRouteCatalogStats())
  vi.mocked(storefrontRouteCatalogApi.list).mockResolvedValue({
    items: [], pagination: { page: 1, page_size: 50, total: 0, total_pages: 0 },
  })
})

describe('synchronizing the URL catalog checks current state', () => {
  it('checks every URL in the selected language even when the displayed list is empty or filtered to old failures', async () => {
    const scope = effectScope()
    const catalog = scope.run(() => useStorefrontRouteCatalog(true))!
    catalog.applyPreset('canonical', false)
    Object.assign(catalog.filters, {
      locale: 'zh_cn', search: 'old error', check_status: 'not_found', source_type: 'product', entry_status: 'stale',
    })
    catalog.pagination.page = 4
    await catalog.syncCatalog()
    expect(operationStore.run).toHaveBeenCalledExactlyOnceWith({ locale: 'zh_cn', include_aliases: true, batch_size: 200 }, 'zh_cn')
    expect(catalog.pagination.page).toBe(1)
    expect(storefrontRouteCatalogApi.stats).toHaveBeenCalled()
    expect(storefrontRouteCatalogApi.list).toHaveBeenCalled()
    expect(catalog.syncing.value).toBe(false)
    scope.stop()
  })

  it('releases the sync state after a failed check so the user can retry', async () => {
    const scope = effectScope()
    const catalog = scope.run(() => useStorefrontRouteCatalog(true))!
    catalog.filters.locale = 'en'
    operationStore.run.mockResolvedValueOnce(false)
    await catalog.syncCatalog()
    expect(catalog.syncing.value).toBe(false)
    await catalog.syncCatalog()
    expect(operationStore.run).toHaveBeenCalledTimes(2)
    scope.stop()
  })

  it('requires edit access and a selected language before starting a check', async () => {
    const scope = effectScope()
    const readOnlyCatalog = scope.run(() => useStorefrontRouteCatalog(false))!
    readOnlyCatalog.filters.locale = 'en'
    await readOnlyCatalog.syncCatalog()
    const editableCatalog = scope.run(() => useStorefrontRouteCatalog(true))!
    await editableCatalog.syncCatalog()
    expect(operationStore.run).not.toHaveBeenCalled()
    scope.stop()
  })
})
