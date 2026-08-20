import { useRuntimeConfig, useState } from '#imports'
import { usePublicApiBase } from '~/composables/usePublicApiBase'
import type { BrandGalleryPhoto } from '~/types/brandGalleryPhotos'
import {
  extractPublicGalleries,
  mapPublicGalleryToBrandPhoto,
  type GalleryDetailEnvelope,
  type GalleryListEnvelope,
  type PublicGallery,
} from '~/utils/brandGalleryPhotos'
import {
  createStorefrontMediaContext,
} from '~/utils/storefrontMedia'

export function useBrandGalleryPhotos() {
  const runtimeConfig = useRuntimeConfig()
  const mediaContext = createStorefrontMediaContext(runtimeConfig)
  const apiBase = usePublicApiBase()
  const brandPhotos = useState<BrandGalleryPhoto[]>('brand-gallery-photos', () => [])
  const brandLoading = useState<boolean>('brand-gallery-photos-loading', () => true)
  const brandError = useState<string | null>('brand-gallery-photos-error', () => null)
  const detailLoads = new Set<string>()

  const fetchBrandPhotos = async () => {
    try {
      brandLoading.value = true
      brandError.value = null

      const response = await $fetch<GalleryListEnvelope | PublicGallery[]>(
        `${apiBase.value}/galleries`,
        {
          headers: { accept: 'application/json' },
          query: { page: 1, page_size: 100 },
        },
      )
      const galleries = extractPublicGalleries(response)

      brandPhotos.value = galleries
        .map((gallery): BrandGalleryPhoto | null => (
          mapPublicGalleryToBrandPhoto(gallery, mediaContext, false)
        ))
        .filter((item): item is BrandGalleryPhoto => item !== null)
    } catch {
      brandError.value = 'load_failed'
      brandPhotos.value = []
    } finally {
      brandLoading.value = false
    }
  }

  const updateGalleryDetailState = (
    index: number,
    photo: BrandGalleryPhoto,
    patch: Partial<BrandGalleryPhoto>,
  ) => {
    const current = brandPhotos.value[index]
    if (!current || current.galleryId !== photo.galleryId) return
    brandPhotos.value.splice(index, 1, { ...current, ...patch })
  }

  const loadBrandGalleryDetails = async (index: number) => {
    const photo = brandPhotos.value[index]
    if (!photo?.galleryId || photo.galleryDetailsLoaded || detailLoads.has(photo.galleryId)) return

    detailLoads.add(photo.galleryId)
    updateGalleryDetailState(index, photo, { galleryDetailsLoading: true })
    try {
      const response = await $fetch<GalleryDetailEnvelope>(
        `${apiBase.value}/galleries/${encodeURIComponent(photo.galleryId)}`,
        { headers: { accept: 'application/json' } },
      )
      const detail = response?.data
      if (!detail) {
        updateGalleryDetailState(index, photo, {
          galleryDetailsLoaded: true,
          galleryDetailsLoading: false,
        })
        return
      }

      const hydratedPhoto = mapPublicGalleryToBrandPhoto(detail, mediaContext, true)
      if (!hydratedPhoto) {
        updateGalleryDetailState(index, photo, {
          galleryDetailsLoaded: true,
          galleryDetailsLoading: false,
        })
        return
      }

      const current = brandPhotos.value[index]
      if (!current || current.galleryId !== photo.galleryId) return
      brandPhotos.value.splice(index, 1, {
        ...hydratedPhoto,
        id: photo.id,
        galleryId: photo.galleryId,
        galleryDetailsLoading: false,
      })
    } catch {
      updateGalleryDetailState(index, photo, {
        galleryDetailsLoading: false,
      })
    } finally {
      updateGalleryDetailState(index, photo, { galleryDetailsLoading: false })
      detailLoads.delete(photo.galleryId)
    }
  }

  return {
    brandPhotos,
    brandLoading,
    brandError,
    fetchBrandPhotos,
    loadBrandGalleryDetails,
  }
}
