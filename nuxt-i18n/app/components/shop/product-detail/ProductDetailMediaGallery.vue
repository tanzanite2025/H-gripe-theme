<template>
  <section class="product-media-column" aria-label="Product media">
    <div class="product-media-layout">
      <div class="product-media-stage">
        <figure v-if="previewMedia?.kind === 'image'" class="product-media-frame">
          <StorefrontImage
            :src="previewMedia.url"
            :alt="previewMedia.alt"
            preset="detail"
            loading="eager"
            fetchpriority="high"
          />
        </figure>
        <figure v-else-if="previewMedia?.kind === 'video'" class="product-media-frame product-media-frame--video">
          <video
            :src="previewMedia.url"
            :poster="previewMedia.poster || undefined"
            controls
            playsinline
            preload="metadata"
          />
        </figure>
        <div v-else class="product-media-placeholder">
          <Icon name="lucide:image-off" class="product-media-placeholder__icon" aria-hidden="true" />
        </div>
        <template v-if="items.length > 1">
          <button
            type="button"
            class="tz-directional-arrow tz-directional-arrow--large product-media-nav product-media-nav--previous"
            aria-label="Previous media"
            @click="emit('previous')"
          >
            <Icon name="lucide:chevron-left" aria-hidden="true" />
          </button>
          <button
            type="button"
            class="tz-directional-arrow tz-directional-arrow--large product-media-nav product-media-nav--next"
            aria-label="Next media"
            @click="emit('next')"
          >
            <Icon name="lucide:chevron-right" aria-hidden="true" />
          </button>
        </template>
      </div>
      <div
        ref="productMediaThumbnailsRef"
        class="product-media-thumbnails"
        :class="{ 'product-media-thumbnails--centered': !slotsOverflowing }"
        aria-label="Media thumbnails"
      >
        <template v-for="(media, index) in slots" :key="media?.id || `media-placeholder-${index}`">
          <button
            v-if="media"
            type="button"
            class="product-media-thumbnail"
            :data-media-id="media.id"
            :class="{ 'product-media-thumbnail--active': selectedMediaId === media.id }"
            :aria-label="`${media.kind === 'video' ? 'View product video' : 'View product image'} ${index + 1}`"
            :aria-pressed="selectedMediaId === media.id"
            @click="emit('select', media.id)"
          >
            <StorefrontImage
              v-if="media.thumbnailUrl"
              :src="media.thumbnailUrl"
              :alt="media.alt"
              preset="thumbnail"
            />
            <span v-else class="product-media-thumbnail__placeholder">
              <Icon
                :name="media.kind === 'video' ? 'lucide:video' : 'lucide:image-off'"
                aria-hidden="true"
              />
            </span>
            <span v-if="media.kind === 'video'" class="product-media-thumbnail__badge" aria-hidden="true">
              <Icon name="lucide:play" />
            </span>
          </button>
          <div v-else class="product-media-thumbnail product-media-thumbnail--placeholder" aria-hidden="true">
            <span class="product-media-thumbnail__placeholder">
              <Icon name="lucide:image" aria-hidden="true" />
            </span>
          </div>
        </template>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from 'vue'
import type {
  ProductGalleryItem,
  ProductPreviewMedia,
} from '~/types/productDetail'

const props = defineProps<{
  items: ProductGalleryItem[]
  slots: Array<ProductGalleryItem | null>
  selectedMediaId: string | null
  previewMedia: ProductPreviewMedia | null
  slotsOverflowing: boolean
}>()

const emit = defineEmits<{
  (event: 'select', mediaId: string): void
  (event: 'previous'): void
  (event: 'next'): void
}>()

const productMediaThumbnailsRef = ref<HTMLElement | null>(null)

const centerSelectedMediaThumbnail = async () => {
  const mediaId = props.selectedMediaId
  if (!mediaId) return

  await nextTick()
  const container = productMediaThumbnailsRef.value
  if (!container) return

  const selected = Array.from(container.querySelectorAll<HTMLElement>('[data-media-id]'))
    .find((element) => element.dataset.mediaId === mediaId)
  if (!selected) return

  const containerRect = container.getBoundingClientRect()
  const selectedRect = selected.getBoundingClientRect()
  const scrollOptions: ScrollToOptions = {
    behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth',
  }

  if (container.scrollHeight > container.clientHeight + 1) {
    scrollOptions.top = container.scrollTop
      + selectedRect.top
      - containerRect.top
      - ((container.clientHeight - selectedRect.height) / 2)
  }

  if (container.scrollWidth > container.clientWidth + 1) {
    scrollOptions.left = container.scrollLeft
      + selectedRect.left
      - containerRect.left
      - ((container.clientWidth - selectedRect.width) / 2)
  }

  container.scrollTo(scrollOptions)
}

watch(() => props.selectedMediaId, () => {
  void centerSelectedMediaThumbnail()
}, { flush: 'post' })

watch(() => props.items, () => {
  void centerSelectedMediaThumbnail()
}, { flush: 'post' })

onMounted(() => {
  void centerSelectedMediaThumbnail()
})
</script>

<style scoped>
.product-media-column {
  display: grid;
  gap: 1rem;
  min-width: 0;
}

.product-media-layout {
  --product-media-thumb-gap: 0.65rem;
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: var(--product-media-thumb-gap);
  min-width: 0;
}

.product-media-stage {
  position: relative;
  width: 100%;
  aspect-ratio: 1 / 1;
  min-width: 0;
}

.product-media-frame,
.product-media-placeholder {
  width: 100%;
  height: 100%;
  min-width: 0;
  max-width: 100%;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0;
  border-radius: 0.75rem;
  overflow: hidden;
  background: var(--tz-surface-subtle);
  border: 1px solid rgba(148, 163, 184, 0.18);
}

.product-media-frame img,
.product-media-frame video {
  width: 100%;
  height: 100%;
  display: block;
  object-fit: contain;
}

.product-media-frame--video {
  background: var(--tz-surface-subtle);
}

.product-media-nav {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  z-index: 1;
}

.product-media-nav--previous {
  left: 0.75rem;
}

.product-media-nav--next {
  right: 0.75rem;
}

.product-media-thumbnails {
  display: flex;
  flex-wrap: nowrap;
  gap: var(--product-media-thumb-gap);
  min-width: 0;
  min-height: 0;
  overflow-x: auto;
  padding: 0.1rem 0.1rem 0.35rem;
  overscroll-behavior: contain;
  scroll-behavior: smooth;
  scroll-padding-inline: 50%;
  scrollbar-width: thin;
  scrollbar-color: var(--tz-border-strong) transparent;
}

.product-media-thumbnail {
  position: relative;
  flex: 0 0 4.75rem;
  width: 4.75rem;
  height: 4.75rem;
  aspect-ratio: 1 / 1;
  overflow: hidden;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.5rem;
  background: var(--tz-surface-muted);
  color: var(--tz-text-primary);
  cursor: pointer;
  padding: 0;
  transition: border-color 0.2s ease, opacity 0.2s ease, transform 0.2s ease;
}

.product-media-thumbnail--placeholder {
  cursor: default;
  border-style: dashed;
  background: var(--tz-surface-subtle);
}

.product-media-thumbnail:hover {
  border-color: var(--tz-border-strong);
  transform: translateY(-1px);
}

.product-media-thumbnail--placeholder:hover {
  border-color: var(--tz-border-subtle);
  transform: none;
}

.product-media-thumbnail--active {
  border-color: rgba(5, 150, 105, 0.88);
  box-shadow: 0 0 0 2px rgba(5, 150, 105, 0.16);
}

.product-media-thumbnail img {
  width: 100%;
  height: 100%;
  display: block;
  object-fit: cover;
}

.product-media-thumbnail__placeholder {
  display: flex;
  width: 100%;
  height: 100%;
  align-items: center;
  justify-content: center;
  color: var(--tz-text-muted);
}

.product-media-thumbnail__placeholder svg {
  width: 1.15rem;
  height: 1.15rem;
}

.product-media-thumbnail__badge {
  position: absolute;
  right: 0.3rem;
  bottom: 0.3rem;
  display: inline-flex;
  width: 1.35rem;
  height: 1.35rem;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--tz-border-strong);
  border-radius: 50%;
  background: var(--tz-surface-muted);
  color: var(--tz-text-primary);
}

.product-media-thumbnail__badge svg {
  width: 0.72rem;
  height: 0.72rem;
}

.product-media-placeholder {
  flex-direction: column;
  padding: 1.5rem;
  text-align: center;
  color: var(--tz-text-secondary);
  border-style: dashed;
}

.product-media-placeholder__icon {
  width: 3rem;
  height: 3rem;
  margin-bottom: 0.85rem;
  color: rgba(5, 150, 105, 0.78);
}

@media (min-width: 900px) {
  .product-media-layout {
    grid-template-columns: clamp(3.5rem, 5vw, 5.25rem) minmax(0, 1fr);
    width: 100%;
    align-items: stretch;
    gap: 1rem;
  }

  .product-media-stage {
    grid-column: 2;
    grid-row: 1;
  }

  .product-media-thumbnails {
    grid-column: 1;
    grid-row: 1;
    align-self: stretch;
    flex-direction: column;
    justify-content: flex-start;
    height: 100%;
    max-height: 100%;
    min-height: 0;
    overflow-x: hidden;
    overflow-y: auto;
    padding: 0.1rem;
    scroll-padding-block: 50%;
  }

  .product-media-thumbnails--centered {
    justify-content: center;
  }

  .product-media-thumbnail {
    flex: 0 0 auto;
    width: 100%;
    height: auto;
    aspect-ratio: 1 / 1;
  }
}

@media (max-width: 767px) {
  .product-media-thumbnail {
    flex-basis: 4.5rem;
    width: 4.5rem;
    height: 4.5rem;
  }

  .product-media-nav {
    width: 2.25rem;
    height: 2.25rem;
  }
}
</style>
