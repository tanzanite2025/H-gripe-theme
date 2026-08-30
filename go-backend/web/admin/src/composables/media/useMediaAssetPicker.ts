import { computed, reactive, ref, unref, watch } from 'vue'
import type { MaybeRefOrGetter, WatchSource } from 'vue'
import mediaApi from '@/api/media'
import type { MediaAsset } from '@/api/media'

interface MediaPickerFilters {
  search: string
  status: string
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
  const loadedMediaType = ref('')
  const filters = reactive<MediaPickerFilters>({ search: '', status: 'active' })
  const pagination = reactive<MediaPickerPagination>({ page: 1, pageSize: 40, total: 0 })
  const resolvedMediaType = computed(() => {
    const value = String(unref(mediaType) || 'image').trim().toLowerCase()
    return ['image', 'video', 'all'].includes(value) ? value : 'image'
  })

  const loadAssets = async (): Promise<void> => {
    loading.value = true
    try {
      const mediaTypeValue = resolvedMediaType.value
      const result = await mediaApi.listAssets({
        page: pagination.page,
        page_size: pagination.pageSize,
        media_type: mediaTypeValue === 'all' ? undefined : mediaTypeValue,
        status: filters.status !== 'all' ? filters.status : undefined,
        visibility: 'public',
        search: filters.search || undefined,
      })
      assets.value = result.assets || []
      pagination.total = result.pagination?.total ?? 0
      loadedMediaType.value = mediaTypeValue
    } catch (error) {
      console.error('Failed to load media assets:', error)
      assets.value = []
      pagination.total = 0
    } finally {
      loading.value = false
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
    if (isOpen && loadedMediaType.value !== resolvedMediaType.value) {
      loadAssets()
    }
  })

  return {
    loading,
    assets,
    filters,
    pagination,
    reload,
    updatePage,
    updatePageSize,
  }
}
