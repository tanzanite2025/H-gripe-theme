import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type {
  BrandGalleryImage,
  BrandGalleryPhoto,
} from '~/types/brandGalleryPhotos'

export function useBrandGalleryImageNavigation(options: {
  open: () => boolean
  gallery: () => BrandGalleryPhoto | null
  onClose: () => void
}) {
  const activeImageIndex = ref(0)

  const images = computed<BrandGalleryImage[]>(() => {
    const gallery = options.gallery()
    if (gallery?.galleryImageItems?.length) return gallery.galleryImageItems

    return (gallery?.galleryImages || []).map((url, index) => ({
      id: `image-${index + 1}`,
      url,
      thumbnail: url,
    }))
  })

  const currentImage = computed(() => images.value[activeImageIndex.value] || null)
  const hasMultipleImages = computed(() => images.value.length > 1)

  const selectImage = (index: number) => {
    if (index < 0 || index >= images.value.length) return
    activeImageIndex.value = index
  }

  const showPreviousImage = () => {
    if (!hasMultipleImages.value) return
    activeImageIndex.value = (
      activeImageIndex.value - 1 + images.value.length
    ) % images.value.length
  }

  const showNextImage = () => {
    if (!hasMultipleImages.value) return
    activeImageIndex.value = (activeImageIndex.value + 1) % images.value.length
  }

  const onKeydown = (event: KeyboardEvent) => {
    if (!options.open()) return
    if (event.key === 'Escape') {
      event.preventDefault()
      options.onClose()
      return
    }
    if (event.key === 'ArrowLeft') {
      event.preventDefault()
      showPreviousImage()
    } else if (event.key === 'ArrowRight') {
      event.preventDefault()
      showNextImage()
    }
  }

  watch(
    () => options.gallery()?.galleryId,
    () => {
      activeImageIndex.value = 0
    },
  )

  watch(
    images,
    (nextImages) => {
      if (!nextImages.length) {
        activeImageIndex.value = 0
        return
      }
      activeImageIndex.value = Math.min(activeImageIndex.value, nextImages.length - 1)
    },
  )

  onMounted(() => window.addEventListener('keydown', onKeydown))
  onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))

  return {
    activeImageIndex,
    currentImage,
    hasMultipleImages,
    images,
    selectImage,
    showNextImage,
    showPreviousImage,
  }
}
