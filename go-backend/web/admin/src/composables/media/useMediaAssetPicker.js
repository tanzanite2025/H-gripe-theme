import { reactive, ref, watch } from 'vue'
import mediaApi from '@/api/media'

export const useMediaAssetPicker = (open) => {
  const loading = ref(false)
  const assets = ref([])
  const loadedOnce = ref(false)
  const filters = reactive({ search: '', status: 'active' })
  const pagination = reactive({ page: 1, pageSize: 40, total: 0 })

  const loadAssets = async () => {
    loading.value = true
    try {
      const result = await mediaApi.listAssets({
        page: pagination.page,
        page_size: pagination.pageSize,
        media_type: 'image',
        status: filters.status !== 'all' ? filters.status : undefined,
        visibility: 'public',
        search: filters.search || undefined,
      })
      assets.value = result.assets || []
      pagination.total = result.pagination?.total ?? 0
      loadedOnce.value = true
    } catch (error) {
      console.error('Failed to load media assets:', error)
      assets.value = []
      pagination.total = 0
    } finally {
      loading.value = false
    }
  }

  const reload = () => {
    pagination.page = 1
    loadAssets()
  }

  const updatePage = (page) => {
    pagination.page = page
    loadAssets()
  }

  const updatePageSize = (pageSize) => {
    pagination.pageSize = pageSize
    pagination.page = 1
    loadAssets()
  }

  watch(open, (isOpen) => {
    if (isOpen && !loadedOnce.value) {
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
