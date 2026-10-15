import { onScopeDispose, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import customerServiceApi from '@/api/customerService'
import type { MediaAsset } from '@/api/media'

// A completed upload belongs to the captured conversation. A switched or
// disposed panel may never send it into whichever conversation is now open.
export function useConversationAttachmentUploadWithinCurrentAuthorizedConversation(currentConversationId: () => number | string | null | undefined) {
  const attachmentUploading = ref(false)
  let requestGeneration = 0
  let activeRequest: AbortController | undefined
  let disposed = false
  const cancelConversationAttachmentUploadForChangedOrDisposedConversation = (): void => {
    requestGeneration++
    activeRequest?.abort()
    attachmentUploading.value = false
  }
  watch(currentConversationId, cancelConversationAttachmentUploadForChangedOrDisposedConversation, { flush: 'sync' })
  onScopeDispose(() => { disposed = true; cancelConversationAttachmentUploadForChangedOrDisposedConversation() })

  const uploadAttachmentForCurrentConversation = async (file: File, mediaType: 'image' | 'video'): Promise<MediaAsset | null> => {
    const conversationId = currentConversationId()
    if (disposed || !conversationId || attachmentUploading.value) return null
    const generation = ++requestGeneration
    const controller = new AbortController()
    activeRequest = controller
    attachmentUploading.value = true
    try {
      const formData = new FormData()
      formData.append('file', file)
      formData.append('media_type', mediaType)
      const asset = await customerServiceApi.uploadConversationAttachment(conversationId, formData, controller.signal)
      if (disposed || controller.signal.aborted || generation !== requestGeneration || String(conversationId) !== String(currentConversationId())) return null
      if (!String(asset.url || '').trim()) throw new Error('Attachment upload did not return a usable receipt or URL')
      return asset
    } catch (error) {
      if (disposed || controller.signal.aborted || generation !== requestGeneration) return null
      console.error('Failed to upload conversation attachment:', error)
      toast.error(mediaType === 'image' ? '图片上传失败' : '视频上传失败')
      return null
    } finally {
      if (!disposed && generation === requestGeneration) attachmentUploading.value = false
    }
  }
  return { attachmentUploading, uploadAttachmentForCurrentConversation }
}
