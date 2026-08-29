<template>
  <div class="space-y-4">
    <AdminPageHeader
      :title="t('mp4ToWebm.title')"
      :description="t('mp4ToWebm.description')"
    >
      <template #actions>
        <Button
          v-if="converting"
          variant="outline"
          size="sm"
          @click="cancelConversion"
        >
          <X class="size-3.5" />
          {{ t('mp4ToWebm.cancel') }}
        </Button>
        <Button
          v-else-if="sourceFile"
          variant="outline"
          size="sm"
          @click="resetTool"
        >
          <RotateCcw class="size-3.5" />
          {{ t('mp4ToWebm.chooseAnother') }}
        </Button>
        <Button
          v-if="outputPreviewUrl"
          size="sm"
          :disabled="converting"
          @click="downloadWebm"
        >
          <Download class="size-3.5" />
          {{ t('mp4ToWebm.download') }}
        </Button>
      </template>
    </AdminPageHeader>

    <div class="grid min-h-0 gap-4 xl:grid-cols-[minmax(0,1fr)_20rem]">
      <Card class="min-h-0">
        <CardHeader class="border-b">
          <div class="flex min-w-0 items-start justify-between gap-3">
            <div class="min-w-0">
              <CardTitle class="flex items-center gap-2">
                <FileVideo class="size-4 text-primary" />
                {{ t('mp4ToWebm.workspaceTitle') }}
              </CardTitle>
              <CardDescription>{{ t('mp4ToWebm.workspaceDescription') }}</CardDescription>
            </div>
            <span class="shrink-0 font-mono text-[10px] font-black tracking-wider text-muted-foreground">
              MP4 / MOV / M4V → WEBM
            </span>
          </div>
        </CardHeader>

        <CardContent class="space-y-4 p-4">
          <div class="grid gap-4 lg:grid-cols-2">
            <section class="min-w-0">
              <div class="mb-2 flex items-center justify-between gap-3">
                <div>
                  <p class="text-xs font-black uppercase tracking-wider">{{ t('mp4ToWebm.sourceTitle') }}</p>
                  <p class="mt-1 text-[10px] font-bold text-muted-foreground">{{ t('mp4ToWebm.sourceDescription') }}</p>
                </div>
                <span
                  v-if="sourceFile"
                  class="shrink-0 rounded-full bg-muted px-2 py-1 font-mono text-[10px] font-black text-muted-foreground"
                >
                  {{ formatBytes(sourceFile.size) }}
                </span>
              </div>

              <div
                class="video-converter-dropzone"
                :class="{
                  'video-converter-dropzone--active': isDragging,
                  'video-converter-dropzone--filled': sourceFile,
                }"
                role="button"
                tabindex="0"
                :aria-label="t('mp4ToWebm.selectFile')"
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
                  accept="video/mp4,video/quicktime,video/x-m4v,video/m4v,.mp4,.mov,.m4v"
                  @change="handleFileInput"
                />

                <div v-if="sourcePreviewUrl && !sourcePreviewError" class="flex h-full min-h-72 w-full flex-col items-center justify-center gap-4 p-4">
                  <div class="video-converter-preview-surface flex min-h-48 w-full items-center justify-center overflow-hidden border border-border/70 p-3">
                    <video
                      ref="sourceVideo"
                      :src="sourcePreviewUrl"
                      :aria-label="sourceFile?.name || t('mp4ToWebm.sourceTitle')"
                      class="video-converter-video"
                      controls
                      playsinline
                      preload="metadata"
                      @loadedmetadata="handleSourceMetadata"
                      @error="handleSourcePreviewError"
                    />
                  </div>
                  <div class="min-w-0 max-w-full text-center">
                    <p class="truncate text-xs font-black">{{ sourceFile?.name }}</p>
                    <p class="mt-1 font-mono text-[10px] font-bold text-muted-foreground">
                      {{ t('mp4ToWebm.sourceFormat', { format: formatSourceType(sourceFile) }) }}
                      <template v-if="sourceWidth && sourceHeight">
                        <span aria-hidden="true"> · </span>
                        {{ t('mp4ToWebm.dimensions', { width: sourceWidth, height: sourceHeight }) }}
                      </template>
                      <template v-if="sourceDuration">
                        <span aria-hidden="true"> · </span>
                        {{ formatDuration(sourceDuration) }}
                      </template>
                    </p>
                  </div>
                </div>

                <div v-else-if="sourcePreviewUrl" class="flex min-h-72 w-full flex-col items-center justify-center gap-3 px-5 text-center">
                  <span class="flex size-12 items-center justify-center rounded-2xl bg-primary/10 text-primary">
                    <FileVideo class="size-6" />
                  </span>
                  <p class="max-w-full truncate text-sm font-black">{{ sourceFile?.name }}</p>
                  <p class="text-[10px] font-bold text-muted-foreground">{{ t('mp4ToWebm.previewUnavailable') }}</p>
                </div>

                <div v-else class="flex min-h-72 flex-col items-center justify-center gap-3 px-5 text-center">
                  <span class="flex size-12 items-center justify-center rounded-2xl bg-primary/10 text-primary">
                    <FileVideo class="size-6" />
                  </span>
                  <p class="text-sm font-black">{{ t('mp4ToWebm.selectFile') }}</p>
                  <p class="text-[10px] font-bold text-muted-foreground">{{ t('mp4ToWebm.dropHint') }}</p>
                </div>
              </div>
            </section>

            <section class="min-w-0">
              <div class="mb-2 flex items-center justify-between gap-3">
                <div>
                  <p class="text-xs font-black uppercase tracking-wider">{{ t('mp4ToWebm.outputTitle') }}</p>
                  <p class="mt-1 text-[10px] font-bold text-muted-foreground">{{ t('mp4ToWebm.outputDescription') }}</p>
                </div>
                <span
                  v-if="outputPreviewUrl"
                  class="shrink-0 rounded-full bg-emerald-500/10 px-2 py-1 font-mono text-[10px] font-black text-emerald-700"
                >
                  {{ t('mp4ToWebm.ready') }}
                </span>
              </div>

              <div class="video-converter-preview-surface flex min-h-72 items-center justify-center overflow-hidden border border-border/70 p-4">
                <video
                  v-if="outputPreviewUrl && !outputPreviewError"
                  ref="outputVideo"
                  :src="outputPreviewUrl"
                  :aria-label="t('mp4ToWebm.outputTitle')"
                  class="video-converter-video"
                  controls
                  playsinline
                  preload="metadata"
                  @loadedmetadata="handleOutputMetadata"
                  @error="outputPreviewError = true"
                />
                <div v-else-if="outputPreviewUrl" class="flex flex-col items-center justify-center gap-3 px-5 text-center text-muted-foreground">
                  <FileVideo class="size-7 opacity-60" />
                  <p class="text-xs font-black">{{ t('mp4ToWebm.previewUnavailable') }}</p>
                  <p class="text-[10px] font-bold">{{ t('mp4ToWebm.downloadStillAvailable') }}</p>
                </div>
                <div v-else class="flex flex-col items-center justify-center gap-3 px-5 text-center text-muted-foreground">
                  <LoaderCircle v-if="converting" class="size-7 animate-spin text-primary" />
                  <FileVideo v-else class="size-7 opacity-50" />
                  <p class="text-xs font-black">
                    {{ converting ? processingLabel : t('mp4ToWebm.waiting') }}
                  </p>
                </div>
              </div>

              <div v-if="outputPreviewUrl" class="mt-3 grid grid-cols-2 gap-x-4 gap-y-2 border-t border-border/70 pt-3">
                <div class="min-w-0">
                  <p class="text-[10px] font-bold text-muted-foreground">{{ t('mp4ToWebm.outputFileSize') }}</p>
                  <p class="mt-1 truncate font-mono text-xs font-black">{{ formatBytes(outputBytes) }}</p>
                </div>
                <div class="min-w-0">
                  <p class="text-[10px] font-bold text-muted-foreground">{{ t('mp4ToWebm.outputDimensions') }}</p>
                  <p class="mt-1 truncate font-mono text-xs font-black">
                    {{ outputWidth && outputHeight ? t('mp4ToWebm.dimensions', { width: outputWidth, height: outputHeight }) : t('mp4ToWebm.notAvailable') }}
                  </p>
                </div>
                <div class="min-w-0">
                  <p class="text-[10px] font-bold text-muted-foreground">{{ t('mp4ToWebm.outputDuration') }}</p>
                  <p class="mt-1 truncate font-mono text-xs font-black">{{ formatDuration(displayOutputDuration) }}</p>
                </div>
                <div class="min-w-0">
                  <p class="text-[10px] font-bold text-muted-foreground">{{ t('mp4ToWebm.sizeChange') }}</p>
                  <p class="mt-1 truncate font-mono text-xs font-black" :class="sizeChangeClass">{{ sizeChangeLabel }}</p>
                </div>
              </div>
            </section>
          </div>

          <div v-if="converting" class="border border-primary/20 bg-primary/5 px-3 py-3">
            <div class="mb-2 flex items-center justify-between gap-3 text-[10px] font-black">
              <span>{{ processingLabel }}</span>
              <span class="font-mono text-primary">{{ progressPercent }}%</span>
            </div>
            <div class="video-converter-progress-track" role="progressbar" :aria-valuenow="progressPercent" aria-valuemin="0" aria-valuemax="100">
              <span class="video-converter-progress-bar" :style="{ width: `${progressPercent}%` }" />
            </div>
          </div>

          <div v-if="errorMessage" class="flex items-start gap-2 border border-rose-500/25 bg-rose-500/5 px-3 py-2.5 text-xs font-bold text-rose-700">
            <TriangleAlert class="mt-0.5 size-4 shrink-0" />
            <span>{{ errorMessage }}</span>
          </div>
        </CardContent>
      </Card>

      <Card class="h-fit">
        <CardHeader class="border-b">
          <CardTitle>{{ t('mp4ToWebm.settingsTitle') }}</CardTitle>
          <CardDescription>{{ t('mp4ToWebm.settingsDescription') }}</CardDescription>
        </CardHeader>

        <CardContent class="p-4">
          <form class="space-y-5" @submit.prevent="convertCurrent">
      <div class="grid gap-1.5">
              <span class="flex items-center justify-between gap-3 text-[10px] font-black uppercase tracking-wider">
                <span>{{ t('mp4ToWebm.outputSize') }}</span>
                <output class="font-mono text-primary">
                  {{ settings.targetWidth && settings.targetHeight
                    ? `${settings.targetWidth} × ${settings.targetHeight} px`
                    : t('mp4ToWebm.waitingDimensions') }}
                </output>
              </span>
              <div class="grid grid-cols-2 gap-2">
                <label class="grid gap-1">
                  <span class="text-[10px] font-bold text-muted-foreground">{{ t('mp4ToWebm.outputWidth') }}</span>
                  <Input
                    v-model.number="settings.targetWidth"
                    type="number"
                    min="2"
                    max="8192"
                    step="2"
                    :disabled="!sourceFile || converting"
                    class="text-right"
                  />
                </label>
                <label class="grid gap-1">
                  <span class="text-[10px] font-bold text-muted-foreground">{{ t('mp4ToWebm.outputHeight') }}</span>
                  <Input
                    v-model.number="settings.targetHeight"
                    type="number"
                    min="2"
                    max="8192"
                    step="2"
                    :disabled="!sourceFile || converting"
                    class="text-right"
                  />
                </label>
              </div>
              <span class="text-[10px] font-bold leading-4 text-muted-foreground">{{ t('mp4ToWebm.outputSizeHelp') }}</span>
            </div>

            <label class="grid gap-1.5">
              <span class="flex items-center justify-between gap-3 text-[10px] font-black uppercase tracking-wider">
                <span>{{ t('mp4ToWebm.quality') }}</span>
                <output class="font-mono text-primary">{{ settings.quality }}</output>
              </span>
              <div class="grid grid-cols-[minmax(0,1fr)_4.5rem] items-center gap-2">
                <input
                  v-model.number="settings.quality"
                  type="range"
                  min="18"
                  max="45"
                  step="1"
                  class="h-2 w-full cursor-pointer accent-primary"
                  :disabled="!sourceFile || converting"
                />
                <Input
                  v-model.number="settings.quality"
                  type="number"
                  min="18"
                  max="45"
                  step="1"
                  :disabled="!sourceFile || converting"
                  class="text-right"
                />
              </div>
              <span class="text-[10px] font-bold leading-4 text-muted-foreground">{{ t('mp4ToWebm.qualityHelp') }}</span>
            </label>

            <div class="flex items-center justify-between gap-4 border border-border/70 bg-muted/20 px-3 py-2.5">
              <div class="min-w-0">
                <p class="text-xs font-black">{{ t('mp4ToWebm.keepAudio') }}</p>
                <p class="mt-1 text-[10px] font-bold leading-4 text-muted-foreground">{{ t('mp4ToWebm.keepAudioHelp') }}</p>
              </div>
              <Switch v-model="settings.keepAudio" size="sm" :disabled="!sourceFile || converting" :aria-label="t('mp4ToWebm.keepAudio')" />
            </div>

            <div class="flex items-start gap-2 border border-emerald-500/20 bg-emerald-500/5 px-3 py-2.5 text-[10px] font-bold leading-4 text-emerald-800">
              <ShieldCheck class="mt-0.5 size-4 shrink-0" />
              <span>{{ t('mp4ToWebm.localOnly') }}</span>
            </div>

            <p class="text-[10px] font-bold leading-4 text-muted-foreground">{{ t('mp4ToWebm.engineNote') }}</p>

            <Button class="w-full" size="lg" type="submit" :disabled="!sourceFile || converting">
              <LoaderCircle v-if="converting" class="size-4 animate-spin" />
              <FileVideo v-else class="size-4" />
              {{ converting ? processingLabel : t('mp4ToWebm.convert') }}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, reactive, ref } from 'vue'
import { FFmpeg } from '@ffmpeg/ffmpeg'
import { fetchFile, toBlobURL } from '@ffmpeg/util'
import coreURL from '@ffmpeg/core?url'
import wasmURL from '@ffmpeg/core/wasm?url'
import { Download, FileVideo, LoaderCircle, RotateCcw, ShieldCheck, TriangleAlert, X } from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { useAdminI18n } from '@/i18n'

type ConversionStage = 'idle' | 'loading' | 'preparing' | 'encoding'

interface VideoSettings {
  targetWidth: number | null
  targetHeight: number | null
  quality: number
  keepAudio: boolean
}

interface NormalizedVideoSettings {
  targetWidth: number
  targetHeight: number
  quality: number
  keepAudio: boolean
}

const MAX_OUTPUT_DIMENSION = 8192
const SUPPORTED_EXTENSIONS = ['mp4', 'mov', 'm4v']
const { t } = useAdminI18n()
const fileInput = ref<HTMLInputElement | null>(null)
const sourceVideo = ref<HTMLVideoElement | null>(null)
const outputVideo = ref<HTMLVideoElement | null>(null)
const sourceFile = ref<File | null>(null)
const sourcePreviewUrl = ref('')
const outputPreviewUrl = ref('')
const outputBlob = ref<Blob | null>(null)
const sourcePreviewError = ref(false)
const outputPreviewError = ref(false)
const sourceWidth = ref(0)
const sourceHeight = ref(0)
const sourceDuration = ref(0)
const outputWidth = ref(0)
const outputHeight = ref(0)
const outputDuration = ref(0)
const isDragging = ref(false)
const converting = ref(false)
const progressPercent = ref(0)
const conversionStage = ref<ConversionStage>('idle')
const errorMessage = ref('')
const conversionRunID = ref(0)
const settings = reactive<VideoSettings>({
  targetWidth: null,
  targetHeight: null,
  quality: 32,
  keepAudio: true,
})

let ffmpeg: FFmpeg | null = null
let ffmpegLoadPromise: Promise<FFmpeg> | null = null
let ffmpegLogTail: string[] = []

const outputBytes = computed(() => outputBlob.value?.size || 0)
const displayOutputDuration = computed(() => outputDuration.value || sourceDuration.value)
const processingLabel = computed(() => {
  if (conversionStage.value === 'loading') return t('mp4ToWebm.loadingEngine')
  if (conversionStage.value === 'preparing') return t('mp4ToWebm.preparing')
  return t('mp4ToWebm.processing')
})
const sizeChangeLabel = computed(() => {
  if (!sourceFile.value || !outputBytes.value || !sourceFile.value.size) return t('mp4ToWebm.notAvailable')
  const difference = ((sourceFile.value.size - outputBytes.value) / sourceFile.value.size) * 100
  return difference >= 0
    ? t('mp4ToWebm.sizeReduced', { percent: difference.toFixed(1) })
    : t('mp4ToWebm.sizeIncreased', { percent: Math.abs(difference).toFixed(1) })
})
const sizeChangeClass = computed(() => {
  if (!sourceFile.value || !outputBytes.value || !sourceFile.value.size) return 'text-muted-foreground'
  return outputBytes.value <= sourceFile.value.size ? 'text-emerald-700' : 'text-amber-700'
})

const revokeObjectURL = (url: string): void => {
  if (url) URL.revokeObjectURL(url)
}

const releaseOutput = (): void => {
  revokeObjectURL(outputPreviewUrl.value)
  outputPreviewUrl.value = ''
  outputBlob.value = null
  outputPreviewError.value = false
  outputWidth.value = 0
  outputHeight.value = 0
  outputDuration.value = 0
}

const releaseSourcePreview = (): void => {
  revokeObjectURL(sourcePreviewUrl.value)
  sourcePreviewUrl.value = ''
  sourcePreviewError.value = false
}

const disposeFfmpeg = (): void => {
  ffmpegLoadPromise = null
  if (!ffmpeg) return
  ffmpeg.terminate()
  ffmpeg = null
}

const formatBytes = (bytes: number): string => {
  if (bytes < 1024) return `${bytes} B`
  const kilobytes = bytes / 1024
  if (kilobytes < 1024) return `${kilobytes.toFixed(kilobytes >= 100 ? 0 : 1)} KB`
  const megabytes = kilobytes / 1024
  return `${megabytes.toFixed(megabytes >= 100 ? 0 : 1)} MB`
}

const formatDuration = (seconds: number): string => {
  if (!Number.isFinite(seconds) || seconds <= 0) return t('mp4ToWebm.notAvailable')
  const totalSeconds = Math.round(seconds)
  const hours = Math.floor(totalSeconds / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const remainingSeconds = totalSeconds % 60
  if (hours > 0) return `${hours}:${String(minutes).padStart(2, '0')}:${String(remainingSeconds).padStart(2, '0')}`
  return `${minutes}:${String(remainingSeconds).padStart(2, '0')}`
}

const formatSourceType = (file: File | null): string => {
  if (!file) return 'VIDEO'
  const extension = file.name.split('.').pop()?.toUpperCase() || ''
  return extension || file.type.replace(/^video\//i, '').toUpperCase() || 'VIDEO'
}

const isSupportedFile = (file: File): boolean => {
  const extension = file.name.split('.').pop()?.toLowerCase() || ''
  const type = file.type.toLowerCase()
  const supportedType = !type
    || type.startsWith('video/')
    || type === 'application/octet-stream'
    || type === 'application/mp4'
  return SUPPORTED_EXTENSIONS.includes(extension) && supportedType
}

const clampInteger = (value: number, min: number, max: number, fallback: number): number => {
  const parsed = Number(value)
  if (!Number.isFinite(parsed)) return fallback
  return Math.round(Math.min(max, Math.max(min, parsed)))
}

const normalizeDimension = (value: number | null, fallback: number): number | null => {
  const parsed = Number(value)
  const candidate = Number.isFinite(parsed) && parsed > 0 ? parsed : fallback
  if (!Number.isFinite(candidate) || candidate <= 0) return null
  const clamped = Math.min(MAX_OUTPUT_DIMENSION, Math.max(2, Math.round(candidate)))
  return clamped % 2 === 0 ? clamped : Math.max(2, clamped - 1)
}

const normalizedSettings = (): NormalizedVideoSettings | null => {
  const targetWidth = normalizeDimension(settings.targetWidth, sourceWidth.value)
  const targetHeight = normalizeDimension(settings.targetHeight, sourceHeight.value)
  if (!targetWidth || !targetHeight) return null

  const normalized = {
    targetWidth,
    targetHeight,
    quality: clampInteger(settings.quality, 18, 45, 32),
    keepAudio: Boolean(settings.keepAudio),
  }
  Object.assign(settings, normalized)
  return normalized
}

const handleFfmpegProgress = ({ progress }: { progress: number }): void => {
  if (!converting.value || !Number.isFinite(progress)) return
  progressPercent.value = Math.min(99, Math.max(0, Math.round(progress * 100)))
}

const handleFfmpegLog = ({ message }: { message: string }): void => {
  if (!message.trim()) return
  ffmpegLogTail = [...ffmpegLogTail.slice(-19), message]
}

const ensureFfmpeg = async (): Promise<FFmpeg> => {
  if (ffmpeg?.loaded) return ffmpeg
  if (ffmpegLoadPromise) return ffmpegLoadPromise

  const instance = ffmpeg || new FFmpeg()
  ffmpeg = instance
  instance.on('progress', handleFfmpegProgress)
  instance.on('log', handleFfmpegLog)

  ffmpegLoadPromise = (async () => {
    let coreBlobURL = ''
    let wasmBlobURL = ''
    try {
      conversionStage.value = 'loading'
      coreBlobURL = await toBlobURL(coreURL, 'text/javascript')
      wasmBlobURL = await toBlobURL(wasmURL, 'application/wasm')
      await instance.load({
        coreURL: coreBlobURL,
        wasmURL: wasmBlobURL,
      })
      return instance
    } catch (error) {
      instance.terminate()
      if (ffmpeg === instance) ffmpeg = null
      throw error
    } finally {
      revokeObjectURL(coreBlobURL)
      revokeObjectURL(wasmBlobURL)
    }
  })()

  try {
    return await ffmpegLoadPromise
  } finally {
    ffmpegLoadPromise = null
  }
}

const cleanupFfmpegFiles = async (instance: FFmpeg, inputName: string, outputName: string): Promise<void> => {
  if (!instance.loaded) return
  await Promise.all([
    instance.deleteFile(inputName).catch(() => undefined),
    instance.deleteFile(outputName).catch(() => undefined),
  ])
}

const chooseFile = (file: File | null): void => {
  if (!file) return
  if (!isSupportedFile(file)) {
    errorMessage.value = t('mp4ToWebm.unsupported')
    return
  }

  ++conversionRunID.value
  errorMessage.value = ''
  releaseSourcePreview()
  releaseOutput()
  sourceFile.value = file
  sourcePreviewUrl.value = URL.createObjectURL(file)
  sourceWidth.value = 0
  sourceHeight.value = 0
  sourceDuration.value = 0
  settings.targetWidth = null
  settings.targetHeight = null
  isDragging.value = false
}

const handleSourceMetadata = (event: Event): void => {
  const video = event.currentTarget instanceof HTMLVideoElement ? event.currentTarget : sourceVideo.value
  if (!video || !video.videoWidth || !video.videoHeight) return

  sourceWidth.value = video.videoWidth
  sourceHeight.value = video.videoHeight
  sourceDuration.value = Number.isFinite(video.duration) ? video.duration : 0
  if (!settings.targetWidth) settings.targetWidth = normalizeDimension(sourceWidth.value, sourceWidth.value)
  if (!settings.targetHeight) settings.targetHeight = normalizeDimension(sourceHeight.value, sourceHeight.value)
}

const handleSourcePreviewError = (): void => {
  sourcePreviewError.value = true
}

const handleOutputMetadata = (event: Event): void => {
  const video = event.currentTarget instanceof HTMLVideoElement ? event.currentTarget : outputVideo.value
  if (!video) return
  if (video.videoWidth && video.videoHeight) {
    outputWidth.value = video.videoWidth
    outputHeight.value = video.videoHeight
  }
  if (Number.isFinite(video.duration)) outputDuration.value = video.duration
}

const convertCurrent = async (): Promise<void> => {
  if (!sourceFile.value) {
    errorMessage.value = t('mp4ToWebm.noFile')
    return
  }
  if (converting.value) return

  const normalized = normalizedSettings()
  if (!normalized) {
    errorMessage.value = t('mp4ToWebm.dimensionsRequired')
    return
  }

  const runID = ++conversionRunID.value
  const inputName = `input.${sourceFile.value.name.split('.').pop()?.toLowerCase() || 'mp4'}`
  const outputName = 'output.webm'
  converting.value = true
  conversionStage.value = 'loading'
  progressPercent.value = 0
  errorMessage.value = ''
  releaseOutput()
  ffmpegLogTail = []

  let instance: FFmpeg | null = null
  try {
    instance = await ensureFfmpeg()
    if (runID !== conversionRunID.value) return

    conversionStage.value = 'preparing'
    await instance.writeFile(inputName, await fetchFile(sourceFile.value))
    if (runID !== conversionRunID.value) return

    conversionStage.value = 'encoding'
    const args = [
      '-i', inputName,
      '-map', '0:v:0',
      ...(normalized.keepAudio ? ['-map', '0:a:0?'] : ['-an']),
      '-vf', `scale=${normalized.targetWidth}:${normalized.targetHeight}:flags=lanczos`,
      '-c:v', 'libvpx',
      '-crf', String(normalized.quality),
      '-b:v', '0',
      '-deadline', 'good',
      '-cpu-used', '4',
      '-pix_fmt', 'yuv420p',
      ...(normalized.keepAudio ? ['-c:a', 'libopus', '-b:a', '96k'] : []),
      '-f', 'webm',
      outputName,
    ]
    const exitCode = await instance.exec(args)
    if (runID !== conversionRunID.value) return
    if (exitCode !== 0) throw new Error('ffmpeg-exec-failed')

    const outputData = await instance.readFile(outputName)
    if (typeof outputData === 'string') throw new Error('ffmpeg-output-invalid')
    const outputBuffer = new ArrayBuffer(outputData.byteLength)
    new Uint8Array(outputBuffer).set(outputData)
    const blob = new Blob([outputBuffer], { type: 'video/webm' })
    if (!blob.size) throw new Error('ffmpeg-output-empty')

    releaseOutput()
    outputBlob.value = blob
    outputWidth.value = normalized.targetWidth
    outputHeight.value = normalized.targetHeight
    outputDuration.value = sourceDuration.value
    outputPreviewUrl.value = URL.createObjectURL(blob)
    progressPercent.value = 100
    await nextTick()
  } catch (error) {
    if (runID !== conversionRunID.value) return
    console.error('[mp4-to-webm]', error, ffmpegLogTail.slice(-8).join('\n'))
    errorMessage.value = error instanceof Error && error.message === 'ffmpeg-output-empty'
      ? t('mp4ToWebm.emptyOutput')
      : t('mp4ToWebm.convertFailed')
  } finally {
    if (instance) await cleanupFfmpegFiles(instance, inputName, outputName)
    if (runID === conversionRunID.value) {
      converting.value = false
      conversionStage.value = 'idle'
    }
  }
}

const cancelConversion = (): void => {
  if (!converting.value) return
  ++conversionRunID.value
  converting.value = false
  conversionStage.value = 'idle'
  progressPercent.value = 0
  errorMessage.value = ''
  disposeFfmpeg()
}

const triggerFilePicker = (): void => {
  if (!converting.value) fileInput.value?.click()
}

const handleFileInput = (event: Event): void => {
  const input = event.target instanceof HTMLInputElement ? event.target : null
  const file = input?.files?.[0] || null
  if (input) input.value = ''
  chooseFile(file)
}

const handleDrop = (event: DragEvent): void => {
  isDragging.value = false
  chooseFile(event.dataTransfer?.files?.[0] || null)
}

const handleDragLeave = (event: DragEvent): void => {
  const currentTarget = event.currentTarget
  const relatedTarget = event.relatedTarget
  if (currentTarget instanceof HTMLElement && relatedTarget instanceof Node && currentTarget.contains(relatedTarget)) return
  isDragging.value = false
}

const downloadWebm = (): void => {
  if (!outputPreviewUrl.value || !outputBlob.value) return
  const baseName = (sourceFile.value?.name || 'video').replace(/\.[^/.]+$/, '') || 'video'
  const anchor = document.createElement('a')
  anchor.href = outputPreviewUrl.value
  anchor.download = `${baseName}.webm`
  document.body.appendChild(anchor)
  anchor.click()
  anchor.remove()
}

const resetTool = (): void => {
  ++conversionRunID.value
  converting.value = false
  conversionStage.value = 'idle'
  progressPercent.value = 0
  sourceFile.value = null
  releaseSourcePreview()
  releaseOutput()
  sourceWidth.value = 0
  sourceHeight.value = 0
  sourceDuration.value = 0
  settings.targetWidth = null
  settings.targetHeight = null
  errorMessage.value = ''
  isDragging.value = false
  disposeFfmpeg()
  if (fileInput.value) fileInput.value.value = ''
}

onBeforeUnmount(() => {
  ++conversionRunID.value
  releaseSourcePreview()
  releaseOutput()
  disposeFfmpeg()
})
</script>

<style scoped>
.video-converter-dropzone {
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

.video-converter-dropzone:hover,
.video-converter-dropzone:focus-visible,
.video-converter-dropzone--active {
  border-color: rgb(5 150 105 / 0.65);
  background: rgb(236 253 245 / 0.7);
  box-shadow: 0 0 0 3px rgb(5 150 105 / 0.1);
  outline: none;
}

.video-converter-dropzone--filled {
  background: rgb(248 250 252 / 0.9);
}

.video-converter-preview-surface {
  min-height: 18rem;
  border-radius: 1rem;
  background: #0f172a;
}

.video-converter-video {
  display: block;
  max-height: 17rem;
  max-width: 100%;
  width: auto;
  object-fit: contain;
}

.video-converter-progress-track {
  height: 0.5rem;
  overflow: hidden;
  border-radius: 999px;
  background: rgb(148 163 184 / 0.28);
}

.video-converter-progress-bar {
  display: block;
  height: 100%;
  min-width: 0.25rem;
  border-radius: inherit;
  background: #10b981;
  transition: width 160ms ease;
}

@media (max-width: 639px) {
  .video-converter-preview-surface,
  .video-converter-dropzone {
    min-height: 15rem;
  }

  .video-converter-video {
    max-height: 14rem;
  }
}
</style>
