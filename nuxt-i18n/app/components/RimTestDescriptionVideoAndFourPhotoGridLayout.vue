<template>
  <section
    v-if="areTestReportMessagesReady"
    :id="usePageAnchor ? testDefinition.sectionId : testReportInstanceId"
    ref="rimTestSectionElement"
    class="rim-test-media-layout"
    :class="{ 'rim-test-media-layout--dialog': presentation === 'dialog' }"
    :data-test-report-id="testDefinition.sectionId"
    :aria-labelledby="`${testReportInstanceId}-heading`"
  >
    <header class="rim-test-media-layout__header">
      <component :is="headingTag" :id="`${testReportInstanceId}-heading`" class="rim-test-media-layout__title">
        {{ testTitle }}
      </component>
      <div class="rim-test-media-layout__header-actions">
        <span v-if="testDefinition.hasStandard" class="rim-test-media-layout__standard">
          <Icon name="lucide:book-open" aria-hidden="true" />
          {{ t(`${testMessagePrefix}.standard`) }}
        </span>
        <slot name="header-actions" />
      </div>
    </header>

    <div class="rim-test-media-layout__layout">
      <div class="rim-test-media-layout__description">
        <div class="rim-test-media-layout__purpose">
          <component :is="subheadingTag" class="rim-test-media-layout__subheading">
            {{ t('testReportRim.labels.purpose') }}
          </component>
          <p>{{ t(`${testMessagePrefix}.purpose`) }}</p>
        </div>

        <div class="rim-test-media-layout__description-group">
          <component :is="subheadingTag" class="rim-test-media-layout__subheading">
            {{ t('testReportRim.labels.testMethod') }}
          </component>
          <ol class="rim-test-media-layout__methods">
            <li v-for="methodIndex in testDefinition.methodCount" :key="methodIndex">
              {{ t(`${testMessagePrefix}.method.${methodIndex - 1}`) }}
            </li>
          </ol>
        </div>

        <div class="rim-test-media-layout__description-group">
          <component :is="subheadingTag" class="rim-test-media-layout__subheading">
            {{ t('testReportRim.labels.evaluationCriteria') }}
          </component>
          <ul class="rim-test-media-layout__criteria">
            <li v-for="criterionIndex in testDefinition.criteriaCount" :key="criterionIndex">
              {{ t(`${testMessagePrefix}.criteria.${criterionIndex - 1}`) }}
            </li>
          </ul>
        </div>
      </div>

      <div class="rim-test-media-layout__media">
        <div
          ref="rimTestMediaRail"
          class="rim-test-media-layout__media-rail"
          @scroll.passive="updateActiveRimTestMediaSlideFromScroll"
        >
          <figure
            :id="`${testReportInstanceId}-video`"
            class="rim-test-media-layout__media-item"
            data-rim-test-media-slide
            :inert="isRimTestMediaMounted && !isWideRimTestLayout && activeRimTestMediaSlideIndex !== 0"
          >
            <component :is="subheadingTag" class="rim-test-media-layout__subheading">
              {{ t('testReportRim.media.videoTitle', { test: testTitle }) }}
            </component>
            <div class="rim-test-media-layout__video-frame">
              <video
                v-if="mediaConfiguration.videoSourcePath"
                ref="rimTestVideo"
                :src="mediaConfiguration.videoSourcePath"
                :poster="mediaConfiguration.videoPosterPath || undefined"
                :aria-label="t('testReportRim.media.videoTitle', { test: testTitle })"
                controls
                playsinline
                preload="none"
              />
              <div v-else class="rim-test-media-layout__placeholder">
                <Icon name="lucide:video" class="rim-test-media-layout__placeholder-icon" aria-hidden="true" />
                <span class="rim-test-media-layout__placeholder-status">{{ t('testReportRim.media.videoPending') }}</span>
              </div>
            </div>
            <figcaption>{{ t(`${testMessagePrefix}.media.videoCaption`) }}</figcaption>
          </figure>

          <figure
            :id="`${testReportInstanceId}-photos`"
            class="rim-test-media-layout__media-item"
            data-rim-test-media-slide
            :inert="isRimTestMediaMounted && !isWideRimTestLayout && activeRimTestMediaSlideIndex !== 1"
          >
            <component :is="subheadingTag" class="rim-test-media-layout__subheading">
              {{ t('testReportRim.media.photoTitle', { test: testTitle }) }}
            </component>
            <div class="rim-test-media-layout__photo-grid">
              <div
                v-for="(photoSourcePath, photoIndex) in mediaConfiguration.photoSourcePaths"
                :key="photoIndex"
                class="rim-test-media-layout__photo-frame"
              >
                <StorefrontImage
                  v-if="photoSourcePath"
                  :src="photoSourcePath"
                  :alt="t('testReportRim.media.photoAlt', { test: testTitle, number: photoIndex + 1 })"
                  preset="content"
                  sizes="50vw lg:16vw"
                />
                <div v-else class="rim-test-media-layout__placeholder rim-test-media-layout__placeholder--photo">
                  <Icon name="lucide:image" class="rim-test-media-layout__placeholder-icon" aria-hidden="true" />
                  <span class="rim-test-media-layout__placeholder-status">
                    {{ t('testReportRim.media.photoPending', { number: photoIndex + 1 }) }}
                  </span>
                </div>
              </div>
            </div>
            <figcaption>{{ t(`${testMessagePrefix}.media.photoCaption`) }}</figcaption>
          </figure>
        </div>

        <div
          class="tz-carousel-pagination rim-test-media-layout__pagination"
          role="group"
          :aria-label="t('testReportRim.media.paginationLabel', { test: testTitle })"
          @keydown="handleRimTestMediaPaginationKeydown"
        >
          <button
            v-for="(mediaSlide, mediaSlideIndex) in rimTestMediaSlides"
            :key="mediaSlide.panelId"
            type="button"
            class="tz-carousel-pagination__dot"
            :class="{ 'is-active': activeRimTestMediaSlideIndex === mediaSlideIndex }"
            :aria-label="t(mediaSlide.titleKey, { test: testTitle })"
            :aria-controls="mediaSlide.panelId"
            :aria-pressed="activeRimTestMediaSlideIndex === mediaSlideIndex"
            @click="scrollToRimTestMediaSlide(mediaSlideIndex)"
          />
        </div>
      </div>
    </div>
  </section>
  <TestReportContentLoadingAndRetryStatus
    v-else
    :failed="Boolean(testReportMessageLoadingError)"
    @retry="retryLoadingTestReportMessages"
  />
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, useId, watch } from 'vue'
import { useResizeObserver } from '@vueuse/core'
import { useI18n } from '#imports'
import StorefrontImage from '~/components/StorefrontImage.vue'
import TestReportContentLoadingAndRetryStatus from '~/components/technology/test-reports/TestReportContentLoadingAndRetryStatus.vue'
import type { RimTestReportTestDefinition } from '~/data/rimTestReportTestDefinitionsAndMediaConfiguration'
import type { TestReportBlockPresentationProps } from '~/types/testReportBlock'
import { useIndependentTestReportMessages } from '~/composables/useIndependentTestReportMessages'

const { headingTag = 'h2', usePageAnchor = false, presentation = 'standalone', testDefinition } = defineProps<TestReportBlockPresentationProps & {
  testDefinition: RimTestReportTestDefinition
}>()

const { t } = useI18n()
const { areTestReportMessagesReady, testReportMessageLoadingError, retryLoadingTestReportMessages } = useIndependentTestReportMessages('testReportRim')

const testReportInstanceId = `${testDefinition.sectionId}-${useId()}`
const subheadingTag = computed(() => ({ h2: 'h3', h3: 'h4', h4: 'h5' })[headingTag])
const testMessagePrefix = computed(() => 'testReportRim.tests.' + testDefinition.id)
const testTitle = computed(() => t(testMessagePrefix.value + '.title'))
const mediaConfiguration = computed(() => testDefinition.mediaConfiguration)
const rimTestSectionElement = ref<HTMLElement | null>(null)
const rimTestMediaRail = ref<HTMLElement | null>(null)
const rimTestVideo = ref<HTMLVideoElement | null>(null)
const activeRimTestMediaSlideIndex = ref(0)
const isRimTestMediaMounted = ref(false)
const isWideRimTestLayout = ref(false)
const rimTestMediaSlides = computed(() => [
  { panelId: testReportInstanceId + '-video', titleKey: 'testReportRim.media.videoTitle' },
  { panelId: testReportInstanceId + '-photos', titleKey: 'testReportRim.media.photoTitle' },
])

onMounted(() => {
  isRimTestMediaMounted.value = true
})

function scrollToRimTestMediaSlide(slideIndex: number, behavior: ScrollBehavior = 'smooth') {
  const mediaRail = rimTestMediaRail.value
  const mediaSlide = mediaRail?.querySelectorAll<HTMLElement>('[data-rim-test-media-slide]')[slideIndex]
  if (!mediaRail || !mediaSlide || isWideRimTestLayout.value) return

  mediaRail.scrollTo({
    left: mediaRail.scrollLeft + mediaSlide.getBoundingClientRect().left - mediaRail.getBoundingClientRect().left,
    behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'instant' : behavior,
  })
}

function updateActiveRimTestMediaSlideFromScroll() {
  const mediaRail = rimTestMediaRail.value
  if (!mediaRail || !mediaRail.clientWidth || isWideRimTestLayout.value) return

  const railCenter = mediaRail.getBoundingClientRect().left + mediaRail.clientWidth / 2
  const mediaSlides = Array.from(mediaRail.querySelectorAll<HTMLElement>('[data-rim-test-media-slide]'))
  let nearestSlideIndex = 0
  let nearestSlideDistance = Number.POSITIVE_INFINITY

  mediaSlides.forEach((mediaSlide, slideIndex) => {
    const slideBounds = mediaSlide.getBoundingClientRect()
    const slideDistance = Math.abs(slideBounds.left + slideBounds.width / 2 - railCenter)
    if (slideDistance < nearestSlideDistance) {
      nearestSlideIndex = slideIndex
      nearestSlideDistance = slideDistance
    }
  })
  activeRimTestMediaSlideIndex.value = nearestSlideIndex
}

function handleRimTestMediaPaginationKeydown(event: KeyboardEvent) {
  if (!['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(event.key)) return

  const pagination = event.currentTarget as HTMLElement
  const paginationButtons = Array.from(pagination.querySelectorAll<HTMLButtonElement>('button'))
  const focusedButtonIndex = paginationButtons.indexOf(event.target as HTMLButtonElement)
  if (focusedButtonIndex < 0) return

  const lastSlideIndex = rimTestMediaSlides.value.length - 1
  const arrowDirection = event.key === 'ArrowRight' ? 1 : -1
  const nextSlideIndex = event.key === 'Home' ? 0
    : event.key === 'End' ? lastSlideIndex
      : (focusedButtonIndex + arrowDirection + rimTestMediaSlides.value.length) % rimTestMediaSlides.value.length

  event.preventDefault()
  paginationButtons[nextSlideIndex]?.focus({ preventScroll: true })
  scrollToRimTestMediaSlide(nextSlideIndex)
}

useResizeObserver(rimTestSectionElement, (entries) => {
  const contentWidth = entries[0]?.contentRect.width
  if (contentWidth === undefined) return
  isWideRimTestLayout.value = contentWidth >= 896
  void nextTick(() => scrollToRimTestMediaSlide(activeRimTestMediaSlideIndex.value, 'instant'))
})

watch([activeRimTestMediaSlideIndex, isWideRimTestLayout], ([slideIndex, isWideLayout]) => {
  if (!isWideLayout && slideIndex !== 0) rimTestVideo.value?.pause()
})

onBeforeUnmount(() => rimTestVideo.value?.pause())
</script>

<style scoped>
.rim-test-media-layout {
  container: rim-test-content / inline-size;
  margin-top: 1rem;
  padding: clamp(1.25rem, 2.5vw, 2rem);
  border: 1px solid var(--tz-border-subtle);
  border-radius: 1rem;
  background: var(--tz-card-surface);
  box-shadow: 0 4px 18px rgba(20, 32, 43, 0.05);
  text-align: start;
  scroll-margin-top: calc(var(--tz-site-header-spacer-height, 120px) + 1rem);
}

.rim-test-media-layout--dialog {
  margin: 0;
  border: 0;
  border-radius: 0;
  box-shadow: none;
}

.rim-test-media-layout__header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding-bottom: 1.25rem;
  border-bottom: 1px solid var(--tz-border-subtle);
}

.rim-test-media-layout__title {
  margin: 0;
  font-size: var(--tz-type-section-title);
  line-height: 1.35;
  font-weight: 600;
  color: var(--tz-text-primary);
}

.rim-test-media-layout__header-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.65rem;
}

.rim-test-media-layout__standard {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4rem 0.65rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.5rem;
  color: var(--tz-text-secondary);
  font-size: 0.75rem;
  line-height: 1.3;
}

.rim-test-media-layout__layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 1.75rem;
  padding-top: 1.5rem;
}

.rim-test-media-layout__description,
.rim-test-media-layout__media {
  min-width: 0;
}

.rim-test-media-layout__description {
  color: var(--tz-text-secondary);
  font-size: 0.9rem;
  line-height: 1.85;
}

.rim-test-media-layout__subheading {
  margin: 0 0 0.6rem;
  color: var(--tz-text-primary);
  font-size: 0.9rem;
  line-height: 1.5;
  font-weight: 600;
}

.rim-test-media-layout__purpose p {
  margin: 0;
}

.rim-test-media-layout__description-group {
  margin-top: 1.5rem;
}

.rim-test-media-layout__methods,
.rim-test-media-layout__criteria {
  display: grid;
  gap: 0.65rem;
  margin: 0;
  padding-inline-start: 1.4rem;
}

.rim-test-media-layout__methods {
  list-style: decimal;
}

.rim-test-media-layout__criteria {
  list-style: disc;
}

.rim-test-media-layout__methods li,
.rim-test-media-layout__criteria li {
  padding-inline-start: 0.25rem;
}

.rim-test-media-layout__methods li::marker,
.rim-test-media-layout__criteria li::marker {
  color: var(--tz-text-accent);
  font-weight: 600;
}

.rim-test-media-layout__media-rail {
  display: grid;
  grid-auto-flow: column;
  grid-auto-columns: 100%;
  align-items: start;
  gap: 1rem;
  overflow-x: auto;
  overscroll-behavior-inline: contain;
  scroll-snap-type: x mandatory;
  scrollbar-width: none;
}

.rim-test-media-layout__media-rail::-webkit-scrollbar {
  display: none;
}

.rim-test-media-layout__media-item {
  min-width: 0;
  margin: 0;
  scroll-snap-align: start;
}

.rim-test-media-layout__video-frame,
.rim-test-media-layout__photo-frame {
  position: relative;
  overflow: hidden;
  border: 1px dashed var(--tz-border-strong);
  border-radius: 0.75rem;
  background: var(--tz-surface-inset);
}

.rim-test-media-layout__video-frame {
  aspect-ratio: 16 / 9;
}

.rim-test-media-layout__photo-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  grid-template-rows: repeat(2, minmax(0, 1fr));
  gap: 0.5rem;
  aspect-ratio: 16 / 9;
}

.rim-test-media-layout__photo-frame {
  min-width: 0;
  min-height: 0;
  border-radius: 0.5rem;
}

.rim-test-media-layout__video-frame video,
.rim-test-media-layout__photo-frame :deep(img) {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.rim-test-media-layout__placeholder {
  display: flex;
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 1rem;
  text-align: center;
}

.rim-test-media-layout__placeholder-icon {
  width: 1.75rem;
  height: 1.75rem;
  margin-bottom: 0.25rem;
  color: var(--tz-text-accent);
}

.rim-test-media-layout__placeholder-status {
  color: var(--tz-text-secondary);
  font-size: 0.8rem;
}

.rim-test-media-layout__placeholder--photo {
  gap: 0.3rem;
  padding: 0.4rem;
}

.rim-test-media-layout__placeholder--photo .rim-test-media-layout__placeholder-icon {
  width: 1.1rem;
  height: 1.1rem;
  margin: 0;
}

.rim-test-media-layout__placeholder--photo .rim-test-media-layout__placeholder-status {
  font-size: 0.7rem;
  line-height: 1.4;
}

.rim-test-media-layout__pagination {
  margin-top: 0.5rem;
}

.rim-test-media-layout__media-item figcaption {
  margin-top: 0.65rem;
  color: var(--tz-text-secondary);
  font-size: 0.8rem;
  line-height: 1.65;
}

/* Keep this container breakpoint in sync with the ResizeObserver width above. */
@container rim-test-content (min-width: 896px) {
  .rim-test-media-layout__layout {
    grid-template-columns: repeat(3, minmax(0, 1fr));
    align-items: start;
    gap: clamp(1.25rem, 2vw, 2rem);
  }

  .rim-test-media-layout__media,
  .rim-test-media-layout__media-rail {
    display: contents;
  }

  .rim-test-media-layout__pagination {
    display: none;
  }
}
</style>
