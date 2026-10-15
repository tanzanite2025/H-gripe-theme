import { onScopeDispose, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import mediaApi from '@/api/media'
import type { MediaAsset, MediaAssetListParams, OperationalMediaUploadContext } from '@/api/media'
import { normalizeMediaLibraryPartitionView, type MediaLibraryPartitionView } from '@/modules/media/mediaLibraryBusinessPartitionUploadOriginAndCustomerServicePurposePresentation'
import { validateUploadFile } from '@/lib/uploadSpecs'

export interface MediaLibraryFilters {
  libraryPartition: MediaLibraryPartitionView
  uploadOrigin: string
  customerServiceMediaPurpose: 'all' | 'conversation_attachment' | 'reusable_reply'
  search: string
  mediaType: string
  status: string
}

export interface MediaLibraryPagination {
  page: number
  pageSize: number
  total: number
}

export const useMediaLibraryAssets = (initialPartition?: unknown) => {
  const loading = ref(false)
  const uploading = ref(false)
  const saving = ref(false)
  const assets = ref<MediaAsset[]>([])
  const loadError = ref('')
  const partitionCounts = ref<Record<string, number>>({})
  const canViewConversationAttachments = ref(false)
  const filters = reactive<MediaLibraryFilters>({ search: '', mediaType: 'all', status: 'all',
    libraryPartition: normalizeMediaLibraryPartitionView(initialPartition), uploadOrigin: 'all', customerServiceMediaPurpose: 'all' })
  const pagination = reactive<MediaLibraryPagination>({ page: 1, pageSize: 40, total: 0 })
  let activeRequest: AbortController | undefined
  let requestGeneration = 0
  let disposed = false

  const fetchAssets = async (): Promise<void> => {
    if (disposed) return
    activeRequest?.abort()
    const controller = new AbortController()
    activeRequest = controller
    const generation = ++requestGeneration
    loading.value = true
    loadError.value = ''
    assets.value = []
    pagination.total = 0
    partitionCounts.value = {}
    canViewConversationAttachments.value = false
    try {
      const params: MediaAssetListParams = {
        library_partition: filters.libraryPartition,
        upload_origin: filters.uploadOrigin !== 'all' ? filters.uploadOrigin : undefined,
        customer_service_media_purpose: filters.customerServiceMediaPurpose !== 'all' ? filters.customerServiceMediaPurpose : undefined,
        page: pagination.page,
        page_size: pagination.pageSize,
        search: filters.search || undefined,
        media_type: filters.mediaType !== 'all' ? filters.mediaType : undefined,
        status: filters.status !== 'all' ? filters.status : undefined,
      }
      const [result, scope] = await Promise.all([
        mediaApi.listAssets(params, controller.signal), mediaApi.countAssetsByLibraryPartition(params, controller.signal),
      ])
      if (disposed || controller.signal.aborted || generation !== requestGeneration) return
      assets.value = result.assets || []
      pagination.total = result.pagination?.total ?? 0
      partitionCounts.value = scope.counts
      canViewConversationAttachments.value = scope.canViewConversationAttachments
    } catch (error) {
      if (disposed || controller.signal.aborted || generation !== requestGeneration) return
      console.error('Failed to load media assets:', error)
      assets.value = []
      pagination.total = 0
      partitionCounts.value = {}
      canViewConversationAttachments.value = false
      loadError.value = '媒体资源加载失败，请重试'
    } finally {
      if (!disposed && generation === requestGeneration) loading.value = false
    }
  }

  onScopeDispose(() => { disposed = true; requestGeneration++; activeRequest?.abort() })

  const selectLibraryPartition = (value: unknown): void => {
    filters.libraryPartition = normalizeMediaLibraryPartitionView(value)
    filters.customerServiceMediaPurpose = 'all'
    applyFilters()
  }

  const applyFilters = (): void => {
    pagination.page = 1
    fetchAssets()
  }

  const resetFilters = (): void => {
    Object.assign(filters, { search: '', mediaType: 'all', status: 'all', uploadOrigin: 'all', customerServiceMediaPurpose: 'all' })
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

  const uploadFile = async (file?: File | null, context?: OperationalMediaUploadContext): Promise<void> => {
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
      await mediaApi.uploadAsset(formData, context)
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
    if (asset?.requires_conversation_access) return
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
    loadError,
    partitionCounts,
    canViewConversationAttachments,
    filters,
    pagination,
    loading,
    uploading,
    saving,
    fetchAssets,
    applyFilters,
    selectLibraryPartition,
    resetFilters,
    updatePage,
    updatePageSize,
    uploadFile,
    saveAsset,
    copyURL,
    downloadCopyrightEvidence,
  }
}
