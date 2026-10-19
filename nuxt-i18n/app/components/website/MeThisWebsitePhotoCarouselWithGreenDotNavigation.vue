<template>
  <section
    v-if="websitePhotos.length"
    class="website-photo-carousel"
    :aria-label="isChineseLocale ? '我与这个网站的照片' : 'Photos behind this website'"
    @keydown.left.prevent="selectWebsitePhotoPage(activeWebsitePhotoPageIndex - 1)"
    @keydown.right.prevent="selectWebsitePhotoPage(activeWebsitePhotoPageIndex + 1)"
    @keydown.home.prevent="selectWebsitePhotoPage(0)"
    @keydown.end.prevent="selectWebsitePhotoPage(websitePhotoPageCount - 1)"
  >
    <div
      :id="websitePhotoStageId"
      class="website-photo-carousel__stage"
      :style="{ '--website-photo-columns': websitePhotosPerPage }"
      aria-live="polite"
      aria-atomic="true"
    >
      <figure
        v-for="(photo, photoIndex) in websitePhotos"
        :key="photo.sourcePath"
        v-show="Math.floor(photoIndex / websitePhotosPerPage) === activeWebsitePhotoPageIndex"
        class="website-photo-carousel__frame"
      >
        <StorefrontImage
          :src="photo.url"
          :alt="photo.description"
          :optimize="false"
          :loading="photoIndex < websitePhotosPerPage ? 'eager' : 'lazy'"
          decoding="async"
          class="website-photo-carousel__image"
        />
      </figure>
    </div>

    <div
      v-if="websitePhotoPageCount > 1"
      class="tz-carousel-pagination website-photo-carousel__pagination"
      role="group"
      :aria-label="isChineseLocale ? '切换照片' : 'Choose a photo'"
    >
      <button
        v-for="pageNumber in websitePhotoPageCount"
        :key="pageNumber"
        type="button"
        class="tz-carousel-pagination__dot"
        :class="{ 'is-active': activeWebsitePhotoPageIndex === pageNumber - 1 }"
        :aria-label="isChineseLocale ? `查看第 ${pageNumber} 组照片，共 ${websitePhotoPageCount} 组` : `Show photo group ${pageNumber} of ${websitePhotoPageCount}`"
        :aria-pressed="activeWebsitePhotoPageIndex === pageNumber - 1"
        :aria-controls="websitePhotoStageId"
        @click="selectWebsitePhotoPage(pageNumber - 1)"
      />
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, useId, watch } from 'vue'
import { useMediaQuery } from '@vueuse/core'
import { useI18n } from '#imports'
import StorefrontImage from '~/components/StorefrontImage.vue'
import { isSimplifiedChineseStorefrontLocale } from '~/utils/storefrontLocales'

// Vite discovers these files during development/build and publishes their URLs.
// Keep images in the route-matching folder; numeric filename prefixes control order.
const websitePhotoAssetUrls = import.meta.glob<string>(
  '../../assets/images/website/me-this-website/*.{avif,gif,jpeg,jpg,png,webp,AVIF,GIF,JPEG,JPG,PNG,WEBP}',
  { eager: true, query: '?url', import: 'default' },
)
const { locale } = useI18n()
const isChineseLocale = computed(() => isSimplifiedChineseStorefrontLocale(locale.value))
const websitePhotoStageId = `website-photo-stage-${useId()}`
const activeWebsitePhotoPageIndex = ref(0)
const isDesktopPhotoLayout = useMediaQuery('(min-width: 1024px)', { ssrWidth: 1280 })
const isTabletPhotoLayout = useMediaQuery('(min-width: 640px)', { ssrWidth: 1280 })
const websitePhotosPerPage = computed(() => isDesktopPhotoLayout.value ? 5 : isTabletPhotoLayout.value ? 3 : 2)
const websitePhotos = computed(() => Object.entries(websitePhotoAssetUrls)
  .sort(([firstPath], [secondPath]) => firstPath.localeCompare(secondPath, 'en', { numeric: true }))
  .map(([sourcePath, url], photoIndex) => ({
    sourcePath,
    url,
    description: sourcePath.split('/').pop()!
      .replace(/\.[^.]+$/, '')
      .replace(/^\d+[-_\s]*/, '')
      .replace(/[-_]+/g, ' ')
      .trim() || (isChineseLocale.value ? `网站照片 ${photoIndex + 1}` : `Website photo ${photoIndex + 1}`),
  })))

const websitePhotoPageCount = computed(() => Math.ceil(websitePhotos.value.length / websitePhotosPerPage.value))

function selectWebsitePhotoPage(pageIndex: number): void {
  const pageCount = websitePhotoPageCount.value
  if (!pageCount) return
  activeWebsitePhotoPageIndex.value = ((pageIndex % pageCount) + pageCount) % pageCount
}

watch(websitePhotosPerPage, (newPageSize, previousPageSize) => {
  const firstVisiblePhotoIndex = activeWebsitePhotoPageIndex.value * previousPageSize
  activeWebsitePhotoPageIndex.value = Math.floor(firstVisiblePhotoIndex / newPageSize)
})

watch(websitePhotoPageCount, (pageCount) => {
  activeWebsitePhotoPageIndex.value = Math.min(activeWebsitePhotoPageIndex.value, Math.max(0, pageCount - 1))
})
</script>

<style scoped>
.website-photo-carousel {
  width: min(100%, 76rem);
  margin-inline: auto;
  padding-top: 0.75rem;
}

.website-photo-carousel__stage {
  display: grid;
  grid-template-columns: repeat(var(--website-photo-columns), minmax(0, 1fr));
  align-items: start;
  gap: 0.5rem;
}

.website-photo-carousel__frame {
  width: 100%;
  min-width: 0;
  aspect-ratio: 1 / 1;
  margin: 0;
  overflow: hidden;
  border-radius: 0.375rem;
}

.website-photo-carousel__image {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.website-photo-carousel__pagination {
  flex-wrap: wrap;
  margin-top: 0.25rem;
}
</style>
