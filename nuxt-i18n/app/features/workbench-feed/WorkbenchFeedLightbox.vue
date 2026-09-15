<template>
  <Teleport to="body">
    <div v-if="open" class="workbench-feed-lightbox" role="dialog" aria-modal="true" aria-label="Workshop photo">
      <button type="button" class="workbench-feed-lightbox__backdrop" aria-label="Close" @click="$emit('close')"></button>
      <div class="workbench-feed-lightbox__panel">
        <button type="button" class="workbench-feed-lightbox__close" aria-label="Close" title="Close" @click="$emit('close')">
          <Icon name="lucide:x" aria-hidden="true" />
        </button>
        <button v-if="media.length > 1" type="button" class="workbench-feed-lightbox__nav workbench-feed-lightbox__nav--prev" aria-label="Previous photo" title="Previous photo" @click="$emit('previous')">
          <Icon name="lucide:chevron-left" aria-hidden="true" />
        </button>
        <StorefrontImage v-if="activeMedia" :src="activeMedia.file_path" :alt="activeMedia.caption || 'Workshop photo'" preset="gallery" loading="eager" class="workbench-feed-lightbox__image" />
        <button v-if="media.length > 1" type="button" class="workbench-feed-lightbox__nav workbench-feed-lightbox__nav--next" aria-label="Next photo" title="Next photo" @click="$emit('next')">
          <Icon name="lucide:chevron-right" aria-hidden="true" />
        </button>
        <p v-if="activeMedia?.caption" class="workbench-feed-lightbox__caption">{{ activeMedia.caption }}</p>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import type { WorkbenchFeedMedia } from './types'

defineProps<{
  open: boolean
  media: WorkbenchFeedMedia[]
  activeMedia: WorkbenchFeedMedia | null
}>()

defineEmits<{
  (event: 'close'): void
  (event: 'previous'): void
  (event: 'next'): void
}>()
</script>

<style scoped>
.workbench-feed-lightbox { position: fixed; inset: 0; z-index: 12000; display: grid; place-items: center; padding: 1rem; }
.workbench-feed-lightbox__backdrop { position: absolute; inset: 0; border: 0; background: rgb(2 6 23 / 82%); }
.workbench-feed-lightbox__panel { position: relative; z-index: 1; display: grid; width: min(92vw, 1000px); max-height: 92vh; place-items: center; }
.workbench-feed-lightbox__image { display: block; width: auto; max-width: 100%; max-height: 82vh; object-fit: contain; }
.workbench-feed-lightbox__close, .workbench-feed-lightbox__nav { position: absolute; z-index: 2; display: grid; width: 2.5rem; height: 2.5rem; place-items: center; border: 1px solid rgb(255 255 255 / 35%); border-radius: 50%; color: #fff; background: rgb(15 23 42 / 75%); }
.workbench-feed-lightbox__close { top: -3rem; right: 0; }
.workbench-feed-lightbox__nav { top: 50%; transform: translateY(-50%); }
.workbench-feed-lightbox__nav--prev { left: .75rem; }
.workbench-feed-lightbox__nav--next { right: .75rem; }
.workbench-feed-lightbox__caption { margin: .75rem 0 0; color: #fff; font-size: .8rem; text-align: center; }
</style>
