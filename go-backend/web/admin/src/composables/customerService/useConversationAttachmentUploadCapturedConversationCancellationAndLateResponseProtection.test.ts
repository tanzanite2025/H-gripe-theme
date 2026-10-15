import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { effectScope, ref, type EffectScope } from 'vue'
import { toast } from 'vue-sonner'
import customerServiceApi from '@/api/customerService'
import type { MediaAsset } from '@/api/media'
import { useConversationAttachmentUploadWithinCurrentAuthorizedConversation } from './useConversationAttachmentUploadWithinCurrentAuthorizedConversation'

vi.mock('@/api/customerService', () => ({ default: { uploadConversationAttachment: vi.fn() } }))
vi.mock('vue-sonner', () => ({ toast: { error: vi.fn() } }))

const testScopes: EffectScope[] = []
const attachmentFile = new File(['image'], 'conversation-attachment.png', { type: 'image/png' })

function createConversationAttachmentUploaderWithDisposableScope() {
  const scope = effectScope()
  testScopes.push(scope)
  const conversationId = ref<number | null>(12)
  const uploader = scope.run(() => useConversationAttachmentUploadWithinCurrentAuthorizedConversation(() => conversationId.value))!
  return { scope, conversationId, uploader }
}

function createDeferredConversationAttachmentUploadResponse() {
  let resolve!: (value: MediaAsset) => void
  let reject!: (error: Error) => void
  const promise = new Promise<MediaAsset>((resolvePromise, rejectPromise) => { resolve = resolvePromise; reject = rejectPromise })
  return { promise, resolve, reject }
}

beforeEach(() => {
  vi.resetAllMocks()
  vi.spyOn(console, 'error').mockImplementation(() => {})
  vi.mocked(customerServiceApi.uploadConversationAttachment).mockResolvedValue({ id: 1, url: '/authorized-preview', media_type: 'image' })
})

afterEach(() => {
  testScopes.splice(0).forEach((scope) => scope.stop())
  vi.restoreAllMocks()
})

describe('conversation attachment uploads remain owned by the captured authorized conversation', () => {
  it('uploads to the captured conversation and returns its authorized preview or compatibility URL', async () => {
    const { uploader } = createConversationAttachmentUploaderWithDisposableScope()
    const asset = await uploader.uploadAttachmentForCurrentConversation(attachmentFile, 'image')
    const [conversationId, formData, signal] = vi.mocked(customerServiceApi.uploadConversationAttachment).mock.calls[0]!
    expect(conversationId).toBe(12)
    expect(formData.get('file')).toBe(attachmentFile)
    expect(formData.get('media_type')).toBe('image')
    expect(signal).toBeInstanceOf(AbortSignal)
    expect(asset?.url).toBe('/authorized-preview')
    expect(uploader.attachmentUploading.value).toBe(false)
  })

  it('does not return an old attachment after switching away and back to the same conversation', async () => {
    const pendingUpload = createDeferredConversationAttachmentUploadResponse()
    vi.mocked(customerServiceApi.uploadConversationAttachment).mockReturnValueOnce(pendingUpload.promise)
    const { conversationId, uploader } = createConversationAttachmentUploaderWithDisposableScope()
    const upload = uploader.uploadAttachmentForCurrentConversation(attachmentFile, 'image')
    const signal = vi.mocked(customerServiceApi.uploadConversationAttachment).mock.calls[0]![2]!
    conversationId.value = 13
    conversationId.value = 12
    expect(signal.aborted).toBe(true)
    pendingUpload.resolve({ id: 9, url: '/old-preview' })
    expect(await upload).toBeNull()
    expect(toast.error).not.toHaveBeenCalled()
  })

  it('keeps the new uploading state when the old conversation request finishes', async () => {
    const oldUpload = createDeferredConversationAttachmentUploadResponse()
    const newUpload = createDeferredConversationAttachmentUploadResponse()
    vi.mocked(customerServiceApi.uploadConversationAttachment).mockReturnValueOnce(oldUpload.promise).mockReturnValueOnce(newUpload.promise)
    const { conversationId, uploader } = createConversationAttachmentUploaderWithDisposableScope()
    const firstUpload = uploader.uploadAttachmentForCurrentConversation(attachmentFile, 'image')
    conversationId.value = 13
    const secondUpload = uploader.uploadAttachmentForCurrentConversation(attachmentFile, 'image')
    oldUpload.resolve({ id: 9, url: '/old-preview' })
    expect(await firstUpload).toBeNull()
    expect(uploader.attachmentUploading.value).toBe(true)
    newUpload.resolve({ id: 2, url: '/current-preview' })
    expect((await secondUpload)?.id).toBe(2)
    expect(uploader.attachmentUploading.value).toBe(false)
  })

  it('does not emit upload errors or return an attachment after disposal', async () => {
    const pendingUpload = createDeferredConversationAttachmentUploadResponse()
    vi.mocked(customerServiceApi.uploadConversationAttachment).mockReturnValueOnce(pendingUpload.promise)
    const { scope, uploader } = createConversationAttachmentUploaderWithDisposableScope()
    const upload = uploader.uploadAttachmentForCurrentConversation(attachmentFile, 'image')
    const signal = vi.mocked(customerServiceApi.uploadConversationAttachment).mock.calls[0]![2]!
    scope.stop()
    pendingUpload.reject(new Error('Late network failure'))
    expect(await upload).toBeNull()
    expect(signal.aborted).toBe(true)
    expect(toast.error).not.toHaveBeenCalled()
    expect(await uploader.uploadAttachmentForCurrentConversation(attachmentFile, 'image')).toBeNull()
    expect(customerServiceApi.uploadConversationAttachment).toHaveBeenCalledTimes(1)
  })

  it('rejects duplicate uploads and responses without a usable URL', async () => {
    const pendingUpload = createDeferredConversationAttachmentUploadResponse()
    vi.mocked(customerServiceApi.uploadConversationAttachment).mockReturnValueOnce(pendingUpload.promise)
    const { uploader } = createConversationAttachmentUploaderWithDisposableScope()
    const upload = uploader.uploadAttachmentForCurrentConversation(attachmentFile, 'image')
    expect(await uploader.uploadAttachmentForCurrentConversation(attachmentFile, 'image')).toBeNull()
    pendingUpload.resolve({ id: 1, url: '  ' })
    expect(await upload).toBeNull()
    expect(toast.error).toHaveBeenCalledTimes(1)
    expect(uploader.attachmentUploading.value).toBe(false)
  })
})
