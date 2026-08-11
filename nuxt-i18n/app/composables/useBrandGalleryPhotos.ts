import { ref } from 'vue'
import { usePublicApiBase } from '~/composables/usePublicApiBase'

export interface PictureWarehouseProductLink {
  product_id: number
  name?: string
  slug?: string
  locale?: string
}

export interface BrandGalleryPhoto {
  id: string
  kind: 'brand'
  title: string
  region: string
  nickname?: string
  galleryId: string
  galleryDetailsLoaded: boolean
  galleryImages: string[]
  productLinks: PictureWarehouseProductLink[]
}

interface PublicGalleryImage {
  id?: number | string
  url?: string
  thumbnail?: string
  title?: string
}

interface PublicGallery {
  id?: number | string
  name?: string
  title?: string
  slug?: string
  description?: string
  cover_image?: string
  images?: PublicGalleryImage[]
  product_links?: PictureWarehouseProductLink[]
}

interface GalleryListEnvelope {
  data?: PublicGallery[]
}

interface GalleryDetailEnvelope {
  data?: PublicGallery
}

const normalizeGalleryId = (value: unknown): string | null => {
  const id = String(value ?? '').trim()
  return id ? id : null
}

const galleryImageUrls = (gallery: PublicGallery): string[] => {
  const images = Array.isArray(gallery.images) ? gallery.images : []
  return images
    .map((image) => image?.url || image?.thumbnail || '')
    .filter((value): value is string => Boolean(value))
}

const galleryProductLinks = (gallery: PublicGallery): PictureWarehouseProductLink[] => {
  return Array.isArray(gallery.product_links)
    ? gallery.product_links.filter((product) => Boolean(product.slug || product.name))
    : []
}

const publicGalleryToBrandPhoto = (
  gallery: PublicGallery,
  detailsLoaded = false,
): BrandGalleryPhoto | null => {
  const id = normalizeGalleryId(gallery.id)
  if (!id) return null

  const images = galleryImageUrls(gallery)
  const cover = gallery.cover_image || images[0] || ''
  if (!cover) return null

  return {
    id: `gallery-${id}`,
    kind: 'brand',
    title: String(gallery.name || gallery.title || 'Brand gallery'),
    region: String(gallery.slug || 'Gallery'),
    nickname: gallery.description ? String(gallery.description) : undefined,
    galleryId: id,
    galleryDetailsLoaded: detailsLoaded,
    galleryImages: images.length ? images : [cover],
    productLinks: galleryProductLinks(gallery),
  }
}

export function useBrandGalleryPhotos() {
  const apiBase = usePublicApiBase()
  const brandPhotos = ref<BrandGalleryPhoto[]>([])
  const brandLoading = ref(true)
  const brandError = ref<string | null>(null)
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
      const galleries = Array.isArray(response)
        ? response
        : Array.isArray(response?.data)
          ? response.data
          : []

      brandPhotos.value = galleries
        .map((gallery): BrandGalleryPhoto | null => publicGalleryToBrandPhoto(gallery, false))
        .filter((item): item is BrandGalleryPhoto => item !== null)
    } catch {
      brandError.value = 'load_failed'
      brandPhotos.value = []
    } finally {
      brandLoading.value = false
    }
  }

  const markGalleryDetailLoaded = (index: number, photo: BrandGalleryPhoto) => {
    const current = brandPhotos.value[index]
    if (!current || current.galleryId !== photo.galleryId) return
    brandPhotos.value.splice(index, 1, { ...current, galleryDetailsLoaded: true })
  }

  const loadBrandGalleryDetails = async (index: number) => {
    const photo = brandPhotos.value[index]
    if (!photo?.galleryId || photo.galleryDetailsLoaded || detailLoads.has(photo.galleryId)) return

    detailLoads.add(photo.galleryId)
    try {
      const response = await $fetch<GalleryDetailEnvelope>(
        `${apiBase.value}/galleries/${encodeURIComponent(photo.galleryId)}`,
        { headers: { accept: 'application/json' } },
      )
      const detail = response?.data
      if (!detail) {
        markGalleryDetailLoaded(index, photo)
        return
      }

      const hydratedPhoto = publicGalleryToBrandPhoto(detail, true)
      if (!hydratedPhoto) {
        markGalleryDetailLoaded(index, photo)
        return
      }

      const current = brandPhotos.value[index]
      if (!current || current.galleryId !== photo.galleryId) return
      brandPhotos.value.splice(index, 1, {
        ...hydratedPhoto,
        id: photo.id,
        galleryId: photo.galleryId,
      })
    } catch {
      markGalleryDetailLoaded(index, photo)
    } finally {
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
