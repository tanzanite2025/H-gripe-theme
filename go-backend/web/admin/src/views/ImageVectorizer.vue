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
                  class="h-auto object-contain"
                  :style="svgPreviewStyle"
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
            <div class="grid gap-1.5">
              <span class="text-[10px] font-black uppercase tracking-wider">{{ t('imageVectorizer.traceMode') }}</span>
              <RadioGroup v-model="traceMode" class="grid grid-cols-3 gap-2">
                <label class="flex h-9 cursor-pointer items-center justify-center rounded-md border border-border/80 px-3 text-[10px] font-black uppercase tracking-wider text-muted-foreground transition-colors has-data-[state=checked]:border-primary has-data-[state=checked]:bg-accent has-data-[state=checked]:text-foreground">
                  <RadioGroupItem class="sr-only" value="professional" />
                  {{ t('imageVectorizer.traceModeProfessional') }}
                </label>
                <label class="flex h-9 cursor-pointer items-center justify-center rounded-md border border-border/80 px-3 text-[10px] font-black uppercase tracking-wider text-muted-foreground transition-colors has-data-[state=checked]:border-primary has-data-[state=checked]:bg-accent has-data-[state=checked]:text-foreground">
                  <RadioGroupItem class="sr-only" value="balanced" />
                  {{ t('imageVectorizer.traceModeBalanced') }}
                </label>
                <label class="flex h-9 cursor-pointer items-center justify-center rounded-md border border-border/80 px-3 text-[10px] font-black uppercase tracking-wider text-muted-foreground transition-colors has-data-[state=checked]:border-primary has-data-[state=checked]:bg-accent has-data-[state=checked]:text-foreground">
                  <RadioGroupItem class="sr-only" value="photo" />
                  {{ t('imageVectorizer.traceModePhoto') }}
                </label>
              </RadioGroup>
            </div>

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
                  step="0.05"
                  class="h-2 w-full cursor-pointer accent-primary"
                />
                <Input v-model.number="settings.ltres" type="number" min="0.1" max="10" step="0.05" class="text-right" />
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
                  step="0.05"
                  class="h-2 w-full cursor-pointer accent-primary"
                />
                <Input v-model.number="settings.qtres" type="number" min="0.1" max="10" step="0.05" class="text-right" />
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
import { computed, nextTick, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { Download, Image as ImageIcon, ImagePlus, LoaderCircle, RotateCcw, ShieldCheck, TriangleAlert, WandSparkles } from '@lucide/vue'
import ImageTracer, { type ImageTracerOptions } from 'imagetracerjs'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
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

type TraceMode = 'professional' | 'balanced' | 'photo'

interface PaletteColor {
  r: number
  g: number
  b: number
  a: number
}

interface ColorBucket {
  r: number
  g: number
  b: number
  count: number
}

interface PaletteCandidate extends PaletteColor {
  count: number
  brightness: number
  saturation: number
  familyKey: string
}

interface Tracedata {
  layers: Array<unknown>
  palette: PaletteColor[]
  width: number
  height: number
}

interface TraceProfile {
  numberofcolors: number
  ltres: number
  qtres: number
  pathomit: number
  blurradius: number
  colorsampling: NonNullable<ImageTracerOptions['colorsampling']>
  colorquantcycles: number
  linefilter: boolean
  rightangleenhance: boolean
  strokewidth: number
  roundcoords: number
  imageSmoothing: boolean
  paletteBucketSize: number
  solidOpaqueColors: boolean
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
const traceMode = ref<TraceMode>('professional')

const traceProfiles: Record<TraceMode, TraceProfile> = {
  professional: {
    numberofcolors: 4,
    ltres: 0.2,
    qtres: 0.2,
    pathomit: 0,
    blurradius: 0,
    colorsampling: 0,
    colorquantcycles: 1,
    linefilter: true,
    rightangleenhance: true,
    strokewidth: 0,
    roundcoords: 1,
    imageSmoothing: false,
    paletteBucketSize: 32,
    solidOpaqueColors: true,
  },
  balanced: {
    numberofcolors: 8,
    ltres: 0.6,
    qtres: 0.6,
    pathomit: 4,
    blurradius: 0,
    colorsampling: 0,
    colorquantcycles: 1,
    linefilter: true,
    rightangleenhance: true,
    strokewidth: 0,
    roundcoords: 1,
    imageSmoothing: true,
    paletteBucketSize: 24,
    solidOpaqueColors: true,
  },
  photo: {
    numberofcolors: 24,
    ltres: 1.2,
    qtres: 1.2,
    pathomit: 8,
    blurradius: 1,
    colorsampling: 0,
    colorquantcycles: 1,
    linefilter: true,
    rightangleenhance: false,
    strokewidth: 0,
    roundcoords: 1,
    imageSmoothing: true,
    paletteBucketSize: 16,
    solidOpaqueColors: true,
  },
}

const settings = reactive<VectorizerSettings>({
  numberofcolors: traceProfiles.professional.numberofcolors,
  ltres: traceProfiles.professional.ltres,
  qtres: traceProfiles.professional.qtres,
  pathomit: traceProfiles.professional.pathomit,
  blurradius: traceProfiles.professional.blurradius,
  maxDimension: 1600,
  outputWidth: 44,
  outputHeight: 44,
})

const outputBytes = computed(() => svgText.value ? new Blob([svgText.value]).size : 0)
const svgPreviewStyle = computed<Record<string, string>>(() => {
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

const syncSettings = (updates: Partial<VectorizerSettings>): void => {
  syncingSettings = true
  Object.assign(settings, updates)
  syncingSettings = false
}

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

const applyTraceProfile = (mode: TraceMode): void => {
  const profile = traceProfiles[mode]
  Object.assign(settings, {
    numberofcolors: profile.numberofcolors,
    ltres: profile.ltres,
    qtres: profile.qtres,
    pathomit: profile.pathomit,
    blurradius: profile.blurradius,
  })
}

watch(traceMode, (mode) => {
  applyTraceProfile(mode)
}, { immediate: true })

const normalizedSettings = (): VectorizerSettings => {
  const normalized = {
    numberofcolors: clampInteger(settings.numberofcolors, 2, 64, 4),
    ltres: clampNumber(settings.ltres, 0.1, 10, 0.2),
    qtres: clampNumber(settings.qtres, 0.1, 10, 0.2),
    pathomit: clampInteger(settings.pathomit, 0, 32, 0),
    blurradius: clampInteger(settings.blurradius, 0, 5, 0),
    maxDimension: clampInteger(settings.maxDimension, 256, 4096, 1600),
    outputWidth: clampInteger(settings.outputWidth, 1, 8192, 44),
    outputHeight: clampInteger(settings.outputHeight, 1, 8192, 44),
  }
  syncSettings(normalized)
  return normalized
}

const loadImage = (url: string): Promise<HTMLImageElement> => new Promise((resolve, reject) => {
  const image = new Image()
  image.onload = () => resolve(image)
  image.onerror = () => reject(new Error('image-load-failed'))
  image.src = url
})

const rasterizeImage = (image: HTMLImageElement, maxDimension: number, imageSmoothing: boolean): ImageData => {
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

  context.imageSmoothingEnabled = imageSmoothing
  if (imageSmoothing) context.imageSmoothingQuality = 'high'
  context.clearRect(0, 0, width, height)
  context.drawImage(image, 0, 0, width, height)

  return context.getImageData(0, 0, width, height)
}

const quantizeColorChannel = (value: number, bucketSize: number): number => {
  const clamped = clampInteger(value, 0, 255, 0)
  if (clamped === 0) return 0
  return Math.min(255, Math.floor(clamped / bucketSize) * bucketSize + Math.floor(bucketSize / 2))
}

const TRACE_ALPHA_THRESHOLD = 8

const buildEmptySvg = (width: number, height: number): string => (
  `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${width} ${height}" version="1.1" width="${width}" height="${height}"></svg>`
)

const normalizeTraceImageData = (imageData: ImageData, solidOpaqueColors: boolean): ImageData => {
  const normalized = new ImageData(new Uint8ClampedArray(imageData.data), imageData.width, imageData.height)
  const { data } = normalized

  for (let index = 0; index < data.length; index += 4) {
    const alpha = data[index + 3]
    if (alpha < TRACE_ALPHA_THRESHOLD) {
      data[index] = 0
      data[index + 1] = 0
      data[index + 2] = 0
      data[index + 3] = 0
      continue
    }

    if (solidOpaqueColors) {
      data[index + 3] = 255
    }
  }

  return normalized
}

const calculateColorSaturation = (r: number, g: number, b: number): number => {
  const maxChannel = Math.max(r, g, b)
  const minChannel = Math.min(r, g, b)
  if (maxChannel <= 0) return 0
  return (maxChannel - minChannel) / maxChannel
}

const calculateColorHue = (r: number, g: number, b: number): number => {
  const maxChannel = Math.max(r, g, b)
  const minChannel = Math.min(r, g, b)
  const delta = maxChannel - minChannel
  if (delta === 0) return 0

  let hue = 0
  if (maxChannel === r) {
    hue = ((g - b) / delta) % 6
  } else if (maxChannel === g) {
    hue = (b - r) / delta + 2
  } else {
    hue = (r - g) / delta + 4
  }

  return ((hue * 60) + 360) % 360
}

const calculatePaletteFamilyKey = (r: number, g: number, b: number, saturation: number): string => {
  const brightnessBucket = Math.floor(Math.max(r, g, b) / 32)
  if (saturation < 0.16) return `gray:${brightnessBucket}`
  return `hue:${Math.floor(calculateColorHue(r, g, b) / 30)}`
}

const calculateColorDistance = (left: PaletteColor, right: PaletteColor): number => (
  Math.abs(left.r - right.r) + Math.abs(left.g - right.g) + Math.abs(left.b - right.b)
)

const selectPaletteCandidates = (candidates: PaletteCandidate[], visibleColorLimit: number): PaletteCandidate[] => {
  if (candidates.length <= visibleColorLimit) {
    return candidates.slice().sort((left, right) => right.count - left.count)
  }

  const remaining = candidates.slice().sort((left, right) => right.count - left.count)
  const selected: PaletteCandidate[] = []
  const familyCounts = new Map<string, number>()

  const firstCandidate = remaining.shift()
  if (firstCandidate) {
    selected.push(firstCandidate)
    familyCounts.set(firstCandidate.familyKey, 1)
  }

  while (selected.length < visibleColorLimit && remaining.length > 0) {
    let bestIndex = 0
    let bestScore = -Infinity

    for (let index = 0; index < remaining.length; index += 1) {
      const candidate = remaining[index]
      const nearestDistance = selected.reduce(
        (nearest, selectedCandidate) => Math.min(nearest, calculateColorDistance(candidate, selectedCandidate)),
        Number.POSITIVE_INFINITY,
      )
      const familyCount = familyCounts.get(candidate.familyKey) || 0
      const countScore = Math.log1p(candidate.count)
      const saturationBoost = 1 + candidate.saturation * 3.25
      const diversityBoost = 0.35 + Math.min(1, nearestDistance / 96) * 1.65
      const familyMultiplier = Math.pow(0.45, familyCount)
      const score = countScore * saturationBoost * diversityBoost * familyMultiplier

      if (score > bestScore) {
        bestScore = score
        bestIndex = index
      }
    }

    const [nextCandidate] = remaining.splice(bestIndex, 1)
    if (nextCandidate) {
      selected.push(nextCandidate)
      familyCounts.set(nextCandidate.familyKey, (familyCounts.get(nextCandidate.familyKey) || 0) + 1)
    }
  }

  return selected.sort((left, right) => right.count - left.count)
}

const buildHistogramPalette = (
  imageData: ImageData,
  visibleColorLimit: number,
  bucketSize: number,
  solidOpaqueColors: boolean,
): PaletteColor[] => {
  const buckets = new Map<string, ColorBucket & { a: number }>()
  const { data, width, height } = imageData
  const sampleStep = Math.max(1, Math.floor(Math.max(width, height) / 768))
  const weight = sampleStep * sampleStep
  let hasTransparentPixels = false

  for (let y = 0; y < height; y += sampleStep) {
    for (let x = 0; x < width; x += sampleStep) {
      const index = (y * width + x) * 4
      const alpha = data[index + 3]
      if (alpha < TRACE_ALPHA_THRESHOLD) {
        hasTransparentPixels = true
        continue
      }

      const r = data[index]
      const g = data[index + 1]
      const b = data[index + 2]
      const key = [
        quantizeColorChannel(r, bucketSize),
        quantizeColorChannel(g, bucketSize),
        quantizeColorChannel(b, bucketSize),
        solidOpaqueColors ? 255 : quantizeColorChannel(alpha, 32),
      ].join(':')
      const bucket = buckets.get(key) || { r: 0, g: 0, b: 0, a: 0, count: 0 }
      bucket.r += r * weight
      bucket.g += g * weight
      bucket.b += b * weight
      bucket.a += alpha * weight
      bucket.count += weight
      buckets.set(key, bucket)
    }
  }

  const visibleCandidates: PaletteCandidate[] = [...buckets.values()]
    .filter((bucket) => bucket.count > 0)
    .map((bucket) => {
      const r = Math.round(bucket.r / bucket.count)
      const g = Math.round(bucket.g / bucket.count)
      const b = Math.round(bucket.b / bucket.count)
      const saturation = calculateColorSaturation(r, g, b)
      return {
        r,
        g,
        b,
        a: solidOpaqueColors ? 255 : Math.round(bucket.a / bucket.count),
        count: bucket.count,
        brightness: Math.max(r, g, b) / 255,
        saturation,
        familyKey: calculatePaletteFamilyKey(r, g, b, saturation),
      }
    })

  const palette: PaletteColor[] = []
  if (hasTransparentPixels) {
    palette.push({ r: 0, g: 0, b: 0, a: 0 })
  }

  for (const candidate of selectPaletteCandidates(visibleCandidates, visibleColorLimit)) {
    palette.push({
      r: candidate.r,
      g: candidate.g,
      b: candidate.b,
      a: candidate.a,
    })
  }

  return palette.length > 0 ? palette : [{ r: 0, g: 0, b: 0, a: 0 }]
}

const pruneTransparentLayers = (tracedata: Tracedata): Tracedata => {
  const palette: PaletteColor[] = []
  const layers: Array<unknown> = []

  tracedata.palette.forEach((color, index) => {
    if (color.a <= 0) return
    palette.push(color)
    layers.push(tracedata.layers[index])
  })

  return {
    ...tracedata,
    palette,
    layers,
  }
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
  const profile = traceProfiles[traceMode.value]
  const imageData = normalizeTraceImageData(rasterizeImage(image, normalized.maxDimension, profile.imageSmoothing), profile.solidOpaqueColors)
  const palette = buildHistogramPalette(imageData, normalized.numberofcolors, profile.paletteBucketSize, profile.solidOpaqueColors)
  const imageTracer = ImageTracer as typeof ImageTracer & {
    imagedataToTracedata: (data: ImageData, options?: ImageTracerOptions | string) => Tracedata
    getsvgstring: (tracedata: Tracedata, options?: ImageTracerOptions | string) => string
  }
  const options: ImageTracerOptions = {
    blurradius: normalized.blurradius,
    colorsampling: profile.colorsampling,
    colorquantcycles: profile.colorquantcycles,
    desc: false,
    layering: 0,
    linefilter: profile.linefilter,
    ltres: normalized.ltres,
    numberofcolors: normalized.numberofcolors,
    pathomit: normalized.pathomit,
    rightangleenhance: profile.rightangleenhance,
    roundcoords: profile.roundcoords,
    scale: 1,
    strokewidth: profile.strokewidth,
    viewbox: true,
    qtres: normalized.qtres,
  }
  options.pal = palette

  const tracedata = pruneTransparentLayers(imageTracer.imagedataToTracedata(imageData, options))
  const svg = setSvgDimensions(
    tracedata.layers.length > 0
      ? imageTracer.getsvgstring(tracedata, options)
      : buildEmptySvg(normalized.outputWidth, normalized.outputHeight),
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
  }, 320)
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
  clearScheduledConversion()
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

watch(
  () => [
    traceMode.value,
    settings.numberofcolors,
    settings.ltres,
    settings.qtres,
    settings.pathomit,
    settings.blurradius,
    settings.maxDimension,
    settings.outputWidth,
    settings.outputHeight,
  ],
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
