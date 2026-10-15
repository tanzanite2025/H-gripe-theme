<template>
  <section
    class="reusable-photo-wall"
    :aria-label="ariaLabel"
    @keydown.left.prevent="selectPhotoWallPage(activePhotoWallPageIndex - 1)"
    @keydown.right.prevent="selectPhotoWallPage(activePhotoWallPageIndex + 1)"
    @keydown.home.prevent="selectPhotoWallPage(0)"
    @keydown.end.prevent="selectPhotoWallPage(photoWallPageCount - 1)"
  >
    <h3 v-if="title" class="reusable-photo-wall__title">{{ title }}</h3>

    <div
      :id="photoWallImageStageId"
      class="reusable-photo-wall__stage"
      :style="{ '--photo-wall-visible-columns': photoWallImagesPerPage }"
      aria-live="polite"
      aria-atomic="true"
    >
      <figure v-for="image in visiblePhotoWallImages" :key="image.id" class="reusable-photo-wall__image-frame">
        <StorefrontImage
          :src="image.src"
          :alt="image.alt"
          :width="image.width"
          :height="image.height"
          class="reusable-photo-wall__image"
          preset="gallery"
          :sizes="image.sizes || 'xs:100vw sm:50vw md:33vw lg:420px'"
          loading="lazy"
          decoding="async"
        />
        <figcaption v-if="image.caption" class="reusable-photo-wall__caption">{{ image.caption }}</figcaption>
      </figure>

      <template v-if="configuredPhotoWallImages.length === 0">
        <div
          v-for="placeholderIndex in photoWallImagesPerPage"
          :key="`photo-wall-empty-frame-${placeholderIndex}`"
          class="reusable-photo-wall__empty-frame"
          aria-hidden="true"
        >
          <svg viewBox="0 0 32 32" fill="none" stroke="currentColor" stroke-width="1.2">
            <rect x="4" y="5" width="24" height="22" rx="3" />
            <circle cx="12" cy="12" r="2.5" />
            <path d="m5 24 7-7 5 5 4-4 6 6" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
        </div>
      </template>
    </div>

    <div class="reusable-photo-wall__pagination-space">
      <div
        v-if="photoWallPageCount > 1"
        class="tz-carousel-pagination reusable-photo-wall__pagination"
        role="group"
        :aria-label="paginationLabel"
      >
        <button
          v-for="pageNumber in photoWallPageCount"
          :key="pageNumber"
          type="button"
          class="tz-carousel-pagination__dot"
          :class="{ 'is-active': activePhotoWallPageIndex === pageNumber - 1 }"
          :aria-label="showPageLabel(pageNumber, photoWallPageCount)"
          :aria-pressed="activePhotoWallPageIndex === pageNumber - 1"
          :aria-controls="photoWallImageStageId"
          @click="selectPhotoWallPage(pageNumber - 1)"
        />
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, useId, watch } from 'vue'
import { useMediaQuery } from '@vueuse/core'
import StorefrontImage from '~/components/StorefrontImage.vue'

export interface PhotoWallImage {
  id: string
  src: string
  alt: string
  caption?: string
  width?: number
  height?: number
  sizes?: string
}

const props = withDefaults(defineProps<{
  images: readonly PhotoWallImage[]
  title?: string
  ariaLabel: string
  paginationLabel: string
  showPageLabel: (pageNumber: number, pageCount: number) => string
  imagesPerPage?: number
}>(), {
  title: '',
  imagesPerPage: 3,
})

const emit = defineEmits<{
  'page-change': [pageIndex: number]
}>()

const photoWallImageStageId = `photo-wall-images-${useId()}`
const isCompactPhotoWall = useMediaQuery('(max-width: 639px)', { ssrWidth: 1280 })
const activePhotoWallPageIndex = ref(0)
const configuredPhotoWallImages = computed(() => props.images.filter(image => image.src.trim()))
const photoWallImagesPerPage = computed(() => isCompactPhotoWall.value ? 1 : Math.max(1, Math.floor(props.imagesPerPage)))
const photoWallPageCount = computed(() => Math.ceil(configuredPhotoWallImages.value.length / photoWallImagesPerPage.value))
const visiblePhotoWallImages = computed(() => {
  const firstVisibleImageIndex = activePhotoWallPageIndex.value * photoWallImagesPerPage.value
  return configuredPhotoWallImages.value.slice(firstVisibleImageIndex, firstVisibleImageIndex + photoWallImagesPerPage.value)
})

function selectPhotoWallPage(pageIndex: number): void {
  const pageCount = photoWallPageCount.value
  if (pageCount === 0) return
  activePhotoWallPageIndex.value = ((pageIndex % pageCount) + pageCount) % pageCount
  emit('page-change', activePhotoWallPageIndex.value)
}

watch(photoWallImagesPerPage, (newImagesPerPage, previousImagesPerPage) => {
  const firstVisibleImageIndex = activePhotoWallPageIndex.value * previousImagesPerPage
  activePhotoWallPageIndex.value = Math.floor(firstVisibleImageIndex / newImagesPerPage)
})

watch(photoWallPageCount, (pageCount) => {
  activePhotoWallPageIndex.value = Math.min(activePhotoWallPageIndex.value, Math.max(0, pageCount - 1))
})
</script>

<style scoped>
.reusable-photo-wall {
  display: grid;
  grid-template-rows: auto minmax(8rem, 1fr) 3rem;
  gap: 0.5rem;
  min-width: 0;
  min-height: 13.5rem;
}

.reusable-photo-wall__title {
  grid-row: 1;
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.8rem;
  font-weight: 500;
  line-height: 1.5;
}

.reusable-photo-wall__stage {
  grid-row: 2;
  container-type: size;
  display: grid;
  grid-template-columns: repeat(var(--photo-wall-visible-columns), minmax(0, 1fr));
  grid-auto-rows: minmax(0, 1fr);
  gap: 0.75rem;
  min-width: 0;
  min-height: 0;
}

.reusable-photo-wall__image-frame,
.reusable-photo-wall__empty-frame {
  align-self: center;
  justify-self: center;
  width: min(100%, 100cqh);
  aspect-ratio: 1 / 1;
  min-width: 0;
  min-height: 0;
  margin: 0;
  overflow: hidden;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.65rem;
  background: var(--tz-card-surface);
}

.reusable-photo-wall__image-frame {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}

.reusable-photo-wall__image {
  position: absolute;
  inset: 0;
  display: block;
  width: 100%;
  height: 100%;
  min-height: 0;
  object-fit: contain;
}

.reusable-photo-wall__caption {
  position: absolute;
  inset-inline: 0;
  bottom: 0;
  padding: 0.45rem 0.6rem;
  background: color-mix(in srgb, var(--tz-card-surface) 92%, transparent);
  color: var(--tz-text-primary);
  font-size: 0.75rem;
  line-height: 1.45;
}

.reusable-photo-wall__empty-frame {
  display: grid;
  place-items: center;
  border-style: dashed;
  background: transparent;
  color: var(--tz-text-muted);
}

.reusable-photo-wall__empty-frame svg {
  width: 2rem;
  height: 2rem;
  opacity: 0.45;
}

.reusable-photo-wall__pagination-space {
  grid-row: 3;
  min-width: 0;
  min-height: 3rem;
}

.reusable-photo-wall__pagination {
  justify-content: safe center;
  overflow-x: auto;
  scrollbar-width: none;
}

.reusable-photo-wall__pagination::-webkit-scrollbar {
  display: none;
}

@media (max-width: 639px) {
  .reusable-photo-wall {
    grid-template-rows: auto minmax(12rem, 1fr) 3rem;
    min-height: 18rem;
  }
}
</style>
