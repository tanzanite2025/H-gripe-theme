<template>
  <div class="my-2 space-y-2 text-sm">
    <div class="flex flex-wrap items-center gap-3">
      <button type="button" class="font-semibold underline disabled:opacity-50" :disabled="loading" @click="loadAuthorizedPrivateAttachmentPreview">{{ loading ? t('privateOrderAttachments.loading', 'Loading…') : label }}</button>
      <button type="button" class="underline disabled:opacity-50" :disabled="loading" @click="downloadAuthorizedPrivateAttachment">{{ t('privateOrderAttachments.download', 'Download') }}</button>
    </div>
    <p v-if="error" role="alert" class="text-red-600">{{ error }}</p>
    <template v-if="previewURL">
      <img v-if="previewKind === 'image'" :src="previewURL" :alt="label" class="max-h-80 max-w-full rounded object-contain">
      <video v-else-if="previewKind === 'video'" :src="previewURL" controls preload="metadata" class="max-h-80 max-w-full rounded" />
      <p v-else>{{ t('privateOrderAttachments.downloadToView', 'Download this file to view it.') }}</p>
      <button type="button" class="block underline" @click="clearAuthorizedPrivateAttachmentPreview">{{ t('privateOrderAttachments.closePreview', 'Close preview') }}</button>
    </template>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import { useAuth } from '~/composables/useAuth'

const props = withDefaults(defineProps<{ path: string; label: string; filename?: string; claimToken?: string }>(), { filename: 'attachment', claimToken: '' })
const { t } = useI18n()
const auth = useAuth()
const loading = ref(false)
const error = ref('')
const previewURL = ref('')
const previewKind = ref('')
let requestVersion = 0
let controller: AbortController | null = null
const downloadURLs = new Map<string, ReturnType<typeof setTimeout>>()

function clearAuthorizedPrivateAttachmentPreview() {
  if (previewURL.value) URL.revokeObjectURL(previewURL.value)
  previewURL.value = ''
  previewKind.value = ''
}

function cancelAuthorizedPrivateAttachmentRequestAndReleaseURLs() {
  requestVersion += 1
  controller?.abort()
  controller = null
  loading.value = false
  error.value = ''
  clearAuthorizedPrivateAttachmentPreview()
  for (const [url, timer] of downloadURLs) { clearTimeout(timer); URL.revokeObjectURL(url) }
  downloadURLs.clear()
}

async function fetchAuthorizedPrivateAttachmentBlob(): Promise<Blob> {
  if (!/^\/(warranty\/(claims|orders)\/|orders\/[^/]+\/after-sales\/attachments\/)/.test(props.path)) throw new Error('Invalid attachment path')
  controller?.abort()
  controller = new AbortController()
  return auth.request<Blob>(props.path, {
    responseType: 'blob', cache: 'no-store', redirect: 'error', signal: controller.signal,
    headers: props.claimToken ? { 'X-Warranty-Claim-Token': props.claimToken } : {},
  })
}

async function loadAuthorizedPrivateAttachmentPreview() {
  const version = ++requestVersion
  loading.value = true
  error.value = ''
  try {
    const blob = await fetchAuthorizedPrivateAttachmentBlob()
    if (version !== requestVersion) return
    clearAuthorizedPrivateAttachmentPreview()
    previewKind.value = /^image\/(jpeg|png|webp|gif|avif)$/.test(blob.type) ? 'image' : /^video\/(mp4|webm|quicktime)$/.test(blob.type) ? 'video' : 'download'
    previewURL.value = URL.createObjectURL(blob)
  } catch {
    if (version === requestVersion) error.value = t('privateOrderAttachments.readFailed', 'This attachment could not be opened. Refresh and try again.')
  } finally { if (version === requestVersion) loading.value = false }
}

function authorizedPrivateAttachmentDownloadFilename(contentType: string): string {
  const extension = ({ 'image/jpeg': '.jpg', 'image/png': '.png', 'image/webp': '.webp', 'image/gif': '.gif', 'image/avif': '.avif', 'video/mp4': '.mp4', 'video/webm': '.webm', 'video/quicktime': '.mov', 'application/pdf': '.pdf' } as Record<string, string>)[contentType] || ''
  return /\.[a-z0-9]{1,10}$/i.test(props.filename) ? props.filename : props.filename + extension
}

async function downloadAuthorizedPrivateAttachment() {
  const version = ++requestVersion
  loading.value = true
  error.value = ''
  try {
    const blob = await fetchAuthorizedPrivateAttachmentBlob()
    if (version !== requestVersion) return
    const url = URL.createObjectURL(blob)
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = authorizedPrivateAttachmentDownloadFilename(blob.type)
    document.body.appendChild(anchor)
    anchor.click()
    anchor.remove()
    downloadURLs.set(url, setTimeout(() => { URL.revokeObjectURL(url); downloadURLs.delete(url) }, 30000))
  } catch {
    if (version === requestVersion) error.value = t('privateOrderAttachments.downloadFailed', 'This attachment could not be downloaded. Refresh and try again.')
  } finally { if (version === requestVersion) loading.value = false }
}

watch(() => [props.path, props.claimToken, auth.isAuthenticated.value, auth.user.value?.id], cancelAuthorizedPrivateAttachmentRequestAndReleaseURLs)
onBeforeUnmount(cancelAuthorizedPrivateAttachmentRequestAndReleaseURLs)
</script>
