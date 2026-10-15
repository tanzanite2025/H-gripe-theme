<template>
  <div class="manufacturing-process-video-column min-w-0" :style="{ '--manufacturing-process-video-aspect-ratio': videoAspectRatio }">
    <button
      type="button"
      class="manufacturing-process-video-preview relative group w-full overflow-hidden rounded-xl border tz-border-subtle shadow-md"
      :aria-label="openDialogLabel"
      aria-haspopup="dialog"
      :aria-controls="isManufacturingProcessVideoDialogOpen ? manufacturingProcessVideoDialogId : undefined"
      @click="openManufacturingProcessVideoDialog"
    >
      <video
        v-if="previewMode === 'muted-video'"
        ref="manufacturingProcessVideoPreviewElement"
        class="manufacturing-process-video-preview-playback"
        :src="videoSourcePath"
        :poster="videoPosterPath"
        :autoplay="isManufacturingProcessVideoPreviewMounted && preferredVideoMotion === 'no-preference'"
        muted
        loop
        playsinline
        preload="metadata"
        aria-hidden="true"
      ></video>
      <img
        v-else
        :src="videoPosterPath"
        :alt="videoPosterDescription"
        class="manufacturing-process-video-preview-playback"
        loading="lazy"
        decoding="async"
      >
      <span class="absolute inset-0 flex items-center justify-center bg-black/30 transition-colors group-hover:bg-black/10" aria-hidden="true">
        <span class="flex h-16 w-16 items-center justify-center rounded-full border tz-border-strong/30 tz-surface-subtle backdrop-blur-sm transition-transform group-hover:scale-110">
          <svg class="ml-1 h-8 w-8 tz-text-primary" fill="currentColor" viewBox="0 0 24 24"><path d="M8 5v14l11-7z" /></svg>
        </span>
      </span>
    </button>
    <p v-if="previewCaption" class="mt-3 text-sm tz-text-secondary leading-relaxed">{{ previewCaption }}</p>

    <Teleport to="body">
      <div
        v-if="isManufacturingProcessVideoDialogOpen"
        :id="manufacturingProcessVideoDialogId"
        :style="{ '--manufacturing-process-video-aspect-ratio': videoAspectRatio, '--manufacturing-process-video-dialog-max-width': videoDialogMaxWidth }"
        ref="manufacturingProcessVideoDialogElement"
        class="tz-standard-modal-mask fixed inset-0 z-[9999] flex items-center justify-center p-4 tz-mobile-safe-modal-mask"
        role="dialog"
        aria-modal="true"
        :aria-label="openDialogLabel"
        tabindex="-1"
        @click.self="closeManufacturingProcessVideoDialog"
      >
        <button
          ref="manufacturingProcessVideoDialogCloseButtonElement"
          type="button"
          class="manufacturing-process-video-dialog-close tz-global-close-btn absolute z-10"
          :aria-label="closeDialogLabel"
          @click="closeManufacturingProcessVideoDialog"
        >
          ×
        </button>
        <div class="manufacturing-process-video-content tz-standard-modal-surface w-full max-w-5xl overflow-hidden">
          <video class="manufacturing-process-video-dialog-playback" :src="videoSourcePath" :poster="videoPosterPath" controls autoplay playsinline></video>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, useId, watch } from 'vue'
import { usePreferredReducedMotion, useScrollLock } from '@vueuse/core'
import { useFocusTrap } from '@vueuse/integrations/useFocusTrap'
import { useDialogStack } from '~/composables/useDialogStack'

withDefaults(defineProps<{
  videoSourcePath: string
  videoPosterPath: string
  previewCaption?: string
  previewMode?: 'muted-video' | 'poster-image'
  videoPosterDescription?: string
  videoAspectRatio?: string
  videoDialogMaxWidth?: string
  openDialogLabel: string
  closeDialogLabel: string
}>(), {
  previewCaption: '',
  previewMode: 'muted-video',
  videoPosterDescription: '',
  videoAspectRatio: '4 / 3',
  videoDialogMaxWidth: '64rem',
})

const isManufacturingProcessVideoDialogOpen = ref(false)
const isManufacturingProcessVideoPreviewMounted = ref(false)
const manufacturingProcessVideoPreviewElement = ref<HTMLVideoElement | null>(null)
const manufacturingProcessVideoDialogElement = ref<HTMLElement | null>(null)
const manufacturingProcessVideoDialogCloseButtonElement = ref<HTMLButtonElement | null>(null)
const manufacturingProcessVideoDialogDocumentBodyElement = ref<HTMLElement | null>(null)
const manufacturingProcessVideoDialogId = `manufacturing-process-video-${useId()}`
const preferredVideoMotion = usePreferredReducedMotion()
const isManufacturingProcessVideoDialogDocumentBodyScrollLocked = useScrollLock(manufacturingProcessVideoDialogDocumentBodyElement)
const manufacturingProcessVideoDialogStack = useDialogStack()
let unregisterManufacturingProcessVideoDialogFromStack: (() => void) | null = null
let shouldResumeManufacturingProcessVideoPreviewAfterDialogCloses = false

const {
  activate: activateManufacturingProcessVideoDialogFocusTrap,
  deactivate: deactivateManufacturingProcessVideoDialogFocusTrap,
} = useFocusTrap(manufacturingProcessVideoDialogElement, {
  escapeDeactivates: false,
  allowOutsideClick: true,
  initialFocus: () => manufacturingProcessVideoDialogCloseButtonElement.value!,
  fallbackFocus: () => manufacturingProcessVideoDialogElement.value!,
})

function openManufacturingProcessVideoDialog(): void {
  const previewVideo = manufacturingProcessVideoPreviewElement.value
  shouldResumeManufacturingProcessVideoPreviewAfterDialogCloses = !!previewVideo && !previewVideo.paused
  previewVideo?.pause()
  isManufacturingProcessVideoDialogOpen.value = true
}

function closeManufacturingProcessVideoDialog(): void {
  isManufacturingProcessVideoDialogOpen.value = false
}

function releaseManufacturingProcessVideoDialogFocusScrollAndKeyboardRegistration(): void {
  deactivateManufacturingProcessVideoDialogFocusTrap()
  unregisterManufacturingProcessVideoDialogFromStack?.()
  unregisterManufacturingProcessVideoDialogFromStack = null
  isManufacturingProcessVideoDialogDocumentBodyScrollLocked.value = false
}

watch(isManufacturingProcessVideoDialogOpen, async (isOpen) => {
  if (!isOpen) {
    releaseManufacturingProcessVideoDialogFocusScrollAndKeyboardRegistration()
    if (shouldResumeManufacturingProcessVideoPreviewAfterDialogCloses && preferredVideoMotion.value === 'no-preference') {
      void manufacturingProcessVideoPreviewElement.value?.play().catch(() => {})
    }
    return
  }

  isManufacturingProcessVideoDialogDocumentBodyScrollLocked.value = true
  unregisterManufacturingProcessVideoDialogFromStack = manufacturingProcessVideoDialogStack.register(
    manufacturingProcessVideoDialogId,
    closeManufacturingProcessVideoDialog,
    { priority: 9999 },
  )
  await nextTick()
  if (isManufacturingProcessVideoDialogOpen.value) activateManufacturingProcessVideoDialogFocusTrap()
}, { flush: 'post' })

watch(preferredVideoMotion, (motionPreference) => {
  if (motionPreference === 'reduce') manufacturingProcessVideoPreviewElement.value?.pause()
})

onMounted(() => {
  manufacturingProcessVideoDialogDocumentBodyElement.value = document.body
  // The server cannot know the browser's motion preference. Start only after
  // hydration so reduced-motion visitors get the same poster markup on both sides.
  isManufacturingProcessVideoPreviewMounted.value = true
  if (preferredVideoMotion.value === 'no-preference') {
    void manufacturingProcessVideoPreviewElement.value?.play().catch(() => {})
  }
})

onBeforeUnmount(releaseManufacturingProcessVideoDialogFocusScrollAndKeyboardRegistration)
</script>

<style scoped>
.manufacturing-process-video-preview {
  display: block;
  cursor: pointer;
}

.manufacturing-process-video-preview:focus-visible {
  outline: 2px solid var(--tz-action-primary);
  outline-offset: 4px;
}

.manufacturing-process-video-preview-playback {
  display: block;
  width: 100%;
  height: auto;
  aspect-ratio: var(--manufacturing-process-video-aspect-ratio, 4 / 3);
  object-fit: cover;
}

.manufacturing-process-video-dialog-close {
  top: max(1rem, calc(1rem + var(--tz-safe-area-top, 0px)));
  right: max(1rem, calc(1rem + var(--tz-safe-area-right, 0px)));
}

.manufacturing-process-video-content {
  max-width: min(var(--manufacturing-process-video-dialog-max-width, 64rem), calc((var(--tz-mobile-safe-viewport-height, 100vh) - 2rem) * (var(--manufacturing-process-video-aspect-ratio, 4 / 3))));
}

.manufacturing-process-video-dialog-playback {
  display: block;
  width: 100%;
  max-height: calc(var(--tz-mobile-safe-viewport-height, 100vh) - 2rem);
  aspect-ratio: var(--manufacturing-process-video-aspect-ratio, 4 / 3);
  object-fit: contain;
  background: #000;
}
</style>
