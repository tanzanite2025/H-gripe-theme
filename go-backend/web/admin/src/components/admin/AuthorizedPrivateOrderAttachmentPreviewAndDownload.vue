<template>
  <div class="space-y-2">
    <div class="flex flex-wrap items-center gap-2">
      <button type="button" class="text-xs font-semibold underline disabled:opacity-50" :disabled="loading || !path" @click="loadAuthorizedPrivateAttachmentPreview">{{ loading ? '加载中…' : label }}</button>
      <button type="button" class="text-xs underline disabled:opacity-50" :disabled="loading || !path" @click="downloadAuthorizedPrivateAttachment">下载</button>
    </div>
    <p v-if="error" role="alert" class="text-xs text-destructive">{{ error }}</p>
    <template v-if="previewURL">
      <img v-if="previewKind === 'image'" :src="previewURL" :alt="label" class="max-h-72 max-w-full rounded object-contain">
      <video v-else-if="previewKind === 'video'" :src="previewURL" controls preload="metadata" class="max-h-72 max-w-full rounded" />
      <p v-else class="text-xs text-muted-foreground">此文件请下载后查看。</p>
      <button type="button" class="block text-xs underline" @click="clearAuthorizedPrivateAttachmentPreview">关闭预览</button>
    </template>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import axios from '@/utils/axios'

const props = withDefaults(defineProps<{ path: string; label: string; filename?: string }>(), { filename: 'attachment' })
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
  if (!/^\/api\/admin\/(warranty\/(claims|shipment-records)\/|after-sales\/)/.test(props.path)) throw new Error('Invalid attachment path')
  controller?.abort()
  controller = new AbortController()
  const response = await axios.get<Blob>(props.path, { responseType: 'blob', signal: controller.signal, suppressGlobalErrorToast: true })
  return response.data
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
    if (version === requestVersion) error.value = '附件无法读取，请刷新后重试。'
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
    if (version === requestVersion) error.value = '附件无法下载，请刷新后重试。'
  } finally { if (version === requestVersion) loading.value = false }
}

watch(() => props.path, cancelAuthorizedPrivateAttachmentRequestAndReleaseURLs)
onBeforeUnmount(cancelAuthorizedPrivateAttachmentRequestAndReleaseURLs)
</script>
