import { reactive } from 'vue'
import { toast } from 'vue-sonner'
import mediaApi from '@/api/media'

export const useMediaDeletion = (refreshAssets) => {
  const deleteDialog = reactive({
    open: false,
    asset: null,
    references: [],
    total: 0,
    loading: false,
    deleting: false,
    confirmation: '',
  })
  let inspectionToken = 0

  const resetDeleteDialog = () => {
    deleteDialog.asset = null
    deleteDialog.references = []
    deleteDialog.total = 0
    deleteDialog.loading = false
    deleteDialog.deleting = false
    deleteDialog.confirmation = ''
  }

  const setDeleteDialogOpen = (open) => {
    deleteDialog.open = open
    if (!open && !deleteDialog.deleting) {
      inspectionToken += 1
      resetDeleteDialog()
    }
  }

  const requestDelete = async (asset) => {
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

  const deleteAsset = async () => {
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
      const payload = error?.response?.data || {}
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
