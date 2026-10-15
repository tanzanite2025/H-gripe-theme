import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { effectScope, ref, type EffectScope } from 'vue'
import mediaApi from '@/api/media'
import type { MediaAssetListResult } from '@/api/media'
import { useMediaAssetPicker } from './useMediaAssetPicker'

vi.mock('@/api/media', () => ({ default: { listAssets: vi.fn() } }))
const testScopes: EffectScope[] = []

function createReusableMediaPickerWithDisposableScope() {
  const scope = effectScope()
  testScopes.push(scope)
  const open = ref(false)
  const mediaType = ref<'image' | 'video'>('image')
  const picker = scope.run(() => useMediaAssetPicker(open, () => mediaType.value))!
  return { scope, open, mediaType, picker }
}

function createDeferredReusableMediaPickerResponse() {
  let resolve!: (value: MediaAssetListResult) => void
  const promise = new Promise<MediaAssetListResult>((resolvePromise) => { resolve = resolvePromise })
  return { promise, resolve }
}

function createReusableMediaPickerListResponse(assetId: number): MediaAssetListResult {
  return { assets: [{ id: assetId }], pagination: { page: 1, page_size: 40, total: 1, total_pages: 1 } }
}

beforeEach(() => {
  vi.resetAllMocks()
  vi.spyOn(console, 'error').mockImplementation(() => {})
  vi.mocked(mediaApi.listAssets).mockResolvedValue(createReusableMediaPickerListResponse(1))
})

afterEach(() => {
  testScopes.splice(0).forEach((scope) => scope.stop())
  vi.restoreAllMocks()
})

describe('reusable media picker only exposes current enabled public selection results', () => {
  it('requests reusable enabled public assets and reloads on reopening instead of trusting cached permissions', async () => {
    const { open, picker } = createReusableMediaPickerWithDisposableScope()
    expect(mediaApi.listAssets).not.toHaveBeenCalled()
    open.value = true
    await vi.waitFor(() => expect(picker.loading.value).toBe(false))
    expect(mediaApi.listAssets).toHaveBeenCalledWith(expect.objectContaining({ media_type: 'image', selection_context: 'reusable', visibility: 'public', status: 'active' }), expect.any(AbortSignal))
    open.value = false
    expect(picker.assets.value).toEqual([])
    open.value = true
    expect(mediaApi.listAssets).toHaveBeenCalledTimes(2)
  })

  it('resets pagination and ignores images arriving after the picker switches to videos', async () => {
    const oldImages = createDeferredReusableMediaPickerResponse()
    const currentVideos = createDeferredReusableMediaPickerResponse()
    vi.mocked(mediaApi.listAssets).mockReturnValueOnce(oldImages.promise).mockReturnValueOnce(currentVideos.promise)
    const { open, mediaType, picker } = createReusableMediaPickerWithDisposableScope()
    open.value = true
    const oldSignal = vi.mocked(mediaApi.listAssets).mock.calls[0]![1]!
    picker.pagination.page = 5
    mediaType.value = 'video'
    expect(picker.pagination.page).toBe(1)
    expect(oldSignal.aborted).toBe(true)
    expect(mediaApi.listAssets).toHaveBeenLastCalledWith(expect.objectContaining({ media_type: 'video', page: 1 }), expect.any(AbortSignal))
    oldImages.resolve(createReusableMediaPickerListResponse(99))
    await Promise.resolve()
    expect(picker.assets.value).toEqual([])
    expect(picker.loading.value).toBe(true)
    currentVideos.resolve(createReusableMediaPickerListResponse(2))
    await vi.waitFor(() => expect(picker.assets.value).toEqual([{ id: 2 }]))
  })

  it('ignores requests that finish after closing and after scope disposal', async () => {
    const pendingImages = createDeferredReusableMediaPickerResponse()
    vi.mocked(mediaApi.listAssets).mockReturnValueOnce(pendingImages.promise)
    const { scope, open, picker } = createReusableMediaPickerWithDisposableScope()
    open.value = true
    const signal = vi.mocked(mediaApi.listAssets).mock.calls[0]![1]!
    open.value = false
    pendingImages.resolve(createReusableMediaPickerListResponse(99))
    await Promise.resolve()
    expect(signal.aborted).toBe(true)
    expect(picker.assets.value).toEqual([])
    scope.stop()
    open.value = true
    picker.reload()
    expect(mediaApi.listAssets).toHaveBeenCalledTimes(1)
  })

  it('shows load failure instead of reporting an empty successful selection', async () => {
    vi.mocked(mediaApi.listAssets).mockRejectedValueOnce(new Error('Selection unavailable'))
    const { open, picker } = createReusableMediaPickerWithDisposableScope()
    open.value = true
    await vi.waitFor(() => expect(picker.loading.value).toBe(false))
    expect(picker.assets.value).toEqual([])
    expect(picker.loadError.value).toContain('加载失败')
    picker.reload()
    await vi.waitFor(() => expect(picker.loading.value).toBe(false))
    expect(picker.loadError.value).toBe('')
    expect(picker.assets.value).toEqual([{ id: 1 }])
  })
})
