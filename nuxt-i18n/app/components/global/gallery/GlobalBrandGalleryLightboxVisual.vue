<template>
  <div class="brand-gallery-lightbox__visual">
    <div class="brand-gallery-lightbox__image-stage">
      <button
        v-if="hasMultipleImages"
        type="button"
        class="brand-gallery-lightbox__arrow brand-gallery-lightbox__arrow--left"
        :aria-label="labels.previousImage"
        @click="emit('previous')"
      >
        <Icon name="lucide:chevron-left" class="size-5" aria-hidden="true" />
      </button>

      <div class="brand-gallery-lightbox__image-frame">
        <StorefrontImage
          v-if="currentImage"
          :src="currentImage.url"
          :alt="currentImage.alt || currentImage.title || galleryTitle"
          class="brand-gallery-lightbox__image"
          preset="gallery"
          loading="eager"
          decoding="async"
        />
        <div v-else class="brand-gallery-lightbox__empty-image">
          {{ labels.noImage }}
        </div>
      </div>

      <button
        v-if="hasMultipleImages"
        type="button"
        class="brand-gallery-lightbox__arrow brand-gallery-lightbox__arrow--right"
        :aria-label="labels.nextImage"
        @click="emit('next')"
      >
        <Icon name="lucide:chevron-right" class="size-5" aria-hidden="true" />
      </button>
    </div>

    <div class="brand-gallery-lightbox__image-meta">
      <div class="flex min-w-0 items-center justify-between gap-3">
        <p class="brand-gallery-lightbox__image-title">
          {{ currentImage?.title || galleryTitle }}
        </p>
        <span v-if="images.length > 1" class="brand-gallery-lightbox__counter">
          {{ activeImageIndex + 1 }} / {{ images.length }}
        </span>
      </div>
      <p v-if="currentImage?.description" class="brand-gallery-lightbox__image-description">
        {{ currentImage.description }}
      </p>
      <div v-if="currentTags.length" class="brand-gallery-lightbox__tags">
        <span v-for="tag in currentTags" :key="tag" class="brand-gallery-lightbox__tag">
          {{ tag }}
        </span>
      </div>
    </div>

    <div
      v-if="hasMultipleImages"
      class="brand-gallery-lightbox__thumbnails"
      :aria-label="labels.imageThumbnails"
    >
      <button
        v-for="(image, index) in images"
        :key="image.id"
        type="button"
        class="brand-gallery-lightbox__thumbnail"
        :class="{ 'is-active': activeImageIndex === index }"
        :aria-label="`${image.title || galleryTitle} ${index + 1}`"
        :aria-current="activeImageIndex === index ? 'true' : undefined"
        @click="emit('select-image', index)"
      >
        <StorefrontImage
          :src="image.thumbnail || image.url"
          :alt="image.alt || image.title || ''"
          class="size-full object-cover"
          preset="thumbnail"
          loading="lazy"
        />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { BrandGalleryImage } from '~/types/brandGalleryPhotos'
import type { GalleryLightboxLabels } from '~/types/brandGalleryLightbox'

const props = defineProps<{
  activeImageIndex: number
  currentImage: BrandGalleryImage | null
  galleryTitle: string
  hasMultipleImages: boolean
  images: BrandGalleryImage[]
  labels: GalleryLightboxLabels
}>()

const emit = defineEmits<{
  next: []
  previous: []
  'select-image': [index: number]
}>()

const currentTags = computed(() =>
  String(props.currentImage?.tags || '')
    .split(',')
    .map((tag) => tag.trim())
    .filter(Boolean),
)
</script>

<style scoped>
.brand-gallery-lightbox__visual {
  min-width: 0;
  padding: 1rem;
}

.brand-gallery-lightbox__image-stage {
  position: relative;
  display: flex;
  min-height: min(58vh, 34rem);
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: var(--tz-image-loading-surface, rgba(255, 255, 255, 0.04));
}

.brand-gallery-lightbox__image-frame {
  display: flex;
  width: 100%;
  height: 100%;
  min-height: inherit;
  align-items: center;
  justify-content: center;
  padding: 2.5rem 3rem;
}

.brand-gallery-lightbox__image {
  display: block;
  max-width: 100%;
  max-height: min(58vh, 34rem);
  object-fit: contain;
}

.brand-gallery-lightbox__empty-image {
  display: grid;
  min-width: min(70%, 20rem);
  min-height: 12rem;
  place-items: center;
  color: rgba(255, 255, 255, 0.48);
  font-size: 0.75rem;
}

.brand-gallery-lightbox__arrow {
  position: absolute;
  z-index: 2;
  display: inline-grid;
  width: 2.25rem;
  height: 2.25rem;
  place-items: center;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 999px;
  background: rgba(2, 6, 23, 0.7);
  color: #fff;
  transition: border-color 160ms ease, background-color 160ms ease, color 160ms ease;
}

.brand-gallery-lightbox__arrow:hover,
.brand-gallery-lightbox__arrow:focus-visible {
  border-color: rgba(181, 255, 109, 0.75);
  background: rgba(181, 255, 109, 0.12);
  color: #b5ff6d;
}

.brand-gallery-lightbox__arrow--left {
  left: 0.75rem;
}

.brand-gallery-lightbox__arrow--right {
  right: 0.75rem;
}

.brand-gallery-lightbox__image-meta {
  min-height: 4.25rem;
  padding: 0.75rem 0.15rem 0;
}

.brand-gallery-lightbox__image-title {
  min-width: 0;
  margin: 0;
  overflow: hidden;
  color: #fff;
  font-size: 0.8125rem;
  font-weight: 700;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.brand-gallery-lightbox__counter {
  flex: none;
  color: rgba(255, 255, 255, 0.5);
  font-family: var(--tz-font-ui);
  font-size: 0.625rem;
}

.brand-gallery-lightbox__image-description {
  margin: 0.35rem 0 0;
  color: rgba(226, 232, 240, 0.72);
  font-size: 0.75rem;
  line-height: 1.55;
  overflow-wrap: anywhere;
}

.brand-gallery-lightbox__tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  margin-top: 0.55rem;
}

.brand-gallery-lightbox__tag {
  border: 1px solid rgba(181, 255, 109, 0.28);
  border-radius: 999px;
  padding: 0.2rem 0.45rem;
  color: #b5ff6d;
  font-size: 0.625rem;
  line-height: 1.2;
}

.brand-gallery-lightbox__thumbnails {
  display: flex;
  gap: 0.5rem;
  max-width: 100%;
  overflow-x: auto;
  padding: 0.2rem 0 0.15rem;
  scrollbar-width: none;
}

.brand-gallery-lightbox__thumbnails::-webkit-scrollbar {
  display: none;
}

.brand-gallery-lightbox__thumbnail {
  position: relative;
  flex: 0 0 3.5rem;
  height: 3.5rem;
  overflow: hidden;
  border: 2px solid transparent;
  border-radius: 0.35rem;
  opacity: 0.58;
  transition: border-color 160ms ease, opacity 160ms ease;
}

.brand-gallery-lightbox__thumbnail:hover,
.brand-gallery-lightbox__thumbnail:focus-visible,
.brand-gallery-lightbox__thumbnail.is-active {
  border-color: #b5ff6d;
  opacity: 1;
}

@media (max-width: 767px) {
  .brand-gallery-lightbox__visual {
    padding: 0.75rem;
  }

  .brand-gallery-lightbox__image-stage {
    min-height: min(56vh, 28rem);
  }

  .brand-gallery-lightbox__image-frame {
    padding-inline: 2.75rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  .brand-gallery-lightbox__arrow,
  .brand-gallery-lightbox__thumbnail {
    transition: none;
  }
}
</style>
