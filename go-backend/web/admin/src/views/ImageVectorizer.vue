<template>
  <div class="space-y-4">
    <AdminPageHeader
      :title="t('imageVectorizer.title')"
      :description="t('imageVectorizer.description')"
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
          {{ t('imageVectorizer.chooseAnother') }}
        </Button>
        <Button
          v-if="svgText"
          size="sm"
          :disabled="converting"
          @click="downloadSvg"
        >
          <Download class="size-3.5" />
          {{ t('imageVectorizer.download') }}
        </Button>
      </template>
    </AdminPageHeader>

    <div class="grid min-h-0 gap-4 xl:grid-cols-[minmax(0,1fr)_20rem]">
      <Card class="min-h-0">
        <CardHeader class="border-b">
          <div class="flex min-w-0 items-start justify-between gap-3">
            <div class="min-w-0">
              <CardTitle class="flex items-center gap-2">
                <ImageIcon class="size-4 text-primary" />
                {{ t('imageVectorizer.workspaceTitle') }}
              </CardTitle>
              <CardDescription>{{ t('imageVectorizer.workspaceDescription') }}</CardDescription>
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
                  <p class="text-xs font-black uppercase tracking-wider">{{ t('imageVectorizer.sourceTitle') }}</p>
                  <p class="mt-1 text-[10px] font-bold text-muted-foreground">{{ t('imageVectorizer.sourceDescription') }}</p>
                </div>
                <span v-if="sourceFile" class="shrink-0 rounded-full bg-muted px-2 py-1 font-mono text-[10px] font-black text-muted-foreground">
                  {{ formatBytes(sourceFile.size) }}
                </span>
              </div>

              <div
                class="image-vectorizer-dropzone"
                :class="{
                  'image-vectorizer-dropzone--active': isDragging,
                  'image-vectorizer-dropzone--filled': sourceFile,
                }"
                role="button"
                tabindex="0"
                :aria-label="t('imageVectorizer.selectFile')"
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
                  <div class="image-vectorizer-preview-surface flex min-h-48 w-full items-center justify-center overflow-hidden border border-border/70 p-3">
                    <img
                      :src="sourcePreviewUrl"
                      :alt="sourceFile?.name || t('imageVectorizer.sourceTitle')"
                      class="max-h-64 max-w-full object-contain"
                    />
                  </div>
                  <div class="min-w-0 max-w-full text-center">
                    <p class="truncate text-xs font-black">{{ sourceFile?.name }}</p>
                    <p v-if="sourceWidth && sourceHeight" class="mt-1 font-mono text-[10px] font-bold text-muted-foreground">
                      {{ t('imageVectorizer.dimensions', { width: sourceWidth, height: sourceHeight }) }}
                    </p>
                  </div>
                </div>

                <div v-else class="flex min-h-72 flex-col items-center justify-center gap-3 px-5 text-center">
                  <span class="flex size-12 items-center justify-center rounded-2xl bg-primary/10 text-primary">
                    <ImagePlus class="size-6" />
                  </span>
                  <p class="text-sm font-black">{{ t('imageVectorizer.selectFile') }}</p>
                  <p class="text-[10px] font-bold text-muted-foreground">{{ t('imageVectorizer.dropHint') }}</p>
                </div>
              </div>
            </section>

            <section class="min-w-0">
              <div class="mb-2 flex items-center justify-between gap-3">
                <div>
                  <p class="text-xs font-black uppercase tracking-wider">{{ t('imageVectorizer.outputTitle') }}</p>
                  <p class="mt-1 text-[10px] font-bold text-muted-foreground">{{ t('imageVectorizer.outputDescription') }}</p>
                </div>
                <span
                  v-if="svgText"
                  class="shrink-0 rounded-full bg-emerald-500/10 px-2 py-1 font-mono text-[10px] font-black text-emerald-700"
                >
                  {{ t('imageVectorizer.ready') }}
                </span>
              </div>

              <div class="image-vectorizer-preview-surface flex min-h-72 items-center justify-center overflow-hidden border border-border/70 p-4">
                <img
                  v-if="svgPreviewUrl"
                  :src="svgPreviewUrl"
                  :alt="t('imageVectorizer.outputTitle')"
                  class="h-auto max-h-64 w-full max-w-full object-contain"
                />
                <div v-else class="flex flex-col items-center justify-center gap-3 px-5 text-center text-muted-foreground">
                  <LoaderCircle v-if="converting" class="size-7 animate-spin text-primary" />
                  <WandSparkles v-else class="size-7 opacity-50" />
                  <p class="text-xs font-black">
                    {{ converting ? t('imageVectorizer.processing') : t('imageVectorizer.waiting') }}
                  </p>
                </div>
              </div>

              <div v-if="svgText" class="mt-3 flex flex-wrap items-center gap-2 text-[10px] font-mono font-bold text-muted-foreground">
                <span>{{ t('imageVectorizer.dimensions', { width: outputWidth, height: outputHeight }) }}</span>
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
          <CardTitle>{{ t('imageVectorizer.settingsTitle') }}</CardTitle>
          <CardDescription>{{ t('imageVectorizer.settingsDescription') }}</CardDescription>
        </CardHeader>

        <CardContent class="p-4">
          <form class="space-y-5" @submit.prevent="convertCurrent">
            <label class="grid gap-1.5">
              <span class="flex items-center justify-between gap-3 text-[10px] font-black uppercase tracking-wider">
                <span>{{ t('imageVectorizer.colors') }}</span>
                <output class="font-mono text-primary">{{ settings.numberofcolors }}</output>
              </span>
              <div class="grid grid-cols-[minmax(0,1fr)_4.5rem] items-center gap-2">
                <input
                  v-model.number="settings.numberofcolors"
                  type="range"
                  min="2"
                  max="64"
                  step="1"
                  class="h-2 w-full cursor-pointer accent-primary"
                />
                <Input v-model.number="settings.numberofcolors" type="number" min="2" max="64" step="1" class="text-right" />
              </div>
              <span class="text-[10px] font-bold leading-4 text-muted-foreground">{{ t('imageVectorizer.colorsHelp') }}</span>
            </label>

            <label class="grid gap-1.5">
              <span class="flex items-center justify-between gap-3 text-[10px] font-black uppercase tracking-wider">
                <span>{{ t('imageVectorizer.linePrecision') }}</span>
                <output class="font-mono text-primary">{{ settings.ltres }}</output>
              </span>
              <div class="grid grid-cols-[minmax(0,1fr)_4.5rem] items-center gap-2">
                <input
                  v-model.number="settings.ltres"
                  type="range"
                  min="0.1"
                  max="10"
                  step="0.1"
                  class="h-2 w-full cursor-pointer accent-primary"
                />
                <Input v-model.number="settings.ltres" type="number" min="0.1" max="10" step="0.1" class="text-right" />
              </div>
              <span class="text-[10px] font-bold leading-4 text-muted-foreground">{{ t('imageVectorizer.linePrecisionHelp') }}</span>
            </label>

            <label class="grid gap-1.5">
              <span class="flex items-center justify-between gap-3 text-[10px] font-black uppercase tracking-wider">
                <span>{{ t('imageVectorizer.curvePrecision') }}</span>
                <output class="font-mono text-primary">{{ settings.qtres }}</output>
              </span>
              <div class="grid grid-cols-[minmax(0,1fr)_4.5rem] items-center gap-2">
                <input
                  v-model.number="settings.qtres"
                  type="range"
                  min="0.1"
                  max="10"
                  step="0.1"
                  class="h-2 w-full cursor-pointer accent-primary"
                />
                <Input v-model.number="settings.qtres" type="number" min="0.1" max="10" step="0.1" class="text-right" />
              </div>
              <span class="text-[10px] font-bold leading-4 text-muted-foreground">{{ t('imageVectorizer.curvePrecisionHelp') }}</span>
            </label>

            <label class="grid gap-1.5">
              <span class="flex items-center justify-between gap-3 text-[10px] font-black uppercase tracking-wider">
                <span>{{ t('imageVectorizer.noiseFilter') }}</span>
                <output class="font-mono text-primary">{{ settings.pathomit }}</output>
              </span>
              <div class="grid grid-cols-[minmax(0,1fr)_4.5rem] items-center gap-2">
                <input
                  v-model.number="settings.pathomit"
                  type="range"
                  min="0"
                  max="32"
                  step="1"
                  class="h-2 w-full cursor-pointer accent-primary"
                />
                <Input v-model.number="settings.pathomit" type="number" min="0" max="32" step="1" class="text-right" />
              </div>
              <span class="text-[10px] font-bold leading-4 text-muted-foreground">{{ t('imageVectorizer.noiseFilterHelp') }}</span>
            </label>

            <label class="grid gap-1.5">
              <span class="flex items-center justify-between gap-3 text-[10px] font-black uppercase tracking-wider">
                <span>{{ t('imageVectorizer.blurRadius') }}</span>
                <output class="font-mono text-primary">{{ settings.blurradius }}</output>
              </span>
              <div class="grid grid-cols-[minmax(0,1fr)_4.5rem] items-center gap-2">
                <input
                  v-model.number="settings.blurradius"
                  type="range"
                  min="0"
                  max="5"
                  step="1"
                  class="h-2 w-full cursor-pointer accent-primary"
                />
                <Input v-model.number="settings.blurradius" type="number" min="0" max="5" step="1" class="text-right" />
              </div>
              <span class="text-[10px] font-bold leading-4 text-muted-foreground">{{ t('imageVectorizer.blurRadiusHelp') }}</span>
            </label>

            <label class="grid gap-1.5">
              <span class="flex items-center justify-between gap-3 text-[10px] font-black uppercase tracking-wider">
                <span>{{ t('imageVectorizer.maxDimension') }}</span>
                <output class="font-mono text-primary">{{ settings.maxDimension }} px</output>
              </span>
              <div class="grid grid-cols-[minmax(0,1fr)_5.5rem] items-center gap-2">
                <input
                  v-model.number="settings.maxDimension"
                  type="range"
                  min="256"
                  max="4096"
                  step="64"
                  class="h-2 w-full cursor-pointer accent-primary"
                />
                <Input v-model.number="settings.maxDimension" type="number" min="256" max="4096" step="64" class="text-right" />
              </div>
              <span class="text-[10px] font-bold leading-4 text-muted-foreground">{{ t('imageVectorizer.maxDimensionHelp') }}</span>
            </label>

            <div class="grid gap-1.5">
              <span class="flex items-center justify-between gap-3 text-[10px] font-black uppercase tracking-wider">
                <span>{{ t('imageVectorizer.outputSize') }}</span>
                <output class="font-mono text-primary">{{ settings.outputWidth }} × {{ settings.outputHeight }} px</output>
              </span>
              <div class="grid grid-cols-2 gap-2">
                <label class="grid gap-1">
                  <span class="text-[10px] font-bold text-muted-foreground">{{ t('imageVectorizer.outputWidth') }}</span>
                  <Input v-model.number="settings.outputWidth" type="number" min="1" max="8192" step="1" class="text-right" />
                </label>
                <label class="grid gap-1">
                  <span class="text-[10px] font-bold text-muted-foreground">{{ t('imageVectorizer.outputHeight') }}</span>
                  <Input v-model.number="settings.outputHeight" type="number" min="1" max="8192" step="1" class="text-right" />
                </label>
              </div>
              <span class="text-[10px] font-bold leading-4 text-muted-foreground">{{ t('imageVectorizer.outputSizeHelp') }}</span>
            </div>

            <div class="flex items-start gap-2 border border-emerald-500/20 bg-emerald-500/5 px-3 py-2.5 text-[10px] font-bold leading-4 text-emerald-800">
              <ShieldCheck class="mt-0.5 size-4 shrink-0" />
              <span>{{ t('imageVectorizer.localOnly') }}</span>
            </div>

            <Button class="w-full" size="lg" type="submit" :disabled="!sourceFile || converting">
              <LoaderCircle v-if="converting" class="size-4 animate-spin" />
              <WandSparkles v-else class="size-4" />
              {{ converting ? t('imageVectorizer.processing') : t('imageVectorizer.convert') }}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, reactive, ref } from 'vue'
import { Download, Image as ImageIcon, ImagePlus, LoaderCircle, RotateCcw, ShieldCheck, TriangleAlert, WandSparkles } from '@lucide/vue'
import ImageTracer, { type ImageTracerOptions } from 'imagetracerjs'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { useAdminI18n } from '@/i18n'

interface VectorizerSettings {
  numberofcolors: number
  ltres: number
  qtres: number
  pathomit: number
  blurradius: number
  maxDimension: number
  outputWidth: number
  outputHeight: number
}

const { t } = useAdminI18n()
const fileInput = ref<HTMLInputElement | null>(null)
const sourceFile = ref<File | null>(null)
const sourcePreviewUrl = ref('')
const svgPreviewUrl = ref('')
const svgText = ref('')
const sourceWidth = ref(0)
const sourceHeight = ref(0)
const outputWidth = ref(0)
const outputHeight = ref(0)
const isDragging = ref(false)
const converting = ref(false)
const errorMessage = ref('')
const conversionRunID = ref(0)
const settings = reactive<VectorizerSettings>({
  numberofcolors: 16,
  ltres: 0.5,
  qtres: 0.5,
  pathomit: 8,
  blurradius: 0,
  maxDimension: 1600,
  outputWidth: 44,
  outputHeight: 44,
})

const outputBytes = computed(() => svgText.value ? new Blob([svgText.value]).size : 0)

const revokeObjectURL = (url: string): void => {
  if (url) URL.revokeObjectURL(url)
}

const releaseSvgPreview = (): void => {
  revokeObjectURL(svgPreviewUrl.value)
  svgPreviewUrl.value = ''
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

const isSupportedFile = (file: File): boolean => {
  const extension = file.name.split('.').pop()?.toLowerCase() || ''
  const supportedType = ['', 'image/png', 'image/jpeg', 'image/jpg', 'image/webp'].includes(file.type.toLowerCase())
  return ['png', 'jpg', 'jpeg', 'webp'].includes(extension) && supportedType
}

const clampNumber = (value: number, min: number, max: number, fallback: number): number => {
  const parsed = Number(value)
  if (!Number.isFinite(parsed)) return fallback
  return Math.min(max, Math.max(min, parsed))
}

const clampInteger = (value: number, min: number, max: number, fallback: number): number => (
  Math.round(clampNumber(value, min, max, fallback))
)

const normalizedSettings = (): VectorizerSettings => {
  const normalized = {
    numberofcolors: clampInteger(settings.numberofcolors, 2, 64, 16),
    ltres: clampNumber(settings.ltres, 0.1, 10, 0.5),
    qtres: clampNumber(settings.qtres, 0.1, 10, 0.5),
    pathomit: clampInteger(settings.pathomit, 0, 32, 8),
    blurradius: clampInteger(settings.blurradius, 0, 5, 0),
    maxDimension: clampInteger(settings.maxDimension, 256, 4096, 1600),
    outputWidth: clampInteger(settings.outputWidth, 1, 8192, 44),
    outputHeight: clampInteger(settings.outputHeight, 1, 8192, 44),
  }
  Object.assign(settings, normalized)
  return normalized
}

const loadImage = (url: string): Promise<HTMLImageElement> => new Promise((resolve, reject) => {
  const image = new Image()
  image.onload = () => resolve(image)
  image.onerror = () => reject(new Error('image-load-failed'))
  image.src = url
})

const rasterizeImage = (image: HTMLImageElement, maxDimension: number): ImageData => {
  const naturalWidth = image.naturalWidth || image.width
  const naturalHeight = image.naturalHeight || image.height
  if (!naturalWidth || !naturalHeight) throw new Error('image-load-failed')

  const scale = Math.min(1, maxDimension / Math.max(naturalWidth, naturalHeight))
  const width = Math.max(1, Math.round(naturalWidth * scale))
  const height = Math.max(1, Math.round(naturalHeight * scale))
  const canvas = document.createElement('canvas')
  canvas.width = width
  canvas.height = height

  const context = canvas.getContext('2d', { willReadFrequently: true })
  if (!context) throw new Error('canvas-unavailable')

  context.imageSmoothingEnabled = true
  context.imageSmoothingQuality = 'high'
  context.clearRect(0, 0, width, height)
  context.drawImage(image, 0, 0, width, height)

  return context.getImageData(0, 0, width, height)
}

const setSvgDimensions = (svg: string, width: number, height: number): string => {
  const document = new DOMParser().parseFromString(svg, 'image/svg+xml')
  const root = document.documentElement
  if (!root || root.nodeName.toLowerCase() !== 'svg') throw new Error('svg-serialize-failed')

  root.setAttribute('width', String(width))
  root.setAttribute('height', String(height))
  return new XMLSerializer().serializeToString(root)
}

const traceImage = async (image: HTMLImageElement, runID: number): Promise<void> => {
  await nextTick()
  if (runID !== conversionRunID.value) return

  const normalized = normalizedSettings()
  const imageData = rasterizeImage(image, normalized.maxDimension)
  const options: ImageTracerOptions = {
    blurradius: normalized.blurradius,
    colorsampling: 2,
    colorquantcycles: 3,
    desc: false,
    layering: 0,
    linefilter: true,
    ltres: normalized.ltres,
    numberofcolors: normalized.numberofcolors,
    pathomit: normalized.pathomit,
    rightangleenhance: true,
    roundcoords: 2,
    scale: 1,
    strokewidth: 0,
    viewbox: true,
    qtres: normalized.qtres,
  }
  const svg = setSvgDimensions(
    ImageTracer.imagedataToSVG(imageData, options),
    normalized.outputWidth,
    normalized.outputHeight,
  )
  if (runID !== conversionRunID.value) return

  releaseSvgPreview()
  svgText.value = svg
  outputWidth.value = normalized.outputWidth
  outputHeight.value = normalized.outputHeight
  svgPreviewUrl.value = URL.createObjectURL(new Blob([svg], { type: 'image/svg+xml;charset=utf-8' }))
}

const chooseFile = async (file: File | null): Promise<void> => {
  if (!file) return
  if (!isSupportedFile(file)) {
    errorMessage.value = t('imageVectorizer.unsupported')
    return
  }

  const runID = ++conversionRunID.value
  converting.value = true
  errorMessage.value = ''
  releaseSourcePreview()
  releaseSvgPreview()
  sourceFile.value = file
  sourcePreviewUrl.value = URL.createObjectURL(file)
  svgText.value = ''
  sourceWidth.value = 0
  sourceHeight.value = 0
  outputWidth.value = 0
  outputHeight.value = 0

  try {
    const image = await loadImage(sourcePreviewUrl.value)
    if (runID !== conversionRunID.value) return
    sourceWidth.value = image.naturalWidth || image.width
    sourceHeight.value = image.naturalHeight || image.height
    await traceImage(image, runID)
  } catch {
    if (runID === conversionRunID.value) {
      errorMessage.value = t('imageVectorizer.convertFailed')
    }
  } finally {
    if (runID === conversionRunID.value) converting.value = false
  }
}

const convertCurrent = async (): Promise<void> => {
  if (!sourceFile.value || !sourcePreviewUrl.value) {
    errorMessage.value = t('imageVectorizer.noFile')
    return
  }
  if (converting.value) return

  const runID = ++conversionRunID.value
  converting.value = true
  errorMessage.value = ''
  releaseSvgPreview()
  svgText.value = ''
  outputWidth.value = 0
  outputHeight.value = 0

  try {
    const image = await loadImage(sourcePreviewUrl.value)
    if (runID !== conversionRunID.value) return
    sourceWidth.value = image.naturalWidth || image.width
    sourceHeight.value = image.naturalHeight || image.height
    await traceImage(image, runID)
  } catch {
    if (runID === conversionRunID.value) {
      errorMessage.value = t('imageVectorizer.convertFailed')
    }
  } finally {
    if (runID === conversionRunID.value) converting.value = false
  }
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

const downloadSvg = (): void => {
  if (!svgText.value || !svgPreviewUrl.value) return
  const baseName = (sourceFile.value?.name || 'image').replace(/\.[^/.]+$/, '') || 'image'
  const anchor = document.createElement('a')
  anchor.href = svgPreviewUrl.value
  anchor.download = `${baseName}.svg`
  document.body.appendChild(anchor)
  anchor.click()
  anchor.remove()
}

const resetTool = (): void => {
  ++conversionRunID.value
  converting.value = false
  sourceFile.value = null
  releaseSourcePreview()
  releaseSvgPreview()
  svgText.value = ''
  sourceWidth.value = 0
  sourceHeight.value = 0
  outputWidth.value = 0
  outputHeight.value = 0
  errorMessage.value = ''
  isDragging.value = false
  if (fileInput.value) fileInput.value.value = ''
}

onBeforeUnmount(() => {
  ++conversionRunID.value
  releaseSourcePreview()
  releaseSvgPreview()
})
</script>

<style scoped>
.image-vectorizer-dropzone {
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

.image-vectorizer-dropzone:hover,
.image-vectorizer-dropzone:focus-visible,
.image-vectorizer-dropzone--active {
  border-color: rgb(5 150 105 / 0.65);
  background: rgb(236 253 245 / 0.7);
  box-shadow: 0 0 0 3px rgb(5 150 105 / 0.1);
  outline: none;
}

.image-vectorizer-dropzone--filled {
  background: rgb(248 250 252 / 0.9);
}

.image-vectorizer-preview-surface {
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
  .image-vectorizer-preview-surface,
  .image-vectorizer-dropzone {
    min-height: 15rem;
  }
}
</style>
