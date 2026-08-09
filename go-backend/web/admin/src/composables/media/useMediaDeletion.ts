import { reactive } from 'vue'
import type { AxiosError } from 'axios'
import { toast } from 'vue-sonner'
import mediaApi from '@/api/media'
import type { MediaAsset, MediaReference } from '@/api/media'

export interface MediaDeleteDialog {
  open: boolean
  asset: MediaAsset | null
  references: MediaReference[]
  total: number
  loading: boolean
  deleting: boolean
  confirmation: string
}

interface MediaAssetInUsePayload {
  code?: string
  references?: MediaReference[]
  total?: number
}

export const useMediaDeletion = (refreshAssets: () => Promise<unknown> | unknown) => {
  const deleteDialog = reactive<MediaDeleteDialog>({
    open: false,
    asset: null,
    references: [],
    total: 0,
    loading: false,
    deleting: false,
    confirmation: '',
  })
  let inspectionToken = 0

  const resetDeleteDialog = (): void => {
    deleteDialog.asset = null
    deleteDialog.references = []
    deleteDialog.total = 0
    deleteDialog.loading = false
    deleteDialog.deleting = false
    deleteDialog.confirmation = ''
  }

  const setDeleteDialogOpen = (open: boolean): void => {
    deleteDialog.open = open
    if (!open && !deleteDialog.deleting) {
      inspectionToken += 1
      resetDeleteDialog()
    }
  }

  const requestDelete = async (asset: MediaAsset): Promise<void> => {
    const token = ++inspectionToken
    deleteDialog.asset = asset
    deleteDialog.references = []
    deleteDialog.total = 0
    deleteDialog.confirmation = ''
    deleteDialog.loading = true
    deleteDialog.open = true
    try {
      const report = await mediaApi.getAssetReferences(asset.id)
      if (token !== inspectionToken) return
      deleteDialog.references = report.references || []
      deleteDialog.total = report.total || 0
    } catch (error) {
      if (token === inspectionToken) {
        console.error('Failed to inspect media references:', error)
        setDeleteDialogOpen(false)
      }
    } finally {
      if (token === inspectionToken) {
        deleteDialog.loading = false
      }
    }
  }

  const deleteAsset = async (): Promise<void> => {
    const asset = deleteDialog.asset
    if (!asset) return
    deleteDialog.deleting = true
    try {
      await mediaApi.deleteAsset(asset.id, { confirm: deleteDialog.confirmation })
      toast.success('媒体资源已永久删除')
      setDeleteDialogOpen(false)
      await refreshAssets()
    } catch (error) {
      console.error('Failed to delete media asset:', error)
      const payload = ((error as AxiosError<MediaAssetInUsePayload>).response?.data || {}) as MediaAssetInUsePayload
      if (payload.code === 'media_asset_in_use') {
        deleteDialog.references = payload.references || []
        deleteDialog.total = payload.total || deleteDialog.references.length
      }
    } finally {
      deleteDialog.deleting = false
    }
  }

  return {
    deleteDialog,
    requestDelete,
    deleteAsset,
    setDeleteDialogOpen,
  }
}
