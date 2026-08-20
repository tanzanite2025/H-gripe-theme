<template>
  <div>
    <GlobalMediaRailCarousel
      :section-id="sectionId"
      :eyebrow="t('home.brandPhotoCarousel.eyebrow')"
      :title="t('home.brandPhotoCarousel.title')"
      :ariaLabel="t('home.brandPhotoCarousel.ariaLabel')"
      :view-more-to="brandPhotoPath"
      :view-more-label="t('home.brandPhotoCarousel.viewMore')"
      :previous-label="t('home.brandPhotoCarousel.previous')"
      :next-label="t('home.brandPhotoCarousel.next')"
      :progress-label="t('home.brandPhotoCarousel.progress')"
      :show-slide-label="showSlideLabel"
      :slides="slides"
      :pending="brandLoading"
      :skeleton-count="3"
      @slide-click="openGallery"
    />
    <GlobalBrandGalleryLightbox
      :open="selectedGalleryIndex !== null"
      :gallery="activeGallery"
      :labels="lightboxLabels"
      @close="closeGallery"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useAsyncData, useI18n, useLocalePath } from '#imports'
import GlobalBrandGalleryLightbox from '~/components/global/gallery/GlobalBrandGalleryLightbox.vue'
import GlobalMediaRailCarousel from '~/components/global/gallery/GlobalMediaRailCarousel.vue'
import type { GlobalMediaRailSlide } from '~/components/global/gallery/GlobalMediaRailCarousel.vue'
import { useBrandGalleryPhotos } from '~/composables/useBrandGalleryPhotos'

interface BrandPhotoSlide {
  id: string
  src: string
  alt: string
  label: string
  description?: string
  clickable: true
}

const sectionId = 'home-brand-photo-carousel'
const { t } = useI18n()
const localePath = useLocalePath()
const brandPhotoPath = localePath('/picture-warehouse/brand')
const {
  brandPhotos,
  brandLoading,
  fetchBrandPhotos,
  loadBrandGalleryDetails,
} = useBrandGalleryPhotos()

await useAsyncData(
  sectionId,
  async () => {
    await fetchBrandPhotos()
    return true
  },
)

const slides = computed<BrandPhotoSlide[]>(() => {
  return brandPhotos.value
    .flatMap((photo) => {
      const cover = photo.coverImage || photo.galleryImageItems[0]?.url || photo.galleryImages[0] || ''
      if (!cover) return []

      return [{
        id: photo.id,
        src: cover,
        alt: photo.title,
        label: photo.title,
        description: photo.description || photo.nickname,
        clickable: true as const,
      }]
    })
})

const selectedGalleryIndex = ref<number | null>(null)
const activeGallery = computed(() =>
  selectedGalleryIndex.value === null
    ? null
    : brandPhotos.value[selectedGalleryIndex.value] || null,
)

const lightboxLabels = computed(() => ({
  close: t('home.brandPhotoCarousel.lightboxClose'),
  previousImage: t('home.brandPhotoCarousel.lightboxPreviousImage'),
  nextImage: t('home.brandPhotoCarousel.lightboxNextImage'),
  imageThumbnails: t('home.brandPhotoCarousel.lightboxImageThumbnails'),
  loadingDetails: t('home.brandPhotoCarousel.lightboxLoadingDetails'),
  noImage: t('home.brandPhotoCarousel.lightboxNoImage'),
  relatedProducts: t('home.brandPhotoCarousel.lightboxRelatedProducts'),
  noRelatedProducts: t('home.brandPhotoCarousel.lightboxNoRelatedProducts'),
}))

const openGallery = (slide: GlobalMediaRailSlide) => {
  const galleryIndex = brandPhotos.value.findIndex((photo) => photo.id === slide.id)
  if (galleryIndex < 0) return

  selectedGalleryIndex.value = galleryIndex
  void loadBrandGalleryDetails(galleryIndex)
}

const closeGallery = () => {
  selectedGalleryIndex.value = null
}

const showSlideLabel = (slide: number) => t('home.brandPhotoCarousel.showSlide', { slide })
</script>
