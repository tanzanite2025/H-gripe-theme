<template>
  <section
    :id="sectionId"
    class="global-media-rail-carousel"
    :aria-labelledby="titleId"
  >
    <div class="global-media-rail-carousel__head">
      <div>
        <p v-if="eyebrow" class="global-media-rail-carousel__eyebrow">
          {{ eyebrow }}
        </p>
        <h3 :id="titleId" class="global-media-rail-carousel__title">
          {{ title }}
        </h3>
      </div>

      <div class="global-media-rail-carousel__actions">
        <NuxtLink
          v-if="showViewMore && viewMoreTo && viewMoreLabel"
          :to="viewMoreTo"
          class="global-media-rail-carousel__more"
        >
          <span>{{ viewMoreLabel }}</span>
          <Icon name="lucide:arrow-right" class="h-3.5 w-3.5" aria-hidden="true" />
        </NuxtLink>

        <div v-if="showControls && hasMultipleSlides" class="global-media-rail-carousel__controls">
          <button
            type="button"
            class="global-media-rail-carousel__arrow"
            :aria-label="previousLabel"
            @click="scrollSlides(-1)"
          >
            <Icon name="lucide:arrow-left" class="h-4 w-4" aria-hidden="true" />
          </button>
          <button
            type="button"
            class="global-media-rail-carousel__arrow"
            :aria-label="nextLabel"
            @click="scrollSlides(1)"
          >
            <Icon name="lucide:arrow-right" class="h-4 w-4" aria-hidden="true" />
          </button>
        </div>
      </div>
    </div>

    <div
      ref="rail"
      class="global-media-rail-carousel__rail"
      role="region"
      :aria-busy="showSkeletons"
      :aria-label="ariaLabel"
      @scroll.passive="updateActiveSlide"
    >
      <template v-if="showSkeletons">
        <div
          v-for="index in skeletonCount"
          :key="index"
          class="global-media-rail-carousel__skeleton"
          aria-hidden="true"
        />
      </template>

      <figure
        v-for="(slide, index) in slides"
        v-else
        :key="slide.id"
        :data-media-slide="index"
        class="global-media-rail-carousel__slide"
        :class="{ 'global-media-rail-carousel__slide--captioned': slide.label || slide.description }"
      >
        <button
          v-if="slide.clickable"
          type="button"
          class="global-media-rail-carousel__slide-button"
          :aria-label="slide.label || slide.alt"
          @click="emit('slide-click', slide, index)"
        >
          <span class="global-media-rail-carousel__media">
            <StorefrontImage
              :src="slide.src"
              :alt="slide.alt"
              class="size-full object-cover"
              preset="gallery"
              :sizes="slide.sizes || defaultSizes"
              :loading="slide.loading || (index < 2 ? 'eager' : 'lazy')"
              decoding="async"
            />
          </span>
          <span v-if="slide.label || slide.description" class="global-media-rail-carousel__caption">
            <span v-if="slide.label" class="global-media-rail-carousel__caption-title">
              {{ slide.label }}
            </span>
            <span v-if="slide.description" class="global-media-rail-carousel__caption-description">
              {{ slide.description }}
            </span>
          </span>
        </button>
        <div v-else class="global-media-rail-carousel__slide-content">
          <span class="global-media-rail-carousel__media">
            <StorefrontImage
              :src="slide.src"
              :alt="slide.alt"
              class="size-full object-cover"
              preset="gallery"
              :sizes="slide.sizes || defaultSizes"
              :loading="slide.loading || (index < 2 ? 'eager' : 'lazy')"
              decoding="async"
            />
          </span>
          <span v-if="slide.label || slide.description" class="global-media-rail-carousel__caption">
            <span v-if="slide.label" class="global-media-rail-carousel__caption-title">
              {{ slide.label }}
            </span>
            <span v-if="slide.description" class="global-media-rail-carousel__caption-description">
              {{ slide.description }}
            </span>
          </span>
        </div>
      </figure>
    </div>

    <div
      v-if="showPagination && hasMultipleSlides"
      class="tz-carousel-pagination global-media-rail-carousel__pagination"
      role="tablist"
      :aria-label="progressLabel"
    >
      <button
        v-for="(slide, index) in slides"
        :key="slide.id"
        type="button"
        class="tz-carousel-pagination__dot"
        :class="{ 'is-active': activeSlide === index }"
        :aria-label="slideLabel(index + 1)"
        :aria-selected="activeSlide === index"
        role="tab"
        @click="scrollToSlide(index)"
      />
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

export interface GlobalMediaRailSlide {
  id: string
  src: string
  alt: string
  label?: string
  description?: string
  clickable?: boolean
  sizes?: string
  loading?: 'eager' | 'lazy'
}

const emit = defineEmits<{
  'slide-click': [slide: GlobalMediaRailSlide, index: number]
}>()

const props = withDefaults(defineProps<{
  sectionId?: string
  eyebrow?: string
  title: string
  ariaLabel: string
  viewMoreTo?: string
  viewMoreLabel?: string
  previousLabel: string
  nextLabel: string
  progressLabel: string
  slides: GlobalMediaRailSlide[]
  pending?: boolean
  skeletonCount?: number
  showViewMore?: boolean
  showControls?: boolean
  showPagination?: boolean
  showSlideLabel?: (slide: number) => string
}>(), {
  sectionId: 'global-media-rail-carousel',
  pending: false,
  skeletonCount: 3,
  showViewMore: true,
  showControls: true,
  showPagination: true,
})

const rail = ref<HTMLElement | null>(null)
const activeSlide = ref(0)
const defaultSizes = 'xs:82vw sm:62vw md:42vw lg:29vw xl:360px'

const titleId = computed(() => `${props.sectionId}-title`)
const showSkeletons = computed(() => props.pending || props.slides.length === 0)
const hasMultipleSlides = computed(() => props.slides.length > 1)

const slideLabel = (slide: number) => {
  return props.showSlideLabel?.(slide) || `Show slide ${slide}`
}

const scrollSlides = (direction: -1 | 1) => {
  const currentRail = rail.value
  if (!currentRail) return

  const firstSlide = currentRail.querySelector<HTMLElement>('[data-media-slide]')
  const slideWidth = firstSlide?.getBoundingClientRect().width ?? currentRail.clientWidth * 0.8
  const gap = Number.parseFloat(
    getComputedStyle(currentRail).columnGap || getComputedStyle(currentRail).gap,
  ) || 0

  currentRail.scrollBy({
    left: direction * (slideWidth + gap),
    behavior: 'smooth',
  })
}

const scrollToSlide = (index: number) => {
  const currentRail = rail.value
  if (!currentRail) return

  const slide = currentRail.querySelector<HTMLElement>(`[data-media-slide="${index}"]`)
  if (!slide) return

  const railRect = currentRail.getBoundingClientRect()
  const slideRect = slide.getBoundingClientRect()
  const maxScrollLeft = Math.max(0, currentRail.scrollWidth - currentRail.clientWidth)
  const targetScrollLeft =
    currentRail.scrollLeft +
    (slideRect.left - railRect.left) -
    (currentRail.clientWidth - slideRect.width) / 2

  currentRail.scrollTo({
    left: Math.min(maxScrollLeft, Math.max(0, targetScrollLeft)),
    behavior: 'smooth',
  })
}

const updateActiveSlide = () => {
  const currentRail = rail.value
  if (!currentRail) return

  const slideElements = [...currentRail.querySelectorAll<HTMLElement>('[data-media-slide]')]
  if (!slideElements.length) return

  const railCenter = currentRail.getBoundingClientRect().left + currentRail.clientWidth / 2
  activeSlide.value = slideElements.reduce((closestIndex, slide, index) => {
    const current = slideElements[closestIndex]
    if (!current) return index

    const slideCenter = slide.getBoundingClientRect().left + slide.clientWidth / 2
    const currentCenter = current.getBoundingClientRect().left + current.clientWidth / 2
    return Math.abs(slideCenter - railCenter) < Math.abs(currentCenter - railCenter)
      ? index
      : closestIndex
  }, 0)
}
</script>

<style scoped>
.global-media-rail-carousel {
  scroll-margin-top: calc(var(--tz-site-header-spacer-height) + 1rem);
}
</style>

<style scoped>
.global-media-rail-carousel {
  padding-top: 0.25rem;
  overflow-x: hidden;
}

.global-media-rail-carousel__head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 1rem;
  padding-bottom: 0.875rem;
  border-bottom: 1px solid rgba(5, 150, 105, 0.22);
}

.global-media-rail-carousel__eyebrow {
  margin: 0;
  color: #059669;
  font-size: 0.6875rem;
  font-weight: 600;
  letter-spacing: 0;
  line-height: 1.2;
  text-transform: uppercase;
}

.global-media-rail-carousel__title {
  margin: 0.375rem 0 0;
  color: var(--tz-text-primary);
  font-size: 1.125rem;
  font-weight: 700;
  letter-spacing: 0;
  line-height: 1.2;
}

.global-media-rail-carousel__actions {
  display: flex;
  flex: none;
  align-items: center;
  gap: 0.625rem;
}

.global-media-rail-carousel__controls {
  display: flex;
  flex: none;
  gap: 0.5rem;
}

.global-media-rail-carousel__more {
  display: inline-flex;
  min-height: 2rem;
  align-items: center;
  gap: 0.35rem;
  padding: 0.35rem 0.65rem;
  border: 1px solid var(--tz-border-strong);
  border-radius: 999px;
  background: var(--tz-surface-card);
  color: var(--tz-text-accent);
  font-size: 0.6875rem;
  font-weight: 700;
  line-height: 1;
  white-space: nowrap;
  transition: border-color 160ms ease, background-color 160ms ease, color 160ms ease;
}

.global-media-rail-carousel__more:hover,
.global-media-rail-carousel__more:focus-visible {
  border-color: rgba(5, 150, 105, 0.72);
  background: rgba(5, 150, 105, 0.08);
  color: var(--tz-text-accent-hover);
}

.global-media-rail-carousel__arrow {
  display: inline-grid;
  width: 2rem;
  height: 2rem;
  place-items: center;
  border: 1px solid var(--tz-border-strong);
  border-radius: 999px;
  background: var(--tz-surface-card);
  color: var(--tz-text-primary);
  transition: border-color 160ms ease, background-color 160ms ease, color 160ms ease;
}

.global-media-rail-carousel__arrow:hover,
.global-media-rail-carousel__arrow:focus-visible {
  border-color: rgba(5, 150, 105, 0.75);
  background: rgba(5, 150, 105, 0.12);
  color: var(--tz-text-accent);
}

.global-media-rail-carousel__rail {
  display: flex;
  gap: 0.875rem;
  margin: 0 -0.5rem;
  padding: 1rem 0.5rem 0.75rem;
  overflow-x: auto;
  overscroll-behavior-inline: contain;
  scroll-padding-inline: 0.5rem;
  scroll-behavior: smooth;
  scroll-snap-type: x mandatory;
  scrollbar-width: none;
  -webkit-overflow-scrolling: touch;
  touch-action: pan-x;
}

.global-media-rail-carousel__rail::-webkit-scrollbar {
  display: none;
}

.global-media-rail-carousel__slide,
.global-media-rail-carousel__skeleton {
  flex: 0 0 clamp(18rem, 29vw, 25rem);
  overflow: hidden;
  scroll-snap-align: start;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.5rem;
  background: var(--tz-image-loading-surface, var(--tz-surface-card));
}

.global-media-rail-carousel__slide,
.global-media-rail-carousel__skeleton {
  aspect-ratio: 4 / 3;
}

.global-media-rail-carousel__slide {
  margin: 0;
}

.global-media-rail-carousel__slide--captioned {
  aspect-ratio: auto;
}

.global-media-rail-carousel__slide-button,
.global-media-rail-carousel__slide-content {
  display: flex;
  width: 100%;
  height: 100%;
  flex-direction: column;
  text-align: left;
}

.global-media-rail-carousel__slide-button {
  color: inherit;
}

.global-media-rail-carousel__slide-button:focus-visible {
  outline: 2px solid #059669;
  outline-offset: -2px;
}

.global-media-rail-carousel__media {
  display: block;
  width: 100%;
  aspect-ratio: 4 / 3;
  flex: none;
  overflow: hidden;
}

.global-media-rail-carousel__caption {
  display: flex;
  min-height: 3.75rem;
  flex-direction: column;
  gap: 0.25rem;
  justify-content: center;
  padding: 0.65rem 0.75rem 0.7rem;
}

.global-media-rail-carousel__caption-title {
  overflow: hidden;
  color: var(--tz-text-primary);
  font-size: 0.75rem;
  font-weight: 700;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.global-media-rail-carousel__caption-description {
  display: -webkit-box;
  overflow: hidden;
  color: var(--tz-text-secondary);
  font-size: 0.625rem;
  line-height: 1.45;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.global-media-rail-carousel__skeleton {
  animation: global-media-rail-carousel-pulse 1.5s ease-in-out infinite;
}

.global-media-rail-carousel__pagination {
  margin-top: 0.125rem;
}

@keyframes global-media-rail-carousel-pulse {
  0%,
  100% {
    opacity: 0.46;
  }

  50% {
    opacity: 0.9;
  }
}

@media (max-width: 767px) {
  .global-media-rail-carousel__head {
    align-items: flex-start;
  }

  .global-media-rail-carousel__actions {
    gap: 0.375rem;
  }

  .global-media-rail-carousel__more {
    padding-inline: 0.55rem;
  }

  .global-media-rail-carousel__slide,
  .global-media-rail-carousel__skeleton {
    flex-basis: min(82vw, 23rem);
  }

  .global-media-rail-carousel__rail {
    padding-inline-end: 1rem;
    scroll-padding-inline-end: 1rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  .global-media-rail-carousel__rail {
    scroll-behavior: auto;
  }

  .global-media-rail-carousel__skeleton {
    animation: none;
  }
}
</style>
