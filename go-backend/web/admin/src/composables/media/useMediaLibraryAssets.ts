import { reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import mediaApi from '@/api/media'
import type { MediaAsset } from '@/api/media'
import { validateUploadFile } from '@/lib/uploadSpecs'

export interface MediaLibraryFilters {
  search: string
  mediaType: string
  status: string
}

export interface MediaLibraryPagination {
  page: number
  pageSize: number
  total: number
}

export const useMediaLibraryAssets = () => {
  const loading = ref(false)
  const uploading = ref(false)
  const saving = ref(false)
  const assets = ref<MediaAsset[]>([])
  const filters = reactive<MediaLibraryFilters>({ search: '', mediaType: 'all', status: 'all' })
  const pagination = reactive<MediaLibraryPagination>({ page: 1, pageSize: 40, total: 0 })

  const fetchAssets = async (): Promise<void> => {
    loading.value = true
    try {
      const result = await mediaApi.listAssets({
        page: pagination.page,
        page_size: pagination.pageSize,
        search: filters.search || undefined,
        media_type: filters.mediaType !== 'all' ? filters.mediaType : undefined,
        status: filters.status !== 'all' ? filters.status : undefined,
      })
      assets.value = result.assets || []
      pagination.total = result.pagination?.total ?? 0
    } catch (error) {
      console.error('Failed to load media assets:', error)
    } finally {
      loading.value = false
    }
  }

  const applyFilters = (): void => {
    pagination.page = 1
    fetchAssets()
  }

  const resetFilters = (): void => {
    Object.assign(filters, { search: '', mediaType: 'all', status: 'all' })
    applyFilters()
  }

  const updatePage = (page: number): void => {
    pagination.page = page
    fetchAssets()
  }

  const updatePageSize = (pageSize: number): void => {
    pagination.pageSize = pageSize
    pagination.page = 1
    fetchAssets()
  }

  const uploadFile = async (file?: File | null): Promise<void> => {
    if (!file) return
    if (file.type.startsWith('image/') || /\.(jpe?g|png|webp)$/i.test(file.name)) {
      const validation = await validateUploadFile(file, 'media_library_image')
      if (!validation.ok) {
        toast.error(validation.error || '媒体库图片不符合上传规范')
        return
      }
      if (validation.warning) toast.warning(validation.warning)
    }
    uploading.value = true
    try {
      const formData = new FormData()
      formData.append('file', file)
      formData.append('media_type', file.type.startsWith('video/') ? 'video' : 'image')
      if (!file.type.startsWith('video/')) formData.append('image_purpose', 'media_library_image')
      await mediaApi.uploadAsset(formData)
      toast.success('媒体资源已上传')
      await fetchAssets()
    } catch (error) {
      console.error('Failed to upload media asset:', error)
    } finally {
      uploading.value = false
    }
  }

  const saveAsset = async (asset: MediaAsset | null, editor: Record<string, unknown>): Promise<boolean> => {
    if (!asset) return false
    saving.value = true
    try {
      await mediaApi.updateAsset(asset.id, { ...editor })
      toast.success('媒体信息已更新')
      await fetchAssets()
      return true
    } catch (error) {
      console.error('Failed to update media asset:', error)
      return false
    } finally {
      saving.value = false
    }
  }

  const copyURL = async (asset?: MediaAsset | null, accessURL?: string | null): Promise<void> => {
    const url = accessURL || asset?.access_url || asset?.url || ''
    if (!url) return
    await navigator.clipboard.writeText(url)
    toast.success('资源 URL 已复制')
  }

  const downloadCopyrightEvidence = async (asset: MediaAsset): Promise<void> => {
    try {
      const response = await mediaApi.downloadCopyrightEvidence(asset.id)
      const objectURL = URL.createObjectURL(response.data)
      const link = document.createElement('a')
      link.href = objectURL
      link.download = `copyright-evidence-${asset.id}.zip`
      document.body.appendChild(link)
      link.click()
      link.remove()
      URL.revokeObjectURL(objectURL)
      toast.success('版权证据包已导出')
    } catch (error) {
      console.error('Failed to export copyright evidence:', error)
    }
  }

  return {
    assets,
    filters,
    pagination,
    loading,
    uploading,
    saving,
    fetchAssets,
    applyFilters,
    resetFilters,
    updatePage,
    updatePageSize,
    uploadFile,
    saveAsset,
    copyURL,
    downloadCopyrightEvidence,
  }
}
