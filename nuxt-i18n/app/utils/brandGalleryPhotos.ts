import type {
  BrandGalleryImage,
  BrandGalleryPhoto,
  PictureWarehouseProductLink,
} from '~/types/brandGalleryPhotos'
import {
  createStorefrontMediaContext,
  normalizeStorefrontMediaUrl,
} from '~/utils/storefrontMedia'

export interface PublicGalleryImage {
  id?: number | string
  url?: string
  thumbnail?: string
  title?: string
  description?: string
  alt?: string
  tags?: string
  order?: number
}

export interface PublicGallery {
  id?: number | string
  name?: string
  title?: string
  slug?: string
  description?: string
  cover_image?: string
  images?: PublicGalleryImage[]
  product_links?: PictureWarehouseProductLink[]
}

export interface GalleryListEnvelope {
  data?: PublicGallery[]
}

export interface GalleryDetailEnvelope {
  data?: PublicGallery
}

type StorefrontMediaContext = ReturnType<typeof createStorefrontMediaContext>

const normalizeGalleryId = (value: unknown): string | null => {
  const id = String(value ?? '').trim()
  return id ? id : null
}

const mapGalleryImageItems = (
  gallery: PublicGallery,
  mediaContext: StorefrontMediaContext,
): BrandGalleryImage[] => {
  const images = Array.isArray(gallery.images) ? gallery.images : []
  return images
    .map((image, index): BrandGalleryImage | null => {
      const url = normalizeStorefrontMediaUrl(
        image?.url || image?.thumbnail,
        mediaContext,
      )
      if (!url) return null

      const thumbnail = normalizeStorefrontMediaUrl(image?.thumbnail, mediaContext) || url
      const imageID = String(image?.id ?? `image-${index + 1}`).trim()

      return {
        id: imageID || `image-${index + 1}`,
        url,
        thumbnail,
        title: image?.title ? String(image.title) : undefined,
        description: image?.description ? String(image.description) : undefined,
        alt: image?.alt ? String(image.alt) : undefined,
        tags: image?.tags ? String(image.tags) : undefined,
        order: typeof image?.order === 'number' ? image.order : undefined,
      }
    })
    .filter((value): value is BrandGalleryImage => value !== null)
}

const mapGalleryProductLinks = (gallery: PublicGallery): PictureWarehouseProductLink[] => {
  return Array.isArray(gallery.product_links)
    ? gallery.product_links.filter((product) => Boolean(product.slug || product.name))
    : []
}

export const extractPublicGalleries = (
  response: GalleryListEnvelope | PublicGallery[],
): PublicGallery[] => {
  if (Array.isArray(response)) return response
  return Array.isArray(response?.data) ? response.data : []
}

export const mapPublicGalleryToBrandPhoto = (
  gallery: PublicGallery,
  mediaContext: StorefrontMediaContext,
  detailsLoaded = false,
): BrandGalleryPhoto | null => {
  const id = normalizeGalleryId(gallery.id)
  if (!id) return null

  const imageItems = mapGalleryImageItems(gallery, mediaContext)
  const cover = normalizeStorefrontMediaUrl(gallery.cover_image, mediaContext)
    || imageItems[0]?.url
    || ''
  if (!cover) return null

  const normalizedImageItems = imageItems.length
    ? imageItems
    : [{
        id: `cover-${id}`,
        url: cover,
        thumbnail: cover,
        alt: String(gallery.name || gallery.title || 'Brand gallery'),
      }]

  return {
    id: `gallery-${id}`,
    kind: 'brand',
    title: String(gallery.name || gallery.title || 'Brand gallery'),
    region: String(gallery.slug || 'Gallery'),
    nickname: gallery.description ? String(gallery.description) : undefined,
    slug: gallery.slug ? String(gallery.slug) : undefined,
    description: gallery.description ? String(gallery.description) : undefined,
    coverImage: cover,
    galleryId: id,
    galleryDetailsLoaded: detailsLoaded,
    galleryDetailsLoading: false,
    galleryImages: normalizedImageItems.map((image) => image.url),
    galleryImageItems: normalizedImageItems,
    productLinks: mapGalleryProductLinks(gallery),
  }
}
