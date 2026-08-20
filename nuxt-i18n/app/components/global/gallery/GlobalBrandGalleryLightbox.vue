<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-opacity duration-200 ease-out"
      leave-active-class="transition-opacity duration-150 ease-in"
      enter-from-class="opacity-0"
      leave-to-class="opacity-0"
    >
      <div
        v-if="props.open && props.gallery"
        class="brand-gallery-lightbox fixed inset-0 z-[1450] flex items-center justify-center px-3 py-4 sm:px-5"
        role="presentation"
        @click.self="close"
      >
        <section
          class="brand-gallery-lightbox__panel"
          role="dialog"
          aria-modal="true"
          :aria-labelledby="titleId"
          tabindex="-1"
        >
          <header class="brand-gallery-lightbox__header">
            <div class="min-w-0">
              <p v-if="props.gallery.slug" class="brand-gallery-lightbox__eyebrow">
                {{ props.gallery.slug }}
              </p>
              <h2 :id="titleId" class="brand-gallery-lightbox__title">
                {{ props.gallery.title }}
              </h2>
            </div>

            <button
              type="button"
              class="brand-gallery-lightbox__close"
              :aria-label="labels.close"
              @click="close"
            >
              <Icon name="lucide:x" class="size-5" aria-hidden="true" />
            </button>
          </header>

          <div class="brand-gallery-lightbox__body">
            <GlobalBrandGalleryLightboxVisual
              :active-image-index="activeImageIndex"
              :current-image="currentImage"
              :gallery-title="props.gallery.title"
              :has-multiple-images="hasMultipleImages"
              :images="images"
              :labels="labels"
              @next="showNextImage"
              @previous="showPreviousImage"
              @select-image="selectImage"
            />
            <GlobalBrandGalleryLightboxDetails
              :gallery="props.gallery"
              :labels="labels"
            />
          </div>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import GlobalBrandGalleryLightboxDetails from '~/components/global/gallery/GlobalBrandGalleryLightboxDetails.vue'
import GlobalBrandGalleryLightboxVisual from '~/components/global/gallery/GlobalBrandGalleryLightboxVisual.vue'
import {
  defaultGalleryLightboxLabels,
  type GalleryLightboxLabels,
} from '~/types/brandGalleryLightbox'
import { useBrandGalleryImageNavigation } from '~/composables/useBrandGalleryImageNavigation'
import type { BrandGalleryPhoto } from '~/types/brandGalleryPhotos'

const props = withDefaults(defineProps<{
  open: boolean
  gallery: BrandGalleryPhoto | null
  labels?: Partial<GalleryLightboxLabels>
}>(), {
  labels: undefined,
})

const emit = defineEmits<{
  close: []
}>()

const titleId = 'brand-gallery-lightbox-title'

const labels = computed<GalleryLightboxLabels>(() => ({
  ...defaultGalleryLightboxLabels,
  ...props.labels,
}))

const close = () => emit('close')

const {
  activeImageIndex,
  currentImage,
  hasMultipleImages,
  images,
  selectImage,
  showNextImage,
  showPreviousImage,
} = useBrandGalleryImageNavigation({
  open: () => props.open,
  gallery: () => props.gallery,
  onClose: close,
})
</script>

<style scoped>
.brand-gallery-lightbox {
  background: rgba(2, 6, 23, 0.84);
  backdrop-filter: blur(8px);
}

.brand-gallery-lightbox__panel {
  display: flex;
  width: min(100%, 70rem);
  max-height: min(92vh, 52rem);
  flex-direction: column;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.14);
  border-radius: 0.75rem;
  background: #020617;
  box-shadow: 0 20px 70px rgba(0, 0, 0, 0.48);
}

.brand-gallery-lightbox__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  padding: 0.85rem 1rem;
}

.brand-gallery-lightbox__eyebrow {
  margin: 0 0 0.25rem;
  overflow: hidden;
  color: #b5ff6d;
  font-size: 0.625rem;
  font-weight: 700;
  line-height: 1.2;
  text-overflow: ellipsis;
  text-transform: uppercase;
  white-space: nowrap;
}

.brand-gallery-lightbox__title {
  margin: 0;
  overflow: hidden;
  color: #fff;
  font-size: 1rem;
  font-weight: 700;
  line-height: 1.25;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.brand-gallery-lightbox__close {
  display: inline-grid;
  flex: none;
  width: 2.25rem;
  height: 2.25rem;
  place-items: center;
  border: 1px solid rgba(255, 255, 255, 0.16);
  border-radius: 999px;
  color: rgba(255, 255, 255, 0.82);
  transition: border-color 160ms ease, background-color 160ms ease, color 160ms ease;
}

.brand-gallery-lightbox__close:hover,
.brand-gallery-lightbox__close:focus-visible {
  border-color: rgba(181, 255, 109, 0.72);
  background: rgba(181, 255, 109, 0.1);
  color: #b5ff6d;
}

.brand-gallery-lightbox__body {
  display: grid;
  min-height: 0;
  overflow: auto;
  grid-template-columns: minmax(0, 1.7fr) minmax(15rem, 0.8fr);
}

@media (max-width: 767px) {
  .brand-gallery-lightbox {
    align-items: flex-start;
    padding: 0;
  }

  .brand-gallery-lightbox__panel {
    width: 100%;
    max-height: 100dvh;
    border-radius: 0;
  }

  .brand-gallery-lightbox__body {
    display: block;
  }
}

@media (prefers-reduced-motion: reduce) {
  .brand-gallery-lightbox__close {
    transition: none;
  }
}
</style>
