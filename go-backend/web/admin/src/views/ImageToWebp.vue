<template>
  <div class="space-y-4">
    <AdminPageHeader
      :title="t('imageToWebp.title')"
      :description="t('imageToWebp.description')"
    >
      <template #actions>
        <Button
          v-if="sourceFile"
          variant="outline"
          size="sm"
          :disabled="converting"
          @click="resetTool"
        >
          <RotateCcw class="size-3.5" />
          {{ t('imageToWebp.chooseAnother') }}
        </Button>
        <Button
          v-if="outputPreviewUrl"
          size="sm"
          :disabled="converting"
          @click="downloadWebp"
        >
          <Download class="size-3.5" />
          {{ t('imageToWebp.download') }}
        </Button>
      </template>
    </AdminPageHeader>

    <div class="grid min-h-0 gap-4 xl:grid-cols-[minmax(0,1fr)_20rem]">
      <Card class="min-h-0">
        <CardHeader class="border-b">
          <div class="flex min-w-0 items-start justify-between gap-3">
            <div class="min-w-0">
              <CardTitle class="flex items-center gap-2">
                <ImageDown class="size-4 text-primary" />
                {{ t('imageToWebp.workspaceTitle') }}
              </CardTitle>
              <CardDescription>{{ t('imageToWebp.workspaceDescription') }}</CardDescription>
            </div>
            <span class="shrink-0 font-mono text-[10px] font-black tracking-wider text-muted-foreground">
              PNG / JPG / JPEG / WEBP
            </span>
          </div>
        </CardHeader>

        <CardContent class="space-y-4 p-4">
          <div class="grid gap-4 lg:grid-cols-2">
            <section class="min-w-0">
              <div class="mb-2 flex items-center justify-between gap-3">
                <div>
                  <p class="text-xs font-black uppercase tracking-wider">{{ t('imageToWebp.sourceTitle') }}</p>
                  <p class="mt-1 text-[10px] font-bold text-muted-foreground">{{ t('imageToWebp.sourceDescription') }}</p>
                </div>
                <span v-if="sourceFile" class="shrink-0 rounded-full bg-muted px-2 py-1 font-mono text-[10px] font-black text-muted-foreground">
                  {{ formatBytes(sourceFile.size) }}
                </span>
              </div>

              <div
                class="image-webp-dropzone"
                :class="{
                  'image-webp-dropzone--active': isDragging,
                  'image-webp-dropzone--filled': sourceFile,
                }"
                role="button"
                tabindex="0"
                :aria-label="t('imageToWebp.selectFile')"
                @click="triggerFilePicker"
                @keydown.enter.prevent="triggerFilePicker"
                @keydown.space.prevent="triggerFilePicker"
                @dragenter.prevent.stop="isDragging = true"
                @dragover.prevent.stop="isDragging = true"
                @dragleave.prevent.stop="handleDragLeave"
                @drop.prevent.stop="handleDrop"
              >
                <input
                  ref="fileInput"
                  class="hidden"
                  type="file"
                  accept="image/png,image/jpeg,image/jpg,image/webp,.png,.jpg,.jpeg,.webp"
                  @change="handleFileInput"
                />

                <div v-if="sourcePreviewUrl" class="flex h-full min-h-72 w-full flex-col items-center justify-center gap-4 p-4">
                  <div class="image-webp-preview-surface flex min-h-48 w-full items-center justify-center overflow-hidden border border-border/70 p-3">
                    <img
                      :src="sourcePreviewUrl"
                      :alt="sourceFile?.name || t('imageToWebp.sourceTitle')"
                      class="max-h-64 max-w-full object-contain"
                    />
                  </div>
                  <div class="min-w-0 max-w-full text-center">
                    <p class="truncate text-xs font-black">{{ sourceFile?.name }}</p>
                    <p v-if="sourceFile && sourceWidth && sourceHeight" class="mt-1 font-mono text-[10px] font-bold text-muted-foreground">
                      {{ t('imageToWebp.sourceFormat', { format: formatSourceType(sourceFile) }) }}
                      <span aria-hidden="true"> · </span>
                      {{ t('imageToWebp.dimensions', { width: sourceWidth, height: sourceHeight }) }}
                    </p>
                  </div>
                </div>

                <div v-else class="flex min-h-72 flex-col items-center justify-center gap-3 px-5 text-center">
                  <span class="flex size-12 items-center justify-center rounded-2xl bg-primary/10 text-primary">
                    <ImagePlus class="size-6" />
                  </span>
                  <p class="text-sm font-black">{{ t('imageToWebp.selectFile') }}</p>
                  <p class="text-[10px] font-bold text-muted-foreground">{{ t('imageToWebp.dropHint') }}</p>
                </div>
              </div>
            </section>

            <section class="min-w-0">
              <div class="mb-2 flex items-center justify-between gap-3">
                <div>
                  <p class="text-xs font-black uppercase tracking-wider">{{ t('imageToWebp.outputTitle') }}</p>
                  <p class="mt-1 text-[10px] font-bold text-muted-foreground">{{ t('imageToWebp.outputDescription') }}</p>
                </div>
                <span
                  v-if="outputPreviewUrl"
                  class="shrink-0 rounded-full bg-emerald-500/10 px-2 py-1 font-mono text-[10px] font-black text-emerald-700"
                >
                  {{ t('imageToWebp.ready') }}
                </span>
              </div>

              <div class="image-webp-preview-surface flex min-h-72 items-center justify-center overflow-hidden border border-border/70 p-4">
                <img
                  v-if="outputPreviewUrl"
                  :src="outputPreviewUrl"
                  :alt="t('imageToWebp.outputTitle')"
                  class="h-auto object-contain"
                  :style="outputPreviewStyle"
                />
                <div v-else class="flex flex-col items-center justify-center gap-3 px-5 text-center text-muted-foreground">
                  <LoaderCircle v-if="converting" class="size-7 animate-spin text-primary" />
                  <ImageDown v-else class="size-7 opacity-50" />
                  <p class="text-xs font-black">
                    {{ converting ? t('imageToWebp.processing') : t('imageToWebp.waiting') }}
                  </p>
                </div>
              </div>

              <div v-if="outputPreviewUrl" class="mt-3 flex flex-wrap items-center gap-2 text-[10px] font-mono font-bold text-muted-foreground">
                <span>{{ t('imageToWebp.dimensions', { width: outputWidth, height: outputHeight }) }}</span>
                <span aria-hidden="true">·</span>
                <span>{{ formatBytes(outputBytes) }}</span>
              </div>
            </section>
          </div>

          <div v-if="errorMessage" class="flex items-start gap-2 border border-rose-500/25 bg-rose-500/5 px-3 py-2.5 text-xs font-bold text-rose-700">
            <TriangleAlert class="mt-0.5 size-4 shrink-0" />
            <span>{{ errorMessage }}</span>
          </div>
        </CardContent>
      </Card>

      <Card class="h-fit">
        <CardHeader class="border-b">
          <CardTitle>{{ t('imageToWebp.settingsTitle') }}</CardTitle>
          <CardDescription>{{ t('imageToWebp.settingsDescription') }}</CardDescription>
        </CardHeader>

        <CardContent class="p-4">
          <form class="space-y-5" @submit.prevent="convertCurrent">
            <label class="grid gap-1.5">
              <span class="flex items-center justify-between gap-3 text-[10px] font-black uppercase tracking-wider">
                <span>{{ t('imageToWebp.quality') }}</span>
                <output class="font-mono text-primary">{{ settings.quality }}%</output>
              </span>
              <div class="grid grid-cols-[minmax(0,1fr)_4.5rem] items-center gap-2">
                <input
                  v-model.number="settings.quality"
                  type="range"
                  min="10"
                  max="100"
                  step="1"
                  class="h-2 w-full cursor-pointer accent-primary"
                />
                <Input v-model.number="settings.quality" type="number" min="10" max="100" step="1" class="text-right" />
              </div>
              <span class="text-[10px] font-bold leading-4 text-muted-foreground">{{ t('imageToWebp.qualityHelp') }}</span>
            </label>

            <div class="grid gap-1.5">
              <span class="flex items-center justify-between gap-3 text-[10px] font-black uppercase tracking-wider">
                <span>{{ t('imageToWebp.outputSize') }}</span>
                <output class="font-mono text-primary">
                  {{ settings.targetWidth && settings.targetHeight
                    ? `${settings.targetWidth} × ${settings.targetHeight} px`
                    : t('imageToWebp.autoSize') }}
                </output>
              </span>
              <div class="grid grid-cols-2 gap-2">
                <label class="grid gap-1">
                  <span class="text-[10px] font-bold text-muted-foreground">{{ t('imageToWebp.outputWidth') }}</span>
                  <Input
                    v-model.number="settings.targetWidth"
                    type="number"
                    min="1"
                    max="16384"
                    step="1"
                    :disabled="!sourceFile || converting"
                    class="text-right"
                  />
                </label>
                <label class="grid gap-1">
                  <span class="text-[10px] font-bold text-muted-foreground">{{ t('imageToWebp.outputHeight') }}</span>
                  <Input
                    v-model.number="settings.targetHeight"
                    type="number"
                    min="1"
                    max="16384"
                    step="1"
                    :disabled="!sourceFile || converting"
                    class="text-right"
                  />
                </label>
              </div>
              <span class="text-[10px] font-bold leading-4 text-muted-foreground">{{ t('imageToWebp.outputSizeHelp') }}</span>
            </div>

            <div class="flex items-start gap-2 border border-emerald-500/20 bg-emerald-500/5 px-3 py-2.5 text-[10px] font-bold leading-4 text-emerald-800">
              <ShieldCheck class="mt-0.5 size-4 shrink-0" />
              <span>{{ t('imageToWebp.localOnly') }}</span>
            </div>

            <Button class="w-full" size="lg" type="submit" :disabled="!sourceFile || converting">
              <LoaderCircle v-if="converting" class="size-4 animate-spin" />
              <ImageDown v-else class="size-4" />
              {{ converting ? t('imageToWebp.processing') : t('imageToWebp.convert') }}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { Download, ImageDown, ImagePlus, LoaderCircle, RotateCcw, ShieldCheck, TriangleAlert } from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { useAdminI18n } from '@/i18n'

interface WebpSettings {
  quality: number
  targetWidth: number | null
  targetHeight: number | null
}

const MAX_OUTPUT_DIMENSION = 16384
const { t } = useAdminI18n()
const fileInput = ref<HTMLInputElement | null>(null)
const sourceFile = ref<File | null>(null)
const sourcePreviewUrl = ref('')
const outputPreviewUrl = ref('')
const outputBlob = ref<Blob | null>(null)
const sourceWidth = ref(0)
const sourceHeight = ref(0)
const outputWidth = ref(0)
const outputHeight = ref(0)
const isDragging = ref(false)
const converting = ref(false)
const errorMessage = ref('')
const conversionRunID = ref(0)
const settings = reactive<WebpSettings>({
  quality: 90,
  targetWidth: null,
  targetHeight: null,
})

const outputBytes = computed(() => outputBlob.value?.size || 0)
const outputPreviewStyle = computed<Record<string, string>>(() => {
  if (!outputWidth.value || !outputHeight.value) return {}
  return {
    aspectRatio: `${outputWidth.value} / ${outputHeight.value}`,
    maxHeight: '16rem',
    maxWidth: '100%',
    width: `${outputWidth.value}px`,
  }
})

let scheduledConversionTimer: number | null = null
let syncingSettings = false

const clearScheduledConversion = (): void => {
  if (scheduledConversionTimer === null) return
  window.clearTimeout(scheduledConversionTimer)
  scheduledConversionTimer = null
}

const syncSettings = (updates: Partial<WebpSettings>): void => {
  syncingSettings = true
  Object.assign(settings, updates)
  syncingSettings = false
}

const revokeObjectURL = (url: string): void => {
  if (url) URL.revokeObjectURL(url)
}

const releaseOutput = (): void => {
  revokeObjectURL(outputPreviewUrl.value)
  outputPreviewUrl.value = ''
  outputBlob.value = null
}

const releaseSourcePreview = (): void => {
  revokeObjectURL(sourcePreviewUrl.value)
  sourcePreviewUrl.value = ''
}

const formatBytes = (bytes: number): string => {
  if (bytes < 1024) return `${bytes} B`
  const kilobytes = bytes / 1024
  if (kilobytes < 1024) return `${kilobytes.toFixed(kilobytes >= 100 ? 0 : 1)} KB`
  const megabytes = kilobytes / 1024
  return `${megabytes.toFixed(megabytes >= 100 ? 0 : 1)} MB`
}

const formatSourceType = (file: File): string => {
  const extension = file.name.split('.').pop()?.toUpperCase() || ''
  if (extension) return extension
  return file.type.replace(/^image\//i, '').toUpperCase() || 'IMAGE'
}

const isSupportedFile = (file: File): boolean => {
  const extension = file.name.split('.').pop()?.toLowerCase() || ''
  const supportedType = ['', 'image/png', 'image/jpeg', 'image/jpg', 'image/webp'].includes(file.type.toLowerCase())
  return ['png', 'jpg', 'jpeg', 'webp'].includes(extension) && supportedType
}

const clampQuality = (value: number): number => {
  const parsed = Number(value)
  if (!Number.isFinite(parsed)) return 90
  return Math.round(Math.min(100, Math.max(10, parsed)))
}

const normalizeDimension = (value: number | null, fallback: number): number => {
  const safeFallback = Math.min(MAX_OUTPUT_DIMENSION, Math.max(1, Math.round(fallback)))
  const parsed = Number(value)
  if (!Number.isFinite(parsed) || parsed <= 0) return safeFallback
  return Math.round(Math.min(MAX_OUTPUT_DIMENSION, Math.max(1, parsed)))
}

const loadImage = (url: string): Promise<HTMLImageElement> => new Promise((resolve, reject) => {
  const image = new Image()
  image.onload = () => resolve(image)
  image.onerror = () => reject(new Error('image-load-failed'))
  image.src = url
})

const imageToWebp = (
  image: HTMLImageElement,
  quality: number,
  targetWidth: number,
  targetHeight: number,
): Promise<Blob> => {
  const naturalWidth = image.naturalWidth || image.width
  const naturalHeight = image.naturalHeight || image.height
  if (!naturalWidth || !naturalHeight) return Promise.reject(new Error('image-load-failed'))

  const canvas = document.createElement('canvas')
  canvas.width = targetWidth
  canvas.height = targetHeight

  const context = canvas.getContext('2d', { willReadFrequently: false })
  if (!context) return Promise.reject(new Error('canvas-unavailable'))

  context.imageSmoothingEnabled = true
  context.imageSmoothingQuality = 'high'
  context.clearRect(0, 0, targetWidth, targetHeight)
  context.drawImage(image, 0, 0, naturalWidth, naturalHeight, 0, 0, targetWidth, targetHeight)

  return new Promise((resolve, reject) => {
    canvas.toBlob((blob) => {
      if (!blob || blob.type.toLowerCase() !== 'image/webp') {
        reject(new Error('webp-unsupported'))
        return
      }
      resolve(blob)
    }, 'image/webp', quality / 100)
  })
}

const convertImage = async (image: HTMLImageElement, runID: number): Promise<void> => {
  await nextTick()
  if (runID !== conversionRunID.value) return

  const quality = clampQuality(settings.quality)
  const naturalWidth = image.naturalWidth || image.width
  const naturalHeight = image.naturalHeight || image.height
  const targetWidth = normalizeDimension(settings.targetWidth, naturalWidth)
  const targetHeight = normalizeDimension(settings.targetHeight, naturalHeight)
  syncSettings({
    quality,
    targetWidth,
    targetHeight,
  })
  const blob = await imageToWebp(image, quality, targetWidth, targetHeight)
  if (runID !== conversionRunID.value) return

  releaseOutput()
  outputBlob.value = blob
  outputWidth.value = targetWidth
  outputHeight.value = targetHeight
  outputPreviewUrl.value = URL.createObjectURL(blob)
}

const chooseFile = async (file: File | null): Promise<void> => {
  if (!file) return
  if (!isSupportedFile(file)) {
    errorMessage.value = t('imageToWebp.unsupported')
    return
  }

  const runID = ++conversionRunID.value
  converting.value = true
  errorMessage.value = ''
  releaseSourcePreview()
  releaseOutput()
  sourceFile.value = file
  sourcePreviewUrl.value = URL.createObjectURL(file)
  sourceWidth.value = 0
  sourceHeight.value = 0
  outputWidth.value = 0
  outputHeight.value = 0

  try {
    const image = await loadImage(sourcePreviewUrl.value)
    if (runID !== conversionRunID.value) return
    sourceWidth.value = image.naturalWidth || image.width
    sourceHeight.value = image.naturalHeight || image.height
    syncSettings({
      targetWidth: normalizeDimension(null, sourceWidth.value),
      targetHeight: normalizeDimension(null, sourceHeight.value),
    })
    await convertImage(image, runID)
  } catch (error) {
    if (runID === conversionRunID.value) {
      errorMessage.value = error instanceof Error && error.message === 'webp-unsupported'
        ? t('imageToWebp.webpUnsupported')
        : t('imageToWebp.convertFailed')
    }
  } finally {
    if (runID === conversionRunID.value) converting.value = false
  }
}

const convertCurrent = async (): Promise<void> => {
  if (!sourceFile.value || !sourcePreviewUrl.value) {
    errorMessage.value = t('imageToWebp.noFile')
    return
  }
  if (converting.value) return

  const runID = ++conversionRunID.value
  converting.value = true
  errorMessage.value = ''
  releaseOutput()
  outputWidth.value = 0
  outputHeight.value = 0

  try {
    const image = await loadImage(sourcePreviewUrl.value)
    if (runID !== conversionRunID.value) return
    sourceWidth.value = image.naturalWidth || image.width
    sourceHeight.value = image.naturalHeight || image.height
    await convertImage(image, runID)
  } catch (error) {
    if (runID === conversionRunID.value) {
      errorMessage.value = error instanceof Error && error.message === 'webp-unsupported'
        ? t('imageToWebp.webpUnsupported')
        : t('imageToWebp.convertFailed')
    }
  } finally {
    if (runID === conversionRunID.value) converting.value = false
  }
}

const scheduleConvertCurrent = (): void => {
  clearScheduledConversion()
  scheduledConversionTimer = window.setTimeout(() => {
    scheduledConversionTimer = null
    if (!sourceFile.value || !sourcePreviewUrl.value) return
    if (converting.value) {
      scheduleConvertCurrent()
      return
    }
    void convertCurrent()
  }, 260)
}

const triggerFilePicker = (): void => {
  if (!converting.value) fileInput.value?.click()
}

const handleFileInput = (event: Event): void => {
  const input = event.target instanceof HTMLInputElement ? event.target : null
  const file = input?.files?.[0] || null
  if (input) input.value = ''
  void chooseFile(file)
}

const handleDrop = (event: DragEvent): void => {
  isDragging.value = false
  void chooseFile(event.dataTransfer?.files?.[0] || null)
}

const handleDragLeave = (event: DragEvent): void => {
  const currentTarget = event.currentTarget
  const relatedTarget = event.relatedTarget
  if (currentTarget instanceof HTMLElement && relatedTarget instanceof Node && currentTarget.contains(relatedTarget)) return
  isDragging.value = false
}

const downloadWebp = (): void => {
  if (!outputPreviewUrl.value || !outputBlob.value) return
  const baseName = (sourceFile.value?.name || 'image').replace(/\.[^/.]+$/, '') || 'image'
  const anchor = document.createElement('a')
  anchor.href = outputPreviewUrl.value
  anchor.download = `${baseName}.webp`
  document.body.appendChild(anchor)
  anchor.click()
  anchor.remove()
}

const resetTool = (): void => {
  ++conversionRunID.value
  clearScheduledConversion()
  converting.value = false
  sourceFile.value = null
  releaseSourcePreview()
  releaseOutput()
  syncSettings({
    targetWidth: null,
    targetHeight: null,
  })
  sourceWidth.value = 0
  sourceHeight.value = 0
  outputWidth.value = 0
  outputHeight.value = 0
  errorMessage.value = ''
  isDragging.value = false
  if (fileInput.value) fileInput.value.value = ''
}

watch(
  () => [settings.quality, settings.targetWidth, settings.targetHeight],
  () => {
    if (syncingSettings || !sourceFile.value || !sourcePreviewUrl.value) return
    scheduleConvertCurrent()
  },
  { flush: 'sync' },
)

onBeforeUnmount(() => {
  ++conversionRunID.value
  clearScheduledConversion()
  releaseSourcePreview()
  releaseOutput()
})
</script>

<style scoped>
.image-webp-dropzone {
  display: flex;
  min-height: 18rem;
  align-items: stretch;
  justify-content: center;
  overflow: hidden;
  border: 1px dashed #cbd5e1;
  border-radius: 1rem;
  background: rgb(248 250 252 / 0.68);
  cursor: pointer;
  transition:
    border-color 160ms ease,
    background-color 160ms ease,
    box-shadow 160ms ease;
}

.image-webp-dropzone:hover,
.image-webp-dropzone:focus-visible,
.image-webp-dropzone--active {
  border-color: rgb(5 150 105 / 0.65);
  background: rgb(236 253 245 / 0.7);
  box-shadow: 0 0 0 3px rgb(5 150 105 / 0.1);
  outline: none;
}

.image-webp-dropzone--filled {
  background: rgb(248 250 252 / 0.9);
}

.image-webp-preview-surface {
  min-height: 18rem;
  border-radius: 1rem;
  background-color: #f8fafc;
  background-image:
    linear-gradient(45deg, #e2e8f0 25%, transparent 25%),
    linear-gradient(-45deg, #e2e8f0 25%, transparent 25%),
    linear-gradient(45deg, transparent 75%, #e2e8f0 75%),
    linear-gradient(-45deg, transparent 75%, #e2e8f0 75%);
  background-position: 0 0, 0 9px, 9px -9px, -9px 0;
  background-size: 18px 18px;
}

@media (max-width: 639px) {
  .image-webp-preview-surface,
  .image-webp-dropzone {
    min-height: 15rem;
  }
}
</style>
