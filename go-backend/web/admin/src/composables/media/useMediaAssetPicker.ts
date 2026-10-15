import { computed, onScopeDispose, reactive, ref, toValue, watch } from 'vue'
import type { MaybeRefOrGetter, WatchSource } from 'vue'
import mediaApi from '@/api/media'
import type { MediaAsset } from '@/api/media'

interface MediaPickerFilters {
  search: string
}

interface MediaPickerPagination {
  page: number
  pageSize: number
  total: number
}

export const useMediaAssetPicker = (
  open: WatchSource<boolean>,
  mediaType: MaybeRefOrGetter<'image' | 'video' | 'all' | string> = 'image',
) => {
  const loading = ref(false)
  const assets = ref<MediaAsset[]>([])
  const loadError = ref('')
  const filters = reactive<MediaPickerFilters>({ search: '' })
  const pagination = reactive<MediaPickerPagination>({ page: 1, pageSize: 40, total: 0 })
  const resolvedMediaType = computed(() => {
    const value = String(toValue(mediaType) || 'image').trim().toLowerCase()
    return ['image', 'video', 'all'].includes(value) ? value : 'image'
  })
  let activeRequest: AbortController | undefined
  let requestGeneration = 0
  let disposed = false
  let isPickerOpen = false

  const cancelReusableMediaPickerRequestAndClearResults = (): void => {
    requestGeneration++
    activeRequest?.abort()
    assets.value = []
    pagination.total = 0
    loadError.value = ''
    loading.value = false
  }

  const loadAssets = async (): Promise<void> => {
    if (disposed || !isPickerOpen) return
    activeRequest?.abort()
    const controller = new AbortController()
    activeRequest = controller
    const generation = ++requestGeneration
    loading.value = true
    assets.value = []
    pagination.total = 0
    loadError.value = ''
    try {
      const mediaTypeValue = resolvedMediaType.value
      const result = await mediaApi.listAssets({
        page: pagination.page,
        page_size: pagination.pageSize,
        media_type: mediaTypeValue === 'all' ? undefined : mediaTypeValue,
        status: 'active',
        visibility: 'public',
        selection_context: 'reusable',
        search: filters.search || undefined,
      }, controller.signal)
      if (disposed || controller.signal.aborted || generation !== requestGeneration) return
      assets.value = result.assets || []
      pagination.total = result.pagination?.total ?? 0
    } catch (error) {
      if (disposed || controller.signal.aborted || generation !== requestGeneration) return
      console.error('Failed to load media assets:', error)
      assets.value = []
      pagination.total = 0
      loadError.value = '可选素材加载失败，请重试'
    } finally {
      if (!disposed && generation === requestGeneration) loading.value = false
    }
  }

  const reload = (): void => {
    pagination.page = 1
    loadAssets()
  }

  const updatePage = (page: number): void => {
    pagination.page = page
    loadAssets()
  }

  const updatePageSize = (pageSize: number): void => {
    pagination.pageSize = pageSize
    pagination.page = 1
    loadAssets()
  }

  watch([open, resolvedMediaType], ([isOpen]) => {
    isPickerOpen = isOpen
    cancelReusableMediaPickerRequestAndClearResults()
    if (isOpen) reload()
  }, { immediate: true, flush: 'sync' })

  onScopeDispose(() => {
    disposed = true
    cancelReusableMediaPickerRequestAndClearResults()
  })

  return {
    loading,
    loadError,
    assets,
    filters,
    pagination,
    reload,
    updatePage,
    updatePageSize,
  }
}
