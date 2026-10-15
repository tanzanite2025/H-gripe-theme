import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { effectScope, type EffectScope } from 'vue'
import mediaApi from '@/api/media'
import type { MediaAssetListResult } from '@/api/media'
import { useMediaLibraryAssets } from './useMediaLibraryAssets'

vi.mock('@/api/media', () => ({ default: { listAssets: vi.fn(), countAssetsByLibraryPartition: vi.fn() } }))
vi.mock('vue-sonner', () => ({ toast: { error: vi.fn(), success: vi.fn(), warning: vi.fn() } }))

const testScopes: EffectScope[] = []

function createMediaLibraryWithDisposableScope(initialPartition?: unknown) {
  const scope = effectScope()
  testScopes.push(scope)
  return { scope, library: scope.run(() => useMediaLibraryAssets(initialPartition))! }
}

function createDeferredMediaLibraryResponse<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((resolvePromise) => { resolve = resolvePromise })
  return { promise, resolve }
}

function createMediaLibraryListResponse(assetId: number): MediaAssetListResult {
  return { assets: [{ id: assetId }], pagination: { page: 1, page_size: 40, total: 1, total_pages: 1 } }
}

beforeEach(() => {
  vi.resetAllMocks()
  vi.spyOn(console, 'error').mockImplementation(() => {})
  vi.mocked(mediaApi.listAssets).mockResolvedValue(createMediaLibraryListResponse(1))
  vi.mocked(mediaApi.countAssetsByLibraryPartition).mockResolvedValue({ counts: { operational: 1 }, canViewConversationAttachments: false })
})

afterEach(() => {
  testScopes.splice(0).forEach((scope) => scope.stop())
  vi.restoreAllMocks()
})

describe('media library business partitions and permission scoped request lifetime', () => {
  it('defaults to operational assets and resets pagination and conversation purpose when changing partitions', async () => {
    const { library } = createMediaLibraryWithDisposableScope()
    await library.fetchAssets()
    expect(mediaApi.listAssets).toHaveBeenCalledWith(expect.objectContaining({ library_partition: 'operational' }), expect.any(AbortSignal))
    library.pagination.page = 8
    library.filters.customerServiceMediaPurpose = 'conversation_attachment'
    library.selectLibraryPartition('catalog')
    expect(library.pagination.page).toBe(1)
    expect(library.filters.customerServiceMediaPurpose).toBe('all')
    expect(mediaApi.listAssets).toHaveBeenLastCalledWith(expect.objectContaining({ library_partition: 'catalog', page: 1, customer_service_media_purpose: undefined }), expect.any(AbortSignal))
    library.filters.search = 'old filter'
    library.resetFilters()
    expect(library.filters.libraryPartition).toBe('catalog')
    expect(library.filters.search).toBe('')
  })

  it('keeps the latest partition and counts when aborted requests still finish late', async () => {
    const oldAssets = createDeferredMediaLibraryResponse<MediaAssetListResult>()
    const oldCounts = createDeferredMediaLibraryResponse<Awaited<ReturnType<typeof mediaApi.countAssetsByLibraryPartition>>>()
    vi.mocked(mediaApi.listAssets).mockReturnValueOnce(oldAssets.promise)
    vi.mocked(mediaApi.countAssetsByLibraryPartition).mockReturnValueOnce(oldCounts.promise)
    const { library } = createMediaLibraryWithDisposableScope()
    const oldRead = library.fetchAssets()
    const oldSignal = vi.mocked(mediaApi.listAssets).mock.calls[0]![1]!
    library.filters.libraryPartition = 'catalog'
    await library.fetchAssets()
    expect(oldSignal.aborted).toBe(true)
    oldAssets.resolve(createMediaLibraryListResponse(99))
    oldCounts.resolve({ counts: { customer_service: 99 }, canViewConversationAttachments: true })
    await oldRead
    expect(library.assets.value).toEqual([{ id: 1 }])
    expect(library.partitionCounts.value).toEqual({ operational: 1 })
    expect(library.canViewConversationAttachments.value).toBe(false)
  })

  it('clears previous totals and access scope immediately, and distinguishes failed counts from an empty library', async () => {
    const { library } = createMediaLibraryWithDisposableScope('customer_service')
    vi.mocked(mediaApi.countAssetsByLibraryPartition).mockResolvedValueOnce({ counts: { customer_service: 1 }, canViewConversationAttachments: true })
    await library.fetchAssets()
    const pendingAssets = createDeferredMediaLibraryResponse<MediaAssetListResult>()
    vi.mocked(mediaApi.listAssets).mockReturnValueOnce(pendingAssets.promise)
    vi.mocked(mediaApi.countAssetsByLibraryPartition).mockRejectedValueOnce(new Error('Counts unavailable'))
    const failedRead = library.fetchAssets()
    expect(library.assets.value).toEqual([])
    expect(library.pagination.total).toBe(0)
    expect(library.partitionCounts.value).toEqual({})
    expect(library.canViewConversationAttachments.value).toBe(false)
    await failedRead
    expect(library.loadError.value).toContain('加载失败')
    pendingAssets.resolve(createMediaLibraryListResponse(99))
    await Promise.resolve()
    expect(library.assets.value).toEqual([])
  })

  it('prevents an old finally from stopping the latest loading indicator', async () => {
    const firstResponse = createDeferredMediaLibraryResponse<MediaAssetListResult>()
    const secondResponse = createDeferredMediaLibraryResponse<MediaAssetListResult>()
    vi.mocked(mediaApi.listAssets).mockReturnValueOnce(firstResponse.promise).mockReturnValueOnce(secondResponse.promise)
    const { library } = createMediaLibraryWithDisposableScope()
    const firstRead = library.fetchAssets()
    const secondRead = library.fetchAssets()
    firstResponse.resolve(createMediaLibraryListResponse(9))
    await firstRead
    expect(library.loading.value).toBe(true)
    secondResponse.resolve(createMediaLibraryListResponse(2))
    await secondRead
    expect(library.loading.value).toBe(false)
    expect(library.assets.value).toEqual([{ id: 2 }])
  })

  it('aborts on disposal and rejects late assets, late permission counts and future reads', async () => {
    const pendingAssets = createDeferredMediaLibraryResponse<MediaAssetListResult>()
    vi.mocked(mediaApi.listAssets).mockReturnValueOnce(pendingAssets.promise)
    const { scope, library } = createMediaLibraryWithDisposableScope()
    const read = library.fetchAssets()
    const signal = vi.mocked(mediaApi.listAssets).mock.calls[0]![1]!
    scope.stop()
    expect(signal.aborted).toBe(true)
    pendingAssets.resolve(createMediaLibraryListResponse(99))
    await read
    await library.fetchAssets()
    expect(library.assets.value).toEqual([])
    expect(library.partitionCounts.value).toEqual({})
    expect(mediaApi.listAssets).toHaveBeenCalledTimes(1)
  })
})
